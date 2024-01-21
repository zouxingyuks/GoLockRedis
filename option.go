package goredislock

import (
	"context"
	"time"
)

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
		m.IsWatchDog = true
	}
}

// WithLockTimeout 设置锁超时时间
// timeout 为0时，表示锁的超时时间为永久，不建议设置锁为永久锁，因为如果锁的粒度过大，会导致锁的争抢过于激烈，影响并发性能
// timeout < 0 时，表示锁的超时时间为默认值
// timeout 单位为毫秒
func WithLockTimeout(timeout time.Duration) OptionFunc {
	if timeout < 0 {
		return func(mutex *mutex) {
			mutex.expire = 30000 * time.Millisecond
		}
	}
	return func(mutex *mutex) {
		mutex.expire = timeout
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
