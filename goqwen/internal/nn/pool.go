package nn

import "sync"

// Pool — параллельный цикл по диапазону [0,n) на заданном числе потоков.
type Pool struct {
	threads int
}

// NewPool создаёт пул; threads < 1 трактуется как 1.
func NewPool(threads int) *Pool {
	if threads < 1 {
		threads = 1
	}
	return &Pool{threads: threads}
}

// Threads возвращает число потоков.
func (p *Pool) Threads() int { return p.threads }

// For делит [0,n) на куски и вызывает fn(lo, hi, worker) параллельно.
// worker — номер потока в [0, Threads()): по нему берут личный рабочий буфер.
// work — грубая оценка объёма работы (число умножений); мелкие задачи
// выполняются в вызывающей горутине, чтобы не платить за запуск потоков.
func (p *Pool) For(n, work int, fn func(lo, hi, worker int)) {
	t := p.threads
	if t > n {
		t = n
	}
	if t <= 1 || work < 1<<15 {
		fn(0, n, 0)
		return
	}
	var wg sync.WaitGroup
	chunk := (n + t - 1) / t
	for w := 1; w < t; w++ {
		lo, hi := w*chunk, min((w+1)*chunk, n)
		if lo >= hi {
			break
		}
		wg.Add(1)
		go func(lo, hi, w int) {
			defer wg.Done()
			fn(lo, hi, w)
		}(lo, hi, w)
	}
	fn(0, min(chunk, n), 0)
	wg.Wait()
}
