package goredislock

import (
	"fmt"
	"github.com/go-redis/redis/v8"
)

// todo 红锁算法
func main() {
	client := redis.NewClient(
		&redis.Options{
			Addr: "localhost:6379",
		})
	pool := NewPool(client)
	m := pool.NewMutex("25345", DefaultKeyGen, WithWatcherDog())
	fmt.Println(m.Lock())
	//time.Sleep(1 * time.Minute)
	//m.Unlock()
	//time.Sleep(1 * time.Minute)
}
