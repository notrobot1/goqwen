package goqwen

import "strings"

// Роли сообщений.
const (
	RoleSystem    = "system"
	RoleUser      = "user"
	RoleAssistant = "assistant"
	RoleTool      = "tool"
)

// Message — одно сообщение диалога.
type Message struct {
	Role      string
	Content   string
	ToolCalls []ToolCall // для роли assistant: вызовы инструментов
	Name      string     // для роли tool: имя инструмента (справочно)

	// raw — дословный текст, сгенерированный моделью. Если задан, в промпт
	// подставляется он: тогда токены совпадают с уже вычисленными и KV-кэш
	// переиспользуется полностью.
	raw string
}

const defaultSystemQwen2 = "You are Qwen, created by Alibaba Cloud. You are a helpful assistant."

// segment — кусок промпта. В special-кусках служебные токены распознаются,
// остальные (текст пользователя, результаты инструментов) кодируются как
// обычный текст — подделать разметку диалога ими нельзя.
type segment struct {
	text    string
	special bool
}

type promptBuilder struct{ segs []segment }

func (b *promptBuilder) add(text string, special bool) {
	if text == "" {
		return
	}
	if n := len(b.segs); n > 0 && b.segs[n-1].special == special {
		b.segs[n-1].text += text
		return
	}
	b.segs = append(b.segs, segment{text, special})
}
func (b *promptBuilder) ctl(text string)  { b.add(text, true) }
func (b *promptBuilder) text(text string) { b.add(text, false) }

// render собирает промпт в формате ChatML так же, как официальный шаблон
// Qwen2.5/Qwen3 (включая описание инструментов и их вызовы/ответы).
func (m *Model) render(system string, tools []Tool, msgs []Message, thinking bool) []segment {
	var b promptBuilder
	if system == "" && m.net.Arch == "qwen2" {
		system = defaultSystemQwen2
	}
	switch {
	case len(tools) > 0:
		b.ctl("<|im_start|>system\n")
		if system != "" {
			b.text(system)
			b.ctl("\n\n")
		}
		b.ctl("# Tools\n\nYou may call one or more functions to assist with the user query.\n\n" +
			"You are provided with function signatures within <tools></tools> XML tags:\n<tools>")
		for _, t := range tools {
			b.ctl("\n")
			b.text(t.schema())
		}
		b.ctl("\n</tools>\n\nFor each function call, return a json object with function name and arguments " +
			"within <tool_call></tool_call> XML tags:\n<tool_call>\n" +
			"{\"name\": <function-name>, \"arguments\": <args-json-object>}\n</tool_call><|im_end|>\n")
	case system != "":
		b.ctl("<|im_start|>system\n")
		b.text(system)
		b.ctl("<|im_end|>\n")
	}

	for i, msg := range msgs {
		switch msg.Role {
		case RoleAssistant:
			b.ctl("<|im_start|>assistant\n")
			switch {
			case msg.raw != "":
				b.ctl(msg.raw)
			default:
				b.text(msg.Content)
				for j, c := range msg.ToolCalls {
					if j > 0 || msg.Content != "" {
						b.ctl("\n")
					}
					b.ctl("<tool_call>\n")
					b.text(`{"name": ` + marshalJSON(c.Name) + `, "arguments": ` + compactArgs(c) + `}`)
					b.ctl("\n</tool_call>")
				}
			}
			b.ctl("<|im_end|>\n")
		case RoleTool:
			if i == 0 || msgs[i-1].Role != RoleTool {
				b.ctl("<|im_start|>user")
			}
			b.ctl("\n<tool_response>\n")
			b.text(msg.Content)
			b.ctl("\n</tool_response>")
			if i == len(msgs)-1 || msgs[i+1].Role != RoleTool {
				b.ctl("<|im_end|>\n")
			}
		case RoleSystem:
			b.ctl("<|im_start|>system\n")
			b.text(msg.Content)
			b.ctl("<|im_end|>\n")
		default:
			b.ctl("<|im_start|>user\n")
			b.text(msg.Content)
			b.ctl("<|im_end|>\n")
		}
	}

	b.ctl("<|im_start|>assistant\n")
	if m.net.Arch == "qwen3" && !thinking && m.thinkOpen >= 0 {
		b.ctl("<think>\n\n</think>\n\n") // пустой блок рассуждений = сразу ответ
	}
	return b.segs
}

func compactArgs(c ToolCall) string {
	s := strings.TrimSpace(string(c.Arguments))
	if s == "" {
		return "{}"
	}
	return s
}

// encode превращает сегменты промпта в токены.
func (m *Model) encode(segs []segment) []int {
	var out []int
	for _, s := range segs {
		out = append(out, m.tok.Encode(s.text, s.special)...)
	}
	return out
}
