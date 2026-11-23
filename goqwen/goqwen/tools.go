package goqwen

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
)

// ToolHandler выполняет вызов инструмента. args — JSON-объект с аргументами,
// которые сформировала модель. Возвращаемая строка передаётся модели как
// результат вызова.
type ToolHandler func(ctx context.Context, args json.RawMessage) (string, error)

// Tool — инструмент (функция), который модель может вызвать.
type Tool struct {
	Name        string
	Description string
	// Parameters — JSON Schema аргументов (объект). nil — без аргументов.
	// Удобно собирать через ObjectSchema и Prop*.
	Parameters map[string]any
	// Handler — обработчик. Если задан, Chat.Send сам выполнит вызов и
	// вернёт модели результат. Если nil, вызов возвращается вызывающей
	// стороне в Response.ToolCalls.
	Handler ToolHandler
}

// ToolFunc оборачивает типизированную функцию в ToolHandler: аргументы
// модели разбираются из JSON в структуру T.
//
//	goqwen.ToolFunc(func(ctx context.Context, a struct{ City string `json:"city"` }) (string, error) { ... })
func ToolFunc[T any](fn func(ctx context.Context, args T) (string, error)) ToolHandler {
	return func(ctx context.Context, raw json.RawMessage) (string, error) {
		var a T
		if len(raw) > 0 {
			if err := json.Unmarshal(raw, &a); err != nil {
				return "", err
			}
		}
		return fn(ctx, a)
	}
}

// ObjectSchema собирает JSON Schema объекта аргументов.
func ObjectSchema(properties map[string]map[string]any, required ...string) map[string]any {
	s := map[string]any{"type": "object", "properties": properties}
	if len(required) > 0 {
		s["required"] = required
	}
	return s
}

// PropString описывает строковый аргумент; enum — допустимые значения.
func PropString(description string, enum ...string) map[string]any {
	p := map[string]any{"type": "string", "description": description}
	if len(enum) > 0 {
		p["enum"] = enum
	}
	return p
}

// PropNumber описывает числовой аргумент.
func PropNumber(description string) map[string]any {
	return map[string]any{"type": "number", "description": description}
}

// PropInteger описывает целочисленный аргумент.
func PropInteger(description string) map[string]any {
	return map[string]any{"type": "integer", "description": description}
}

// PropBool описывает логический аргумент.
func PropBool(description string) map[string]any {
	return map[string]any{"type": "boolean", "description": description}
}

// PropArray описывает аргумент-массив с элементами вида items.
func PropArray(description string, items map[string]any) map[string]any {
	return map[string]any{"type": "array", "description": description, "items": items}
}

// ToolCall — запрос модели на вызов инструмента.
type ToolCall struct {
	Name      string
	Arguments json.RawMessage // JSON-объект
}

// Bind разбирает аргументы вызова в v (указатель на структуру или map).
func (c ToolCall) Bind(v any) error { return json.Unmarshal(c.Arguments, v) }

// ToolExecution — вызов инструмента, который Chat выполнил сам.
type ToolExecution struct {
	Call   ToolCall
	Result string // то, что было передано модели
	Err    error  // ошибка обработчика (модель получила её текст)
}

func marshalJSON(v any) string {
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return "{}"
	}
	return strings.TrimRight(b.String(), "\n")
}

// schema возвращает описание инструмента в формате, которому обучены Qwen.
func (t Tool) schema() string {
	params := t.Parameters
	if params == nil {
		params = map[string]any{"type": "object", "properties": map[string]any{}}
	}
	type fn struct {
		Name        string         `json:"name"`
		Description string         `json:"description"`
		Parameters  map[string]any `json:"parameters"`
	}
	return marshalJSON(struct {
		Type     string `json:"type"`
		Function fn     `json:"function"`
	}{"function", fn{t.Name, t.Description, params}})
}

// parseOutput разбирает сырой текст модели на рассуждения (<think>),
// вызовы инструментов (<tool_call>) и обычный ответ.
func parseOutput(raw string) (content, reasoning string, calls []ToolCall) {
	const thinkOpen, thinkClose = "<think>", "</think>"
	if i := strings.Index(raw, thinkClose); i >= 0 {
		reasoning = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(raw[:i]), thinkOpen))
		raw = raw[i+len(thinkClose):]
	} else if t := strings.TrimSpace(raw); strings.HasPrefix(t, thinkOpen) {
		return "", strings.TrimSpace(t[len(thinkOpen):]), nil // ответ оборвался внутри рассуждений
	}

	const open, closeTag = "<tool_call>", "</tool_call>"
	var text strings.Builder
	for {
		i := strings.Index(raw, open)
		if i < 0 {
			text.WriteString(raw)
			break
		}
		text.WriteString(raw[:i])
		body := raw[i+len(open):]
		rest := ""
		if j := strings.Index(body, closeTag); j >= 0 {
			body, rest = body[:j], body[j+len(closeTag):]
		}
		if call, ok := parseToolCall(body); ok {
			calls = append(calls, call)
		} else {
			text.WriteString(open + body) // не JSON — оставляем как текст
		}
		raw = rest
	}
	return strings.TrimSpace(text.String()), reasoning, calls
}

func parseToolCall(body string) (ToolCall, bool) {
	var v struct {
		Name       string          `json:"name"`
		Arguments  json.RawMessage `json:"arguments"`
		Parameters json.RawMessage `json:"parameters"`
	}
	if err := json.Unmarshal([]byte(strings.TrimSpace(body)), &v); err != nil || v.Name == "" {
		return ToolCall{}, false
	}
	args := v.Arguments
	if len(args) == 0 {
		args = v.Parameters
	}
	// Иногда модель кладёт аргументы строкой с JSON внутри.
	if len(args) > 0 && args[0] == '"' {
		var s string
		if json.Unmarshal(args, &s) == nil && json.Valid([]byte(s)) {
			args = json.RawMessage(s)
		}
	}
	if len(args) == 0 || string(args) == "null" {
		args = json.RawMessage("{}")
	}
	return ToolCall{Name: v.Name, Arguments: args}, true
}
