//go:build arm64

#include "textflag.h"

// func dotNEON(a, b *float32, n int, acc *[16]float32)
// n > 0, кратно 16. Четыре независимых аккумулятора по 4 float32.
TEXT ·dotNEON(SB), NOSPLIT, $0-32
	MOVD a+0(FP), R0
	MOVD b+8(FP), R1
	MOVD n+16(FP), R2
	MOVD acc+24(FP), R3
	VEOR V0.B16, V0.B16, V0.B16
	VEOR V1.B16, V1.B16, V1.B16
	VEOR V2.B16, V2.B16, V2.B16
	VEOR V3.B16, V3.B16, V3.B16

dot_loop:
	VLD1.P 64(R0), [V4.S4, V5.S4, V6.S4, V7.S4]
	VLD1.P 64(R1), [V16.S4, V17.S4, V18.S4, V19.S4]
	VFMLA V4.S4, V16.S4, V0.S4
	VFMLA V5.S4, V17.S4, V1.S4
	VFMLA V6.S4, V18.S4, V2.S4
	VFMLA V7.S4, V19.S4, V3.S4
	SUBS $16, R2, R2
	BNE  dot_loop

	VST1 [V0.S4, V1.S4, V2.S4, V3.S4], (R3)
	RET

// func axpyNEON(alpha float32, x, y *float32, n int)
// n > 0, кратно 8.
TEXT ·axpyNEON(SB), NOSPLIT, $0-32
	MOVWU alpha+0(FP), R4
	MOVD x+8(FP), R0
	MOVD y+16(FP), R1
	MOVD n+24(FP), R2
	VDUP R4, V1.S4

axpy_loop:
	VLD1.P 32(R0), [V2.S4, V3.S4]
	VLD1 (R1), [V4.S4, V5.S4]
	VFMLA V2.S4, V1.S4, V4.S4
	VFMLA V3.S4, V1.S4, V5.S4
	VST1.P [V4.S4, V5.S4], 32(R1)
	SUBS $8, R2, R2
	BNE  axpy_loop
	RET
