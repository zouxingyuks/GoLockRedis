package main

import (
	"github.com/go-redis/redis/v8"
)

type Pool struct {
	//client redis客户端
	client *redis.Client
}

// keyGen 生成锁的key
type keyGen func(lockKey string) string

// DefaultKeyGen 默认的key生成器
var DefaultKeyGen = func(lockKey string) string {
	return "grlock:" + lockKey
}

func NewPool(client *redis.Client) *Pool {
	return &Pool{
		client: client,
	}
}

// NewMutex 创建一个分布式锁
// client redis客户端
// lockKey 锁的key
func (p *Pool) NewMutex(lockKey string, gen keyGen, opts ...Option) Mutex {
	m := newMutex(p.client, gen(lockKey), "", opts...)
	return m
}

// NewRWMutex 创建一个分布式读写锁
// client redis客户端
// lockKey 锁的key
// lockValue 锁的value（用于实现可重入式锁）
func (p *Pool) NewRWMutex(lockKey string, opts ...Option) RWMutex {
	m := newRWMutex(p.client, lockKey, opts...)
	return m
}
