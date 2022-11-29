package main

type Job interface {
	Do()
}

type Worker struct {
	WorkerPool chan chan Job
	JobChannel chan Job
	quit       chan bool
}
type Supervisor struct {
	MaxWorkers int
	WorkerPool chan chan Job
	quit       chan bool
}

func NewWorker(workerPool chan chan Job) Worker {
	return Worker{
		WorkerPool: workerPool,
		JobChannel: make(chan Job),
		quit:       make(chan bool),
	}
}

// executes jobs
func (w Worker) Start() {
	go func() {
		for {
			w.WorkerPool <- w.JobChannel
			// take the Job channel and put it in workerpool
			select {
			case job := <-w.JobChannel:
				// if a job is received then execute it
				job.Do()
			case <-w.quit:
				// if quit is received then stop the worker
				return
			}
		}
	}()
}

func (w Worker) Stop() {
	go func() {
		w.quit <- true
	}()
}

func NewSupervisor(maxWorkers int) Supervisor {
	pool := make(chan chan Job, maxWorkers)
	return Supervisor{
		MaxWorkers: maxWorkers,
		WorkerPool: pool,
		quit:       make(chan bool),
	}
}

func (s Supervisor) Run() {
	for i := 0; i < s.MaxWorkers; i++ {
		worker := NewWorker(s.WorkerPool)
		worker.Start()
	}
	go s.dispatch()
}

var JobQueue chan Job

func (s Supervisor) dispatch() {
	for {
		select {
		case job := <-JobQueue:
			go func(job Job) {
				JobChannel := <-s.WorkerPool
				JobChannel <- job
			}(job)
		case <-s.quit:
			return
		}
	}
}

func (s Supervisor) Stop() {
	go func() {
		s.quit <- true
	}()
}

func runSuperVisor() {
	JobQueue = make(chan Job)

	supervisor := NewSupervisor(4)
	supervisor.Run()

}
