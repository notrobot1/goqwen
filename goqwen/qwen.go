// Package goqwen запускает языковые модели Qwen (Qwen2.5, Qwen2.5-Coder,
// Qwen3) локально, на чистом Go: без cgo, без llama.cpp и без каких-либо
// сторонних зависимостей — только стандартная библиотека.
//
// Модели читаются из официальных файлов GGUF. Вся арифметика вынесена в
// пакет goqwen/vec, который сам выбирает реализацию под процессор
// (AVX-512, AVX2+FMA, NEON или переносимый Go).
//
// Минимальный пример:
//
//	path, _ := goqwen.Models.Download(ctx, goqwen.Qwen25_05B, goqwen.Q8_0, nil)
//	model, _ := goqwen.Load(path, nil)
//	chat := model.NewChat(nil)
//	resp, _ := chat.Send(ctx, "Привет! Кто ты?")
//	fmt.Println(resp.Content)
package goqwen

import (
	"context"
	"errors"
	"fmt"
	"runtime"

	"goqwen/internal/gguf"
	"goqwen/internal/nn"
	"goqwen/internal/tokenizer"
	"goqwen/vec"
)

// Options — параметры загрузки модели. Нулевое значение (или nil) годится.
type Options struct {
	// Threads — число потоков вычислений. 0 — по числу ядер.
	Threads int
	// ContextSize — размер окна контекста в токенах. 0 — 4096 (но не больше,
	// чем поддерживает модель). Память под контекст выделяется по мере роста
	// диалога, а не сразу.
	ContextSize int
	// ExpandWeights распаковывает квантованные веса в float32 при загрузке:
	// генерация заметно быстрее для форматов без прямого SIMD-ядра (все,
	// кроме Q8_0 на x86), но памяти нужно в 4–8 раз больше размера файла.
	ExpandWeights bool
	// Progress вызывается при загрузке весов (может быть nil).
	Progress func(done, total int)
}

// Info — сведения о загруженной модели.
type Info struct {
	Name         string // имя из метаданных файла
	Architecture string // "qwen2" (Qwen2/2.5) или "qwen3"
	Layers       int
	Dim          int
	Heads        int
	Vocab        int
	ContextSize  int    // действующее окно контекста
	TrainContext int    // максимальный контекст модели
	WeightFormat string // формат основных матриц: "Q8_0", "Q4_K", "F32", ...
	WeightBytes  int64  // память под веса
	Threads      int
	SIMD         string // набор инструкций пакета vec: "avx512", "avx2", "neon", "generic"
}

// Model — загруженная модель. Потокобезопасна: из одной модели можно
// создавать сколько угодно независимых диалогов (Chat).
type Model struct {
	net  *nn.Model
	tok  *tokenizer.Tokenizer
	ctx  int
	stop map[int]bool // токены конца ответа

	thinkOpen, thinkClose, toolOpen, toolClose int // -1, если таких токенов нет
}

// Load загружает модель из файла GGUF. Для модели, разбитой на части
// (…-00001-of-0000N.gguf), достаточно указать любую часть — остальные
// должны лежать рядом.
func Load(path string, opts *Options) (*Model, error) {
	var o Options
	if opts != nil {
		o = *opts
	}
	if o.Threads <= 0 {
		o.Threads = runtime.NumCPU()
	}

	f, err := gguf.Open(path)
	if err != nil {
		return nil, fmt.Errorf("goqwen: %w", err)
	}
	defer f.Close()

	tokens, _ := f.Meta["tokenizer.ggml.tokens"].([]string)
	types, _ := f.Meta["tokenizer.ggml.token_type"].([]int32)
	merges, _ := f.Meta["tokenizer.ggml.merges"].([]string)
	if len(tokens) == 0 || len(merges) == 0 {
		return nil, errors.New("goqwen: в файле нет словаря токенизатора (нужен BPE-словарь tokenizer.ggml.*)")
	}

	net, err := nn.Load(f, o.Threads, o.ExpandWeights, o.Progress)
	if err != nil {
		return nil, fmt.Errorf("goqwen: %w", err)
	}
	if net.Vocab != len(tokens) {
		return nil, errors.New("goqwen: размер словаря не совпадает с размером матрицы эмбеддингов")
	}

	m := &Model{net: net, tok: tokenizer.New(tokens, types, merges), stop: map[int]bool{}}
	m.ctx = o.ContextSize
	if m.ctx <= 0 {
		m.ctx = 4096
	}
	if m.ctx > net.CtxTrain {
		m.ctx = net.CtxTrain
	}

	for _, t := range []string{"<|im_end|>", "<|endoftext|>"} {
		if id, ok := m.tok.ID(t); ok {
			m.stop[id] = true
		}
	}
	if id, ok := gguf.Int(f.Meta, "tokenizer.ggml.eos_token_id"); ok {
		m.stop[id] = true
	}
	id := func(t string) int {
		if v, ok := m.tok.ID(t); ok {
			return v
		}
		return -1
	}
	m.thinkOpen, m.thinkClose = id("<think>"), id("</think>")
	m.toolOpen, m.toolClose = id("<tool_call>"), id("</tool_call>")
	return m, nil
}

// LoadFromCatalog скачивает модель из встроенного каталога (если её ещё нет
// на диске) и загружает её. q == "" — рекомендуемое квантование.
func LoadFromCatalog(ctx context.Context, id string, q Quant, opts *Options, dl *DownloadOptions) (*Model, error) {
	path, err := Models.Download(ctx, id, q, dl)
	if err != nil {
		return nil, err
	}
	return Load(path, opts)
}

// Info возвращает сведения о модели.
func (m *Model) Info() Info {
	n := m.net
	return Info{
		Name: n.Name, Architecture: n.Arch, Layers: n.Layers, Dim: n.Dim, Heads: n.Heads,
		Vocab: n.Vocab, ContextSize: m.ctx, TrainContext: n.CtxTrain,
		WeightFormat: n.WeightFormat(), WeightBytes: int64(n.WeightBytes()),
		Threads: n.Pool.Threads(), SIMD: vec.Level(),
	}
}

// Tokenize разбивает текст на токены. Служебные токены вида <|im_start|>
// распознаются.
func (m *Model) Tokenize(text string) []int { return m.tok.Encode(text, true) }

// Detokenize собирает текст из токенов.
func (m *Model) Detokenize(tokens []int) string { return m.tok.Decode(tokens) }

// Ask — самый короткий путь: один вопрос, один ответ, без истории.
func (m *Model) Ask(ctx context.Context, prompt string) (string, error) {
	resp, err := m.NewChat(nil).Send(ctx, prompt)
	if err != nil {
		return "", err
	}
	return resp.Content, nil
}

// Complete продолжает произвольный текст без шаблона диалога («сырое»
// завершение). onText (может быть nil) получает текст по мере генерации.
func (m *Model) Complete(ctx context.Context, prompt string, s *Sampling, onText func(string)) (string, error) {
	tokens := m.tok.Encode(prompt, true)
	if len(tokens) == 0 {
		return "", errors.New("goqwen: пустой промпт")
	}
	if len(tokens) >= m.ctx {
		return "", ErrContextOverflow
	}
	g := &generator{model: m, state: m.net.NewState(m.ctx), sampler: newSampler(s)}
	out, err := g.run(ctx, tokens, onText)
	if err != nil {
		return "", err
	}
	return out.raw, nil
}
