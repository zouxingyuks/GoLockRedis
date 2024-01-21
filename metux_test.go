package goredislock

import (
	"github.com/go-redis/redis/v8"
	"testing"
)

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
	}))
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := pool.NewMutex(tt.name, DefaultKeyGen)
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
