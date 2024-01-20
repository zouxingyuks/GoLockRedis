package main

import "github.com/go-redis/redis/v8"

// 此处是一些lua脚本
// lua 脚本的作用是保证原子性
var (
	//lua脚本
	unlockScript = redis.NewScript(`
		if redis.call("get",KEYS[1]) == ARGV[1] then
			return redis.call("del",KEYS[1])
		else
			return 0
		end
	`)
)