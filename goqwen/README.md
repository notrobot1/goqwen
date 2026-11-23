# goqwen

Запуск языковых моделей [Qwen](https://qwenlm.github.io/) локально, на чистом Go — без cgo, без llama.cpp и без сторонних зависимостей. Только стандартная библиотека.

Поддерживаются Qwen2.5, Qwen2.5-Coder и Qwen3. Модели читаются из файлов GGUF. Векторная арифметика автоматически выбирает лучшее ядро под процессор: AVX-512, AVX2+FMA, NEON или переносимый Go.

Опционально: GPU-ускорение через **llama.cpp** (Metal на Apple Silicon, CUDA на NVIDIA, Vulkan на AMD/Intel) — подключается одним флагом при сборке, без изменений в коде.

---

## Установка

```bash
go get github.com/notrobot1/goqwen/goqwen
```

Требуется Go 1.21+.

---

## Сценарии использования

### 1. Минимальный пример

Скачать модель из каталога и задать один вопрос:

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/notrobot1/goqwen/goqwen"
)

func main() {
    ctx := context.Background()

    path, err := goqwen.Models.Download(ctx, goqwen.Qwen25_05B, goqwen.Q8_0, nil)
    if err != nil {
        log.Fatal(err)
    }

    model, err := goqwen.Load(path, nil)
    if err != nil {
        log.Fatal(err)
    }

    chat := model.NewChat(nil)
    resp, err := chat.Send(ctx, "Привет! Кто ты?")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(resp.Content)
}
```

---

### 2. Загрузка локального файла GGUF

Если модель уже скачана вручную, передайте путь напрямую:

```go
model, err := goqwen.Load("/path/to/model.gguf", nil)
```

Для модели, разбитой на части (`…-00001-of-0000N.gguf`), достаточно указать любую часть — остальные должны лежать рядом.

---

### 3. Выбор модели и квантования

Посмотреть доступные модели и их квантования:

```go
// Все модели каталога
for _, m := range goqwen.Models.Models {
    fmt.Println(m)
}

// Только семейство Qwen2.5
for _, m := range goqwen.Models.Family("Qwen2.5") {
    fmt.Println(m.ID, m.Quants())
}

// Модели, которые поместятся в N байт оперативной памяти
for _, m := range goqwen.Models.FitsIn(8 * 1024 * 1024 * 1024) { // 8 ГБ
    fmt.Println(m.ID)
}
```

Константы идентификаторов:

| Константа | Модель |
|---|---|
| `goqwen.Qwen25_05B` … `Qwen25_72B` | Qwen2.5 Instruct 0.5B–72B |
| `goqwen.Qwen25Coder_05B` … `Qwen25Coder_32B` | Qwen2.5-Coder 0.5B–32B |
| `goqwen.Qwen3_06B` … `Qwen3_32B` | Qwen3 0.6B–32B |

Доступные квантования: `Q4_0`, `Q4_1`, `Q5_0`, `Q5_1`, `Q8_0`, `Q4_K_M`, `Q5_K_M`, `F16`, `BF16`.

---

### 4. Параметры загрузки модели

```go
model, err := goqwen.Load(path, &goqwen.Options{
    Threads:       8,     // число потоков (0 — по числу ядер)
    ContextSize:   8192,  // окно контекста в токенах (0 — 4096)
    ExpandWeights: true,  // распаковать веса в float32: быстрее, но больше RAM
    Verbose:       false, // показывать внутренние логи llama.cpp (по умолчанию false)
    Progress: func(done, total int) {
        fmt.Printf("\rзагрузка: %d / %d МБ", done/1e6, total/1e6)
    },
})
```

Получить информацию о загруженной модели:

```go
info := model.Info()
fmt.Printf("%s: %d слоёв, веса %s (%.0f МБ), SIMD %s, потоков %d\n",
    info.Name, info.Layers, info.WeightFormat,
    float64(info.WeightBytes)/1e6, info.SIMD, info.Threads)
```

---

### 5. Диалог с историей

Chat сохраняет историю сообщений автоматически. При приближении к границе контекста старые сообщения вытесняются:

```go
chat := model.NewChat(&goqwen.ChatOptions{
    System: "Ты — полезный ассистент. Отвечай кратко.",
})
defer chat.Close() // освобождает KV-кэш (важно при GPU-бэкенде)

resp, _ := chat.Send(ctx, "Назови столицу Франции.")
fmt.Println(resp.Content) // Париж

resp, _ = chat.Send(ctx, "А Германии?") // модель помнит контекст
fmt.Println(resp.Content) // Берлин

// Сбросить историю, но сохранить системный промпт
chat.Reset()
```

---

### 6. Потоковая генерация

Получать токены по мере генерации:

```go
resp, err := chat.SendStream(ctx, "Объясни теорему Пифагора", func(token string) {
    fmt.Print(token)
})
fmt.Println()
fmt.Printf("[%d токенов, %.1f ток/с]\n",
    resp.Usage.CompletionTokens, resp.TokensPerSecond())
```

---

### 7. Параметры сэмплинга

```go
s := goqwen.DefaultSampling() // рекомендованные авторами Qwen параметры
s.Temperature = 0.9
s.TopP = 0.9
s.TopK = 50
s.MaxTokens = 512

chat := model.NewChat(&goqwen.ChatOptions{Sampling: &s})
```

Детерминированный режим (жадный декодинг):

```go
greedy := goqwen.Greedy()
chat := model.NewChat(&goqwen.ChatOptions{Sampling: &greedy})
```

---

### 8. Режим рассуждений (Qwen3)

Qwen3 поддерживает цепочку размышлений перед ответом (`<think>…</think>`):

```go
model, _ := goqwen.Load(path, nil) // модель Qwen3

chat := model.NewChat(&goqwen.ChatOptions{Thinking: true})
resp, _ := chat.Send(ctx, "Сколько простых чисел до 100?")

fmt.Println("Ответ:", resp.Content)
if resp.Reasoning != "" {
    fmt.Println("Рассуждение:", resp.Reasoning)
}
```

---

### 9. Вызов инструментов (function calling)

#### Автоматический вызов (с обработчиком)

Библиотека сама вызывает функцию и передаёт результат модели:

```go
type weatherArgs struct {
    City string `json:"city"`
}

weather := goqwen.Tool{
    Name:        "get_weather",
    Description: "Get the current weather in a given city",
    Parameters: goqwen.ObjectSchema(map[string]map[string]any{
        "city": goqwen.PropString("City name, e.g. Kyiv"),
    }, "city"),
    Handler: goqwen.ToolFunc(func(ctx context.Context, a weatherArgs) (string, error) {
        return fmt.Sprintf(`{"city":%q,"temperature_c":7,"condition":"rain"}`, a.City), nil
    }),
}

chat := model.NewChat(&goqwen.ChatOptions{Tools: []goqwen.Tool{weather}})
resp, _ := chat.Send(ctx, "What's the weather in Kyiv?")
fmt.Println(resp.Content)
fmt.Println("Выполнено вызовов:", len(resp.Executed))
```

#### Ручной вызов (без обработчика)

```go
calc := goqwen.Tool{
    Name:        "multiply",
    Description: "Multiply two numbers",
    Parameters: goqwen.ObjectSchema(map[string]map[string]any{
        "a": goqwen.PropNumber("first factor"),
        "b": goqwen.PropNumber("second factor"),
    }, "a", "b"),
}

chat := model.NewChat(&goqwen.ChatOptions{Tools: []goqwen.Tool{calc}})
resp, _ := chat.Send(ctx, "Вычисли 1234 * 5678.")

for _, call := range resp.ToolCalls {
    var args struct{ A, B float64 }
    call.Bind(&args)
    chat.AddToolResult(call.Name, fmt.Sprint(args.A*args.B))
}
if len(resp.ToolCalls) > 0 {
    resp, _ = chat.Continue(ctx, nil)
}
fmt.Println(resp.Content)
```

#### Вспомогательные функции для схем

```go
goqwen.PropString("описание", "опция1", "опция2") // строка, опционально с enum
goqwen.PropNumber("описание")                      // число
goqwen.PropInteger("описание")                     // целое
goqwen.PropBool("описание")                        // булево
goqwen.PropArray("описание", items)                // массив
```

---

### 10. Управление сообщениями вручную

Прочитать, заменить или восстановить историю диалога:

```go
// Прочитать историю
msgs := chat.Messages()

// Восстановить сохранённую историю
chat.SetMessages(msgs)

// Добавить системный промпт позже
chat.SetSystem("Ты эксперт по Go.")

// Добавить или заменить инструменты
chat.SetTools(tool1, tool2)
chat.AddTool(tool3)
```

---

### 11. Низкоуровневое завершение (completion)

Прямое завершение произвольного промпта без шаблона диалога:

```go
result, err := model.Complete(ctx, "Translate to English:\nПривет, мир!\n",
    &goqwen.Sampling{Temperature: 0, MaxTokens: 64},
    func(token string) { fmt.Print(token) },
)
```

Быстрый однострочный вариант (без потока):

```go
result, err := model.Ask(ctx, "What is 2+2?")
```

---

### 12. Токенизатор

```go
ids := model.Tokenize("Привет, мир!")
fmt.Println(ids) // []int{...}

text := model.Detokenize(ids)
fmt.Println(text) // Привет, мир!
```

---

## CLI — команда `qwen`

```bash
go install github.com/notrobot1/goqwen/goqwen/cmd/qwen@latest
```

| Флаг | По умолчанию | Описание |
|---|---|---|
| `-list` | — | Показать каталог моделей и выйти |
| `-model` | `qwen2.5-0.5b-instruct` | Идентификатор модели из каталога |
| `-quant` | авто | Квантование: `q4_k_m`, `q8_0`, … |
| `-file` | — | Путь к локальному файлу GGUF |
| `-dir` | — | Каталог для скачанных моделей |
| `-system` | — | Системный промпт |
| `-p` | — | Один запрос (без интерактивного режима) |
| `-threads` | по числу ядер | Число потоков вычислений |
| `-ctx` | 4096 | Размер окна контекста в токенах |
| `-temp` | 0.7 | Температура (0 — детерминированно) |
| `-max` | 2048 | Максимум токенов в ответе |
| `-think` | false | Режим рассуждений (Qwen3) |
| `-expand` | false | Распаковать веса в float32 |
| `-verbose` | false | Показывать внутренние логи llama.cpp |

**Примеры:**

```bash
# Показать список моделей
qwen -list

# Интерактивный чат
qwen -model qwen2.5-0.5b-instruct

# Один вопрос без интерактивного режима
qwen -model qwen2.5-7b-instruct -p "Объясни, что такое горутина"

# Локальный файл, конкретный системный промпт
qwen -file ./model.gguf -system "Ты помощник программиста" -p "Напиши Hello World на Go"

# Qwen3 с режимом рассуждений
qwen -model qwen3-8b -think

# Максимально быстрый ответ (expand + greedy)
qwen -model qwen2.5-3b-instruct -expand -temp 0 -p "2+2?"
```

Команды в интерактивном режиме:

| Команда | Действие |
|---|---|
| `/reset` | Очистить историю диалога |
| `/exit` или `/quit` | Выйти |

---

## CLI — бенчмарк векторных ядер `vecbench`

Измеряет производительность SIMD-ядер пакета `vec` на текущем процессоре.

```bash
go run github.com/notrobot1/goqwen/goqwen/cmd/vecbench@latest
```

| Флаг | По умолчанию | Описание |
|---|---|---|
| `-ops` | `all` | Операции: `dot`, `axpy`, `matvec`, `q8`, `rmsnorm`, `rope` |
| `-lvl` | `all` | Уровни SIMD: `avx512`, `avx2`, `neon`, `generic` или `all` |
| `-n` | 4096 | Размер вектора |
| `-sec` | 1.0 | Время замера каждой операции (секунды) |
| `-json` | false | Вывод в JSON |

**Примеры:**

```bash
# Все операции, все доступные уровни
go run ./cmd/vecbench

# Только dot и axpy, вектор 8192
go run ./cmd/vecbench -ops dot,axpy -n 8192

# Точный замер 2 секунды, только NEON
go run ./cmd/vecbench -lvl neon -sec 2

# JSON для обработки скриптом
go run ./cmd/vecbench -json | jq '.[] | select(.op=="dot")'

# Сравнить конкретные уровни
go run ./cmd/vecbench -ops dot -lvl avx2,generic -n 4096 -sec 2
```

Пример вывода на Apple M-series (ARM64 / NEON):

```
goqwen vecbench  arch=arm64  os=darwin  GOMAXPROCS=8
доступные уровни: neon, generic  активный: neon

операция     уровень         N        Мop/с       ГБ/с      нс/op
------------------------------------------------------------------
dot          neon         4096         2.78     91.00      360.1
axpy         neon         4096         3.69    181.15      271.3
matvec       neon         4096         0.00     91.05  1474178.1
q8           neon         4096         0.50     10.34     2006.4
rmsnorm      neon         4096         0.60     29.27     1679.4
rope         neon         4096         0.81     26.52     1235.5
dot          generic      4096         0.51     16.85     1944.2
...
```

---

## GPU-ускорение через llama.cpp

По умолчанию библиотека работает на чистом Go (CPU NEON/AVX). Для GPU-инференса
подключается бэкенд llama.cpp — он берёт вычисления на себя, API не меняется.

### Поддерживаемые платформы

| Платформа | Ускоритель | Что нужно |
|---|---|---|
| macOS (Apple Silicon) | Metal GPU | Xcode Command Line Tools |
| macOS (Intel) | Metal GPU | Xcode Command Line Tools |
| Linux | CUDA (NVIDIA) | CUDA Toolkit |
| Linux | Vulkan (AMD / Intel) | Vulkan SDK |
| Linux / Windows | CPU (AVX2/AVX-512) | только C++ компилятор |

### Шаг 1 — собрать библиотеки llama.cpp

Выполнить один раз на каждой целевой платформе:

```bash
# macOS (Metal GPU включён автоматически)
./scripts/build_llama_libs.sh

# Linux с CUDA
LLAMA_CUDA=ON ./scripts/build_llama_libs.sh

# Linux с Vulkan
LLAMA_VULKAN=ON ./scripts/build_llama_libs.sh
```

Результат — набор `.a` файлов в `internal/llamacpp/libs/<GOOS>_<GOARCH>/`.
Эти файлы коммитятся в репозиторий: пользователи библиотеки собирают проект
без cmake и без установки llama.cpp.

### Шаг 2 — собрать приложение

```bash
# С GPU через llama.cpp
CGO_ENABLED=1 go build -tags llamacpp .

# Без GPU (чистый Go, все платформы, без CGo)
go build .
```

При запуске в stderr выводится активный бэкенд:

```
goqwen: бэкенд llama.cpp (GPU/CPU), ctx=8192
# или
goqwen: бэкенд чистый Go (CPU NEON), ctx=8192
```

### Пример кода

Код приложения не меняется — бэкенд выбирается только флагом при сборке:

```go
model, err := goqwen.Load(path, &goqwen.Options{
    ContextSize: 8192,
    // Verbose: true, // показать логи GPU: компиляцию Metal-шейдеров,
    //                // распределение слоёв по устройствам и т.д.
})
defer model.Close()

chat := model.NewChat(nil)
defer chat.Close() // освобождает KV-кэш GPU для этого диалога

resp, _ := chat.Send(ctx, "Привет!")
fmt.Println(resp.Content)
```

Каждый `Chat` получает **собственный** KV-кэш на GPU — несколько параллельных
диалогов (например, планировщик + исполнитель в агентских сценариях) не мешают
друг другу.

### Управление логами

Внутренние логи llama.cpp (компиляция Metal-шейдеров, `sched_reserve`,
`ggml_metal_library_compile_pipeline` и т.д.) по умолчанию подавлены.

```go
// Включить логи для отладки GPU (до вызова Load):
model, err := goqwen.Load(path, &goqwen.Options{
    Verbose: true,
})

// Переключить после загрузки не нужно — Verbose действует на весь процесс.
```

### Производительность (Apple M1 Pro, Qwen-7B Q4_K)

| Операция | CPU NEON | GPU (Metal) | Ускорение |
|---|---|---|---|
| matmul F32 4096×4096 | ~1 300 мкс | ~660 мкс | **2×** |
| matmul Q8_0 4096×11008 | ~3 500 мкс | ~380 мкс | **9×** |
| matmul Q4_K 4096×11008 | ~3 200 мкс | ~640 мкс | **5×** |

---

## Скачивание моделей

Модели скачиваются с Hugging Face автоматически при первом запуске:

```go
path, err := goqwen.Models.Download(ctx, goqwen.Qwen25_7B, goqwen.Q4_K_M, &goqwen.DownloadOptions{
    Dir: "/path/to/models", // куда сохранять (по умолчанию — $HOME/.cache/goqwen)
    Progress: func(p goqwen.DownloadProgress) {
        fmt.Fprintf(os.Stderr, "\r%s: %.1f%%", p.File, 100*float64(p.Done)/float64(p.Total))
    },
})
```

Или сразу загрузить в одну строку:

```go
model, err := goqwen.LoadFromCatalog(ctx, goqwen.Qwen25_15B, goqwen.Q8_0, nil, nil)
```

---

## Требования к памяти (ориентировочно)

| Модель | Q4_K_M | Q8_0 | F16 |
|---|---|---|---|
| 0.5B | ~0.5 ГБ | ~0.8 ГБ | ~1.3 ГБ |
| 1.5B | ~1.1 ГБ | ~1.9 ГБ | ~3.5 ГБ |
| 7B | ~4.4 ГБ | ~7.7 ГБ | ~15 ГБ |
| 14B | ~8.5 ГБ | ~15 ГБ | ~29 ГБ |
| 32B | ~19 ГБ | ~34 ГБ | ~65 ГБ |
