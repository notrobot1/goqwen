package vec

import (
	"encoding/binary"
	"fmt"
	"math"
)

// Format — формат хранения весов (совпадает с нумерацией типов тензоров GGUF).
type Format uint32

const (
	F32  Format = 0
	F16  Format = 1
	Q4_0 Format = 2
	Q4_1 Format = 3
	Q5_0 Format = 6
	Q5_1 Format = 7
	Q8_0 Format = 8
	Q2_K Format = 10
	Q3_K Format = 11
	Q4_K Format = 12
	Q5_K Format = 13
	Q6_K Format = 14
	BF16 Format = 30
)

var formatInfo = map[Format]struct {
	name         string
	elems, bytes int
}{
	F32: {"F32", 1, 4}, F16: {"F16", 1, 2}, BF16: {"BF16", 1, 2},
	Q4_0: {"Q4_0", 32, 18}, Q4_1: {"Q4_1", 32, 20},
	Q5_0: {"Q5_0", 32, 22}, Q5_1: {"Q5_1", 32, 24},
	Q8_0: {"Q8_0", 32, 34},
	Q2_K: {"Q2_K", 256, 84}, Q3_K: {"Q3_K", 256, 110},
	Q4_K: {"Q4_K", 256, 144}, Q5_K: {"Q5_K", 256, 176}, Q6_K: {"Q6_K", 256, 210},
}

// Supported сообщает, умеет ли пакет читать этот формат.
func (f Format) Supported() bool { _, ok := formatInfo[f]; return ok }

// Block возвращает размер блока формата: число элементов и число байт.
func (f Format) Block() (elems, bytes int) {
	fi := formatInfo[f]
	return fi.elems, fi.bytes
}

func (f Format) String() string {
	if fi, ok := formatInfo[f]; ok {
		return fi.name
	}
	return fmt.Sprintf("GGML(%d)", uint32(f))
}

// f16Table — таблица перевода всех 65536 значений float16 в float32.
var f16Table [1 << 16]float32

func init() {
	for i := range f16Table {
		f16Table[i] = halfToFloat(uint16(i))
	}
}

func halfToFloat(h uint16) float32 {
	sign := uint32(h>>15) << 31
	exp := uint32(h>>10) & 0x1f
	man := uint32(h) & 0x3ff
	switch exp {
	case 0:
		if man == 0 {
			return math.Float32frombits(sign)
		}
		e := uint32(113) // денормализованное число: нормализуем мантиссу
		for man&0x400 == 0 {
			man <<= 1
			e--
		}
		return math.Float32frombits(sign | e<<23 | (man&0x3ff)<<13)
	case 31:
		return math.Float32frombits(sign | 0xff<<23 | man<<13)
	}
	return math.Float32frombits(sign | (exp+112)<<23 | man<<13)
}

func h16(b []byte) float32 { return f16Table[uint16(b[0])|uint16(b[1])<<8] }

// DotRow возвращает скалярное произведение строки весов row (в формате f) на
// вектор x. scratch — рабочий буфер длиной не меньше len(x); он нужен
// форматам, у которых нет прямого SIMD-ядра (строка сначала распаковывается).
func DotRow(f Format, row []byte, x, scratch []float32) float32 {
	if f == Q8_0 && q8Direct {
		return dotQ8Impl(row, x)
	}
	s := scratch[:len(x)]
	Dequantize(f, row, s)
	return dotImpl(s, x)
}

// Dequantize распаковывает src (формат f) в dst. len(dst) должно быть кратно
// размеру блока формата.
func Dequantize(f Format, src []byte, dst []float32) {
	switch f {
	case F32:
		for i := range dst {
			dst[i] = math.Float32frombits(binary.LittleEndian.Uint32(src[4*i:]))
		}
	case F16:
		for i := range dst {
			dst[i] = f16Table[uint16(src[2*i])|uint16(src[2*i+1])<<8]
		}
	case BF16:
		for i := range dst {
			dst[i] = math.Float32frombits(uint32(uint16(src[2*i])|uint16(src[2*i+1])<<8) << 16)
		}
	case Q8_0:
		for ; len(dst) >= 32; dst, src = dst[32:], src[34:] {
			d := h16(src)
			q, y := src[2:34], dst[:32]
			for j := range y {
				y[j] = d * float32(int8(q[j]))
			}
		}
	case Q4_0:
		for ; len(dst) >= 32; dst, src = dst[32:], src[18:] {
			d := h16(src)
			q, y := src[2:18], dst[:32]
			for j, b := range q {
				y[j] = d * float32(int(b&0xF)-8)
				y[j+16] = d * float32(int(b>>4)-8)
			}
		}
	case Q4_1:
		for ; len(dst) >= 32; dst, src = dst[32:], src[20:] {
			d, m := h16(src), h16(src[2:])
			q, y := src[4:20], dst[:32]
			for j, b := range q {
				y[j] = d*float32(b&0xF) + m
				y[j+16] = d*float32(b>>4) + m
			}
		}
	case Q5_0:
		for ; len(dst) >= 32; dst, src = dst[32:], src[22:] {
			d := h16(src)
			qh := binary.LittleEndian.Uint32(src[2:])
			q, y := src[6:22], dst[:32]
			for j, b := range q {
				h0 := uint8(qh>>uint(j)) & 1 << 4
				h1 := uint8(qh>>uint(j+16)) & 1 << 4
				y[j] = d * float32(int(b&0xF|h0)-16)
				y[j+16] = d * float32(int(b>>4|h1)-16)
			}
		}
	case Q5_1:
		for ; len(dst) >= 32; dst, src = dst[32:], src[24:] {
			d, m := h16(src), h16(src[2:])
			qh := binary.LittleEndian.Uint32(src[4:])
			q, y := src[8:24], dst[:32]
			for j, b := range q {
				h0 := uint8(qh>>uint(j)) & 1 << 4
				h1 := uint8(qh>>uint(j+16)) & 1 << 4
				y[j] = d*float32(b&0xF|h0) + m
				y[j+16] = d*float32(b>>4|h1) + m
			}
		}
	case Q2_K:
		for ; len(dst) >= 256; dst, src = dst[256:], src[84:] {
			deqQ2K(src[:84], dst[:256])
		}
	case Q3_K:
		for ; len(dst) >= 256; dst, src = dst[256:], src[110:] {
			deqQ3K(src[:110], dst[:256])
		}
	case Q4_K:
		for ; len(dst) >= 256; dst, src = dst[256:], src[144:] {
			deqQ4K(src[:144], dst[:256])
		}
	case Q5_K:
		for ; len(dst) >= 256; dst, src = dst[256:], src[176:] {
			deqQ5K(src[:176], dst[:256])
		}
	case Q6_K:
		for ; len(dst) >= 256; dst, src = dst[256:], src[210:] {
			deqQ6K(src[:210], dst[:256])
		}
	default:
		panic("vec: неподдерживаемый формат весов " + f.String())
	}
}

// scaleMinK4 достаёт 6-битные масштаб и минимум под-блока j из 12 байт
// упакованных масштабов (форматы Q4_K и Q5_K).
func scaleMinK4(j int, q []byte) (sc, m float32) {
	if j < 4 {
		return float32(q[j] & 63), float32(q[j+4] & 63)
	}
	return float32(q[j+4]&0xF | q[j-4]>>6<<4), float32(q[j+4]>>4 | q[j]>>6<<4)
}

func deqQ4K(b []byte, y []float32) {
	d, dmin := h16(b), h16(b[2:])
	scales, q := b[4:16], b[16:144]
	for g := 0; g < 4; g++ {
		sc, m := scaleMinK4(2*g, scales)
		d1, m1 := d*sc, dmin*m
		sc, m = scaleMinK4(2*g+1, scales)
		d2, m2 := d*sc, dmin*m
		qq := q[32*g : 32*g+32]
		y1, y2 := y[64*g:64*g+32], y[64*g+32:64*g+64]
		for l, v := range qq {
			y1[l] = d1*float32(v&0xF) - m1
			y2[l] = d2*float32(v>>4) - m2
		}
	}
}

func deqQ5K(b []byte, y []float32) {
	d, dmin := h16(b), h16(b[2:])
	scales, qh, q := b[4:16], b[16:48], b[48:176]
	for g := 0; g < 4; g++ {
		sc, m := scaleMinK4(2*g, scales)
		d1, m1 := d*sc, dmin*m
		sc, m = scaleMinK4(2*g+1, scales)
		d2, m2 := d*sc, dmin*m
		qq := q[32*g : 32*g+32]
		y1, y2 := y[64*g:64*g+32], y[64*g+32:64*g+64]
		s1, s2 := uint(2*g), uint(2*g+1)
		for l, v := range qq {
			h := qh[l]
			y1[l] = d1*float32(v&0xF|(h>>s1&1)<<4) - m1
			y2[l] = d2*float32(v>>4|(h>>s2&1)<<4) - m2
		}
	}
}

func deqQ6K(b []byte, y []float32) {
	ql, qh, sc := b[0:128], b[128:192], b[192:208]
	d := h16(b[208:])
	for n := 0; n < 2; n++ {
		l0, h0, s0, y0 := ql[64*n:64*n+64], qh[32*n:32*n+32], sc[8*n:8*n+8], y[128*n:128*n+128]
		for l := 0; l < 32; l++ {
			is := l / 16
			h := h0[l]
			q1 := int(l0[l]&0xF|(h&3)<<4) - 32
			q2 := int(l0[l+32]&0xF|(h>>2&3)<<4) - 32
			q3 := int(l0[l]>>4|(h>>4&3)<<4) - 32
			q4 := int(l0[l+32]>>4|(h>>6&3)<<4) - 32
			y0[l] = d * float32(int8(s0[is])) * float32(q1)
			y0[l+32] = d * float32(int8(s0[is+2])) * float32(q2)
			y0[l+64] = d * float32(int8(s0[is+4])) * float32(q3)
			y0[l+96] = d * float32(int8(s0[is+6])) * float32(q4)
		}
	}
}

func deqQ2K(b []byte, y []float32) {
	scales, qs := b[0:16], b[16:80]
	d, dmin := h16(b[80:]), h16(b[82:])
	is := 0
	for n := 0; n < 2; n++ {
		q := qs[32*n : 32*n+32]
		for j := 0; j < 4; j++ {
			shift := uint(2 * j)
			for half := 0; half < 2; half++ {
				sc := scales[is]
				is++
				dl, ml := d*float32(sc&0xF), dmin*float32(sc>>4)
				qq := q[16*half : 16*half+16]
				yy := y[128*n+32*j+16*half:]
				for l, v := range qq {
					yy[l] = dl*float32(v>>shift&3) - ml
				}
			}
		}
	}
}

func deqQ3K(b []byte, y []float32) {
	hm, qs, sb := b[0:32], b[32:96], b[96:108]
	dAll := h16(b[108:])

	const kmask1, kmask2 = 0x03030303, 0x0f0f0f0f
	var aux [4]uint32
	aux[0] = binary.LittleEndian.Uint32(sb[0:])
	aux[1] = binary.LittleEndian.Uint32(sb[4:])
	tmp := binary.LittleEndian.Uint32(sb[8:])
	aux[2] = (aux[0]>>4)&kmask2 | ((tmp>>4)&kmask1)<<4
	aux[3] = (aux[1]>>4)&kmask2 | ((tmp>>6)&kmask1)<<4
	aux[0] = aux[0]&kmask2 | ((tmp>>0)&kmask1)<<4
	aux[1] = aux[1]&kmask2 | ((tmp>>2)&kmask1)<<4
	var scales [16]int8
	for i := range scales {
		scales[i] = int8(aux[i/4] >> (8 * uint(i%4)))
	}

	is := 0
	m := uint8(1)
	for n := 0; n < 2; n++ {
		q := qs[32*n : 32*n+32]
		for j := 0; j < 4; j++ {
			shift := uint(2 * j)
			for half := 0; half < 2; half++ {
				dl := dAll * float32(int(scales[is])-32)
				is++
				qq := q[16*half : 16*half+16]
				hh := hm[16*half : 16*half+16]
				yy := y[128*n+32*j+16*half:]
				for l, v := range qq {
					val := int(v >> shift & 3)
					if hh[l]&m == 0 {
						val -= 4
					}
					yy[l] = dl * float32(val)
				}
			}
			m <<= 1
		}
	}
}
