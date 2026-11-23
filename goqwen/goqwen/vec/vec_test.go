package vec

import (
	"math"
	"math/rand"
	"testing"
)

func randVec(r *rand.Rand, n int) []float32 {
	v := make([]float32, n)
	for i := range v {
		v[i] = r.Float32()*2 - 1
	}
	return v
}

func close32(a, b, tol float32) bool {
	d := float64(a - b)
	return math.Abs(d) <= float64(tol)*(1+math.Abs(float64(b)))
}

// Каждая доступная на машине реализация сверяется с переносимой на векторах
// всех длин от 0 до 300 (чтобы покрыть все «хвосты» циклов) и со смещением
// начала (невыровненные адреса).
func TestKernelsMatchGeneric(t *testing.T) {
	saved := Level()
	defer SetLevel(saved)
	r := rand.New(rand.NewSource(1))
	for _, lvl := range Levels() {
		if err := SetLevel(lvl); err != nil {
			t.Fatal(err)
		}
		for n := 0; n <= 300; n++ {
			for off := 0; off < 3; off++ {
				a, b := randVec(r, n+off)[off:], randVec(r, n+off)[off:]
				want := dotGeneric(a, b)
				if got := Dot(a, b); !close32(got, want, 1e-4) {
					t.Fatalf("%s: Dot n=%d: got %v want %v", lvl, n, got, want)
				}
				y1 := append([]float32(nil), b...)
				y2 := append([]float32(nil), b...)
				axpyGeneric(0.37, a, y1)
				Axpy(0.37, a, y2)
				for i := range y1 {
					if !close32(y2[i], y1[i], 1e-5) {
						t.Fatalf("%s: Axpy n=%d i=%d: got %v want %v", lvl, n, i, y2[i], y1[i])
					}
				}
			}
		}
		// Q8_0: прямое ядро против «распаковать и перемножить».
		for blocks := 1; blocks <= 40; blocks++ {
			row := make([]byte, blocks*34)
			r.Read(row)
			for b := 0; b < blocks; b++ { // масштаб — обычное положительное float16
				row[b*34+1] = 0x30 + byte(r.Intn(8))
			}
			x := randVec(r, blocks*32)
			deq := make([]float32, blocks*32)
			Dequantize(Q8_0, row, deq)
			want := dotGeneric(deq, x)
			scratch := make([]float32, blocks*32)
			if got := DotRow(Q8_0, row, x, scratch); !close32(got, want, 1e-3) {
				t.Fatalf("%s: DotRow Q8_0 blocks=%d: got %v want %v", lvl, blocks, got, want)
			}
			if got := dotQ8Generic(row, x); !close32(got, want, 1e-3) {
				t.Fatalf("dotQ8Generic blocks=%d: got %v want %v", blocks, got, want)
			}
		}
	}
}

func TestHalfToFloat(t *testing.T) {
	cases := map[uint16]float32{
		0x0000: 0, 0x3c00: 1, 0xc000: -2, 0x3555: 0.33325195, 0x7bff: 65504,
		0x0001: 5.9604645e-08, 0x0400: 6.1035156e-05, 0x03ff: 6.097555e-05,
	}
	for h, want := range cases {
		if got := halfToFloat(h); got != want {
			t.Errorf("halfToFloat(%#04x) = %v, want %v", h, got, want)
		}
	}
	if !math.IsInf(float64(halfToFloat(0x7c00)), 1) || !math.IsNaN(float64(halfToFloat(0x7e00))) {
		t.Error("inf/nan")
	}
}

func TestSoftmaxRMSNorm(t *testing.T) {
	x := []float32{1, 2, 3, 4}
	Softmax(x)
	var s float32
	for _, v := range x {
		s += v
	}
	if !close32(s, 1, 1e-6) || !(x[3] > x[2] && x[2] > x[1]) {
		t.Fatalf("softmax: %v", x)
	}
	v := []float32{3, 4}
	w := []float32{1, 1}
	RMSNorm(v, v, w, 0)
	if !close32(v[0], 3/float32(math.Sqrt(12.5)), 1e-6) {
		t.Fatalf("rmsnorm: %v", v)
	}
}

func benchDot(b *testing.B, lvl string, n int) {
	saved := Level()
	defer SetLevel(saved)
	if SetLevel(lvl) != nil {
		b.Skip("нет на этой машине")
	}
	r := rand.New(rand.NewSource(1))
	x, y := randVec(r, n), randVec(r, n)
	b.SetBytes(int64(n * 4))
	b.ResetTimer()
	var s float32
	for i := 0; i < b.N; i++ {
		s += Dot(x, y)
	}
	_ = s
}

func BenchmarkDotGeneric(b *testing.B) { benchDot(b, "generic", 4096) }
func BenchmarkDotAVX2(b *testing.B)    { benchDot(b, "avx2", 4096) }
func BenchmarkDotAVX512(b *testing.B)  { benchDot(b, "avx512", 4096) }
func BenchmarkDotNEON(b *testing.B)    { benchDot(b, "neon", 4096) }

func benchQ8(b *testing.B, lvl string) {
	saved := Level()
	defer SetLevel(saved)
	if SetLevel(lvl) != nil {
		b.Skip("нет на этой машине")
	}
	const blocks = 128
	r := rand.New(rand.NewSource(1))
	row := make([]byte, blocks*34)
	r.Read(row)
	x := randVec(r, blocks*32)
	scratch := make([]float32, blocks*32)
	b.SetBytes(blocks * 32 * 4)
	b.ResetTimer()
	var s float32
	for i := 0; i < b.N; i++ {
		s += DotRow(Q8_0, row, x, scratch)
	}
	_ = s
}

func BenchmarkQ8Generic(b *testing.B) { benchQ8(b, "generic") }
func BenchmarkQ8AVX2(b *testing.B)    { benchQ8(b, "avx2") }
func BenchmarkQ8AVX512(b *testing.B)  { benchQ8(b, "avx512") }
