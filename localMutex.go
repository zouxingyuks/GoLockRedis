package goredislock

import (
	"context"
	"sync"
)

type localMutex struct {
	mutex *sync.Mutex
	ctx   *context.Context
}

// todo 本地锁
func (l *localMutex) Lock() {
	//TODO implement me
	panic("implement me")
}

func (l *localMutex) Unlock() {
	//TODO implement me
	panic("implement me")
}
