package main

import (
	"context"
	"github.com/go-redis/redis/v8"
)

type Pool struct {
	//client redis客户端
	client *redis.Client
}

func NewPool(client *redis.Client) *Pool {
	return &Pool{client: client}
}

// NewMutex 创建一个分布式锁
// client redis客户端
// lockKey 锁的key
// lockValue 锁的value（用于实现可重入式锁）
func (p *Pool) NewMutex(ctx context.Context, lockKey string, lockValue any, opts ...Option) Mutex {
	//todo 支持自定义uuid
	m := newMutex(p.client, lockKey, lockValue, opts...)
	return m
}
