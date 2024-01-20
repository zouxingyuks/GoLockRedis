package main

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"time"
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

// todo 红锁算法
func main() {
	client := redis.NewClient(
		&redis.Options{
			Addr: "localhost:6379",
		})
	pool := NewPool(client)
	m := pool.NewMutex(context.Background(), "lock", "uuid", WithWatcherDog())
	fmt.Println(m.Lock())
	time.Sleep(1 * time.Minute)
	m.Unlock()
}
