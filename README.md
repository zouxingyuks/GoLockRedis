# GoLockRedis

## 简介
这是一个基于Redis的分布式锁实现，用于在分布式应用中确保对共享资源的互斥访问。

## 安装
首先，确保你已经安装了Go语言和Redis。

使用以下命令安装该库：

```bash
go get  github.com/zouxingyuks/GoLockRedis
```

## DEMO

基于 goredislock 实现的防超卖示例

```go
package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	goredislock "github.com/zouxingyuks/GoLockRedis"
	"gorm.io/gorm"
	"strconv"
	"time"
)

type Product struct {
	gorm.Model
	Stock int
}

func main() {

	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	pool := goredislock.NewPool(client, goredislock.DefaultKeyGen, goredislock.DefaultValueGen)

	r := gin.Default()
	r.POST("/set-stock/:id", func(c *gin.Context) {
		id := c.Param("id")
		var input struct {
			Stock int `json:"stock"`
		}
		if err := c.BindJSON(&input); err != nil {
			c.JSON(400, gin.H{"error": "Invalid input"})
			return
		}
		ctx := c.Request.Context()
		// 将库存值设置到 Redis 中
		err := client.Set(ctx, "stock:"+id, input.Stock, 0).Err() // 0 表示永不过期
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to set stock in cache"})
			return
		}

		c.JSON(200, gin.H{"message": "Stock updated successfully"})
	})

	r.POST("/order/:id", func(c *gin.Context) {
		id := c.Param("id")

		// 获取分布式锁
		gid := uuid.New().String()
		ctx := context.WithValue(c.Request.Context(), "gid", gid)
		lock := pool.NewMutex(ctx, id, goredislock.WithWatcherDog())
		if _, err := lock.TryLock(200*time.Millisecond, 5*time.Second); err != nil {
			c.JSON(500, gin.H{"error": "Failed to acquire lock"})
			return
		}
		// 从 Redis 中获取库存
		stockVal, err := client.Get(c.Request.Context(), "stock:"+id).Result()
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to get stock from cache"})
			return
		}

		stock, _ := strconv.Atoi(stockVal)
		// 检查库存并处理订单
		if stock > 0 {
			stock -= 1
			client.Set(c.Request.Context(), "stock:"+id, stock, 0) // 更新库存
			err = lock.Unlock()
			if err != nil {
				c.JSON(500, gin.H{"error": "Failed to release lock", "err": err})
				return
			}
			c.JSON(200, gin.H{
				"message": "Order processed successfully",
			})
		} else {
			err = lock.Unlock()
			if err != nil {
				c.JSON(500, gin.H{"error": "Failed to release lock", "err": err})
				return
			}
			c.JSON(500, gin.H{
				"message": "Not enough stock",
			})
		}

	})

	r.Run(":8080") // 在默认端口启动服务
}

```

## 注意事项

- 请根据你的实际情况修改Redis客户端的配置。
- 在实际使用中，你可能需要处理锁超时的情况。
- 请确保适当地处理错误，以便在出现问题时进行恢复或报告。

## 贡献

如果你发现了bug或者有改进建议，请随时提交issue或者发起pull request。

