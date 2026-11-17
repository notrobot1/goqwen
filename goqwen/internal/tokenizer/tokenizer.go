// Package tokenizer реализует байтовый BPE-токенизатор Qwen (семейство GPT-2)
// по словарю и списку слияний, которые хранятся прямо в файле GGUF.
package tokenizer

import (
	"sort"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"
)

// Типы токенов GGUF.
const (
	typeControl     = 3
	typeUserDefined = 4
)

// Tokenizer потокобезопасен.
type Tokenizer struct {
	tokens  []string
	types   []int32
	index   map[string]int
	ranks   map[[2]string]int
	decoded [][]byte // байтовое представление каждого токена

	special      map[byte][]string // спец-токены по первому байту, длинные первыми
	specialFirst string            // множество первых байтов (для IndexAny)

	byteSym [256]string // байт -> символ словаря GPT-2

	mu    sync.Mutex
	cache map[string][]int
}

// New строит токенизатор из массивов tokenizer.ggml.* файла GGUF.
func New(tokens []string, types []int32, merges []string) *Tokenizer {
	t := &Tokenizer{
		tokens:  tokens,
		types:   types,
		index:   make(map[string]int, len(tokens)),
		ranks:   make(map[[2]string]int, len(merges)),
		special: map[byte][]string{},
		cache:   map[string][]int{},
	}

	// Таблица GPT-2: печатные байты отображаются сами в себя, остальные —
	// в символы начиная с U+0100.
	symByte := map[rune]byte{}
	n := 0
	for b := 0; b < 256; b++ {
		r := rune(b)
		if !(b >= 33 && b <= 126 || b >= 161 && b <= 172 || b >= 174) {
			r = rune(256 + n)
			n++
		}
		t.byteSym[b] = string(r)
		symByte[r] = byte(b)
	}

	t.decoded = make([][]byte, len(tokens))
	for id, tok := range tokens {
		t.index[tok] = id
		if t.isSpecial(id) {
			t.decoded[id] = []byte(tok)
			if tok != "" {
				t.special[tok[0]] = append(t.special[tok[0]], tok)
			}
			continue
		}
		raw := make([]byte, 0, len(tok))
		ok := true
		for _, r := range tok {
			b, found := symByte[r]
			if !found {
				ok = false
				break
			}
			raw = append(raw, b)
		}
		if !ok {
			raw = []byte(tok)
		}
		t.decoded[id] = raw
	}
	first := make([]byte, 0, len(t.special))
	for b, list := range t.special {
		sort.Slice(list, func(i, j int) bool { return len(list[i]) > len(list[j]) })
		first = append(first, b)
	}
	t.specialFirst = string(first)

	for rank, m := range merges {
		if i := strings.IndexByte(m, ' '); i > 0 {
			t.ranks[[2]string{m[:i], m[i+1:]}] = rank
		}
	}
	return t
}

func (t *Tokenizer) isSpecial(id int) bool {
	if id >= len(t.types) {
		return false
	}
	return t.types[id] == typeControl || t.types[id] == typeUserDefined
}

// IsControl сообщает, является ли токен управляющим (<|im_end|> и т.п.).
func (t *Tokenizer) IsControl(id int) bool {
	return id >= 0 && id < len(t.types) && t.types[id] == typeControl
}

// Size возвращает размер словаря.
func (t *Tokenizer) Size() int { return len(t.tokens) }

// ID возвращает идентификатор токена по его тексту в словаре.
func (t *Tokenizer) ID(token string) (int, bool) {
	id, ok := t.index[token]
	return id, ok
}

// Bytes возвращает байты, которые токен добавляет к тексту.
func (t *Tokenizer) Bytes(id int) []byte {
	if id < 0 || id >= len(t.decoded) {
		return nil
	}
	return t.decoded[id]
}

// Decode собирает текст из токенов.
func (t *Tokenizer) Decode(ids []int) string {
	var b []byte
	for _, id := range ids {
		b = append(b, t.Bytes(id)...)
	}
	return string(b)
}

// Encode разбивает текст на токены. При parseSpecial=true служебные токены
// (<|im_start|>, <tool_call>, ...) распознаются в тексте как единое целое;
// при false текст кодируется «как есть» — так кодируется пользовательский
// ввод, чтобы им нельзя было подделать разметку диалога.
func (t *Tokenizer) Encode(text string, parseSpecial bool) []int {
	var out []int
	if !parseSpecial {
		return t.encodePlain(out, text)
	}
	for len(text) > 0 {
		i, tok := t.findSpecial(text)
		if i < 0 {
			out = t.encodePlain(out, text)
			break
		}
		out = t.encodePlain(out, text[:i])
		out = append(out, t.index[tok])
		text = text[i+len(tok):]
	}
	return out
}

// findSpecial ищет ближайший спец-токен в тексте.
func (t *Tokenizer) findSpecial(text string) (int, string) {
	off := 0
	for {
		i := strings.IndexAny(text[off:], t.specialFirst)
		if i < 0 {
			return -1, ""
		}
		off += i
		for _, tok := range t.special[text[off]] {
			if strings.HasPrefix(text[off:], tok) {
				return off, tok
			}
		}
		off++
	}
}

func (t *Tokenizer) encodePlain(out []int, text string) []int {
	for _, w := range preTokenize(text) {
		out = append(out, t.bpe(w)...)
	}
	return out
}

// bpe кодирует один фрагмент претокенизации.
func (t *Tokenizer) bpe(word string) []int {
	t.mu.Lock()
	ids, ok := t.cache[word]
	t.mu.Unlock()
	if ok {
		return ids
	}

	syms := make([]string, len(word))
	for i := 0; i < len(word); i++ {
		syms[i] = t.byteSym[word[i]]
	}
	for len(syms) > 1 {
		best, bestRank := -1, int(^uint(0)>>1)
		for i := 0; i+1 < len(syms); i++ {
			if r, ok := t.ranks[[2]string{syms[i], syms[i+1]}]; ok && r < bestRank {
				best, bestRank = i, r
			}
		}
		if best < 0 {
			break
		}
		a, b := syms[best], syms[best+1]
		merged := a + b
		j := 0
		for i := 0; i < len(syms); i++ {
			if i+1 < len(syms) && syms[i] == a && syms[i+1] == b {
				syms[j] = merged
				i++
			} else {
				syms[j] = syms[i]
			}
			j++
		}
		syms = syms[:j]
	}

	ids = make([]int, 0, len(syms))
	for _, s := range syms {
		if id, ok := t.index[s]; ok {
			ids = append(ids, id)
			continue
		}
		// Символа нет в словаре целиком — раскладываем на отдельные байты.
		for _, r := range s {
			if id, ok := t.index[string(r)]; ok {
				ids = append(ids, id)
			}
		}
	}

	t.mu.Lock()
	if len(t.cache) > 1<<16 {
		t.cache = map[string][]int{}
	}
	t.cache[word] = ids
	t.mu.Unlock()
	return ids
}

// preTokenize — ручная реализация регулярного выражения претокенизатора Qwen
// (в стандартном regexp нет опережающей проверки (?!\S)):
//
//	(?i:'s|'t|'re|'ve|'m|'ll|'d) | [^\r\n\p{L}\p{N}]?\p{L}+ | \p{N} |
//	 ?[^\s\p{L}\p{N}]+[\r\n]* | \s*[\r\n]+ | \s+(?!\S) | \s+
func preTokenize(text string) []string {
	rs := make([]rune, 0, len(text))
	offs := make([]int, 0, len(text)+1) // байтовое смещение каждой руны
	for i := 0; i < len(text); {
		r, size := utf8.DecodeRuneInString(text[i:])
		rs = append(rs, r)
		offs = append(offs, i)
		i += size
	}
	offs = append(offs, len(text))
	n := len(rs)

	isL := func(i int) bool { return i < n && unicode.IsLetter(rs[i]) }
	isN := func(i int) bool { return i < n && unicode.IsNumber(rs[i]) }
	isS := func(i int) bool { return i < n && unicode.IsSpace(rs[i]) }
	isNL := func(i int) bool { return i < n && (rs[i] == '\r' || rs[i] == '\n') }
	isOther := func(i int) bool { return i < n && !isS(i) && !isL(i) && !isN(i) }

	var out []string
	for i := 0; i < n; {
		j := i
		switch {
		case rs[i] == '\'' && contraction(rs[i+1:]) > 0:
			j = i + 1 + contraction(rs[i+1:])
		case !isNL(i) && !isL(i) && !isN(i) && isL(i+1):
			for j = i + 2; isL(j); j++ {
			}
		case isL(i):
			for j = i + 1; isL(j); j++ {
			}
		case isN(i):
			j = i + 1
		case isOther(i) || rs[i] == ' ' && isOther(i+1):
			for j = i + 1; isOther(j); j++ {
			}
			for ; isNL(j); j++ {
			}
		default: // пробельные символы
			end := i
			for ; isS(end); end++ {
			}
			lastNL := -1
			for k := end - 1; k >= i; k-- {
				if isNL(k) {
					lastNL = k
					break
				}
			}
			switch {
			case lastNL >= 0: // \s*[\r\n]+
				j = lastNL + 1
			case end == n: // \s+(?!\S) в конце текста
				j = end
			case end-i >= 2: // \s+(?!\S): оставляем последний пробел слову
				j = end - 1
			default: // \s+
				j = end
			}
		}
		if j <= i {
			j = i + 1
		}
		out = append(out, text[offs[i]:offs[j]])
		i = j
	}
	return out
}

// contraction возвращает длину английского сокращения после апострофа
// (s, t, re, ve, m, ll, d) или 0.
func contraction(rs []rune) int {
	lower := func(i int) rune {
		if i < len(rs) {
			return unicode.ToLower(rs[i])
		}
		return 0
	}
	switch a, b := lower(0), lower(1); {
	case a == 's' || a == 't':
		return 1
	case a == 'r' && b == 'e', a == 'v' && b == 'e':
		return 2
	case a == 'm':
		return 1
	case a == 'l' && b == 'l':
		return 2
	case a == 'd':
		return 1
	}
	return 0
}
