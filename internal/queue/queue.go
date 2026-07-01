package queue

import (
	"context"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

type Queue struct {
	Client    *asynq.Client
	Inspector *asynq.Inspector
	redis     *redis.Client
}

func Connect(redisURL string) (*Queue, error) {
	opt, err := asynq.ParseRedisURI(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}

	rdb, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parse redis url for client: %w", err)
	}

	return &Queue{
		Client:    asynq.NewClient(opt),
		Inspector: asynq.NewInspector(opt),
		redis:     redis.NewClient(rdb),
	}, nil
}

func (q *Queue) Ping(ctx context.Context) error {
	return q.redis.Ping(ctx).Err()
}

func (q *Queue) Redis() *redis.Client {
	return q.redis
}

func (q *Queue) Close() error {
	if err := q.Client.Close(); err != nil {
		return err
	}
	if err := q.Inspector.Close(); err != nil {
		return err
	}
	return q.redis.Close()
}

func (q *Queue) NewServer(redisURL string, concurrency int) (*asynq.Server, error) {
	opt, err := asynq.ParseRedisURI(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}

	return asynq.NewServer(opt, asynq.Config{
		Concurrency: concurrency,
	}), nil
}

func (q *Queue) NewScheduler(redisURL string) (*asynq.Scheduler, error) {
	opt, err := asynq.ParseRedisURI(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}

	return asynq.NewScheduler(opt, nil), nil
}
