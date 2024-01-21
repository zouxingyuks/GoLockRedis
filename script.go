package goredislock

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
	lockScript = redis.NewScript(`
		local key = KEYS[1]
        local value = ARGV[1]
        local expire = ARGV[2]

        local set_result = redis.call('SETNX', key, value)
        if set_result == 1 then
            redis.call('PEXPIRE', key, expire)
            return {1, value}
        else
            return {0, redis.call('GET', key)}
        end
	`)
)
