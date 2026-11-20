// vecbench — CLI-бенчмарк векторных ядер пакета vec.
//
// Запуск:
//
//	go run ./cmd/vecbench [флаги]
//
// Флаги:
//
//	-ops   dot,axpy,matvec,q8,rmsnorm,rope  (по умолчанию: все)
//	-lvl   avx512,avx2,neon,generic,all     (по умолчанию: all)
//	-n     размер вектора (по умолчанию: 4096)
//	-sec   время прогрева+замера в секундах (по умолчанию: 1)
//	-json  вывод в JSON
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/notrobot1/goqwen/goqwen/vec"
)

type result struct {
	Op      string  `json:"op"`
	Level   string  `json:"level"`
	N       int     `json:"n"`
	OpsPerS float64 `json:"ops_per_sec"`
	GBperS  float64 `json:"gb_per_sec"`
	NsPerOp float64 `json:"ns_per_op"`
}

func randVec(r *rand.Rand, n int) []float32 {
	v := make([]float32, n)
	for i := range v {
		v[i] = r.Float32()*2 - 1
	}
	return v
}

func randQ8Row(r *rand.Rand, blocks int) []byte {
	row := make([]byte, blocks*34)
	r.Read(row)
	for b := 0; b < blocks; b++ {
		row[b*34+1] = 0x30 + byte(r.Intn(8))
	}
	return row
}

// bench запускает fn не менее minDur и возвращает (итераций, суммарное время).
func bench(minDur time.Duration, fn func()) (iters int64, elapsed time.Duration) {
	// прогрев
	for i := 0; i < 5; i++ {
		fn()
	}
	runtime.GC()

	start := time.Now()
	for {
		fn()
		iters++
		elapsed = time.Since(start)
		if elapsed >= minDur {
			break
		}
	}
	return
}

func run(ops []string, levels []string, n int, dur time.Duration, asJSON bool) {
	r := rand.New(rand.NewSource(42))
	blocks := n / 32
	if blocks == 0 {
		blocks = 1
	}

	x := randVec(r, n)
	y := randVec(r, n)

	rows := n       // строк матрицы
	mat := randVec(r, rows*n)
	dst := make([]float32, rows)

	q8row := randQ8Row(r, blocks)
	scratch := make([]float32, blocks*32)

	weight := randVec(r, n)
	xnorm := randVec(r, n)
	cosv := randVec(r, n/2)
	sinv := randVec(r, n/2)
	xrope := randVec(r, n)

	wantOps := map[string]bool{}
	for _, o := range ops {
		wantOps[o] = true
	}

	wantLevels := map[string]bool{}
	for _, l := range levels {
		wantLevels[l] = true
	}

	var results []result

	if !asJSON {
		fmt.Printf("%-12s %-8s %8s %12s %10s %10s\n",
			"операция", "уровень", "N", "Мop/с", "ГБ/с", "нс/op")
		fmt.Println(strings.Repeat("-", 66))
	}

	saved := vec.Level()
	defer vec.SetLevel(saved)

	for _, lvl := range vec.Levels() {
		if !wantLevels[lvl] && !wantLevels["all"] {
			continue
		}
		if err := vec.SetLevel(lvl); err != nil {
			continue
		}

		type opDef struct {
			name  string
			bytes int64 // байт/вызов (для ГБ/с)
			fn    func()
		}

		var sink float32
		var sinkB int

		opDefs := []opDef{
			{
				name:  "dot",
				bytes: int64(n) * 4 * 2,
				fn:    func() { sink += vec.Dot(x, y) },
			},
			{
				name:  "axpy",
				bytes: int64(n) * 4 * 3,
				fn:    func() { vec.Axpy(0.5, x, y) },
			},
			{
				name:  "matvec",
				bytes: int64(rows) * int64(n) * 4 * 2,
				fn:    func() { vec.MatVec(dst, mat, x) },
			},
			{
				name:  "q8",
				bytes: int64(blocks)*34 + int64(blocks)*32*4,
				fn: func() {
					sink += vec.DotRow(vec.Q8_0, q8row, scratch[:blocks*32], scratch)
				},
			},
			{
				name:  "rmsnorm",
				bytes: int64(n) * 4 * 3,
				fn:    func() { vec.RMSNorm(xnorm, xnorm, weight, 1e-6) },
			},
			{
				name:  "rope",
				bytes: int64(n) * 4 * 2,
				fn:    func() { vec.RotateHalf(xrope, cosv, sinv) },
			},
		}
		_ = sink
		_ = sinkB

		for _, op := range opDefs {
			if !wantOps[op.name] && !wantOps["all"] {
				continue
			}

			iters, elapsed := bench(dur, op.fn)

			opsPerS := float64(iters) / elapsed.Seconds()
			gbPerS := float64(op.bytes) * float64(iters) / elapsed.Seconds() / 1e9
			nsPerOp := elapsed.Seconds() * 1e9 / float64(iters)

			res := result{
				Op:      op.name,
				Level:   lvl,
				N:       n,
				OpsPerS: opsPerS,
				GBperS:  gbPerS,
				NsPerOp: nsPerOp,
			}
			results = append(results, res)

			if !asJSON {
				fmt.Printf("%-12s %-8s %8d %12.2f %10.3f %10.1f\n",
					op.name, lvl, n, opsPerS/1e6, gbPerS, nsPerOp)
			}
		}
	}

	if asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(results)
	}
}

func main() {
	opsFlag := flag.String("ops", "all", "операции через запятую: dot,axpy,matvec,q8,rmsnorm,rope или all")
	lvlFlag := flag.String("lvl", "all", "уровни через запятую: avx512,avx2,neon,generic или all")
	nFlag := flag.Int("n", 4096, "размер вектора")
	secFlag := flag.Float64("sec", 1.0, "время замера каждой операции (секунды)")
	jsonFlag := flag.Bool("json", false, "вывод в JSON")
	flag.Parse()

	ops := strings.Split(strings.ToLower(*opsFlag), ",")
	levels := strings.Split(strings.ToLower(*lvlFlag), ",")
	dur := time.Duration(*secFlag * float64(time.Second))

	if !*jsonFlag {
		fmt.Printf("goqwen vecbench  arch=%s  os=%s  GOMAXPROCS=%d\n",
			runtime.GOARCH, runtime.GOOS, runtime.GOMAXPROCS(0))
		fmt.Printf("доступные уровни: %s  активный: %s\n\n",
			strings.Join(vec.Levels(), ", "), vec.Level())
	}

	run(ops, levels, *nFlag, dur, *jsonFlag)
}
