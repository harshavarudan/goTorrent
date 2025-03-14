package worker

import (
	"fmt"
	"sync"
)

type Worker struct {
	workerID  int
	jobSignal chan Job
	quit      chan bool
}

type Dispatcher struct {
	id      int
	workers []*Worker
	jobChan chan Job
	mu      *sync.Mutex
}

type Job interface {
	Job()
}

func NewDispatcher(maxWorkers int) *Dispatcher {
	pool := make([]*Worker, maxWorkers)
	for i := 0; i < maxWorkers; i++ {
		pool[i] = &Worker{
			workerID:  i,
			jobSignal: make(chan Job),
			quit:      make(chan bool),
		}
	}
	return &Dispatcher{
		id:      maxWorkers,
		workers: pool,
		jobChan: make(chan Job),
		mu:      &sync.Mutex{},
	}
}
func (d *Dispatcher) Run() {
	// Start all workers
	for _, worker := range d.workers {
		fmt.Println("Starting worker", worker.workerID)
		go worker.Job()
	}

	// Dispatch jobs to workers in a round-robin fashion
	go func() {
		workerIndex := 0
		for dt := range d.jobChan {
			d.mu.Lock()
			workerIndex++
			workerIndex %= len(d.workers)
			fmt.Printf("Dispatcher got signal:%d sending to worker %d\n", dt, workerIndex)
			go func() {
				d.workers[workerIndex].jobSignal <- dt
			}()

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
		workerID:  d.id,
		jobSignal: make(chan Job),
		quit:      make(chan bool),
	}
	d.id = d.id + 1
	d.mu.Lock()
	defer d.mu.Unlock()
	d.workers = append(d.workers, newWorker)
	go newWorker.Job()

}
func (d *Dispatcher) SendJob(job Job) {
	d.jobChan <- job
	fmt.Println("Added job", job)
}

func (w *Worker) Job() {
	fmt.Printf("Worker %d is ready\n", w.workerID)

	for {
		select {
		case dt := <-w.jobSignal:

			dt.Job()
			fmt.Println("Worker ", w.workerID, " processing signal:")

		case <-w.quit:

			fmt.Printf("Worker %d is quitting\n\n", w.workerID)
			close(w.jobSignal)

			return
		}
	}
}
