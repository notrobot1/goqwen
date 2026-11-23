package goqwen

import (
	"fmt"
	"sort"
	"strings"
)

// Quant — вариант квантования весов (чем меньше бит, тем меньше файл и
// память, но ниже качество).
type Quant string

const (
	Q2K  Quant = "q2_k"
	Q3KM Quant = "q3_k_m"
	Q4_0 Quant = "q4_0"
	Q4KM Quant = "q4_k_m" // хороший баланс размера и качества
	Q5_0 Quant = "q5_0"
	Q5KM Quant = "q5_k_m"
	Q6K  Quant = "q6_k"
	Q8_0 Quant = "q8_0" // почти без потерь; самый быстрый формат на x86 (прямое SIMD-ядро)
	FP16 Quant = "fp16" // исходная точность
)

// Variant — один вариант квантования модели: файлы и их суммарный размер.
type Variant struct {
	Quant Quant
	Size  int64    // байт; примерно столько же нужно оперативной памяти
	Files []string // один файл или части разбитой модели по порядку
}

// ModelInfo — описание модели, которую можно загрузить.
type ModelInfo struct {
	ID          string // идентификатор в каталоге, например "qwen2.5-7b-instruct"
	Family      string // "Qwen2.5", "Qwen2.5-Coder" или "Qwen3"
	Parameters  string // размер модели: "0.5B", "7B", ...
	Repo        string // репозиторий на Hugging Face
	Thinking    bool   // поддерживает режим рассуждений (<think>)
	Description string
	Variants    []Variant
}

// Catalog — список моделей Qwen, которые умеет загружать и запускать
// библиотека. Каталог можно дополнять своими записями (поле Models).
type Catalog struct {
	Models []ModelInfo
}

// Models — встроенный каталог: официальные GGUF-сборки Qwen2.5,
// Qwen2.5-Coder и Qwen3 из репозиториев Qwen на Hugging Face.
var Models = Catalog{Models: catalogData}

// IDs возвращает идентификаторы всех моделей каталога.
func (c Catalog) IDs() []string {
	out := make([]string, len(c.Models))
	for i, m := range c.Models {
		out[i] = m.ID
	}
	return out
}

// Find ищет модель по идентификатору (без учёта регистра).
func (c Catalog) Find(id string) (ModelInfo, bool) {
	for _, m := range c.Models {
		if strings.EqualFold(m.ID, id) {
			return m, true
		}
	}
	return ModelInfo{}, false
}

// Family возвращает модели одного семейства ("Qwen2.5", "Qwen2.5-Coder", "Qwen3").
func (c Catalog) Family(family string) []ModelInfo {
	var out []ModelInfo
	for _, m := range c.Models {
		if strings.EqualFold(m.Family, family) {
			out = append(out, m)
		}
	}
	return out
}

// FitsIn возвращает модели, у которых есть вариант не больше maxBytes
// (удобно для выбора модели под доступную память).
func (c Catalog) FitsIn(maxBytes int64) []ModelInfo {
	var out []ModelInfo
	for _, m := range c.Models {
		for _, v := range m.Variants {
			if v.Size <= maxBytes {
				out = append(out, m)
				break
			}
		}
	}
	return out
}

// Variant возвращает вариант с заданным квантованием.
func (m ModelInfo) Variant(q Quant) (Variant, bool) {
	q = Quant(strings.ToLower(string(q)))
	for _, v := range m.Variants {
		if v.Quant == q {
			return v, true
		}
	}
	return Variant{}, false
}

// Quants возвращает доступные варианты квантования модели.
func (m ModelInfo) Quants() []Quant {
	out := make([]Quant, len(m.Variants))
	for i, v := range m.Variants {
		out[i] = v.Quant
	}
	return out
}

// DefaultQuant возвращает рекомендуемое квантование: q4_k_m, а если его
// нет — q8_0, иначе самый маленький из имеющихся вариантов.
func (m ModelInfo) DefaultQuant() Quant {
	for _, q := range []Quant{Q4KM, Q8_0} {
		if _, ok := m.Variant(q); ok {
			return q
		}
	}
	if len(m.Variants) == 0 {
		return ""
	}
	vs := append([]Variant(nil), m.Variants...)
	sort.Slice(vs, func(i, j int) bool { return vs[i].Size < vs[j].Size })
	return vs[0].Quant
}

func (m ModelInfo) String() string {
	qs := make([]string, len(m.Variants))
	for i, v := range m.Variants {
		qs[i] = fmt.Sprintf("%s (%.1f ГБ)", v.Quant, float64(v.Size)/1e9)
	}
	return fmt.Sprintf("%s — %s: %s", m.ID, m.Description, strings.Join(qs, ", "))
}
