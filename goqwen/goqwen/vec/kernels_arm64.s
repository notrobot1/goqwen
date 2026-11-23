//go:build arm64

#include "textflag.h"

// func dotNEON(a, b *float32, n int, acc *[32]float32)
// n > 0, кратно 32. Восемь независимых аккумуляторов × 4 float32 = 32 частичных суммы.
// На M1: 2 FMA-юнита × 4-цикловый latency → 8 цепочек полностью заполняют конвейер.
TEXT ·dotNEON(SB), NOSPLIT, $0-32
	MOVD a+0(FP), R0
	MOVD b+8(FP), R1
	MOVD n+16(FP), R2
	MOVD acc+24(FP), R3
	VEOR V0.B16, V0.B16, V0.B16
	VEOR V1.B16, V1.B16, V1.B16
	VEOR V2.B16, V2.B16, V2.B16
	VEOR V3.B16, V3.B16, V3.B16
	VEOR V4.B16, V4.B16, V4.B16
	VEOR V5.B16, V5.B16, V5.B16
	VEOR V6.B16, V6.B16, V6.B16
	VEOR V7.B16, V7.B16, V7.B16

dot_loop:
	VLD1.P 64(R0), [V8.S4, V9.S4, V10.S4, V11.S4]
	VLD1.P 64(R0), [V12.S4, V13.S4, V14.S4, V15.S4]
	VLD1.P 64(R1), [V16.S4, V17.S4, V18.S4, V19.S4]
	VLD1.P 64(R1), [V20.S4, V21.S4, V22.S4, V23.S4]
	VFMLA V8.S4,  V16.S4, V0.S4
	VFMLA V9.S4,  V17.S4, V1.S4
	VFMLA V10.S4, V18.S4, V2.S4
	VFMLA V11.S4, V19.S4, V3.S4
	VFMLA V12.S4, V20.S4, V4.S4
	VFMLA V13.S4, V21.S4, V5.S4
	VFMLA V14.S4, V22.S4, V6.S4
	VFMLA V15.S4, V23.S4, V7.S4
	SUBS $32, R2, R2
	BNE  dot_loop

	VST1 [V0.S4, V1.S4, V2.S4, V3.S4], (R3)
	ADD  $64, R3, R3
	VST1 [V4.S4, V5.S4, V6.S4, V7.S4], (R3)
	RET

// func scaleMulNEON(dst, x, weight *float32, scale float32, n int)
// dst[i] = x[i] * scale * weight[i], n > 0, кратно 8.
TEXT ·scaleMulNEON(SB), NOSPLIT, $0-40
	MOVD  dst+0(FP), R0
	MOVD  x+8(FP), R1
	MOVD  weight+16(FP), R2
	MOVWU scale+24(FP), R4
	MOVD  n+32(FP), R3
	VDUP  R4, V0.S4            // V0 = {scale, scale, scale, scale}

scalemul_loop:
	VLD1.P 32(R1), [V1.S4, V2.S4]  // x[0:8]
	VLD1.P 32(R2), [V3.S4, V4.S4]  // weight[0:8]
	// tmp = x * scale  (используем FMUL через нулевой аккумулятор)
	VEOR V5.B16, V5.B16, V5.B16
	VEOR V6.B16, V6.B16, V6.B16
	VFMLA V1.S4, V0.S4, V5.S4      // V5 = x[0:4] * scale
	VFMLA V2.S4, V0.S4, V6.S4      // V6 = x[4:8] * scale
	// dst = tmp * weight
	VEOR V7.B16, V7.B16, V7.B16
	VEOR V8.B16, V8.B16, V8.B16
	VFMLA V5.S4, V3.S4, V7.S4      // V7 = tmp[0:4] * weight[0:4]
	VFMLA V6.S4, V4.S4, V8.S4      // V8 = tmp[4:8] * weight[4:8]
	VST1.P [V7.S4, V8.S4], 32(R0)
	SUBS $8, R3, R3
	BNE  scalemul_loop
	RET

// func rotateHalfNEON(xlo, xhi, cos, sin *float32, n int)
// Применяет RoPE к n элементам (кратно 8):
//   xlo[i] = a*c - b*s,  xhi[i] = b*c + a*s
// Обрабатывает 8 элементов за итерацию (4 независимых цепочки),
// чтобы скрыть 4-цикловый latency VFMLA на Apple M-серии.
TEXT ·rotateHalfNEON(SB), NOSPLIT, $0-40
	MOVD xlo+0(FP), R0
	MOVD xhi+8(FP), R1
	MOVD cos+16(FP), R2
	MOVD sin+24(FP), R3
	MOVD n+32(FP), R4

rope_loop:
	// Загрузить 8 элементов из каждого массива.
	// xlo/xhi — без инкремента (запись пойдёт через VST1.P);
	// cos/sin — с пост-инкрементом (читаются один раз).
	VLD1 (R0), [V0.S4, V1.S4]       // a0 = xlo[0..3], a1 = xlo[4..7]
	VLD1 (R1), [V2.S4, V3.S4]       // b0 = xhi[0..3], b1 = xhi[4..7]
	VLD1.P 32(R2), [V4.S4, V5.S4]   // c0, c1
	VLD1.P 32(R3), [V6.S4, V7.S4]   // s0, s1

	// Четыре независимых аккумулятора — OOO-ядро M1 выполняет их параллельно.
	VEOR V8.B16, V8.B16, V8.B16     // new_a0 = 0
	VEOR V9.B16, V9.B16, V9.B16     // new_a1 = 0
	VEOR V10.B16, V10.B16, V10.B16  // new_b0 = 0
	VEOR V11.B16, V11.B16, V11.B16  // new_b1 = 0

	// Первый проход: все четыре цепочки стартуют одновременно.
	VFMLA V4.S4, V0.S4, V8.S4   // V8  = a0*c0
	VFMLA V5.S4, V1.S4, V9.S4   // V9  = a1*c1
	VFMLA V4.S4, V2.S4, V10.S4  // V10 = b0*c0
	VFMLA V5.S4, V3.S4, V11.S4  // V11 = b1*c1

	// Второй проход: завершает каждую цепочку.
	VFMLS V6.S4, V2.S4, V8.S4   // V8  = a0*c0 - b0*s0
	VFMLS V7.S4, V3.S4, V9.S4   // V9  = a1*c1 - b1*s1
	VFMLA V6.S4, V0.S4, V10.S4  // V10 = b0*c0 + a0*s0
	VFMLA V7.S4, V1.S4, V11.S4  // V11 = b1*c1 + a1*s1

	// Запись: пост-инкремент продвигает R0/R1 на 32 байта.
	VST1.P [V8.S4, V9.S4], 32(R0)
	VST1.P [V10.S4, V11.S4], 32(R1)

	SUBS $8, R4, R4
	BNE  rope_loop

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
