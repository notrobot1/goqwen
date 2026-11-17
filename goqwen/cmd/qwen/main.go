// Команда qwen — консольный чат с локальной моделью Qwen.
//
//	qwen -list                               показать каталог моделей
//	qwen -model qwen2.5-0.5b-instruct        скачать (если нужно) и начать диалог
//	qwen -file ./model.gguf -p "Привет"      один вопрос к локальному файлу
package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"goqwen"
)

func main() {
	var (
		list    = flag.Bool("list", false, "показать каталог моделей и выйти")
		id      = flag.String("model", goqwen.Qwen25_05B, "идентификатор модели из каталога")
		quant   = flag.String("quant", "", "квантование (q4_k_m, q8_0, ...); по умолчанию рекомендуемое")
		file    = flag.String("file", "", "путь к локальному файлу GGUF (вместо -model)")
		dir     = flag.String("dir", "", "каталог для скачанных моделей")
		system  = flag.String("system", "", "системный промпт")
		prompt  = flag.String("p", "", "один запрос (без интерактивного режима)")
		threads = flag.Int("threads", 0, "число потоков (0 — по числу ядер)")
		ctxSize = flag.Int("ctx", 0, "окно контекста в токенах (0 — 4096)")
		temp    = flag.Float64("temp", 0.7, "температура (0 — детерминированно)")
		maxTok  = flag.Int("max", 2048, "максимум токенов в ответе")
		think   = flag.Bool("think", false, "режим рассуждений (Qwen3)")
		expand  = flag.Bool("expand", false, "распаковать веса в float32 (быстрее, но больше памяти)")
	)
	flag.Parse()

	if *list {
		for _, m := range goqwen.Models.Models {
			fmt.Println(m)
		}
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	path := *file
	if path == "" {
		var err error
		path, err = goqwen.Models.Download(ctx, *id, goqwen.Quant(*quant), &goqwen.DownloadOptions{
			Dir: *dir,
			Progress: func(p goqwen.DownloadProgress) {
				fmt.Fprintf(os.Stderr, "\rзагрузка %s (%d/%d): %5.1f%%", p.File, p.FileIndex, p.FileCount, 100*float64(p.Done)/float64(p.Total))
			},
		})
		fmt.Fprintln(os.Stderr)
		if err != nil {
			fatal(err)
		}
	}

	model, err := goqwen.Load(path, &goqwen.Options{Threads: *threads, ContextSize: *ctxSize, ExpandWeights: *expand})
	if err != nil {
		fatal(err)
	}
	info := model.Info()
	fmt.Fprintf(os.Stderr, "%s: %d слоёв, веса %s (%.0f МБ), потоков %d, SIMD %s, контекст %d\n",
		info.Name, info.Layers, info.WeightFormat, float64(info.WeightBytes)/1e6, info.Threads, info.SIMD, info.ContextSize)

	s := goqwen.DefaultSampling()
	s.Temperature = float32(*temp)
	s.MaxTokens = *maxTok
	chat := model.NewChat(&goqwen.ChatOptions{System: *system, Sampling: &s, Thinking: *think})

	ask := func(q string) {
		resp, err := chat.SendStream(ctx, q, func(t string) { fmt.Print(t) })
		fmt.Println()
		if err != nil {
			fmt.Fprintln(os.Stderr, "ошибка:", err)
			return
		}
		if resp.Reasoning != "" {
			fmt.Fprintf(os.Stderr, "[рассуждения: %s]\n", resp.Reasoning)
		}
		fmt.Fprintf(os.Stderr, "[%d+%d токенов, %.1f ток/с]\n", resp.Usage.PromptTokens, resp.Usage.CompletionTokens, resp.TokensPerSecond())
	}

	if *prompt != "" {
		ask(*prompt)
		return
	}
	fmt.Fprintln(os.Stderr, "Введите сообщение (/reset — очистить историю, /exit — выход).")
	in := bufio.NewScanner(os.Stdin)
	in.Buffer(make([]byte, 1<<20), 1<<20)
	for {
		fmt.Print("> ")
		if !in.Scan() {
			return
		}
		switch q := strings.TrimSpace(in.Text()); q {
		case "":
		case "/exit", "/quit":
			return
		case "/reset":
			chat.Reset()
		default:
			ask(q)
		}
		if ctx.Err() != nil {
			return
		}
	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "ошибка:", err)
	os.Exit(1)
}
