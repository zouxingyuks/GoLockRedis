package goredislock

import (
	"context"
	"github.com/go-redis/redis/v8"
)

type Pool struct {
	//client redis客户端
	client *redis.Client
	kGen   keyGen
	vGen   valueGen
	m      map[string]*Mutex
}

// keyGen 生成锁的key
type keyGen func(lockKey string) string

// DefaultKeyGen 默认的key生成器
var DefaultKeyGen = func(lockKey string) string {
	return "grlock:" + lockKey
}

// valueGen 生成锁的value
type valueGen func(ctx context.Context) string

// DefaultValueGen 默认的value生成器
// 此处传入的 ctx 需要包含一个 gid，用于标识当前进程
var DefaultValueGen = func(ctx context.Context) string {
	return "grlock:" + ctx.Value("gid").(string)
}

func NewPool(client *redis.Client, kGen keyGen, vGen valueGen) *Pool {
	return &Pool{
		client: client,
		kGen:   kGen,
		vGen:   vGen,
	}
}

// NewMutex 创建一个分布式锁
// ctx 此处的 ctx 仅用于生成当前进程的唯一标识，用于实现可重入式锁。
// client redis客户端
// lockKey 临界区资源的唯一标识
// kGen 生成锁的key
// vFen 生成锁的value
func (p *Pool) NewMutex(ctx context.Context, lockKey string, opts ...Option) Mutex {
	m := NewMutex(p.client, p.kGen(lockKey), p.vGen(ctx), opts...)
	return m
}
