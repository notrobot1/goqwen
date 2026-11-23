//go:build !amd64 && !arm64

package vec

// На остальных архитектурах (riscv64, wasm, 386, ...) используются
// переносимые реализации из generic.go.
