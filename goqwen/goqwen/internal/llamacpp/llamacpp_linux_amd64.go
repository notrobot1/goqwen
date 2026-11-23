//go:build linux && amd64 && cgo && llamacpp

package llamacpp

/*
#cgo CFLAGS: -I${SRCDIR}/include
#cgo LDFLAGS: ${SRCDIR}/libs/linux_amd64/libllama.a
#cgo LDFLAGS: ${SRCDIR}/libs/linux_amd64/libggml-cpu.a
#cgo LDFLAGS: ${SRCDIR}/libs/linux_amd64/libggml-base.a
#cgo LDFLAGS: ${SRCDIR}/libs/linux_amd64/libggml.a
#cgo LDFLAGS: -lstdc++ -lm
*/
import "C"
