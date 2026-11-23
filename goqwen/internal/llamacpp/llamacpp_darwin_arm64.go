//go:build darwin && arm64 && cgo && llamacpp

package llamacpp

/*
#cgo CFLAGS: -I${SRCDIR}/include
#cgo LDFLAGS: ${SRCDIR}/libs/darwin_arm64/libllama.a
#cgo LDFLAGS: ${SRCDIR}/libs/darwin_arm64/libggml-metal.a
#cgo LDFLAGS: ${SRCDIR}/libs/darwin_arm64/libggml-cpu.a
#cgo LDFLAGS: ${SRCDIR}/libs/darwin_arm64/libggml-base.a
#cgo LDFLAGS: ${SRCDIR}/libs/darwin_arm64/libggml.a
#cgo LDFLAGS: -lstdc++
#cgo LDFLAGS: -framework Foundation -framework Metal -framework MetalKit -framework Accelerate
*/
import "C"
