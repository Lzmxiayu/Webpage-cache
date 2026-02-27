package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"webpage-cache/internal/model"

	"github.com/redis/go-redis/v9"
)

type RedisQueue struct {
	client *redis.Client
	key    string
}

func NewRedisQueue(client *redis.Client, key string) *RedisQueue {
	return &RedisQueue{
		client: client,
		key:    key,
	}
}

func (q *RedisQueue) Push(task model.Task) error {
	payload, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("marshal task: %w", err)
	}

	if err := q.client.LPush(context.Background(), q.key, payload).Err(); err != nil {
		return fmt.Errorf("redis lpush: %w", err)
	}
	return nil
}

func (q *RedisQueue) Pop() (model.Task, error) {
	// 0 timeout means block until an item is available.
	result, err := q.client.BRPop(context.Background(), 0*time.Second, q.key).Result()
	if err != nil {
		return model.Task{}, fmt.Errorf("redis brpop: %w", err)
	}
	if len(result) != 2 {
		return model.Task{}, fmt.Errorf("unexpected brpop result length: %d", len(result))
	}

	var task model.Task
	if err := json.Unmarshal([]byte(result[1]), &task); err != nil {
		return model.Task{}, fmt.Errorf("unmarshal task: %w", err)
	}
	return task, nil
}

