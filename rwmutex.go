package goredislock

import "github.com/go-redis/redis/v8"

//只有取到锁的部分才可以使用可重入锁

// todo 读写锁
type RWMutex interface {
	RUnlock()
	RLock()
	Lock()
	Unlock()
	//todo 读写锁的可重入
}
type rwmutex struct {
	mutex
}

func newRWMutex(client *redis.Client, lockKey string, opts ...Option) RWMutex {
	m := &rwmutex{}
	return m
}

func (r *rwmutex) RUnlock() {
	//TODO implement me
	panic("implement me")
}

func (r *rwmutex) RLock() {
	//TODO implement me
	panic("implement me")
}

func (r *rwmutex) Lock() {
	//TODO implement me
	panic("implement me")
}

func (r *rwmutex) Unlock() {
	//TODO implement me
	panic("implement me")
}
