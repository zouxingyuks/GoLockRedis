package goredislock

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"testing"
)

// 测试正常的加锁和解锁
func Test_mutex_Lock_Unlock(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "test1",
			want: "test1",
		},
	}

	pool := NewPool(redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	}), DefaultKeyGen, DefaultValueGen)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "gid", "test1")
			m := pool.NewMutex(ctx, tt.name)
			_, err := m.Lock()
			if err != nil {
				t.Errorf("Lock() error = %v", err)
				return
			}

			err = m.Unlock()
			if err != nil {
				t.Errorf("Unlock() error = %v", err)
				return
			}
		})
	}
}

// 测试重复加锁
func Test_mutex_repeat_Lock(t *testing.T) {
	tests := []struct {
		name string
		gid  string
	}{
		{
			name: "test1",
			gid:  "test1",
		},
	}

	pool := NewPool(redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	}), DefaultKeyGen, DefaultValueGen)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), "gid", tt.gid)
			wantGid := DefaultValueGen(ctx)
			m := pool.NewMutex(ctx, tt.name)
			_, err := m.Lock()
			if err != nil {
				t.Errorf("Lock() error = %v", err)
				return
			}
			fmt.Println("第一次加锁成功")
			gid, err := m.Lock()
			if err != nil && !errors.Is(err, ErrMutexHasLocked) {
				t.Errorf("Lock() error = %v", err)
				return
			}
			if wantGid != gid {
				t.Errorf("gid is %v, want %v", gid, DefaultValueGen(ctx))
				return
			}
			fmt.Println("重复加锁成功")

		})
	}
}
