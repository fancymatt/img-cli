// Package concurrent provides utilities for parallel processing and worker pools.
// It includes a configurable worker pool for concurrent image processing tasks.
package concurrent

import (
	"context"
	"img-cli/pkg/logger"
	"sync"
)

// Task represents a unit of work to be processed
type Task interface {
	Process(ctx context.Context) error
	GetID() string
}

// Result wraps the outcome of a task execution
type Result struct {
	TaskID string
	Error  error
}

// WorkerPool manages concurrent task execution
type WorkerPool struct {
	workers    int
	taskQueue  chan Task
	results    chan Result
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewWorkerPool creates a new worker pool with the specified number of workers
func NewWorkerPool(workers int) *WorkerPool {
	if workers <= 0 {
		workers = 1
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &WorkerPool{
		workers:   workers,
		taskQueue: make(chan Task, workers*2), // Buffer for efficiency
		results:   make(chan Result, workers*2),
		ctx:       ctx,
		cancel:    cancel,
	}
}

// Start begins processing tasks
func (p *WorkerPool) Start() {
	logger.Info("Starting worker pool", "workers", p.workers)

	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}
}

// worker processes tasks from the queue
func (p *WorkerPool) worker(id int) {
	defer p.wg.Done()

	logger.Debug("Worker started", "worker_id", id)

	for {
		select {
		case task, ok := <-p.taskQueue:
			if !ok {
				logger.Debug("Worker stopping - queue closed", "worker_id", id)
				return
			}

			logger.Debug("Processing task",
				"worker_id", id,
				"task_id", task.GetID())

			err := task.Process(p.ctx)

			p.results <- Result{
				TaskID: task.GetID(),
				Error:  err,
			}

			if err != nil {
				logger.Error("Task failed",
					"worker_id", id,
					"task_id", task.GetID(),
					"error", err)
			} else {
				logger.Debug("Task completed",
					"worker_id", id,
					"task_id", task.GetID())
			}

		case <-p.ctx.Done():
			logger.Debug("Worker stopping - context cancelled", "worker_id", id)
			return
		}
	}
}

// Submit adds a task to the processing queue
func (p *WorkerPool) Submit(task Task) {
	select {
	case p.taskQueue <- task:
		logger.Debug("Task submitted", "task_id", task.GetID())
	case <-p.ctx.Done():
		logger.Warn("Cannot submit task - pool is shutting down", "task_id", task.GetID())
	}
}

// Results returns the results channel
func (p *WorkerPool) Results() <-chan Result {
	return p.results
}

// Wait blocks until all tasks are processed
func (p *WorkerPool) Wait() {
	close(p.taskQueue)
	p.wg.Wait()
	close(p.results)
}

// Shutdown gracefully stops the worker pool
func (p *WorkerPool) Shutdown() {
	logger.Info("Shutting down worker pool")
	p.cancel()
	close(p.taskQueue)
	p.wg.Wait()
	close(p.results)
}

// ProcessBatch processes a batch of tasks concurrently
func ProcessBatch(ctx context.Context, tasks []Task, workers int) []Result {
	if len(tasks) == 0 {
		return nil
	}

	// Limit workers to number of tasks
	if workers > len(tasks) {
		workers = len(tasks)
	}

	pool := NewWorkerPool(workers)
	pool.Start()

	// Submit all tasks
	go func() {
		for _, task := range tasks {
			pool.Submit(task)
		}
		pool.Wait()
	}()

	// Collect results
	results := make([]Result, 0, len(tasks))
	for result := range pool.Results() {
		results = append(results, result)
	}

	return results
}

// ImageProcessingTask implements Task for image processing
type ImageProcessingTask struct {
	ID          string
	InputPath   string
	OutputPath  string
	ProcessFunc func(ctx context.Context, input, output string) error
}

// Process executes the image processing task
func (t *ImageProcessingTask) Process(ctx context.Context) error {
	return t.ProcessFunc(ctx, t.InputPath, t.OutputPath)
}

// GetID returns the task identifier
func (t *ImageProcessingTask) GetID() string {
	return t.ID
}

// ParallelMap applies a function to items in parallel
func ParallelMap[T any, R any](ctx context.Context, items []T, workers int, fn func(context.Context, T) (R, error)) ([]R, error) {
	if len(items) == 0 {
		return nil, nil
	}

	type indexedResult struct {
		index  int
		result R
		err    error
	}

	resultsChan := make(chan indexedResult, len(items))
	var wg sync.WaitGroup

	// Create a semaphore to limit concurrency
	sem := make(chan struct{}, workers)

	for i, item := range items {
		wg.Add(1)
		go func(idx int, val T) {
			defer wg.Done()

			select {
			case sem <- struct{}{}: // Acquire semaphore
				defer func() { <-sem }() // Release semaphore

				result, err := fn(ctx, val)
				resultsChan <- indexedResult{
					index:  idx,
					result: result,
					err:    err,
				}

			case <-ctx.Done():
				resultsChan <- indexedResult{
					index: idx,
					err:   ctx.Err(),
				}
			}
		}(i, item)
	}

	// Wait for all goroutines and close channel
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect and sort results
	results := make([]R, len(items))
	var firstErr error

	for res := range resultsChan {
		if res.err != nil && firstErr == nil {
			firstErr = res.err
		}
		results[res.index] = res.result
	}

	return results, firstErr
}