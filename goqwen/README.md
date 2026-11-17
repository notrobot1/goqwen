# goqwen

Локальный запуск моделей **Qwen** (Qwen2.5, Qwen2.5-Coder, Qwen3) на чистом Go.

- Только стандартная библиотека: без cgo, без llama.cpp, без сторонних модулей.
- Читает официальные файлы GGUF (в том числе разбитые на части).
- Векторные вычисления вынесены в пакет `vec/` и выбирают реализацию под процессор.
- Диалог с историей, потоковый вывод, инструменты (tools), режим рассуждений Qwen3.
- Встроенный каталог моделей и загрузчик с Hugging Face с докачкой.

Требуется Go 1.21+. Имя модуля — `goqwen`; при публикации замените его в `go.mod` и импортах на свой путь.

## Быстрый старт

```go
ctx := context.Background()

// Скачать модель из каталога (если её ещё нет) и загрузить.
path, err := goqwen.Models.Download(ctx, goqwen.Qwen25_05B, goqwen.Q8_0, nil)
model, err := goqwen.Load(path, nil) // или goqwen.Load("/path/model.gguf", nil)

chat := model.NewChat(&goqwen.ChatOptions{System: "Отвечай кратко."})

resp, err := chat.Send(ctx, "Привет! Кто ты?")
fmt.Println(resp.Content)

// Потоковый вывод; диалог помнит предыдущие реплики.
resp, err = chat.SendStream(ctx, "Расскажи подробнее", func(text string) { fmt.Print(text) })
```

Один вопрос без истории: `answer, err := model.Ask(ctx, "…")`.
Продолжение произвольного текста без шаблона диалога: `model.Complete(ctx, prompt, nil, nil)`.

## Каталог моделей

`goqwen.Models` — структура `Catalog` со списком моделей, которые можно загрузить:

```go
for _, m := range goqwen.Models.Models {
    fmt.Println(m.ID, m.Family, m.Parameters, m.Quants())
}
m, ok := goqwen.Models.Find("qwen2.5-7b-instruct")
small := goqwen.Models.FitsIn(4 << 30)        // модели, у которых есть вариант до 4 ГБ
path, err := goqwen.Models.Download(ctx, goqwen.Qwen3_4B, goqwen.Q4KM, &goqwen.DownloadOptions{
    Dir:      "./models",
    Progress: func(p goqwen.DownloadProgress) { /* p.Done / p.Total */ },
})
```

| Семейство | Модели | Константы |
|---|---|---|
| Qwen2.5 | 0.5B, 1.5B, 3B, 7B, 14B, 32B, 72B | `Qwen25_05B` … `Qwen25_72B` |
| Qwen2.5-Coder | 0.5B, 1.5B, 3B, 7B, 14B, 32B | `Qwen25Coder_05B` … `Qwen25Coder_32B` |
| Qwen3 | 0.6B, 1.7B, 4B, 8B, 14B, 32B | `Qwen3_06B` … `Qwen3_32B` |

Квантования: `Q2K`, `Q3KM`, `Q4_0`, `Q4KM`, `Q5_0`, `Q5KM`, `Q6K`, `Q8_0`, `FP16` (набор зависит от модели).
Каталог можно дополнять своими записями — это обычный срез. Любой другой GGUF архитектуры `qwen2`/`qwen3` загружается через `Load` по пути.

Модели скачиваются в `<кэш пользователя>/goqwen` (переопределяется `DownloadOptions.Dir` или переменной `GOQWEN_MODELS`). Зеркало Hugging Face задаётся `HF_ENDPOINT`, токен — `HF_TOKEN`.

## Инструменты (tools)

Инструмент с обработчиком библиотека вызывает сама и возвращает результат модели:

```go
weather := goqwen.Tool{
    Name:        "get_weather",
    Description: "Get the current weather in a given city",
    Parameters: goqwen.ObjectSchema(map[string]map[string]any{
        "city": goqwen.PropString("City name"),
    }, "city"),
    Handler: goqwen.ToolFunc(func(ctx context.Context, a struct {
        City string `json:"city"`
    }) (string, error) {
        return `{"temperature_c": 7, "condition": "rain"}`, nil
    }),
}
chat := model.NewChat(&goqwen.ChatOptions{Tools: []goqwen.Tool{weather}})
resp, _ := chat.Send(ctx, "What's the weather in Kyiv?")
fmt.Println(resp.Content, len(resp.Executed))
```

Инструмент без `Handler` возвращается вызывающей стороне:

```go
resp, _ := chat.Send(ctx, "…")
for _, call := range resp.ToolCalls {        // resp.FinishReason == "tool_calls"
    var args struct{ A, B float64 }
    call.Bind(&args)
    chat.AddToolResult(call.Name, fmt.Sprint(args.A*args.B))
}
resp, _ = chat.Continue(ctx, nil)
```

Список инструментов можно менять на ходу: `chat.SetTools(...)`, `chat.AddTool(...)`.
Маленькие модели (0.5B) вызывают инструменты неуверенно — для этого лучше брать 1.5B и больше.

## Настройки

```go
model, _ := goqwen.Load(path, &goqwen.Options{
    Threads:       8,     // 0 — по числу ядер
    ContextSize:   8192,  // 0 — 4096
    ExpandWeights: false, // true — распаковать веса в float32 (быстрее, больше памяти)
})

s := goqwen.DefaultSampling() // temperature 0.7, top-k 20, top-p 0.8, штраф за повторы 1.05
s.MaxTokens = 512
chat := model.NewChat(&goqwen.ChatOptions{
    Sampling: &s,          // goqwen.Greedy() — детерминированный ответ
    Thinking: true,        // режим рассуждений Qwen3 → resp.Reasoning
})
```

`Response` содержит `Content`, `Reasoning`, `ToolCalls`, `Executed`, `FinishReason` (`stop` / `length` / `tool_calls`), `Usage` и `Elapsed`.
Когда история перестаёт помещаться в окно контекста, самые старые реплики отбрасываются.
`Model` потокобезопасна: из одной модели можно создавать много независимых `Chat`.

## Пакет `vec`: оптимизация под железо

Вся арифметика (скалярное произведение, axpy, умножение квантованной строки на вектор, RMSNorm, softmax, RoPE, распаковка квантованных форматов) лежит в `vec/`. Реализация выбирается при старте:

| Платформа | Реализация | Как выбирается |
|---|---|---|
| amd64 | AVX-512F | CPUID + XGETBV во время выполнения |
| amd64 | AVX2 + FMA + F16C | CPUID + XGETBV во время выполнения |
| arm64 | NEON | всегда (обязательная часть ARMv8) |
| остальные | чистый Go | — |

Ассемблер — штатный ассемблер Go (файлы `kernels_*.s`), определение возможностей процессора написано в самом пакете.
`vec.Level()` показывает активный набор, `vec.SetLevel("generic")` переключает его принудительно (для отладки и замеров).

Замеры на одном ядре Xeon (скалярное произведение 4096 элементов): чистый Go — 1136 нс, AVX2 — 135 нс, AVX-512 — 89 нс.

## Скорость и память

- Память ≈ размер файла модели + KV-кэш (растёт вместе с диалогом).
- Самый быстрый формат на x86 — **Q8_0**: для него есть прямое SIMD-ядро. Qwen2.5-0.5B Q8_0 даёт около 14 токенов/с на одном ядре; вычисления распараллелены по потокам.
- Остальные квантованные форматы (Q4_K_M и т.д.) распаковываются построчно обычным Go-кодом и работают в несколько раз медленнее. `Options.ExpandWeights` распаковывает веса один раз при загрузке: скорость близка к Q8_0, но памяти нужно 4 байта на параметр.
- На arm64 прямого ядра для Q8_0 нет: строка распаковывается, затем перемножается через NEON.

## Что проверено

- Логиты совпадают с llama.cpp (Qwen2.5 и Qwen3; на модели fp16 расхождение ≤ 0.023 при диапазоне ~30), одинаково на AVX-512, AVX2 и чистом Go.
- Токенизатор совпал с эталонным токенизатором Hugging Face на 437 строках (кириллица, CJK, эмодзи, код, пробелы).
- Распаковка всех форматов квантования побитно совпала с эталонной реализацией из пакета `gguf`.
- `go test ./...`; сквозной тест на модели: `GOQWEN_TEST_MODEL=/path/model.gguf go test -run Model`.

## Ограничения

- NEON-ядра для arm64 собираются и проходят `go vet`, но на реальном ARM-процессоре не запускались.
- Нет NFC-нормализации Unicode в токенизаторе (в стандартной библиотеке её нет); на обычном тексте это не влияет.
- Не поддерживаются MoE-модели (Qwen3-30B-A3B и т.п.), мультимодальные (VL), Embedding и форматы квантования IQ*.
- Промпт вычисляется по одному токену, без пакетной обработки и без GPU.

## Консольная утилита

```
go run ./cmd/qwen -list
go run ./cmd/qwen -model qwen2.5-1.5b-instruct -quant q8_0
go run ./cmd/qwen -file ./model.gguf -p "Привет" -temp 0
```

## Структура

```
qwen.go, chat.go, tools.go, template.go, sampler.go   публичный API
catalog.go, catalog_data.go, download.go               каталог моделей и загрузка
vec/                                                   векторные вычисления + SIMD
internal/gguf, internal/tokenizer, internal/nn         формат файла, BPE, трансформер
cmd/qwen, examples/                                    утилита и примеры
```
