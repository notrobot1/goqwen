package goqwen

import (
	"math"
	"math/rand"
	"sort"
	"time"

	"github.com/notrobot1/goqwen/goqwen/vec"
)

// Sampling — параметры выбора следующего токена.
type Sampling struct {
	// Temperature — «температура». 0 и меньше — всегда самый вероятный токен
	// (детерминированный ответ).
	Temperature float32
	// TopK — выбирать только из K самых вероятных токенов (0 — без ограничения).
	TopK int
	// TopP — ядерная выборка: наименьший набор токенов с суммарной
	// вероятностью TopP (0 или 1 — выключено).
	TopP float32
	// RepeatPenalty — штраф за повтор недавних токенов (1 — выключен).
	RepeatPenalty float32
	// RepeatLastN — сколько последних токенов учитывать для штрафа.
	RepeatLastN int
	// MaxTokens — предел длины ответа в токенах (0 — до конца контекста).
	MaxTokens int
	// Seed — зерно генератора случайных чисел (0 — случайное).
	Seed int64
}

// DefaultSampling возвращает параметры, рекомендованные авторами Qwen.
func DefaultSampling() Sampling {
	return Sampling{Temperature: 0.7, TopK: 20, TopP: 0.8, RepeatPenalty: 1.05, RepeatLastN: 64, MaxTokens: 2048}
}

// Greedy возвращает параметры детерминированной генерации.
func Greedy() Sampling {
	return Sampling{Temperature: 0, RepeatPenalty: 1, MaxTokens: 2048}
}

type candidate struct {
	id int
	p  float32
}

type sampler struct {
	opt  Sampling
	rng  *rand.Rand
	cand []candidate
}

func newSampler(s *Sampling) *sampler {
	opt := DefaultSampling()
	if s != nil {
		opt = *s
	}
	seed := opt.Seed
	if seed == 0 {
		seed = time.Now().UnixNano()
	}
	return &sampler{opt: opt, rng: rand.New(rand.NewSource(seed))}
}

// sample выбирает следующий токен. logits изменяются на месте.
func (s *sampler) sample(logits []float32, history []int) int {
	o := &s.opt
	if o.RepeatPenalty > 0 && o.RepeatPenalty != 1 && o.RepeatLastN > 0 {
		recent := history
		if len(recent) > o.RepeatLastN {
			recent = recent[len(recent)-o.RepeatLastN:]
		}
		seen := make(map[int]struct{}, len(recent))
		for _, id := range recent {
			if _, dup := seen[id]; dup {
				continue
			}
			seen[id] = struct{}{}
			if logits[id] > 0 {
				logits[id] /= o.RepeatPenalty
			} else {
				logits[id] *= o.RepeatPenalty
			}
		}
	}
	if o.Temperature <= 0 {
		return vec.ArgMax(logits)
	}

	// Отбор K лучших: упорядоченный по убыванию список фиксированной длины.
	k := o.TopK
	if k <= 0 || k > len(logits) {
		k = len(logits)
	}
	c := s.cand[:0]
	if k < 256 {
		for id, l := range logits {
			if len(c) == k && l <= c[k-1].p {
				continue
			}
			i := len(c)
			if i < k {
				c = append(c, candidate{})
			} else {
				i = k - 1
			}
			for ; i > 0 && c[i-1].p < l; i-- {
				c[i] = c[i-1]
			}
			c[i] = candidate{id, l}
		}
	} else {
		for id, l := range logits {
			c = append(c, candidate{id, l})
		}
		sort.Slice(c, func(i, j int) bool { return c[i].p > c[j].p })
		c = c[:k]
	}
	s.cand = c

	// Softmax с температурой.
	mx := c[0].p
	var sum float32
	for i := range c {
		c[i].p = float32(math.Exp(float64((c[i].p - mx) / o.Temperature)))
		sum += c[i].p
	}
	// Top-P: отсекаем хвост распределения.
	if o.TopP > 0 && o.TopP < 1 {
		var acc float32
		for i := range c {
			acc += c[i].p
			if acc >= o.TopP*sum {
				c = c[:i+1]
				sum = acc
				break
			}
		}
	}
	r := s.rng.Float32() * sum
	for _, x := range c {
		if r -= x.p; r <= 0 {
			return x.id
		}
	}
	return c[len(c)-1].id
}
