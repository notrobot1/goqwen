package nn

import (
	"fmt"
	"unsafe"

	"github.com/notrobot1/goqwen/goqwen/internal/gguf"
	"github.com/notrobot1/goqwen/goqwen/vec"
)

// Matrix — матрица весов rows×cols, хранящаяся построчно либо как float32,
// либо в исходном квантованном виде (тогда строки распаковываются на лету).
type Matrix struct {
	Rows, Cols int
	Format     vec.Format
	F32        []float32 // для Format == F32
	Raw        []byte    // для остальных форматов
	rowBytes   int
}

// nativeLittleEndian: можно ли читать float32 из файла напрямую в память.
var nativeLittleEndian = func() bool {
	x := uint16(1)
	return *(*byte)(unsafe.Pointer(&x)) == 1
}()

// loadMatrix читает тензор. Одномерный тензор трактуется как одна строка.
// При expand=true квантованные веса сразу распаковываются в float32.
func loadMatrix(t *gguf.Tensor, expand bool) (*Matrix, error) {
	f := vec.Format(t.Type)
	if !f.Supported() {
		return nil, fmt.Errorf("тензор %s: формат весов %s не поддерживается", t.Name, f)
	}
	m := &Matrix{Format: f}
	switch len(t.Dims) {
	case 1:
		m.Rows, m.Cols = 1, int(t.Dims[0])
	case 2:
		m.Rows, m.Cols = int(t.Dims[1]), int(t.Dims[0])
	default:
		return nil, fmt.Errorf("тензор %s: неожиданная размерность %v", t.Name, t.Dims)
	}
	elems, bytes := f.Block()
	if m.Cols%elems != 0 {
		return nil, fmt.Errorf("тензор %s: ширина %d не кратна блоку %s", t.Name, m.Cols, f)
	}
	m.rowBytes = m.Cols / elems * bytes

	if f == vec.F32 && nativeLittleEndian {
		m.F32 = make([]float32, m.Rows*m.Cols)
		buf := unsafe.Slice((*byte)(unsafe.Pointer(unsafe.SliceData(m.F32))), len(m.F32)*4)
		return m, t.ReadInto(buf)
	}
	m.Raw = make([]byte, m.Rows*m.rowBytes)
	if err := t.ReadInto(m.Raw); err != nil {
		return nil, err
	}
	if expand || f == vec.F32 {
		m.expand()
	}
	return m, nil
}

// expand переводит матрицу в float32.
func (m *Matrix) expand() {
	if m.Format == vec.F32 && m.F32 != nil {
		return
	}
	out := make([]float32, m.Rows*m.Cols)
	for i := 0; i < m.Rows; i++ {
		vec.Dequantize(m.Format, m.Raw[i*m.rowBytes:(i+1)*m.rowBytes], out[i*m.Cols:(i+1)*m.Cols])
	}
	m.F32, m.Raw, m.Format, m.rowBytes = out, nil, vec.F32, m.Cols*4
}

// Bytes возвращает объём памяти, занятый весами.
func (m *Matrix) Bytes() int { return len(m.F32)*4 + len(m.Raw) }

// Row записывает строку i в dst в виде float32.
func (m *Matrix) Row(i int, dst []float32) {
	if m.Format == vec.F32 {
		copy(dst, m.F32[i*m.Cols:(i+1)*m.Cols])
		return
	}
	vec.Dequantize(m.Format, m.Raw[i*m.rowBytes:(i+1)*m.rowBytes], dst[:m.Cols])
}

// MulVec вычисляет dst = M·x параллельно по строкам. scratch[worker] —
// личный буфер потока длиной не меньше Cols.
func (m *Matrix) MulVec(p *Pool, dst, x []float32, scratch [][]float32) {
	x = x[:m.Cols]
	p.For(m.Rows, m.Rows*m.Cols, func(lo, hi, w int) {
		if m.Format == vec.F32 {
			vec.MatVec(dst[lo:hi], m.F32[lo*m.Cols:hi*m.Cols], x)
			return
		}
		s := scratch[w]
		rb := m.rowBytes
		for i := lo; i < hi; i++ {
			dst[i] = vec.DotRow(m.Format, m.Raw[i*rb:(i+1)*rb], x, s)
		}
	})
}
