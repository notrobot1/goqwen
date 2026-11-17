package tokenizer

import (
	"reflect"
	"strings"
	"testing"
)

func TestPreTokenize(t *testing.T) {
	cases := map[string][]string{
		"Hello, world!":     {"Hello", ",", " world", "!"},
		"I'm we'LL":         {"I", "'m", " we", "'LL"},
		"a  b":              {"a", " ", " b"},
		"x \n\n y":          {"x", " \n\n", " y"},
		"end  ":             {"end", "  "},
		"12 3":              {"1", "2", " ", "3"},
		"Привет, мир":       {"Привет", ",", " мир"},
		"f(x):\n    return": {"f", "(x", "):\n", "   ", " return"},
		"!!!\n\nnext":       {"!!!\n\n", "next"},
		"":                  nil,
	}
	for in, want := range cases {
		got := preTokenize(in)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("preTokenize(%q) = %q, want %q", in, got, want)
		}
		if strings.Join(got, "") != in {
			t.Errorf("preTokenize(%q) теряет текст", in)
		}
	}
}

// Мини-словарь: проверяем слияния BPE, спец-токены и обратное декодирование.
func TestEncodeDecode(t *testing.T) {
	tokens := []string{"h", "e", "l", "o", "Ġ", "he", "ll", "hell", "hello", "Ġhello", "<|im_end|>", "!"}
	types := []int32{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 3, 1}
	merges := []string{"h e", "l l", "he ll", "hell o", "Ġ hello"}
	tk := New(tokens, types, merges)

	got := tk.Encode("hello hello<|im_end|>!", true)
	want := []int{8, 9, 10, 11}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Encode = %v, want %v", got, want)
	}
	if s := tk.Decode(got); s != "hello hello<|im_end|>!" {
		t.Fatalf("Decode = %q", s)
	}
	// Без parseSpecial служебный токен не должен распознаваться.
	for _, id := range tk.Encode("<|im_end|>", false) {
		if id == 10 {
			t.Fatal("спец-токен распознан в пользовательском тексте")
		}
	}
	if !tk.IsControl(10) || tk.IsControl(0) {
		t.Fatal("IsControl")
	}
}
