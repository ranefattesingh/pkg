package main

import (
	"context"
	"fmt"
	"time"

	"github.com/ranefattesingh/pkg/workerpool" // Import your workerpool package
)

func main() {
	wp := workerpool.NewWorkerPool(5, 1)

	taskDone := make(chan struct{}) // Channel to signal task completion

	go func() {
		<-time.After(10 * time.Second)
		wp.AddTask(func(ctx context.Context) error {
			fmt.Println("Hi 1")
			close(taskDone) // Signal that the task is done
			return nil
		})
	}()

	go func() {
		err := wp.StartAndExitOnErr(context.Background())
		if err != nil {
			fmt.Println(err)
		}
	}()

	<-taskDone // Wait for the task to complete
}
