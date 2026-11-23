//go:build arm64

package vec

import "unsafe"

// dotNEON обрабатывает n элементов (n > 0 и кратно 32) и записывает в acc
// 32 частичные суммы (8 аккумуляторов × 4 float32).
//
//go:noescape
func dotNEON(a, b *float32, n int, acc *[32]float32)

// scaleMulNEON: dst[i] = x[i] * scale * weight[i], n > 0, кратно 8.
//
//go:noescape
func scaleMulNEON(dst, x, weight *float32, scale float32, n int)

// axpyNEON выполняет y += alpha*x для n элементов (n > 0 и кратно 8).
//
//go:noescape
func axpyNEON(alpha float32, x, y *float32, n int)

// rotateHalfNEON применяет RoPE к n элементам (n > 0, кратно 8).
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
		scaleMulImpl = scaleMulNeon
		levelName = "neon"
	}
	levels["neon"]()
}

func dotNeon(a, b []float32) float32 {
	n := len(a) &^ 31 // кратно 32
	var s float32
	if n > 0 {
		var acc [32]float32
		dotNEON(unsafe.SliceData(a), unsafe.SliceData(b), n, &acc)
		s = (((acc[0]+acc[1])+(acc[2]+acc[3]))+((acc[4]+acc[5])+(acc[6]+acc[7]))) +
			(((acc[8]+acc[9])+(acc[10]+acc[11]))+((acc[12]+acc[13])+(acc[14]+acc[15]))) +
			(((acc[16]+acc[17])+(acc[18]+acc[19]))+((acc[20]+acc[21])+(acc[22]+acc[23]))) +
			(((acc[24]+acc[25])+(acc[26]+acc[27]))+((acc[28]+acc[29])+(acc[30]+acc[31])))
	}
	b = b[:len(a)]
	for i := n; i < len(a); i++ {
		s += a[i] * b[i]
	}
	return s
}

// scaleMulNeon: dst[i] = x[i] * scale * weight[i]. Вызывается из RMSNorm.
func scaleMulNeon(dst, x, weight []float32, scale float32) {
	n := len(x) &^ 7
	if n > 0 {
		scaleMulNEON(
			unsafe.SliceData(dst),
			unsafe.SliceData(x),
			unsafe.SliceData(weight),
			scale, n,
		)
	}
	weight = weight[:len(x)]
	for i := n; i < len(x); i++ {
		dst[i] = x[i] * scale * weight[i]
	}
}

func rotateHalfNeon(x, cos, sin []float32) {
	half := len(cos)
	n := half &^ 7
	if n > 0 {
		rotateHalfNEON(
			unsafe.SliceData(x),
			unsafe.SliceData(x[half:]),
			unsafe.SliceData(cos),
			unsafe.SliceData(sin),
			n,
		)
	}
	hi := x[half : 2*half]
	for i := n; i < half; i++ {
		a, b, c, s := x[i], hi[i], cos[i], sin[i]
		x[i] = a*c - b*s
		hi[i] = b*c + a*s
	}
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
