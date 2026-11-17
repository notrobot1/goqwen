//go:build amd64

#include "textflag.h"

// func cpuid(eaxArg, ecxArg uint32) (eax, ebx, ecx, edx uint32)
TEXT ·cpuid(SB), NOSPLIT, $0-24
	MOVL eaxArg+0(FP), AX
	MOVL ecxArg+4(FP), CX
	CPUID
	MOVL AX, eax+8(FP)
	MOVL BX, ebx+12(FP)
	MOVL CX, ecx+16(FP)
	MOVL DX, edx+20(FP)
	RET

// func xgetbv() (eax, edx uint32)
TEXT ·xgetbv(SB), NOSPLIT, $0-8
	XORL CX, CX
	XGETBV
	MOVL AX, eax+0(FP)
	MOVL DX, edx+4(FP)
	RET

// ---------------------------------------------------------------------------
// AVX2 + FMA
// ---------------------------------------------------------------------------

// func dotAVX2(a, b *float32, n int) float32
TEXT ·dotAVX2(SB), NOSPLIT, $0-28
	MOVQ a+0(FP), SI
	MOVQ b+8(FP), DI
	MOVQ n+16(FP), CX
	VXORPS Y0, Y0, Y0
	VXORPS Y1, Y1, Y1
	VXORPS Y2, Y2, Y2
	VXORPS Y3, Y3, Y3
	CMPQ CX, $32
	JLT  dot2_tail8

dot2_loop32:
	VMOVUPS (SI), Y4
	VMOVUPS 32(SI), Y5
	VMOVUPS 64(SI), Y6
	VMOVUPS 96(SI), Y7
	VFMADD231PS (DI), Y4, Y0
	VFMADD231PS 32(DI), Y5, Y1
	VFMADD231PS 64(DI), Y6, Y2
	VFMADD231PS 96(DI), Y7, Y3
	ADDQ $128, SI
	ADDQ $128, DI
	SUBQ $32, CX
	CMPQ CX, $32
	JGE  dot2_loop32

dot2_tail8:
	CMPQ CX, $8
	JLT  dot2_reduce

dot2_loop8:
	VMOVUPS (SI), Y4
	VFMADD231PS (DI), Y4, Y0
	ADDQ $32, SI
	ADDQ $32, DI
	SUBQ $8, CX
	CMPQ CX, $8
	JGE  dot2_loop8

dot2_reduce:
	VADDPS Y1, Y0, Y0
	VADDPS Y3, Y2, Y2
	VADDPS Y2, Y0, Y0
	VEXTRACTF128 $1, Y0, X1
	VADDPS X1, X0, X0
	VHADDPS X0, X0, X0
	VHADDPS X0, X0, X0
	TESTQ CX, CX
	JZ   dot2_done

dot2_tail1:
	VMOVSS (SI), X1
	VFMADD231SS (DI), X1, X0
	ADDQ $4, SI
	ADDQ $4, DI
	DECQ CX
	JNZ  dot2_tail1

dot2_done:
	VZEROUPPER
	MOVSS X0, ret+24(FP)
	RET

// func axpyAVX2(alpha float32, x, y *float32, n int)
TEXT ·axpyAVX2(SB), NOSPLIT, $0-32
	VBROADCASTSS alpha+0(FP), Y0
	MOVQ x+8(FP), SI
	MOVQ y+16(FP), DI
	MOVQ n+24(FP), CX
	CMPQ CX, $16
	JLT  axpy2_tail8

axpy2_loop16:
	VMOVUPS (DI), Y1
	VMOVUPS 32(DI), Y2
	VFMADD231PS (SI), Y0, Y1
	VFMADD231PS 32(SI), Y0, Y2
	VMOVUPS Y1, (DI)
	VMOVUPS Y2, 32(DI)
	ADDQ $64, SI
	ADDQ $64, DI
	SUBQ $16, CX
	CMPQ CX, $16
	JGE  axpy2_loop16

axpy2_tail8:
	CMPQ CX, $8
	JLT  axpy2_tail

	VMOVUPS (DI), Y1
	VFMADD231PS (SI), Y0, Y1
	VMOVUPS Y1, (DI)
	ADDQ $32, SI
	ADDQ $32, DI
	SUBQ $8, CX

axpy2_tail:
	TESTQ CX, CX
	JZ   axpy2_done

axpy2_tail1:
	VMOVSS (DI), X1
	VFMADD231SS (SI), X0, X1
	VMOVSS X1, (DI)
	ADDQ $4, SI
	ADDQ $4, DI
	DECQ CX
	JNZ  axpy2_tail1

axpy2_done:
	VZEROUPPER
	RET

// func dotQ8AVX2(blocks *byte, x *float32, nblocks int) float32
//
// Блок Q8_0 = 2 байта масштаба (float16) + 32 байта int8. Для каждого блока:
// int8 расширяются до int32, переводятся в float32, перемножаются с x и
// суммируются; сумма блока умножается на масштаб и добавляется к общей.
TEXT ·dotQ8AVX2(SB), NOSPLIT, $0-28
	MOVQ blocks+0(FP), SI
	MOVQ x+8(FP), DI
	MOVQ nblocks+16(FP), CX
	VXORPS Y0, Y0, Y0

q8_2_loop:
	MOVWLZX (SI), AX
	VMOVD AX, X1
	VCVTPH2PS X1, X1
	VBROADCASTSS X1, Y1
	VPMOVSXBD 2(SI), Y2
	VPMOVSXBD 10(SI), Y3
	VPMOVSXBD 18(SI), Y4
	VPMOVSXBD 26(SI), Y5
	VCVTDQ2PS Y2, Y2
	VCVTDQ2PS Y3, Y3
	VCVTDQ2PS Y4, Y4
	VCVTDQ2PS Y5, Y5
	VMULPS (DI), Y2, Y2
	VMULPS 32(DI), Y3, Y3
	VFMADD231PS 64(DI), Y4, Y2
	VFMADD231PS 96(DI), Y5, Y3
	VADDPS Y3, Y2, Y2
	VFMADD231PS Y2, Y1, Y0
	ADDQ $34, SI
	ADDQ $128, DI
	DECQ CX
	JNZ  q8_2_loop

	VEXTRACTF128 $1, Y0, X1
	VADDPS X1, X0, X0
	VHADDPS X0, X0, X0
	VHADDPS X0, X0, X0
	VZEROUPPER
	MOVSS X0, ret+24(FP)
	RET

// ---------------------------------------------------------------------------
// AVX-512F
// ---------------------------------------------------------------------------

// func dotAVX512(a, b *float32, n int) float32
TEXT ·dotAVX512(SB), NOSPLIT, $0-28
	MOVQ a+0(FP), SI
	MOVQ b+8(FP), DI
	MOVQ n+16(FP), CX
	VPXORD Z0, Z0, Z0
	VPXORD Z1, Z1, Z1
	VPXORD Z2, Z2, Z2
	VPXORD Z3, Z3, Z3
	CMPQ CX, $64
	JLT  dot5_tail16

dot5_loop64:
	VMOVUPS (SI), Z4
	VMOVUPS 64(SI), Z5
	VMOVUPS 128(SI), Z6
	VMOVUPS 192(SI), Z7
	VFMADD231PS (DI), Z4, Z0
	VFMADD231PS 64(DI), Z5, Z1
	VFMADD231PS 128(DI), Z6, Z2
	VFMADD231PS 192(DI), Z7, Z3
	ADDQ $256, SI
	ADDQ $256, DI
	SUBQ $64, CX
	CMPQ CX, $64
	JGE  dot5_loop64

dot5_tail16:
	CMPQ CX, $16
	JLT  dot5_reduce

dot5_loop16:
	VMOVUPS (SI), Z4
	VFMADD231PS (DI), Z4, Z0
	ADDQ $64, SI
	ADDQ $64, DI
	SUBQ $16, CX
	CMPQ CX, $16
	JGE  dot5_loop16

dot5_reduce:
	VADDPS Z1, Z0, Z0
	VADDPS Z3, Z2, Z2
	VADDPS Z2, Z0, Z0
	VEXTRACTF64X4 $1, Z0, Y1
	VADDPS Y1, Y0, Y0
	CMPQ CX, $8
	JLT  dot5_fold

	VMOVUPS (SI), Y4
	VFMADD231PS (DI), Y4, Y0
	ADDQ $32, SI
	ADDQ $32, DI
	SUBQ $8, CX

dot5_fold:
	VEXTRACTF128 $1, Y0, X1
	VADDPS X1, X0, X0
	VHADDPS X0, X0, X0
	VHADDPS X0, X0, X0
	TESTQ CX, CX
	JZ   dot5_done

dot5_tail1:
	VMOVSS (SI), X1
	VFMADD231SS (DI), X1, X0
	ADDQ $4, SI
	ADDQ $4, DI
	DECQ CX
	JNZ  dot5_tail1

dot5_done:
	VZEROUPPER
	MOVSS X0, ret+24(FP)
	RET

// func axpyAVX512(alpha float32, x, y *float32, n int)
TEXT ·axpyAVX512(SB), NOSPLIT, $0-32
	VBROADCASTSS alpha+0(FP), Z0
	MOVQ x+8(FP), SI
	MOVQ y+16(FP), DI
	MOVQ n+24(FP), CX
	CMPQ CX, $16
	JLT  axpy5_tail8

axpy5_loop16:
	VMOVUPS (DI), Z1
	VFMADD231PS (SI), Z0, Z1
	VMOVUPS Z1, (DI)
	ADDQ $64, SI
	ADDQ $64, DI
	SUBQ $16, CX
	CMPQ CX, $16
	JGE  axpy5_loop16

axpy5_tail8:
	CMPQ CX, $8
	JLT  axpy5_tail

	VMOVUPS (DI), Y1
	VFMADD231PS (SI), Y0, Y1
	VMOVUPS Y1, (DI)
	ADDQ $32, SI
	ADDQ $32, DI
	SUBQ $8, CX

axpy5_tail:
	TESTQ CX, CX
	JZ   axpy5_done

axpy5_tail1:
	VMOVSS (DI), X1
	VFMADD231SS (SI), X0, X1
	VMOVSS X1, (DI)
	ADDQ $4, SI
	ADDQ $4, DI
	DECQ CX
	JNZ  axpy5_tail1

axpy5_done:
	VZEROUPPER
	RET

// func dotQ8AVX512(blocks *byte, x *float32, nblocks int) float32
TEXT ·dotQ8AVX512(SB), NOSPLIT, $0-28
	MOVQ blocks+0(FP), SI
	MOVQ x+8(FP), DI
	MOVQ nblocks+16(FP), CX
	VPXORD Z0, Z0, Z0

q8_5_loop:
	MOVWLZX (SI), AX
	VMOVD AX, X1
	VCVTPH2PS X1, X1
	VBROADCASTSS X1, Z1
	VPMOVSXBD 2(SI), Z2
	VPMOVSXBD 18(SI), Z3
	VCVTDQ2PS Z2, Z2
	VCVTDQ2PS Z3, Z3
	VMULPS (DI), Z2, Z2
	VFMADD231PS 64(DI), Z3, Z2
	VFMADD231PS Z2, Z1, Z0
	ADDQ $34, SI
	ADDQ $128, DI
	DECQ CX
	JNZ  q8_5_loop

	VEXTRACTF64X4 $1, Z0, Y1
	VADDPS Y1, Y0, Y0
	VEXTRACTF128 $1, Y0, X1
	VADDPS X1, X0, X0
	VHADDPS X0, X0, X0
	VHADDPS X0, X0, X0
	VZEROUPPER
	MOVSS X0, ret+24(FP)
	RET
