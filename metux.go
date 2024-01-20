package main

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"sync"
	"time"
)

//todo 只有取到锁的才可以使用可重入锁

// Option 选项接口
type Option interface {
	apply(mutex *mutex)
}

// OptionFunc 选项函数
type OptionFunc func(*mutex)

func (f OptionFunc) apply(mutex *mutex) {
	f(mutex)
}

// WithWatcherDog 启用看门狗
func WithWatcherDog() OptionFunc {
	return func(m *mutex) {
		m.Fns = append(m.Fns, func() {
			// 如果锁本身是永不过期的，不启用看门狗
			if m.expire == 0 {
				return
			}
			interval := m.expire * 2 / 3
			//为了防止一些过小的时间
			if interval == 0 {
				interval = m.expire
			}
			// 默认续费间隔为锁过期时间的 2/3
			go m.watchDog(m.watch, interval)
		})
	}
}

// 看门狗机制
// 定期自动续费
func (m *mutex) watchDog(watch <-chan bool, interval time.Duration) {
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-watch:
			{
				//todo 结束
				return
			}
			// 定时续费
		case <-ticker.C:
			{
				m.Refresh()
			}
		}

	}
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

// WithContext 设置上下文
func WithContext(ctx context.Context) OptionFunc {
	return func(mutex *mutex) {
		mutex.ctx = ctx
	}
}

// WithTries 设置重试次数
func WithTries(tries int) OptionFunc {
	return func(mutex *mutex) {

	}
}

//todo 如果有时间再做一个读写锁分离的分布式锁

// newMutex 创建一个分布式锁
// client redis客户端
// lockKey 锁的key
// lockValue 锁的value（用于实现可重入式锁）
func newMutex(client *redis.Client, lockKey string, lockValue any, opts ...Option) Mutex {
	//todo 设置默认值
	m := &mutex{
		//默认redis客户端为本地客户端
		client: client,
		//默认上下文为后台上下文
		ctx: context.Background(),
		//默认过期时间为30s
		expire: 30000 * time.Millisecond,
		//默认锁的key为lock
		lockKey: lockKey,
		//默认锁的value为uuid
		lockValue: lockValue,
	}
	//选项模式
	for _, opt := range opts {
		opt.apply(m)
	}
	return m
}

type Mutex interface {
	//Lock 加锁
	Lock() (string, error)
	//TryLock 自旋锁
	TryLock(timeout time.Duration) (string, error)
	//Unlock 释放锁
	Unlock() error
	//ForceUnlock 强制释放锁
	ForceUnlock() error
	//Refresh 手动刷新锁的过期时间
	Refresh() error
}
type mutex struct {
	//name 锁的名称
	name string
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
	//watch 看门狗
	watch chan bool
	// Fns
	Fns []func()
}

// keyGen 生成锁的key
type keyGen func() string

// valueGen 生成锁的value
type valueGen func() any

//  todo 对非原子操作的函数加锁与回溯

// Lock 加锁
// 1. 先判断本地锁
// 2. 判断 redis 上是否锁住了
// 3. 如果 redis 上成功，则返回当前锁的 key，否则返回空
func (m *mutex) Lock() (string, error) {
	//使用本地锁来降低redis的压力，以及防止锁的重入
	m.mutex.Lock()
	defer m.mutex.Unlock()
	// 执行所有的外挂函数
	for _, fn := range m.Fns {
		fn()
	}
	//基于 redis 的 setnx 命令实现分布式锁
	ok, err := m.client.SetNX(m.ctx, m.lockKey, m.lockValue, m.expire).Result()
	if err != nil {
		return "", errors.Wrap(ErrMutexHasLocked, err.Error())
		//todo 如何额外处理错误
	}
	//如果获取锁成功，则返回uuid
	if ok {
		return m.lockKey, nil
	}
	return "", ErrMutexHasLocked
}

// TryLock 自旋锁
// 在一段时间内反复尝试上锁
// 持续时间，间隔
func (m *mutex) TryLock(timeout time.Duration) (string, error) {
	// 持续时间，间隔
	ticker := time.NewTimer(timeout)
	jiangeL := time.NewTicker(30 * time.Second)
	for {
		select {

		// 持续时间结束，自旋锁失败
		case <-jiangeL.C:
			{
				return "", ErrMutexHasLocked
			}
		// 间隔一段时间开始加锁
		case <-ticker.C:
			{
				m.Lock()
			}
		}
	}
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
	// 刷新锁的过期时间
	success, err := m.client.Expire(m.ctx, m.lockKey, m.expire).Result()
	if err != nil {
		return errors.Wrap(ErrRefreshFailed, err.Error())
	}
	if !success {
		return ErrRefreshFailed
	}
	return nil

}

// ForceUnlock 强制释放锁
func (m *mutex) ForceUnlock() error {
	// 如果锁本身已经空了，那么也就
	if m.lockValue == nil {

		return nil
	}
	panic("implement me")
}
