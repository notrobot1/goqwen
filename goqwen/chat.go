package goqwen

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"goqwen/internal/nn"
)

// ErrContextOverflow возвращается, когда промпт не помещается в окно
// контекста (увеличьте Options.ContextSize или сократите запрос).
var ErrContextOverflow = errors.New("goqwen: промпт не помещается в окно контекста")

// ChatOptions — параметры диалога. Нулевое значение (или nil) годится.
type ChatOptions struct {
	// System — системный промпт. Пустая строка — стандартный для модели.
	System string
	// Tools — инструменты, доступные модели.
	Tools []Tool
	// Sampling — параметры генерации. nil — DefaultSampling().
	Sampling *Sampling
	// Thinking включает режим рассуждений у Qwen3 (текст рассуждений
	// попадает в Response.Reasoning). По умолчанию выключен — модель
	// отвечает сразу. Для Qwen2.5 не действует.
	Thinking bool
	// MaxToolRounds — сколько раз подряд модель может вызывать инструменты в
	// рамках одного Send (защита от зацикливания). 0 — 8.
	MaxToolRounds int
}

// Usage — расход токенов на один ответ.
type Usage struct {
	PromptTokens     int // длина промпта
	CachedTokens     int // из них взято из KV-кэша (не пересчитывалось)
	CompletionTokens int // сгенерировано
}

// Response — ответ модели.
type Response struct {
	// Content — текст ответа (без рассуждений и вызовов инструментов).
	Content string
	// Reasoning — рассуждения модели (Qwen3 с ChatOptions.Thinking).
	Reasoning string
	// ToolCalls — вызовы инструментов, которые должна выполнить вызывающая
	// сторона (инструменты без Handler). Результаты передаются через
	// Chat.AddToolResult, затем Chat.Continue.
	ToolCalls []ToolCall
	// Executed — вызовы, которые Chat выполнил сам (инструменты с Handler).
	Executed []ToolExecution
	// FinishReason: "stop" — модель закончила ответ; "length" — достигнут
	// MaxTokens или конец контекста; "tool_calls" — модель ждёт результатов
	// инструментов.
	FinishReason string
	Usage        Usage
	Elapsed      time.Duration
}

// TokensPerSecond возвращает среднюю скорость (промпт + ответ).
func (r *Response) TokensPerSecond() float64 {
	if r.Elapsed <= 0 {
		return 0
	}
	n := r.Usage.PromptTokens - r.Usage.CachedTokens + r.Usage.CompletionTokens
	return float64(n) / r.Elapsed.Seconds()
}

// Chat — диалог с моделью: хранит историю сообщений и KV-кэш, поэтому
// каждая следующая реплика пересчитывает только новые токены.
// Методы Chat можно вызывать из разных горутин, но выполняются они по очереди.
type Chat struct {
	mu       sync.Mutex
	model    *Model
	system   string
	tools    []Tool
	thinking bool
	rounds   int
	messages []Message
	gen      *generator
}

// NewChat создаёт новый диалог.
func (m *Model) NewChat(opts *ChatOptions) *Chat {
	var o ChatOptions
	if opts != nil {
		o = *opts
	}
	if o.MaxToolRounds <= 0 {
		o.MaxToolRounds = 8
	}
	return &Chat{
		model:    m,
		system:   o.System,
		tools:    append([]Tool(nil), o.Tools...),
		thinking: o.Thinking,
		rounds:   o.MaxToolRounds,
		gen:      &generator{model: m, state: m.net.NewState(m.ctx), sampler: newSampler(o.Sampling)},
	}
}

// SetSystem меняет системный промпт.
func (c *Chat) SetSystem(system string) {
	c.mu.Lock()
	c.system = system
	c.mu.Unlock()
}

// SetTools заменяет список инструментов.
func (c *Chat) SetTools(tools ...Tool) {
	c.mu.Lock()
	c.tools = append([]Tool(nil), tools...)
	c.mu.Unlock()
}

// AddTool добавляет инструмент (или заменяет одноимённый).
func (c *Chat) AddTool(t Tool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for i := range c.tools {
		if c.tools[i].Name == t.Name {
			c.tools[i] = t
			return
		}
	}
	c.tools = append(c.tools, t)
}

// SetSampling меняет параметры генерации.
func (c *Chat) SetSampling(s Sampling) {
	c.mu.Lock()
	c.gen.sampler = newSampler(&s)
	c.mu.Unlock()
}

// Messages возвращает копию истории диалога.
func (c *Chat) Messages() []Message {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]Message(nil), c.messages...)
}

// SetMessages заменяет историю диалога (например, восстановленную из БД).
func (c *Chat) SetMessages(msgs []Message) {
	c.mu.Lock()
	c.messages = append([]Message(nil), msgs...)
	c.mu.Unlock()
}

// Reset очищает историю. Системный промпт и инструменты сохраняются.
func (c *Chat) Reset() {
	c.mu.Lock()
	c.messages = nil
	c.mu.Unlock()
}

// Send отправляет сообщение пользователя и возвращает ответ модели.
//
// Если модель вызывает инструменты с заданным Handler, они выполняются
// автоматически, результаты возвращаются модели, и так до итогового ответа.
func (c *Chat) Send(ctx context.Context, prompt string) (*Response, error) {
	return c.SendStream(ctx, prompt, nil)
}

// SendStream — как Send, но onText получает текст ответа по мере генерации.
func (c *Chat) SendStream(ctx context.Context, prompt string, onText func(string)) (*Response, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.run(ctx, onText, Message{Role: RoleUser, Content: prompt})
}

// AddToolResult добавляет в историю результат вызова инструмента,
// выполненного вызывающей стороной. После добавления всех результатов
// вызовите Continue.
func (c *Chat) AddToolResult(name, result string) {
	c.mu.Lock()
	c.messages = append(c.messages, Message{Role: RoleTool, Name: name, Content: result})
	c.mu.Unlock()
}

// Continue просит модель продолжить диалог по текущей истории (обычно —
// после AddToolResult). onText может быть nil.
func (c *Chat) Continue(ctx context.Context, onText func(string)) (*Response, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.run(ctx, onText)
}

func (c *Chat) run(ctx context.Context, onText func(string), add ...Message) (resp *Response, err error) {
	start := time.Now()
	saved := len(c.messages)
	c.messages = append(c.messages, add...)
	defer func() {
		if err != nil && len(c.messages) > saved {
			c.messages = c.messages[:saved] // при ошибке история не меняется
		}
	}()

	total := &Response{}
	for round := 0; ; round++ {
		r, err := c.generate(ctx, onText)
		if err != nil {
			return nil, err
		}
		total.Content, total.Reasoning, total.FinishReason = r.Content, r.Reasoning, r.FinishReason
		total.Usage.PromptTokens += r.Usage.PromptTokens
		total.Usage.CachedTokens += r.Usage.CachedTokens
		total.Usage.CompletionTokens += r.Usage.CompletionTokens
		total.Elapsed = time.Since(start)
		if len(r.ToolCalls) == 0 {
			return total, nil
		}

		// Вызовы с обработчиками выполняем сами; если хотя бы один вызов
		// должен выполнить вызывающий код — отдаём ему все.
		auto := round < c.rounds
		for _, call := range r.ToolCalls {
			if t := c.tool(call.Name); t != nil && t.Handler == nil {
				auto = false
			}
		}
		if !auto {
			total.ToolCalls = r.ToolCalls
			total.FinishReason = "tool_calls"
			return total, nil
		}
		for _, call := range r.ToolCalls {
			ex := ToolExecution{Call: call}
			if t := c.tool(call.Name); t == nil {
				ex.Err = fmt.Errorf("unknown tool %q", call.Name)
			} else {
				ex.Result, ex.Err = t.Handler(ctx, call.Arguments)
			}
			if ex.Err != nil {
				if ctx.Err() != nil {
					return nil, ctx.Err()
				}
				ex.Result = "error: " + ex.Err.Error()
			}
			total.Executed = append(total.Executed, ex)
			c.messages = append(c.messages, Message{Role: RoleTool, Name: call.Name, Content: ex.Result})
		}
	}
}

func (c *Chat) tool(name string) *Tool {
	for i := range c.tools {
		if c.tools[i].Name == name {
			return &c.tools[i]
		}
	}
	return nil
}

// generate выполняет один ход модели и добавляет его в историю.
func (c *Chat) generate(ctx context.Context, onText func(string)) (*Response, error) {
	m := c.model
	reserve := min(256, m.ctx/4) // место под ответ
	var tokens []int
	for {
		tokens = m.encode(m.render(c.system, c.tools, c.messages, c.thinking))
		if len(tokens)+reserve <= m.ctx {
			break
		}
		if !c.dropOldest() {
			return nil, ErrContextOverflow
		}
	}

	out, err := c.gen.run(ctx, tokens, onText)
	if err != nil {
		return nil, err
	}
	content, reasoning, calls := parseOutput(out.raw)
	msg := Message{Role: RoleAssistant, Content: content, ToolCalls: calls}
	if reasoning == "" && !strings.Contains(out.raw, "<think>") {
		msg.raw = out.raw
	}
	c.messages = append(c.messages, msg)

	finish := out.finish
	if len(calls) > 0 {
		finish = "tool_calls"
	}
	return &Response{
		Content: content, Reasoning: reasoning, ToolCalls: calls, FinishReason: finish,
		Usage: Usage{PromptTokens: len(tokens), CachedTokens: out.cached, CompletionTokens: out.generated},
	}, nil
}

// dropOldest выбрасывает самую старую реплику (вместе с ответом на неё),
// чтобы освободить место в окне контекста. Последнее сообщение не трогает.
func (c *Chat) dropOldest() bool {
	if len(c.messages) <= 1 {
		return false
	}
	n := 1
	for n < len(c.messages)-1 && c.messages[n].Role != RoleUser {
		n++
	}
	c.messages = c.messages[n:]
	return true
}

// generator — цикл генерации поверх состояния модели с KV-кэшем.
type generator struct {
	model   *Model
	state   *nn.State
	sampler *sampler
	cached  []int // токены, уже вычисленные и лежащие в KV-кэше
}

type genOutput struct {
	raw       string // весь сгенерированный текст, включая <think> и <tool_call>
	content   string // только видимый текст (то, что уходило в onText)
	finish    string
	cached    int
	generated int
}

// run вычисляет промпт (переиспользуя общий с прошлым вызовом префикс) и
// генерирует ответ до токена конца, MaxTokens или конца контекста.
func (g *generator) run(ctx context.Context, prompt []int, onText func(string)) (genOutput, error) {
	m := g.model
	var out genOutput

	// Общий префикс с тем, что уже в кэше. Последний токен промпта всегда
	// пересчитывается — нужны его логиты.
	p := 0
	for p < len(g.cached) && p < len(prompt)-1 && g.cached[p] == prompt[p] {
		p++
	}
	g.state.Truncate(p)
	g.cached = g.cached[:p]
	out.cached = p

	var logits []float32
	var err error
	for i := p; i < len(prompt); i++ {
		if err = ctx.Err(); err != nil {
			return out, err
		}
		if logits, err = g.state.Eval(prompt[i], i == len(prompt)-1); err != nil {
			if errors.Is(err, nn.ErrContextFull) {
				err = ErrContextOverflow
			}
			return out, err
		}
		g.cached = append(g.cached, prompt[i])
	}

	var raw, pending []byte
	var visible strings.Builder
	inThink, inTool, started := false, false, false
	emit := func(b []byte) {
		if len(b) == 0 {
			return
		}
		s := string(b)
		if !started { // ведущие пробелы и переводы строк не показываем
			if s = strings.TrimLeft(s, " \t\r\n"); s == "" {
				return
			}
			started = true
		}
		visible.WriteString(s)
		if onText != nil {
			onText(s)
		}
	}

	limit := g.sampler.opt.MaxTokens
	for {
		if err = ctx.Err(); err != nil {
			return out, err
		}
		id := g.sampler.sample(logits, g.cached)
		if m.stop[id] {
			out.finish = "stop"
			break
		}
		out.generated++
		piece := m.tok.Bytes(id)
		raw = append(raw, piece...)

		switch {
		case id == m.thinkOpen:
			inThink = true
		case id == m.thinkClose:
			inThink, started = false, false
		case id == m.toolOpen:
			inTool = true
		case id == m.toolClose:
			inTool = false
		case !inThink && !inTool:
			// Токен может оканчиваться на середине UTF-8 символа — отдаём
			// наружу только целые символы.
			pending = append(pending, piece...)
			n := completeUTF8(pending)
			emit(pending[:n])
			pending = append(pending[:0], pending[n:]...)
		}

		if limit > 0 && out.generated >= limit || g.state.Len() >= g.state.MaxContext() {
			out.finish = "length"
			break
		}
		if logits, err = g.state.Eval(id, true); err != nil {
			return out, err
		}
		g.cached = append(g.cached, id)
	}
	emit(pending)
	out.raw = string(raw)
	out.content = strings.TrimRight(visible.String(), " \t\r\n")
	return out, nil
}

// completeUTF8 возвращает длину префикса b, не обрывающего UTF-8 символ.
func completeUTF8(b []byte) int {
	for i := len(b) - 1; i >= 0 && i >= len(b)-utf8.UTFMax; i-- {
		if utf8.RuneStart(b[i]) {
			if utf8.FullRune(b[i:]) {
				return len(b)
			}
			return i
		}
	}
	return len(b)
}
