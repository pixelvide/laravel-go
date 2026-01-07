package cache

import (
	"context"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
)

type MemcachedStore struct {
	client *memcache.Client
}

func NewMemcachedStore(client *memcache.Client) *MemcachedStore {
	return &MemcachedStore{client: client}
}

func (s *MemcachedStore) Get(ctx context.Context, key string) (string, error) {
	item, err := s.client.Get(key)
	if err != nil {
		return "", err
	}
	return string(item.Value), nil
}

func (s *MemcachedStore) Put(ctx context.Context, key string, value string, ttl time.Duration) error {
	return s.client.Set(&memcache.Item{
		Key:        key,
		Value:      []byte(value),
		Expiration: int32(ttl.Seconds()),
	})
}

func (s *MemcachedStore) Forget(ctx context.Context, key string) error {
	return s.client.Delete(key)
}

func (s *MemcachedStore) Flush(ctx context.Context) error {
	return s.client.DeleteAll()
}
