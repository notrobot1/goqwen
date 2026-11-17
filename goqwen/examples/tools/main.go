// Пример работы с инструментами (tools / function calling).
//
//	go run ./examples/tools -file /path/to/model.gguf
package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"goqwen"
)

type weatherArgs struct {
	City string `json:"city"`
}

func main() {
	file := flag.String("file", "", "локальный файл GGUF (если не задан — модель скачивается)")
	flag.Parse()
	ctx := context.Background()

	path := *file
	if path == "" {
		var err error
		// Для инструментов лучше брать модель от 1.5B и больше.
		if path, err = goqwen.Models.Download(ctx, goqwen.Qwen25_15B, goqwen.Q8_0, nil); err != nil {
			log.Fatal(err)
		}
	}
	model, err := goqwen.Load(path, nil)
	if err != nil {
		log.Fatal(err)
	}
	greedy := goqwen.Greedy()

	// 1) Инструмент с обработчиком: библиотека сама вызовет функцию и
	//    вернёт модели результат.
	weather := goqwen.Tool{
		Name:        "get_weather",
		Description: "Get the current weather in a given city",
		Parameters: goqwen.ObjectSchema(map[string]map[string]any{
			"city": goqwen.PropString("City name, e.g. Kyiv"),
		}, "city"),
		Handler: goqwen.ToolFunc(func(ctx context.Context, a weatherArgs) (string, error) {
			fmt.Printf("  [вызван get_weather(city=%q)]\n", a.City)
			return fmt.Sprintf(`{"city": %q, "temperature_c": 7, "condition": "light rain"}`, a.City), nil
		}),
	}
	chat := model.NewChat(&goqwen.ChatOptions{Tools: []goqwen.Tool{weather}, Sampling: &greedy})
	resp, err := chat.Send(ctx, "What's the weather in Kyiv right now?")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Ответ:", resp.Content)
	fmt.Println("Выполнено вызовов:", len(resp.Executed))

	// 2) Инструмент без обработчика: вызов возвращается нам, мы выполняем
	//    его сами и передаём результат обратно.
	calc := goqwen.Tool{
		Name:        "multiply",
		Description: "Multiply two numbers",
		Parameters: goqwen.ObjectSchema(map[string]map[string]any{
			"a": goqwen.PropNumber("first factor"),
			"b": goqwen.PropNumber("second factor"),
		}, "a", "b"),
	}
	chat = model.NewChat(&goqwen.ChatOptions{Tools: []goqwen.Tool{calc}, Sampling: &greedy})
	resp, err = chat.Send(ctx, "Use the tool to compute 1234 * 5678.")
	if err != nil {
		log.Fatal(err)
	}
	for _, call := range resp.ToolCalls {
		var args struct{ A, B float64 }
		if err := call.Bind(&args); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("  [модель просит %s(%v, %v)]\n", call.Name, args.A, args.B)
		chat.AddToolResult(call.Name, fmt.Sprint(args.A*args.B))
	}
	if len(resp.ToolCalls) > 0 {
		if resp, err = chat.Continue(ctx, nil); err != nil {
			log.Fatal(err)
		}
	}
	fmt.Println("Ответ:", resp.Content)
}
