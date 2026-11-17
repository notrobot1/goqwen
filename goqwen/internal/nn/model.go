// Package nn реализует прямой проход трансформера Qwen (Qwen2/Qwen2.5/Qwen3):
// RMSNorm, внимание с групповыми запросами (GQA), RoPE, SwiGLU и KV-кэш.
// Вся арифметика делегируется пакету vec.
package nn

import (
	"errors"
	"fmt"
	"math"

	"github.com/notrobot1/goqwen/goqwen/internal/gguf"
)

// Config — гиперпараметры модели из метаданных GGUF.
type Config struct {
	Arch     string // "qwen2" или "qwen3"
	Name     string
	Layers   int
	Dim      int
	Heads    int
	HeadsKV  int
	HeadDim  int
	FF       int
	CtxTrain int // длина контекста, на которой обучена модель
	Vocab    int
	Eps      float32
	RopeBase float64
}

// Layer — веса одного блока трансформера.
type Layer struct {
	AttnNorm, FFNNorm   []float32
	Q, K, V, O          *Matrix
	QBias, KBias, VBias []float32 // есть в Qwen2/2.5, нет в Qwen3
	QNorm, KNorm        []float32 // есть только в Qwen3
	Gate, Up, Down      *Matrix
}

// Model — загруженные веса. После загрузки только читается, поэтому одну
// модель можно использовать из нескольких State одновременно.
type Model struct {
	Config
	Embd    *Matrix
	Blocks  []Layer
	OutNorm []float32
	Out     *Matrix
	Pool    *Pool

	invFreq []float64
	maxCols int
}

// Load читает веса из открытого GGUF. progress (может быть nil) вызывается
// после каждого прочитанного тензора.
func Load(f *gguf.File, threads int, expand bool, progress func(done, total int)) (*Model, error) {
	arch := gguf.String(f.Meta, "general.architecture")
	if arch != "qwen2" && arch != "qwen3" {
		return nil, fmt.Errorf("архитектура %q не поддерживается (нужна qwen2 или qwen3)", arch)
	}
	m := &Model{Pool: NewPool(threads)}
	m.Arch = arch
	m.Name = gguf.String(f.Meta, "general.name")

	geti := func(key string) (int, error) {
		v, ok := gguf.Int(f.Meta, arch+"."+key)
		if !ok {
			return 0, fmt.Errorf("в метаданных нет %s.%s", arch, key)
		}
		return v, nil
	}
	var err error
	if m.Layers, err = geti("block_count"); err != nil {
		return nil, err
	}
	if m.Dim, err = geti("embedding_length"); err != nil {
		return nil, err
	}
	if m.Heads, err = geti("attention.head_count"); err != nil {
		return nil, err
	}
	if m.HeadsKV, _ = geti("attention.head_count_kv"); m.HeadsKV == 0 {
		m.HeadsKV = m.Heads
	}
	m.FF, _ = geti("feed_forward_length")
	if m.CtxTrain, _ = geti("context_length"); m.CtxTrain == 0 {
		m.CtxTrain = 32768
	}
	m.Eps = 1e-6
	if v, ok := gguf.Float(f.Meta, arch+".attention.layer_norm_rms_epsilon"); ok {
		m.Eps = float32(v)
	}
	m.RopeBase = 1e6
	if v, ok := gguf.Float(f.Meta, arch+".rope.freq_base"); ok {
		m.RopeBase = v
	}
	if m.Heads <= 0 || m.HeadsKV <= 0 || m.Heads%m.HeadsKV != 0 {
		return nil, fmt.Errorf("некорректное число голов внимания: %d/%d", m.Heads, m.HeadsKV)
	}

	total := len(f.Tensors)
	done := 0
	mat := func(name string, required bool) (*Matrix, error) {
		t, ok := f.Tensors[name]
		if !ok {
			if required {
				return nil, fmt.Errorf("в модели нет тензора %s", name)
			}
			return nil, nil
		}
		mx, err := loadMatrix(t, expand)
		if err != nil {
			return nil, err
		}
		if mx.Cols > m.maxCols {
			m.maxCols = mx.Cols
		}
		done++
		if progress != nil {
			progress(done, total)
		}
		return mx, nil
	}
	vecOf := func(name string, required bool) ([]float32, error) {
		mx, err := mat(name, required)
		if err != nil || mx == nil {
			return nil, err
		}
		mx.expand()
		return mx.F32, nil
	}

	if m.Embd, err = mat("token_embd.weight", true); err != nil {
		return nil, err
	}
	if m.Embd.Cols != m.Dim {
		return nil, errors.New("размер эмбеддингов не совпадает с метаданными")
	}
	m.Vocab = m.Embd.Rows

	m.Blocks = make([]Layer, m.Layers)
	for i := range m.Blocks {
		l := &m.Blocks[i]
		p := fmt.Sprintf("blk.%d.", i)
		steps := []func() error{
			func() (e error) { l.AttnNorm, e = vecOf(p+"attn_norm.weight", true); return },
			func() (e error) { l.Q, e = mat(p+"attn_q.weight", true); return },
			func() (e error) { l.K, e = mat(p+"attn_k.weight", true); return },
			func() (e error) { l.V, e = mat(p+"attn_v.weight", true); return },
			func() (e error) { l.O, e = mat(p+"attn_output.weight", true); return },
			func() (e error) { l.QBias, e = vecOf(p+"attn_q.bias", false); return },
			func() (e error) { l.KBias, e = vecOf(p+"attn_k.bias", false); return },
			func() (e error) { l.VBias, e = vecOf(p+"attn_v.bias", false); return },
			func() (e error) { l.QNorm, e = vecOf(p+"attn_q_norm.weight", false); return },
			func() (e error) { l.KNorm, e = vecOf(p+"attn_k_norm.weight", false); return },
			func() (e error) { l.FFNNorm, e = vecOf(p+"ffn_norm.weight", true); return },
			func() (e error) { l.Gate, e = mat(p+"ffn_gate.weight", true); return },
			func() (e error) { l.Up, e = mat(p+"ffn_up.weight", true); return },
			func() (e error) { l.Down, e = mat(p+"ffn_down.weight", true); return },
		}
		for _, step := range steps {
			if err := step(); err != nil {
				return nil, err
			}
		}
		if i == 0 {
			m.HeadDim = l.Q.Rows / m.Heads
		}
		kv := m.HeadsKV * m.HeadDim
		switch {
		case l.Q.Rows != m.Heads*m.HeadDim, l.K.Rows != kv, l.V.Rows != kv,
			l.O.Cols != m.Heads*m.HeadDim, l.O.Rows != m.Dim, l.Down.Rows != m.Dim,
			l.Gate.Rows != l.Up.Rows, l.Down.Cols != l.Gate.Rows, m.HeadDim%2 != 0,
			l.QNorm != nil && len(l.QNorm) != m.HeadDim, l.KNorm != nil && len(l.KNorm) != m.HeadDim:
			return nil, fmt.Errorf("слой %d: неожиданные размеры тензоров", i)
		}
		m.FF = l.Gate.Rows
	}
	if m.OutNorm, err = vecOf("output_norm.weight", true); err != nil {
		return nil, err
	}
	// У части моделей выходная матрица совпадает с матрицей эмбеддингов и в
	// файле отдельно не хранится.
	if m.Out, err = mat("output.weight", false); err != nil {
		return nil, err
	}
	if m.Out == nil {
		m.Out = m.Embd
	}

	half := m.HeadDim / 2
	m.invFreq = make([]float64, half)
	for i := range m.invFreq {
		m.invFreq[i] = math.Pow(m.RopeBase, -float64(2*i)/float64(m.HeadDim))
	}
	return m, nil
}

// WeightBytes возвращает объём памяти, занятый весами.
func (m *Model) WeightBytes() int {
	n := m.Embd.Bytes()
	if m.Out != m.Embd {
		n += m.Out.Bytes()
	}
	for i := range m.Blocks {
		l := &m.Blocks[i]
		for _, mx := range []*Matrix{l.Q, l.K, l.V, l.O, l.Gate, l.Up, l.Down} {
			n += mx.Bytes()
		}
	}
	return n
}

// WeightFormat возвращает формат основных матриц (например "Q8_0").
func (m *Model) WeightFormat() string { return m.Blocks[0].Gate.Format.String() }
