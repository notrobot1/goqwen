//go:build linux && arm64 && cgo && llamacpp

package llamacpp

/*
#cgo CFLAGS: -I${SRCDIR}/include
#cgo LDFLAGS: ${SRCDIR}/libs/linux_arm64/libllama.a
#cgo LDFLAGS: ${SRCDIR}/libs/linux_arm64/libggml-cpu.a
#cgo LDFLAGS: ${SRCDIR}/libs/linux_arm64/libggml-base.a
#cgo LDFLAGS: ${SRCDIR}/libs/linux_arm64/libggml.a
#cgo LDFLAGS: -lstdc++ -lm
*/
import "C"
