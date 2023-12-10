package workerpool

import (
	"context"
)

type workerPool struct {
	workerCount int
	taskFnChan  chan func(ctx context.Context) error
	errChan     chan error
}

func NewWorkerPool(wc int) *workerPool {
	return &workerPool{
		workerCount: wc,
		taskFnChan:  make(chan func(ctx context.Context) error, 1),
		errChan:     make(chan error, 1),
	}
}

// Allows submition of new task to the pool.
func (wp *workerPool) AddTask(taskFn func(ctx context.Context) error) {
	wp.taskFnChan <- taskFn
}

func (wp *workerPool) StartAndExitOnErr(ctx context.Context) error {
	defer func() {
		close(wp.errChan)
		close(wp.taskFnChan)
	}()

	wrappedCtx, cancelFn := context.WithCancel(ctx)
	defer cancelFn()

	wp.start(wrappedCtx)

	if err := <-wp.errChan; err != nil {
		return err
	}

	return nil
}

func (wp *workerPool) StartAndLogOnErr(ctx context.Context) {
	go func(taskFnChan chan func(ctx context.Context) error, errChan chan error) {
		defer func() {
			close(wp.errChan)
			close(wp.taskFnChan)
		}()

		wp.start(ctx)

		for {
			select {
			case <-ctx.Done():
				return

			case taskFn, ok := <-wp.taskFnChan:
				if ok {
					if err := taskFn(ctx); err != nil {
						errChan <- err
					}
				}
			}

		}
	}(wp.taskFnChan, wp.errChan)
}

func (wp *workerPool) start(ctx context.Context) {
	for i := 0; i < wp.workerCount; i++ {
		go worker(ctx, wp.taskFnChan, wp.errChan)
	}
}

func worker(ctx context.Context, taskFnChan <-chan func(ctx context.Context) error, errChan chan<- error) {
	for {
		select {
		case taskFn := <-taskFnChan:
			if err := taskFn(ctx); err != nil {
				errChan <- err
			}

		case <-ctx.Done():
			return
		}
	}
}
