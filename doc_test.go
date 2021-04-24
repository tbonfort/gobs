package gobs

import (
	"fmt"
	"time"
)

func ExampleNewPool() {
	pool := NewPool(2)
	pool.Submit(func() error {
		time.Sleep(10 * time.Millisecond)
		return nil
	})
	pool.Submit(func() error {
		time.Sleep(10 * time.Millisecond)
		return nil
	})
	pool.Stop()
	// both goroutines are now complete
}

func ExamplePool_Submit() {
	pool := NewPool(2)
	status := pool.Submit(func() error {
		time.Sleep(10 * time.Millisecond)
		return fmt.Errorf("an error occurred")
	})

	//do some other stuff while waiting for job to complete

	err := status.Wait()
	if err != nil {
		//err.Error() == "an error occurred"
	}
}

func ExamplePool_Batch() {
	pool := NewPool(2)
	status := pool.Submit(func() error {
		time.Sleep(100 * time.Millisecond)
		return nil
	})
	batch := pool.Batch()
	batch.Submit(func() error {
		time.Sleep(10 * time.Millisecond)
		return fmt.Errorf("an error 1 occurred")
	})
	batch.Submit(func() error {
		time.Sleep(10 * time.Millisecond)
		return fmt.Errorf("an error 2 occurred")
	})
	batch.Submit(func() error {
		time.Sleep(10 * time.Millisecond)
		return nil
	})
	err := batch.Wait()
	if err != nil {
		//err.Error() is "an error 1 occurred (and 1 other errors)" or "an error 2 occurred (and 1 other errors)"
		//err.(MultiError).Errors() is a slice of 2 errors
	}

	// this job is still running in the pool, i.e. not affected by batch.Wait()
	_ = status.Wait()
}
