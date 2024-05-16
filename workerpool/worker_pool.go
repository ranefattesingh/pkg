package workerpool

import (
	"context"
	"sync"
)

type workerPool struct {
	workerCount int
	taskFnChan  chan func(ctx context.Context) error
	errChan     chan error
	wg          sync.WaitGroup
}

func NewWorkerPool(wc int, taskBufferSize int) *workerPool {
	return &workerPool{
		workerCount: wc,
		taskFnChan:  make(chan func(ctx context.Context) error, taskBufferSize),
		errChan:     make(chan error, wc), // Buffer size set to the number of workers
	}
}

// Allows submission of a new task to the pool.
func (wp *workerPool) AddTask(taskFn func(ctx context.Context) error) {
	wp.taskFnChan <- taskFn
}

func (wp *workerPool) StartAndExitOnErr(ctx context.Context) error {
	defer close(wp.errChan)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	wp.start(ctx)

	// Wait for all tasks to be completed
	wp.wg.Wait()

	return <-wp.errChan
}

func (wp *workerPool) StartAndIgnoreErr(ctx context.Context) {
	defer close(wp.errChan)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	wp.start(ctx)

	// Wait for all tasks to be completed
	wp.wg.Wait()
}

func (wp *workerPool) start(ctx context.Context) {
	wp.wg.Add(wp.workerCount)
	for i := 0; i < wp.workerCount; i++ {
		go func() {
			defer wp.wg.Done()
			worker(ctx, wp.taskFnChan, wp.errChan)
		}()
	}
}

func worker(ctx context.Context, taskFnChan <-chan func(ctx context.Context) error, errChan chan<- error) {
	for {
		select {
		case taskFn, ok := <-taskFnChan:
			if !ok {
				return
			}
			if err := taskFn(ctx); err != nil {
				errChan <- err
			}

		case <-ctx.Done():
			return
		}
	}
}

func (wp *workerPool) RecieveErr() <-chan error {
	return wp.errChan
}
