package nn

// Pool — параллельный цикл на фиксированных pre-spawned горутинах-воркерах.
// Горутины создаются один раз в NewPool и живут до вызова Close.
// Это избегает повторного создания горутин в горячем пути инференса.
type Pool struct {
	threads int
	work    chan poolJob
	done    chan struct{}
}

type poolJob struct {
	fn     func(lo, hi, w int)
	lo, hi int
	w      int
}

// NewPool создаёт пул; threads < 1 трактуется как 1.
func NewPool(threads int) *Pool {
	if threads < 1 {
		threads = 1
	}
	p := &Pool{
		threads: threads,
		// Буфер = threads: For() отправляет все задания без блокировки
		// и сразу начинает обрабатывать свой кусок.
		work: make(chan poolJob, threads),
		done: make(chan struct{}, threads),
	}
	for i := 1; i < threads; i++ {
		go p.runWorker()
	}
	return p
}

func (p *Pool) runWorker() {
	for j := range p.work {
		j.fn(j.lo, j.hi, j.w)
		p.done <- struct{}{}
	}
}

// Threads возвращает число потоков.
func (p *Pool) Threads() int { return p.threads }

// Close останавливает воркеры. Нельзя вызывать одновременно с For.
func (p *Pool) Close() {
	close(p.work)
}

// For делит [0,n) на куски и вызывает fn(lo, hi, worker) параллельно.
// worker — номер потока в [0, Threads()): по нему берут личный рабочий буфер.
// work — грубая оценка объёма работы (число умножений); мелкие задачи
// выполняются в вызывающей горутине, чтобы не платить за синхронизацию.
func (p *Pool) For(n, work int, fn func(lo, hi, worker int)) {
	t := p.threads
	if t > n {
		t = n
	}
	if t <= 1 || work < 1<<15 {
		fn(0, n, 0)
		return
	}
	chunk := (n + t - 1) / t
	sent := 0
	for w := 1; w < t; w++ {
		lo, hi := w*chunk, min((w+1)*chunk, n)
		if lo >= hi {
			break
		}
		p.work <- poolJob{fn: fn, lo: lo, hi: hi, w: w}
		sent++
	}
	// Главная горутина обрабатывает кусок 0 пока воркеры работают.
	fn(0, min(chunk, n), 0)
	for i := 0; i < sent; i++ {
		<-p.done
	}
}
