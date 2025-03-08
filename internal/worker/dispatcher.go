package worker

import (
	"fmt"
	"sync"
)

type Worker struct {
	workerID int
	signal   chan int
	quit     chan bool
}

type Dispatcher struct {
	id      int
	workers []*Worker
	data    chan int
	mu      *sync.Mutex
	job     func()
}

func NewDispatcher(maxWorkers int, job func()) *Dispatcher {
	pool := make([]*Worker, maxWorkers)
	for i := 0; i < maxWorkers; i++ {
		pool[i] = &Worker{
			workerID: i,
			signal:   make(chan int),
			quit:     make(chan bool),
		}
	}
	return &Dispatcher{
		id:      maxWorkers,
		workers: pool,
		data:    make(chan int),
		mu:      &sync.Mutex{},
		job:     job,
	}
}
func (d *Dispatcher) Run() {
	// Start all workers
	for _, worker := range d.workers {
		fmt.Println("Starting worker", worker.workerID)
		go worker.job(d.job)
	}

	// Dispatch jobs to workers in a round-robin fashion
	go func() {
		i := 0
		for dt := range d.data {
			d.mu.Lock()
			workerIndex := i % len(d.workers)
			i++
			fmt.Printf("Dispatcher got signal:%d sending to worker %d\n", dt, i)
			d.workers[workerIndex].signal <- dt
			fmt.Printf("Dispatcher sent signal: %d to worker %d\n", dt, workerIndex)
			d.mu.Unlock()
		}
	}()
}

func (d *Dispatcher) Stop(i int) {
	// Tell each worker to quit
	d.mu.Lock()
	defer d.mu.Unlock()
	if i > len(d.workers) {
		i = len(d.workers)
	}
	for _, worker := range d.workers[:i] {
		worker.quit <- true
	}
	d.workers = d.workers[i:]

}
func (d *Dispatcher) AddWorker() {
	// Add a new worker to the pool

	newWorker := &Worker{
		workerID: d.id,
		signal:   make(chan int),
		quit:     make(chan bool),
	}
	d.id = d.id + 1
	d.mu.Lock()
	defer d.mu.Unlock()
	d.workers = append(d.workers, newWorker)
	go newWorker.job(d.job)

}
func (d *Dispatcher) SendSignal(data int) {
	d.data <- data
	fmt.Println("Added job", data)
}

func (w *Worker) job(j func()) {
	fmt.Printf("Worker %d is ready\n", w.workerID)
	var lock sync.Mutex
	for {
		select {
		case dt := <-w.signal:
			lock.Lock()
			j()
			fmt.Printf("Worker %d processing signal: %d\n", w.workerID, dt)
			lock.Unlock()
		case <-w.quit:
			lock.Lock()
			fmt.Printf("Worker %d is quitting\n", w.workerID)
			close(w.signal)
			lock.Unlock()
			return
		}
	}
}
