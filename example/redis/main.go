package main

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	_ "net/http/pprof"

	"github.com/go-redis/redis/v8"
	"github.com/nozo-moto/flushprint"
	"github.com/nozo-moto/lock"
	"golang.org/x/sync/errgroup"
)

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	go func() {
		t := time.NewTicker(time.Duration(1) * time.Second)
		for {
			select {
			case _ = <-t.C:
				//flushprint.Print("goroutine count ", runtime.NumGoroutine())
			}
		}
	}()

	redisLocker := lock.NewRedisLocker(
		redis.NewClient(
			&redis.Options{
				Addr:     "localhost:6379",
				Password: "",
				DB:       0,
			},
		),
	)

	eg, ctx := errgroup.WithContext(context.Background())
	for i := 0; i < 10; i++ {
		i := i
		eg.Go(func() error {
			worker := NewWorker(
				redisLocker,
			)
			return worker.Run(i, ctx)
		})
	}

	if err := eg.Wait(); err != nil {
		panic(err)
	}
}

type Count struct {
	count int
	mux   sync.Mutex
}

var count Count

type Worker struct {
	locker         lock.Locker
	getLockTimeout time.Duration
	id             int
	count          *Count
}

func NewWorker(locker lock.Locker) *Worker {
	return &Worker{
		getLockTimeout: time.Second * 2,
		count:          &count,
		locker:         locker,
	}
}

func (w *Worker) Run(id int, ctx context.Context) error {
	w.id = id
	t := time.NewTicker(time.Duration(10) * time.Millisecond)
	defer func() {
		t.Stop()
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case now := <-t.C:
			ctx := context.Background()
			if err := w.run(ctx, now, id); err != nil {
				return err
			}
		}
	}
}

func (w *Worker) run(ctx context.Context, now time.Time, id int) error {
	hasLock, err := w.locker.HasLock(ctx)
	if err != nil {
		return err
	}

	if hasLock {
		return nil
	}

	err = w.do(ctx)
	if err != nil {
	}
	return nil
}

func (w *Worker) do(ctx context.Context) (err error) {
	err = w.locker.GetLock(ctx, w.getLockTimeout)
	if err != nil {
		return err
	}
	defer func() {
		err = w.locker.ReleaseLock(context.Background())
		if err != nil {
		}
	}()

	w.increment()

	return nil
}

func (w *Worker) increment() {
	// w.count.mux.Lock()
	// defer w.count.mux.Unlock()
	time.Sleep(time.Millisecond * 3)
	w.count.count++
	flushprint.Print("count ", w.count.count, " id ", w.id)
}
