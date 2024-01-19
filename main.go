package GoLockRedis

import (
	"context"
	"errors"
	"github.com/go-redis/redis/v8"
	"sync"
	"time"
)

var (
	//ErrMutexHasLocked 锁已经被占用
	ErrMutexHasLocked = errors.New("mutex has locked")
)
var (
	//lua脚本
	script = redis.NewScript(`
		if redis.call("get",KEYS[1]) == ARGV[1] then
			return redis.call("del",KEYS[1])
		else
			return 0
		end
	`)
)

//todo 如果有时间再做一个读写锁分离的分布式锁

type Mutex interface {
	//Lock 加锁
	Lock(opts ...Option) (string, error)
	//Unlock 释放锁
	Unlock() error
	//ForceUnlock 强制释放锁
	ForceUnlock() error
	//Refresh 手动刷新锁的过期时间
	Refresh() error
}
type mutex struct {
	//mutex 需要一个本地锁来节省redis的开销
	mutex sync.Mutex
	//client redis客户端
	client *redis.Client
	//lockKey 锁的key
	lockKey string
	//lockValue 锁的value
	lockValue any
	//expire 锁的过期时间
	expire time.Duration
	//ctx 上下文
	ctx context.Context
}

// NewMutex 创建一个分布式锁
// client redis客户端
// lockKey 锁的key
// lockValue 锁的value（用于实现可重入式锁）
func NewMutex(client *redis.Client, lockKey string, lockValue any) Mutex {
	//todo 设置默认值
	m := &mutex{
		client: client,
		//默认过期时间为30s
		expire: 30000 * time.Millisecond,
		//默认锁的key为lock
		lockKey: lockKey,
		//默认锁的value为uuid
		lockValue: lockValue,
	}
	return m
}

//  todo 对非原子操作的函数加锁与回溯

// Lock 加锁
func (m *mutex) Lock(opts ...Option) (string, error) {
	//使用本地锁来降低redis的压力，以及防止锁的重入
	m.mutex.Lock()
	defer m.mutex.Unlock()
	//选项模式
	for _, opt := range opts {
		opt.apply(m)
	}
	uuid := ""
	//todo 自旋锁
	//基于 redis 的 setnx 命令实现分布式锁
	ok, err := m.client.SetNX(m.ctx, m.lockKey, m.lockValue, m.expire).Result()
	if err != nil {
		//todo 如何额外处理错误
	}
	//如果获取锁成功，则返回uuid
	if ok {
		return uuid, nil
	}
	return uuid, ErrMutexHasLocked
}

// Unlock 释放锁
func (m *mutex) Unlock() error {
	//使用 lua 脚本实现原子操作
	//todo 通过lua脚本实现原子操作

	panic("implement me")
	return nil

}

// Refresh 手动刷新锁的过期时间
func (m *mutex) Refresh() error {
	//如果过期时间为0，则不需要刷新,直接返回
	if m.expire == 0 {
		return nil
	}

	//todo 通过lua脚本实现原子操作
	panic("implement me")
	return nil

}

// ForceUnlock 强制释放锁
func (m *mutex) ForceUnlock() error {
	panic("implement me")
}

// Option 选项接口
type Option interface {
	apply(mutex *mutex)
}

// OptionFunc 选项函数
type OptionFunc func(*mutex)

func (f OptionFunc) apply(mutex *mutex) {
	f(mutex)
}

// WithRedisClient 设置redis客户端
func WithRedisClient(client *redis.Client) OptionFunc {
	return func(mutex *mutex) {
		mutex.client = client
	}
}

// WithWatcherDog 启用看门狗
func WithWatcherDog() OptionFunc {
	//todo 如果过期时间为0，则不需要启用看门狗

	panic("implement me")
	return nil
}

// WithLockTimeout 设置锁超时时间
// timeout 为0时，表示锁的超时时间为永久，不建议设置锁为永久锁，因为如果锁的粒度过大，会导致锁的争抢过于激烈，影响并发性能
// timeout < 0 时，表示锁的超时时间为默认值
// timeout 单位为毫秒
func WithLockTimeout(timeout int) OptionFunc {
	if timeout < 0 {
		return func(mutex *mutex) {
			mutex.expire = 30000 * time.Millisecond
		}
	}
	return func(mutex *mutex) {
		mutex.expire = time.Duration(timeout) * time.Millisecond
	}
}

// todo 红锁算法
func main() {
	//client := redis.NewClient(
	//	&redis.Options{
	//		Addr: "localhost:6379",
	//	})

}
