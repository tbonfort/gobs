# Gobs is a tiny worker pool
[![Go Reference](https://pkg.go.dev/badge/github.com/tbonfort/gobs.svg)](https://pkg.go.dev/github.com/tbonfort/gobs)
[![License](https://img.shields.io/github/license/tbonfort/gobs.svg)](https://github.com/tbonfort/gobs/blob/main/LICENSE)
[![Build Status](https://github.com/tbonfort/gobs/workflows/build/badge.svg?branch=main&event=push)](https://github.com/tbonfort/gobs/actions?query=workflow%3Abuild+event%3Apush+branch%3Amain)
[![Coverage Status](https://coveralls.io/repos/github/tbonfort/gobs/badge.svg?branch=main)](https://coveralls.io/github/tbonfort/gobs?branch=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/tbonfort/gobs)](https://goreportcard.com/report/github.com/tbonfort/gobs)

Package gobs implements a simple job pool where each individual job is run
concurrently in its own goroutine while ensuring that no more than a given number
of jobs can be ran at a time. It provides methods to ensure all jobs have been
completed, and to capture errors.

To submit jobs to a pool:
```go
func doStuff() error {
    time.Sleep(time.Second)
    return nil
}

//create a pool with max concurrency of 2
pool := gobs.NewPool(2)

st1 := pool.Submit(doStuff)
st2 := pool.Submit(doStuff)
st3 := pool.Submit(doStuff)

//wait for 1st job to terminate
err := st1.Wait()
if err!=nil { /* handle error */ }

//wait for all jobs to terminate
pool.Stop()
```

To submit a batch of jobs while making sure they have all terminated successfully:
```go
func doStuff() error {
    time.Sleep(time.Second)
    return nil
}

//create a pool with max concurrency of 2
pool := gobs.NewPool(2)

//create a holder for a batch of jobs
batch := pool.NewBatch()

batch.Submit(doStuff)
batch.Submit(doStuff)
batch.Submit(doStuff)

//wait for all batch jobs to terminate
err := batch.Wait()
if err!=nil { 
    /* handle error globally */
    
    /* or handle individual errors */
    var errors []error = err.(gobs.MultiError).Errors()
}

```

## Caveat

Standard care concerning [closures used with goroutines](https://golang.org/doc/faq#closures_and_goroutines)
should be taken. For example, consider the following code:
```go
pool := gobs.NewPool(10)
mu := sync.Mutex{}
ret := []int{}
for i := 0; i < 10; i++ {
    pool.Submit(func() error {
        mu.Lock()
        ret = append(ret, i)
        mu.Unlock()
        return nil
    })
}
pool.Stop()
sort.Slice(ret, func(i, j int) bool { return ret[i] < ret[j] })
fmt.Println(ret)
```
While the expected output would be `[0 1 2 3 4 5 6 7 8 9]` this program actually prints
an array containing numbers between 0 and 10, e.g. `[10 10 10 10 10 10 10 10 10 10]`!

To avoid this you should make sure that the loop index cannot be re-used from one iteration
to another by rewriting your code as:
```go
pool := gobs.NewPool(10)
mu := sync.Mutex{}
ret := []int{}
for i := 0; i < 10; i++ {
    idx := i
    pool.Submit(func() error {
        mu.Lock()
        ret = append(ret, idx)
        mu.Unlock()
        return nil
    })
}
pool.Stop()
sort.Slice(ret, func(i, j int) bool { return ret[i] < ret[j] })
fmt.Println(ret)
```

