//go:build amd64

package vec

import "unsafe"

// Реализованы в kernels_amd64.s.

//go:noescape
func cpuid(eaxArg, ecxArg uint32) (eax, ebx, ecx, edx uint32)

//go:noescape
func xgetbv() (eax, edx uint32)

//go:noescape
func dotAVX2(a, b *float32, n int) float32

//go:noescape
func axpyAVX2(alpha float32, x, y *float32, n int)

//go:noescape
func dotQ8AVX2(blocks *byte, x *float32, nblocks int) float32

//go:noescape
func dotAVX512(a, b *float32, n int) float32

//go:noescape
func axpyAVX512(alpha float32, x, y *float32, n int)

//go:noescape
func dotQ8AVX512(blocks *byte, x *float32, nblocks int) float32

func init() {
	hasAVX2, hasAVX512 := detectX86()
	if hasAVX2 {
		levels["avx2"] = func() {
			dotImpl = func(a, b []float32) float32 {
				return dotAVX2(unsafe.SliceData(a), unsafe.SliceData(b), len(a))
			}
			axpyImpl = func(alpha float32, x, y []float32) {
				axpyAVX2(alpha, unsafe.SliceData(x), unsafe.SliceData(y), len(x))
			}
			dotQ8Impl = func(row []byte, x []float32) float32 {
				n := min(len(row)/34, len(x)/32)
				if n == 0 {
					return 0
				}
				return dotQ8AVX2(unsafe.SliceData(row), unsafe.SliceData(x), n)
			}
			q8Direct = true
			levelName = "avx2"
		}
		levels["avx2"]()
	}
	if hasAVX512 {
		levels["avx512"] = func() {
			dotImpl = func(a, b []float32) float32 {
				return dotAVX512(unsafe.SliceData(a), unsafe.SliceData(b), len(a))
			}
			axpyImpl = func(alpha float32, x, y []float32) {
				axpyAVX512(alpha, unsafe.SliceData(x), unsafe.SliceData(y), len(x))
			}
			dotQ8Impl = func(row []byte, x []float32) float32 {
				n := min(len(row)/34, len(x)/32)
				if n == 0 {
					return 0
				}
				return dotQ8AVX512(unsafe.SliceData(row), unsafe.SliceData(x), n)
			}
			q8Direct = true
			levelName = "avx512"
		}
		levels["avx512"]()
	}
}

// detectX86 определяет по CPUID/XGETBV, какие расширения можно использовать:
// нужна поддержка и процессором, и операционной системой (сохранение
// регистров YMM/ZMM при переключении контекста).
func detectX86() (avx2, avx512 bool) {
	maxLeaf, _, _, _ := cpuid(0, 0)
	if maxLeaf < 7 {
		return
	}
	_, _, ecx1, _ := cpuid(1, 0)
	const (
		fma     = 1 << 12
		osxsave = 1 << 27
		avx     = 1 << 28
		f16c    = 1 << 29
	)
	if ecx1&osxsave == 0 || ecx1&avx == 0 {
		return
	}
	xcr0, _ := xgetbv()
	if xcr0&0x6 != 0x6 { // ОС сохраняет XMM и YMM
		return
	}
	_, ebx7, _, _ := cpuid(7, 0)
	avx2 = ebx7&(1<<5) != 0 && ecx1&fma != 0 && ecx1&f16c != 0
	// AVX-512F + ОС сохраняет opmask и ZMM (биты 5..7 XCR0).
	avx512 = avx2 && ebx7&(1<<16) != 0 && xcr0&0xE0 == 0xE0
	return
}
