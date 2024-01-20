package main

import "github.com/pkg/errors"

var (
	//ErrMutexHasLocked 锁已经被占用
	ErrMutexHasLocked = errors.New("mutex has locked")
	ErrRefreshFailed  = errors.New("refresh failed")
)
