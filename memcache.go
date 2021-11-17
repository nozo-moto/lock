package lock

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
)

type MemcacheLocker struct {
	key    string
	client *memcache.Client
}

func NewMemcacheLocker(client *memcache.Client) *MemcacheLocker {
	return &MemcacheLocker{
		key:    lockKey,
		client: client,
	}
}

func (m *MemcacheLocker) GetLock(ctx context.Context, timeout time.Duration) error {
	m.client.Timeout = timeout
	err := m.client.Set(
		&memcache.Item{
			Key:   m.key,
			Value: []byte(locked),
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func (m *MemcacheLocker) HasLock(ctx context.Context) (bool, error) {
	item, err := m.client.Get(m.key)
	if err != nil {
		if err == memcache.ErrCacheMiss || err == io.EOF {
			return false, nil
		}
		return true, err
	}
	if bytes.Compare(item.Value, []byte(locked)) == 0 {
		return true, nil
	}

	return true, fmt.Errorf("value is missing")
}

func (m *MemcacheLocker) ReleaseLock(ctx context.Context) error {
	return m.client.Delete(m.key)
}
