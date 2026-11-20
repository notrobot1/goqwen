// Package vec содержит все векторные вычисления библиотеки.
//
// Здесь собраны «горячие» ядра (скалярное произведение, axpy, произведение
// квантованной строки на вектор) и их аппаратно-оптимизированные варианты:
//
//	amd64: AVX-512F и AVX2+FMA+F16C — выбираются во время выполнения по CPUID
//	arm64: NEON (ASIMD) — есть на любом ARMv8
//	прочие архитектуры и старые x86: переносимая реализация на чистом Go
//
// Никаких cgo и внешних пакетов: ассемблер — штатный ассемблер Go, определение
// возможностей процессора реализовано в этом же пакете.
package vec

import (
	"fmt"
	"math"
	"strings"
)

// Активные реализации ядер. По умолчанию — переносимые; init() конкретной
// архитектуры заменяет их на оптимизированные.
var (
	dotImpl          func(a, b []float32) float32          = dotGeneric
	axpyImpl         func(alpha float32, x, y []float32)   = axpyGeneric
	dotQ8Impl        func(row []byte, x []float32) float32 = dotQ8Generic
	rotateHalfImpl   func(x, cos, sin []float32)           = rotateHalfGeneric
	levelName                                              = "generic"
	// q8Direct: есть ли прямое SIMD-ядро для строк Q8_0 (без распаковки).
	q8Direct = false
	levels   = map[string]func(){"generic": useGeneric}
)

func useGeneric() {
	dotImpl, axpyImpl, dotQ8Impl = dotGeneric, axpyGeneric, dotQ8Generic
	rotateHalfImpl = rotateHalfGeneric
	q8Direct = false
	levelName = "generic"
}

// Level возвращает имя используемого набора инструкций:
// "avx512", "avx2", "neon" или "generic".
func Level() string { return levelName }

// Levels возвращает наборы инструкций, доступные на этой машине,
// от самого быстрого к самому медленному.
func Levels() []string {
	out := make([]string, 0, len(levels))
	for _, n := range []string{"avx512", "avx2", "neon", "generic"} {
		if _, ok := levels[n]; ok {
			out = append(out, n)
		}
	}
	return out
}

// SetLevel принудительно выбирает набор инструкций (для отладки и бенчмарков).
// Выбрать можно только то, что реально поддерживает процессор.
// Нельзя вызывать одновременно с вычислениями.
func SetLevel(name string) error {
	f, ok := levels[strings.ToLower(name)]
	if !ok {
		return fmt.Errorf("vec: набор инструкций %q недоступен (доступны: %s)", name, strings.Join(Levels(), ", "))
	}
	f()
	return nil
}

// Dot возвращает скалярное произведение a·b по первым len(a) элементам.
func Dot(a, b []float32) float32 {
	if len(a) == 0 {
		return 0
	}
	return dotImpl(a, b[:len(a)])
}

// Axpy выполняет y += alpha*x по первым len(x) элементам.
func Axpy(alpha float32, x, y []float32) {
	if len(x) == 0 {
		return
	}
	axpyImpl(alpha, x, y[:len(x)])
}

// Add выполняет dst += x.
func Add(dst, x []float32) { Axpy(1, x, dst) }

// MatVec вычисляет dst[i] = w[i*cols:(i+1)*cols] · x для плотной матрицы
// float32, хранящейся построчно (cols = len(x)).
func MatVec(dst, w, x []float32) {
	cols := len(x)
	if cols == 0 {
		return
	}
	for i := range dst {
		dst[i] = dotImpl(w[i*cols:(i+1)*cols], x)
	}
}

// RMSNorm: dst = x / sqrt(mean(x²)+eps) * weight. dst и x могут совпадать.
func RMSNorm(dst, x, weight []float32, eps float32) {
	ss := Dot(x, x)
	scale := float32(1 / math.Sqrt(float64(ss/float32(len(x))+eps)))
	weight = weight[:len(x)]
	dst = dst[:len(x)]
	for i, v := range x {
		dst[i] = v * scale * weight[i]
	}
}

// Softmax преобразует x в распределение вероятностей на месте.
func Softmax(x []float32) {
	if len(x) == 0 {
		return
	}
	mx := x[0]
	for _, v := range x[1:] {
		if v > mx {
			mx = v
		}
	}
	var sum float32
	for i, v := range x {
		e := float32(math.Exp(float64(v - mx)))
		x[i] = e
		sum += e
	}
	inv := 1 / sum
	for i := range x {
		x[i] *= inv
	}
}

// SiLUMul выполняет g[i] = silu(g[i]) * u[i] — ядро блока SwiGLU.
func SiLUMul(g, u []float32) {
	u = u[:len(g)]
	for i, v := range g {
		g[i] = v / (1 + float32(math.Exp(float64(-v)))) * u[i]
	}
}

// RotateHalf применяет RoPE в варианте «половинами» (как в Qwen) к вектору x
// длины 2*len(cos): пара (x[i], x[i+half]) поворачивается на угол с косинусом
// cos[i] и синусом sin[i].
func RotateHalf(x, cos, sin []float32) {
	rotateHalfImpl(x, cos, sin)
}

// ArgMax возвращает индекс максимального элемента.
func ArgMax(x []float32) int {
	best := 0
	for i, v := range x {
		if v > x[best] {
			best = i
		}
	}
	return best
}
