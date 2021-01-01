package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"not-for-work/aviasales_test/config"
	"not-for-work/aviasales_test/internal/uniq"

	"github.com/go-redis/redis/v8"
)

var (
	NotUniqWordErr   = errors.New("cache: word is not uniq")
	EmptyResultErr   = errors.New("cache: empty result")
	NotValidCacheErr = errors.New("cache: not valid cache")
)

type Cache struct {
	client *redis.Client

	retries int
	ttl     time.Duration
}

func New(cfg config.Redis, dropData bool) (*Cache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	err := client.Ping(context.Background()).Err()
	if err != nil {
		return nil, fmt.Errorf("try to ping to redis: %w", err)
	}

	if dropData {
		keys, err := client.Keys(context.Background(), "*").Result()
		if err != nil {
			return nil, fmt.Errorf("drop datal keys: %w", err)
		}

		for _, k := range keys {
			client.Del(context.Background(), k)
		}
		log.Println("CACHE DATA DROPPED")
	}

	c := &Cache{
		client:  client,
		retries: cfg.CacheRetries,
		ttl:     time.Duration(cfg.CacheTTL) * time.Second,
	}

	return c, nil
}

func (s *Cache) Close() error {
	return s.client.Close()
}

func (s *Cache) Get(ctx context.Context, key string) ([]byte, error) {
	data, err := s.client.Get(ctx, key).Bytes()
	switch {
	case errors.Is(err, redis.Nil):
		return nil, EmptyResultErr
	case err != nil:
		return nil, fmt.Errorf("redis: Get value: %w", err)
	}

	var words []string
	if err := json.Unmarshal(data, &words); err != nil {
		return nil, fmt.Errorf("cache: unmarshal result: %w", err)
	}

	counts, err := s.client.Get(ctx, newCountsKey(key)).Int()
	if err != nil {
		return nil, fmt.Errorf("get key: %w", err)
	}

	if counts != len(words) {
		return nil, NotValidCacheErr
	}

	return data, nil
}

func (s *Cache) Put(ctx context.Context, key string, values ...string) error {
	txf := func(tx *redis.Tx) error {
		var words []string
		data, err := tx.Get(ctx, key).Bytes()
		switch {
		case errors.Is(err, redis.Nil):
			words = append(words, values...)
		case err != nil:
			return fmt.Errorf("redis: Get value: %w", err)
		default:
			if err := json.Unmarshal(data, &words); err != nil {
				return fmt.Errorf("unmarshal data from redis: %w", err)
			}
			var ok bool
			words, ok = uniq.Find(words, values)
			if !ok {
				return NotUniqWordErr
			}
		}

		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			bytes, err := json.Marshal(words)
			if err != nil {
				return fmt.Errorf("marshal words to redis: %w", err)
			}

			if err := tx.Set(ctx, key, bytes, s.ttl).Err(); err != nil {
				return fmt.Errorf("redis: set bytes: %w", err)
			}

			return nil
		})

		return err
	}

	for i := 0; i < s.retries; i++ {
		log.Printf("cache retry: %d\n", i)

		err := s.client.Watch(context.Background(), txf, key)
		switch {
		case errors.Is(err, redis.TxFailedErr): // Optimistic lock lost. Retry.
			continue
		case err != nil:
			return fmt.Errorf("redis watch: %w", err)
		default:
			return nil
		}
	}

	return nil
}

func (s *Cache) Set(ctx context.Context, key string, value interface{}) error {
	err := s.client.Set(ctx, newCountsKey(key), value, redis.KeepTTL).Err()
	if err != nil {
		return fmt.Errorf("cache: set: %w", err)
	}

	return nil
}

func newCountsKey(original string) string {
	return fmt.Sprintf("%s-counts", original)
}
