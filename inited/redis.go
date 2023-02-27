package inited

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"time"
)

var Redis *RedisClient

type RedisClient struct {
	*redis.Client
}

func InitRedis(host string, pwd string, db int, poolSize int) {
	if Redis != nil {
		return
	}
	if poolSize == 0 {
		poolSize = 100
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s", host),
		Password: pwd,      // no password set
		DB:       db,       // use default DB
		PoolSize: poolSize, // 连接池大小
	})

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		panic(err)
	}
	Redis = &RedisClient{rdb}
	return
}

// Lock
// Zero expiration means the key has no expiration time.
func (rs *RedisClient) Lock(key string, value interface{}, exp time.Duration) (bool, error) {
	ctx := context.Background()
	return rs.SetNX(ctx, key, value, exp).Result()
}

func (rs *RedisClient) Unlock(key string) (bool, error) {
	ctx := context.Background()
	res := rs.Del(ctx, key)
	ok, err := res.Result()
	if err != nil {
		return false, err
	}
	if ok == 1 {
		return true, nil
	}
	return false, nil
}
