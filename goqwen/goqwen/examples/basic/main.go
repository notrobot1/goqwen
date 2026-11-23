// Простейший пример: скачать модель из каталога, загрузить и поговорить.
//
//	go run ./examples/basic
//	go run ./examples/basic -file /path/to/model.gguf
package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/notrobot1/goqwen/goqwen"
)

func main() {
	file := flag.String("file", "", "локальный файл GGUF (если не задан — модель скачивается)")
	flag.Parse()
	ctx := context.Background()

	// Каталог моделей, которые можно загрузить.
	for _, m := range goqwen.Models.Family("Qwen2.5") {
		fmt.Println(" •", m.ID, m.Quants())
	}

	path := *file
	if path == "" {
		var err error
		path, err = goqwen.Models.Download(ctx, goqwen.Qwen25_05B, goqwen.Q8_0, &goqwen.DownloadOptions{
			Progress: func(p goqwen.DownloadProgress) {
				fmt.Printf("\rзагрузка: %.0f%%", 100*float64(p.Done)/float64(p.Total))
			},
		})
		fmt.Println()
		if err != nil {
			log.Fatal(err)
		}
	}

	model, err := goqwen.Load(path, nil)
	if err != nil {
		log.Fatal(err)
	}
	info := model.Info()
	fmt.Printf("%s, веса %s, SIMD: %s, потоков: %d\n", info.Name, info.WeightFormat, info.SIMD, info.Threads)

	chat := model.NewChat(&goqwen.ChatOptions{System: "Отвечай кратко и только по-русски."})

	// Обычный запрос.
	resp, err := chat.Send(ctx, "Назови столицу Франции одним словом.")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Ответ:", resp.Content)

	// Потоковый запрос; диалог помнит предыдущие реплики.
	fmt.Print("Ответ: ")
	resp, err = chat.SendStream(ctx, "А Германии?", func(text string) { fmt.Print(text) })
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\n[%d токенов промпта (%d из кэша), %d сгенерировано, %.1f ток/с]\n",
		resp.Usage.PromptTokens, resp.Usage.CachedTokens, resp.Usage.CompletionTokens, resp.TokensPerSecond())
}
