package nn

import (
	"errors"
	"math"

	"github.com/notrobot1/goqwen/goqwen/vec"
)

// ErrContextFull возвращается, когда в окне контекста не осталось места.
var ErrContextFull = errors.New("окно контекста заполнено")

// State — состояние одного диалога: KV-кэш внимания и рабочие буферы.
// Не потокобезопасен; у каждого диалога свой State.
type State struct {
	m      *Model
	maxCtx int
	n      int // сколько позиций вычислено и лежит в кэше
	cap    int // на сколько позиций выделена память
	kc, vc [][]float32

	x, h, q, att, o, g, u []float32
	cos, sin              []float32
	scores                [][]float32
	scratch               [][]float32
	logits                []float32
}

// NewState создаёт состояние с окном контекста maxCtx токенов.
func (m *Model) NewState(maxCtx int) *State {
	t := m.Pool.Threads()
	s := &State{
		m:       m,
		maxCtx:  maxCtx,
		kc:      make([][]float32, m.Layers),
		vc:      make([][]float32, m.Layers),
		x:       make([]float32, m.Dim),
		h:       make([]float32, m.Dim),
		o:       make([]float32, m.Dim),
		q:       make([]float32, m.Heads*m.HeadDim),
		att:     make([]float32, m.Heads*m.HeadDim),
		g:       make([]float32, m.FF),
		u:       make([]float32, m.FF),
		cos:     make([]float32, m.HeadDim/2),
		sin:     make([]float32, m.HeadDim/2),
		scores:  make([][]float32, t),
		scratch: make([][]float32, t),
		logits:  make([]float32, m.Vocab),
	}
	for i := range s.scratch {
		s.scratch[i] = make([]float32, m.maxCols)
	}
	return s
}

// Len возвращает число токенов в кэше.
func (s *State) Len() int { return s.n }

// MaxContext возвращает размер окна контекста.
func (s *State) MaxContext() int { return s.maxCtx }

// Truncate отбрасывает всё после первых n токенов (кэш префикса сохраняется).
func (s *State) Truncate(n int) {
	if n < s.n {
		s.n = max(n, 0)
	}
}

// grow увеличивает KV-кэш (память выделяется по мере роста диалога).
func (s *State) grow() {
	newCap := min(max(256, s.cap*2), s.maxCtx)
	kv := s.m.HeadsKV * s.m.HeadDim
	for l := range s.kc {
		k := make([]float32, newCap*kv)
		v := make([]float32, newCap*kv)
		copy(k, s.kc[l][:s.n*kv])
		copy(v, s.vc[l][:s.n*kv])
		s.kc[l], s.vc[l] = k, v
	}
	for i := range s.scores {
		s.scores[i] = make([]float32, newCap)
	}
	s.cap = newCap
}

// Eval прогоняет один токен через модель на следующей позиции. Если
// wantLogits=true, возвращает логиты следующего токена (срез принадлежит
// State и действителен до следующего вызова); иначе выходной слой
// не вычисляется — так обрабатывается промпт.
func (s *State) Eval(token int, wantLogits bool) ([]float32, error) {
	m := s.m
	if token < 0 || token >= m.Vocab {
		return nil, errors.New("токен вне словаря")
	}
	if s.n >= s.maxCtx {
		return nil, ErrContextFull
	}
	if s.n >= s.cap {
		s.grow()
	}
	pos := s.n
	hd, half := m.HeadDim, m.HeadDim/2
	kv := m.HeadsKV * hd
	rep := m.Heads / m.HeadsKV
	scale := float32(1 / math.Sqrt(float64(hd)))
	pool := m.Pool

	for i, f := range m.invFreq {
		sn, cs := math.Sincos(float64(pos) * f)
		s.sin[i], s.cos[i] = float32(sn), float32(cs)
	}

	m.Embd.Row(token, s.x)

	for l := range m.Blocks {
		b := &m.Blocks[l]
		kcur := s.kc[l][pos*kv : (pos+1)*kv]
		vcur := s.vc[l][pos*kv : (pos+1)*kv]

		// --- внимание ---
		vec.RMSNorm(s.h, s.x, b.AttnNorm, m.Eps)
		b.Q.MulVec(pool, s.q, s.h, s.scratch)
		b.K.MulVec(pool, kcur, s.h, s.scratch)
		b.V.MulVec(pool, vcur, s.h, s.scratch)
		if b.QBias != nil {
			vec.Add(s.q, b.QBias)
		}
		if b.KBias != nil {
			vec.Add(kcur, b.KBias)
		}
		if b.VBias != nil {
			vec.Add(vcur, b.VBias)
		}
		for h := 0; h < m.Heads; h++ {
			qh := s.q[h*hd : (h+1)*hd]
			if b.QNorm != nil {
				vec.RMSNorm(qh, qh, b.QNorm, m.Eps)
			}
			vec.RotateHalf(qh, s.cos[:half], s.sin)
		}
		for h := 0; h < m.HeadsKV; h++ {
			kh := kcur[h*hd : (h+1)*hd]
			if b.KNorm != nil {
				vec.RMSNorm(kh, kh, b.KNorm, m.Eps)
			}
			vec.RotateHalf(kh, s.cos[:half], s.sin)
		}

		kc, vc := s.kc[l], s.vc[l]
		pool.For(m.Heads, m.Heads*(pos+1)*hd*2, func(lo, hi, w int) {
			sc := s.scores[w][:pos+1]
			for h := lo; h < hi; h++ {
				qh := s.q[h*hd : (h+1)*hd]
				off := (h / rep) * hd // несколько голов запросов делят одну голову K/V
				for t := range sc {
					sc[t] = vec.Dot(qh, kc[t*kv+off:t*kv+off+hd]) * scale
				}
				vec.Softmax(sc)
				out := s.att[h*hd : (h+1)*hd]
				clear(out)
				for t, p := range sc {
					vec.Axpy(p, vc[t*kv+off:t*kv+off+hd], out)
				}
			}
		})
		b.O.MulVec(pool, s.o, s.att, s.scratch)
		vec.Add(s.x, s.o)

		// --- полносвязный блок SwiGLU ---
		vec.RMSNorm(s.h, s.x, b.FFNNorm, m.Eps)
		b.Gate.MulVec(pool, s.g, s.h, s.scratch)
		b.Up.MulVec(pool, s.u, s.h, s.scratch)
		vec.SiLUMul(s.g, s.u)
		b.Down.MulVec(pool, s.o, s.g, s.scratch)
		vec.Add(s.x, s.o)
	}
	s.n++

	if !wantLogits {
		return nil, nil
	}
	vec.RMSNorm(s.h, s.x, m.OutNorm, m.Eps)
	m.Out.MulVec(pool, s.logits, s.h, s.scratch)
	return s.logits, nil
}
