package goredislock

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"time"
)

type Mutex interface {
	//Lock 加锁
	Lock() (string, error)
	//TryLock 自旋锁
	TryLock(tickerTime time.Duration, timerTime time.Duration) (string, error)
	//Unlock 释放锁
	Unlock() error
	//Refresh 手动刷新锁的过期时间
	Refresh() error
}

type mutex struct {
	//name 锁的名称
	name string
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
	//close
	cancel context.CancelFunc
	// IsWatchDog 是否启用看门狗
	IsWatchDog bool
}

// newMutex 创建一个分布式锁
// client redis客户端
// lockKey 锁的key
// lockValue 锁的value（用于实现可重入式锁）
func newMutex(client *redis.Client, lockKey string, lockValue any, opts ...Option) Mutex {
	//todo 设置默认值
	m := &mutex{
		//默认redis客户端为本地客户端
		client: client,
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

// Lock 加锁
// 1. 判断 redis 上是否锁住了
// 2. 如果 redis 上成功，则返回当前锁的 key，否则返回空
func (m *mutex) Lock() (string, error) {
	//基于 redis 的 setNX 命令实现分布式锁
	ok, err := m.client.SetNX(context.Background(), m.lockKey, m.lockValue, m.expire).Result()
	if err != nil {
		//加锁失败了，尝试释放锁
		terr := m.Unlock()
		if terr != nil {
			err = errors.Wrap(err, terr.Error())
		}
		return "", errors.Wrap(ErrMutexHasLocked, err.Error())
	}
	//如果获取锁成功，则返回uuid
	if ok {
		//此处使用了 context 传递参数
		m.ctx, m.cancel = context.WithCancel(context.Background())
		//启动看门狗,同时如果过期时间为0，则不需要启动看门狗
		if m.IsWatchDog && m.expire != 0 {
			interval := m.expire * 2 / 3
			//为了防止一些过小的时间
			if interval == 0 {
				interval = m.expire
			}
			// 默认续费间隔为锁过期时间的 2/3
			go m.watchDog(interval)
		}
		return m.lockKey, nil
	}

	return "", ErrMutexHasLocked
}

// TryLock 自旋锁
// tickerTime 间隔,timerTime 持续时间
func (m *mutex) TryLock(tickerTime time.Duration, timerTime time.Duration) (string, error) {
	// 持续时间，间隔
	ticker := time.NewTimer(tickerTime)
	timer := time.NewTicker(timerTime)
	for {
		select {
		// 持续时间结束，自旋锁失败
		case <-timer.C:
			{
				return "", ErrMutexHasLocked
			}
		// 间隔一段时间开始加锁
		case <-ticker.C:
			{
				str, err := m.Lock()
				// 加锁成功
				if err == nil {
					return str, nil
				}
				continue
			}
		}
	}
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

// watchDog 看门狗机制
// 定期自动续费
func (m *mutex) watchDog(interval time.Duration) {
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-m.ctx.Done():
			{
				return
			}
			// 定时续费
		case <-ticker.C:
			{
				err := m.Refresh()
				if err != nil {
					//todo 如何取应对看门狗续费失败的情况
					return
				}
			}
		}

	}
}

// Unlock 释放锁
func (m *mutex) Unlock() error {
	//1. 先检测是否上锁，如果没有上锁，则直接返回
	if m.cancel == nil {
		return nil
	}
	//关闭看门狗
	m.cancel()
	m.cancel = nil
	// 不管怎么样，本地锁都要释放，远端锁释放失败可以靠超时机制来兜底释放
	//使用 lua 脚本实现原子操作
	_, err := unlockScript.Run(m.ctx, m.client, []string{m.lockKey}, m.lockValue).Result()
	if err != nil {
		return errors.Wrap(ErrUnlockFailed, err.Error())
	}
	return nil

}
