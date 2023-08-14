package main

import (
	"log"
	"os"
	"sync"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"sticky_queues/file_processing"
)

func main() {
	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	hostname, err := os.Hostname()
	if err != nil {
		log.Fatalln("Unable to get hostname", err)
	}

	stickTaskQueue := file_processing.StickyTaskQueue{
		MultiTaskQueue:  hostname + "-multi",
		SingleTaskQueue: hostname + "-single",
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		w := worker.New(c, "activities-sticky-queues", worker.Options{})
		w.RegisterWorkflow(file_processing.FileProcessingWorkflow)

		w.RegisterActivityWithOptions(stickTaskQueue.GetStickyTaskQueue, activity.RegisterOptions{
			Name: "GetStickyTaskQueue",
		})
		err = w.Run(worker.InterruptCh())
		if err != nil {
			log.Fatalln("Unable to start worker", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		// Create a new worker listening on the multi-stick queue
		stickWorker := worker.New(c, stickTaskQueue.MultiTaskQueue, worker.Options{})

		stickWorker.RegisterActivity(file_processing.DownloadFile)
		stickWorker.RegisterActivity(file_processing.DeleteFile)

		err = stickWorker.Run(worker.InterruptCh())
		if err != nil {
			log.Fatalln("Unable to start worker", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		// Create a new worker listening on the single-stick queue
		stickWorker := worker.New(c, stickTaskQueue.SingleTaskQueue, worker.Options{
			MaxConcurrentActivityExecutionSize: 1,
		})

		stickWorker.RegisterActivity(file_processing.ProcessFile)

		err = stickWorker.Run(worker.InterruptCh())
		if err != nil {
			log.Fatalln("Unable to start worker", err)
		}
	}()

	// Wait for workers to close
	wg.Wait()
}
