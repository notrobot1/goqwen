//go:build arm64

package vec

import "unsafe"

// Реализованы в kernels_arm64.s. NEON (ASIMD) — обязательная часть ARMv8-A,
// поэтому определять его наличие во время выполнения не требуется.

// dotNEON обрабатывает n элементов (n > 0 и кратно 16) и записывает в acc
// шестнадцать частичных сумм.
//
//go:noescape
func dotNEON(a, b *float32, n int, acc *[16]float32)

// axpyNEON выполняет y += alpha*x для n элементов (n > 0 и кратно 8).
//
//go:noescape
func axpyNEON(alpha float32, x, y *float32, n int)

// rotateHalfNEON применяет RoPE к n элементам (n > 0, кратно 4).
//
//go:noescape
func rotateHalfNEON(xlo, xhi, cos, sin *float32, n int)

func init() {
	levels["neon"] = func() {
		dotImpl = dotNeon
		axpyImpl = axpyNeon
		dotQ8Impl = dotQ8Generic
		q8Direct = false
		rotateHalfImpl = rotateHalfNeon
		levelName = "neon"
	}
	levels["neon"]()
}

func rotateHalfNeon(x, cos, sin []float32) {
	half := len(cos)
	n := half &^ 7 // округлить вниз до кратного 8
	if n > 0 {
		rotateHalfNEON(
			unsafe.SliceData(x),
			unsafe.SliceData(x[half:]),
			unsafe.SliceData(cos),
			unsafe.SliceData(sin),
			n,
		)
	}
	// хвост (если half не кратно 8)
	hi := x[half : 2*half]
	for i := n; i < half; i++ {
		a, b, c, s := x[i], hi[i], cos[i], sin[i]
		x[i] = a*c - b*s
		hi[i] = b*c + a*s
	}
}

func dotNeon(a, b []float32) float32 {
	n := len(a) &^ 15
	var s float32
	if n > 0 {
		var acc [16]float32
		dotNEON(unsafe.SliceData(a), unsafe.SliceData(b), n, &acc)
		s = ((acc[0] + acc[1]) + (acc[2] + acc[3])) + ((acc[4] + acc[5]) + (acc[6] + acc[7])) +
			((acc[8] + acc[9]) + (acc[10] + acc[11])) + ((acc[12] + acc[13]) + (acc[14] + acc[15]))
	}
	b = b[:len(a)]
	for i := n; i < len(a); i++ {
		s += a[i] * b[i]
	}
	return s
}

func axpyNeon(alpha float32, x, y []float32) {
	n := len(x) &^ 7
	if n > 0 {
		axpyNEON(alpha, unsafe.SliceData(x), unsafe.SliceData(y), n)
	}
	y = y[:len(x)]
	for i := n; i < len(x); i++ {
		y[i] += alpha * x[i]
	}
}
