package gobs

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMultiError(t *testing.T) {
	me := &multiErr{}
	wg := sync.WaitGroup{}
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func(i int) {
			defer wg.Done()
			if i%2 == 0 {
				me.add(fmt.Errorf("err %d", i))
			} else {
				me.add(nil)
			}
		}(i)
	}
	wg.Wait()
	assert.Contains(t, me.Error(), "(and 4 more errors)")
	assert.Len(t, me.Errors(), 5)
}

func TestMisc(t *testing.T) {
	assert.Panics(t, func() { NewPool(0) })
	assert.Panics(t, func() { NewPool(-1) })
	assert.NotPanics(t, func() { NewPool(1) })
}

func TestJobs(t *testing.T) {
	wg := sync.WaitGroup{}
	pool := NewPool(2)

	//check we don't catch a non-error
	wg.Add(1)
	go func() {
		defer wg.Done()
		st := pool.Submit(func() error {
			time.Sleep(200 * time.Millisecond)
			return nil
		})
		assert.NoError(t, st.Wait())
	}()

	//check we catch an error
	wg.Add(1)
	go func() {
		defer wg.Done()
		st := pool.Submit(func() error {
			time.Sleep(200 * time.Millisecond)
			return assert.AnError
		})
		assert.Equal(t, assert.AnError, st.Wait())
	}()

	//make sure previous goroutines have started
	time.Sleep(time.Millisecond)

	//check our job submission was queued
	wg.Add(1)
	go func() {
		now := time.Now()
		defer wg.Done()
		pool.Submit(func() error {
			return nil
		})
		assert.Greater(t, time.Since(now), 150*time.Millisecond)
	}()
	wg.Wait()

	now := time.Now()
	pool.Submit(func() error {
		time.Sleep(100 * time.Millisecond)
		return nil
	})
	pool.Stop()
	assert.Greater(t, time.Since(now), 100*time.Millisecond)
}

func TestBatch(t *testing.T) {
	p := NewPool(2)
	batch := p.Batch()
	now := time.Now()
	batch.Submit(func() error {
		time.Sleep(50 * time.Millisecond)
		return assert.AnError
	})
	batch.Submit(func() error {
		time.Sleep(50 * time.Millisecond)
		return nil
	})
	batch.Submit(func() error {
		time.Sleep(50 * time.Millisecond)
		return assert.AnError
	})
	batch.Submit(func() error {
		time.Sleep(50 * time.Millisecond)
		return nil
	})
	//check batch respect concurrency / i.e. jobs have been queued
	dur := time.Since(now)
	assert.Greater(t, dur, 50*time.Millisecond)
	assert.Less(t, dur, 100*time.Millisecond)

	err := batch.Wait()
	dur = time.Since(now)
	assert.Greater(t, dur, 100*time.Millisecond)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), " (and 1 more errors)")
	assert.Len(t, err.(MultiError).Errors(), 2)

	batch = p.Batch()
	batch.Submit(func() error {
		return nil
	})
	batch.Submit(func() error {
		return nil
	})
	assert.NoError(t, batch.Wait())
}
