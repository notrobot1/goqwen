package goqwen

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseOutput(t *testing.T) {
	raw := "<think>\nнадо вызвать погоду\n</think>\n\nСейчас узнаю.\n<tool_call>\n{\"name\": \"get_weather\", \"arguments\": {\"city\": \"Kyiv\"}}\n</tool_call>\n<tool_call>\n{\"name\": \"f\", \"arguments\": \"{\\\"x\\\": 1}\"}\n</tool_call>"
	content, reasoning, calls := parseOutput(raw)
	if content != "Сейчас узнаю." || reasoning != "надо вызвать погоду" {
		t.Fatalf("content=%q reasoning=%q", content, reasoning)
	}
	if len(calls) != 2 || calls[0].Name != "get_weather" || string(calls[1].Arguments) != `{"x": 1}` {
		t.Fatalf("calls=%+v", calls)
	}
	var a struct{ City string }
	if err := calls[0].Bind(&a); err != nil || a.City != "Kyiv" {
		t.Fatalf("Bind: %v %+v", err, a)
	}
	// Не-JSON внутри тега остаётся текстом; незакрытый think — всё рассуждения.
	if c, _, calls := parseOutput("x <tool_call>oops</tool_call>"); len(calls) != 0 || !strings.Contains(c, "oops") {
		t.Fatalf("битый вызов: %q %v", c, calls)
	}
	if c, r, _ := parseOutput("<think>думаю"); c != "" || r != "думаю" {
		t.Fatalf("незакрытый think: %q %q", c, r)
	}
}

func TestToolSchema(t *testing.T) {
	tool := Tool{Name: "f", Description: "a <b>", Parameters: ObjectSchema(map[string]map[string]any{"x": PropInteger("n")}, "x")}
	want := `{"type":"function","function":{"name":"f","description":"a <b>","parameters":{"properties":{"x":{"description":"n","type":"integer"}},"required":["x"],"type":"object"}}}`
	if got := tool.schema(); got != want {
		t.Fatalf("schema:\n got %s\nwant %s", got, want)
	}
}

func TestSampler(t *testing.T) {
	logits := func() []float32 { return []float32{0, 5, 1, 4, -2} }
	g := Greedy()
	if id := newSampler(&g).sample(logits(), nil); id != 1 {
		t.Fatalf("greedy = %d", id)
	}
	s := newSampler(&Sampling{Temperature: 1, TopK: 2, Seed: 7})
	seen := map[int]int{}
	for i := 0; i < 500; i++ {
		seen[s.sample(logits(), nil)]++
	}
	if len(seen) != 2 || seen[1] <= seen[3] || seen[3] == 0 {
		t.Fatalf("top-k=2: %v", seen)
	}
	// Top-P = 0.5: остаётся только самый вероятный токен.
	s = newSampler(&Sampling{Temperature: 1, TopP: 0.5, Seed: 7})
	for i := 0; i < 50; i++ {
		if id := s.sample(logits(), nil); id != 1 {
			t.Fatalf("top-p: %d", id)
		}
	}
}

func TestCatalog(t *testing.T) {
	if len(Models.Models) == 0 {
		t.Fatal("каталог пуст")
	}
	ids := map[string]bool{}
	for _, m := range Models.Models {
		if ids[m.ID] || len(m.Variants) == 0 || m.Repo == "" {
			t.Fatalf("плохая запись каталога: %+v", m)
		}
		ids[m.ID] = true
		if _, ok := m.Variant(m.DefaultQuant()); !ok {
			t.Fatalf("%s: нет варианта по умолчанию", m.ID)
		}
	}
	m, ok := Models.Find("QWEN2.5-7B-Instruct")
	if !ok || m.ID != Qwen25_7B {
		t.Fatal("Find")
	}
	if v, _ := m.Variant(Q4KM); len(v.Files) != 2 {
		t.Fatalf("ожидалась модель из двух частей: %v", v.Files)
	}
}

// Загрузка с докачкой: первый запрос обрывается на середине, второй
// продолжает с помощью заголовка Range.
func TestDownloadResume(t *testing.T) {
	data := []byte(strings.Repeat("0123456789", 5000))
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.URL.Path != "/Org/Repo-GGUF/resolve/main/m.gguf" {
			http.NotFound(w, r)
			return
		}
		var from int
		if h := r.Header.Get("Range"); h != "" {
			fmt.Sscanf(h, "bytes=%d-", &from)
			w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", from, len(data)-1, len(data)))
			w.WriteHeader(http.StatusPartialContent)
			w.Write(data[from:])
			return
		}
		w.Header().Set("Content-Length", fmt.Sprint(len(data)))
		w.Write(data[:len(data)/2]) // обрыв соединения на середине
	}))
	defer srv.Close()

	cat := Catalog{Models: []ModelInfo{{ID: "m", Repo: "Org/Repo-GGUF", Variants: []Variant{{Quant: Q8_0, Size: int64(len(data)), Files: []string{"m.gguf"}}}}}}
	opts := &DownloadOptions{Dir: t.TempDir(), Endpoint: srv.URL}
	if _, err := cat.Download(context.Background(), "m", "", opts); err == nil {
		t.Fatal("ожидалась ошибка обрыва")
	}
	path, err := cat.Download(context.Background(), "m", "", opts)
	if err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(path)
	if string(got) != string(data) || filepath.Base(filepath.Dir(path)) != "Repo-GGUF" {
		t.Fatalf("файл скачан неверно: %d байт, %s", len(got), path)
	}
	before := calls
	if _, err := cat.Download(context.Background(), "m", "", opts); err != nil || calls != before {
		t.Fatal("повторная загрузка не должна ходить в сеть")
	}
}

// Сквозной тест на настоящей модели. Запуск:
//
//	GOQWEN_TEST_MODEL=/path/to/qwen.gguf go test -run Model -v
func TestModel(t *testing.T) {
	path := os.Getenv("GOQWEN_TEST_MODEL")
	if path == "" {
		t.Skip("GOQWEN_TEST_MODEL не задан")
	}
	m, err := Load(path, &Options{ContextSize: 512})
	if err != nil {
		t.Fatal(err)
	}
	g := Greedy()
	g.MaxTokens = 32
	ctx := context.Background()

	chat := m.NewChat(&ChatOptions{Sampling: &g})
	var streamed strings.Builder
	r, err := chat.SendStream(ctx, "What is the capital of France? Answer with one word.", func(s string) { streamed.WriteString(s) })
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(r.Content, "Paris") || strings.TrimSpace(streamed.String()) != r.Content {
		t.Fatalf("content=%q streamed=%q", r.Content, streamed.String())
	}
	r2, err := chat.Send(ctx, "And of Italy?")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(r2.Content, "Rome") || r2.Usage.CachedTokens < r.Usage.PromptTokens {
		t.Fatalf("content=%q usage=%+v", r2.Content, r2.Usage)
	}

	called := ""
	tool := Tool{
		Name: "get_weather", Description: "Get the current weather in a given city",
		Parameters: ObjectSchema(map[string]map[string]any{"city": PropString("City name")}, "city"),
		Handler: ToolFunc(func(ctx context.Context, a struct {
			City string `json:"city"`
		}) (string, error) {
			called = a.City
			return `{"temperature_c": 7, "condition": "rain"}`, nil
		}),
	}
	g.MaxTokens = 64
	chat = m.NewChat(&ChatOptions{Tools: []Tool{tool}, Sampling: &g})
	r, err = chat.Send(ctx, "What's the weather in Kyiv right now?")
	if err != nil {
		t.Fatal(err)
	}
	if called != "Kyiv" || len(r.Executed) != 1 || !strings.Contains(r.Content, "7") {
		t.Fatalf("called=%q executed=%d content=%q", called, len(r.Executed), r.Content)
	}

	// Переполнение контекста: старые реплики выбрасываются, диалог живёт.
	chat = m.NewChat(&ChatOptions{Sampling: &g})
	for i := 0; i < 8; i++ {
		if _, err := chat.Send(ctx, strings.Repeat("Tell me about number ", 8)+fmt.Sprint(i)); err != nil {
			t.Fatalf("реплика %d: %v", i, err)
		}
	}
	if n := len(chat.Messages()); n >= 16 {
		t.Fatalf("история не обрезана: %d сообщений", n)
	}
	// Отмена контекста прерывает генерацию и не портит историю.
	cctx, cancel := context.WithCancel(ctx)
	cancel()
	before := len(chat.Messages())
	if _, err := chat.Send(cctx, "hi"); err == nil || len(chat.Messages()) != before {
		t.Fatal("отмена не сработала")
	}
}
