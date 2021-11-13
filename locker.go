package lock

import (
	"context"
	"time"
)

const (
	locked  = "locked"
	lockKey = "lock"
)

type Locker interface {
	GetLock(ctx context.Context, timeout time.Duration) error
	HasLock(ctx context.Context) (bool, error)
	ReleaseLock(ctx context.Context) error
}
