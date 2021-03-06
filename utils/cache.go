package utils

import (
	"fmt"
	"strings"
	"third/redigo/redis"
	"time"
)

type Cache struct {
	redisPool *redis.Pool
	Redis     RedisConfig
}

const (
	Success     int = 1
	KeyNotFound int = 2
	RedisError  int = 3
)

func CheckRedisReturnValue(err error) int {
	if err != nil && strings.Contains(err.Error(), "nil returned") {
		return KeyNotFound
	} else if err == nil {
		return Success
	} else {
		return RedisError
	}
}

func InitRedisPool(my_redis *RedisConfig) (*Cache, error) {
	cache := new(Cache)
	cache.Redis = *my_redis
	cache.RedisPool()
	return cache, nil
}

func (cache *Cache) RedisPool() *redis.Pool {
	if cache.redisPool == nil {
		cache.NewRedisPool(&cache.Redis)
	}
	return cache.redisPool
}

func (cache *Cache) NewRedisPool(my_redis *RedisConfig) {
	cache.redisPool = &redis.Pool{
		Dial: func() (redis.Conn, error) {
			var connect_timeout time.Duration = time.Duration(my_redis.ConnectTimeout) * time.Second
			var read_timeout = time.Duration(my_redis.ReadTimeout) * time.Second
			var write_timeout = time.Duration(my_redis.WriteTimeout) * time.Second

			c, err := redis.DialTimeout("tcp", my_redis.RedisConn, connect_timeout, read_timeout, write_timeout)

			if err != nil {
				fmt.Println("DialTimeout")
				return nil, err
			}
			if len(my_redis.RedisPasswd) > 0 {
				if _, err := c.Do("AUTH", my_redis.RedisPasswd); err != nil {
					c.Close()
					return nil, err
				}
			}

			if my_redis.RedisDb != "" {
				if _, err := c.Do("SELECT", my_redis.RedisDb); err != nil {
					c.Close()
					return nil, err
				}
			}

			return c, err
		},
		MaxIdle:     my_redis.MaxIdle,
		MaxActive:   my_redis.MaxActive,
		IdleTimeout: time.Duration(my_redis.IdleTimeout) * time.Second,
		Wait:        true,
	}
}

func (cache *Cache) Get(key string) ([]byte, error) {
	conn := cache.RedisPool().Get()
	defer conn.Close()

	res, err := redis.Bytes(conn.Do("GET", key))
	if nil != err && !strings.Contains(err.Error(), "nil returned") {
		err = NewInternalError(CacheErrCode, err)
	}
	return res, err
}

func (cache *Cache) Incr(key string) (int, error) {
	conn := cache.RedisPool().Get()
	defer conn.Close()
	res, err := redis.Int(conn.Do("INCR", key))

	if nil != err && !strings.Contains(err.Error(), "nil returned") {
		err = NewInternalError(CacheErrCode, err)
	}
	return res, err
}

func (cache *Cache) MGet(key []interface{}) (interface{}, error) {
	conn := cache.RedisPool().Get()
	defer conn.Close()
	res, err := conn.Do("MGET", key...)
	if nil != err && !strings.Contains(err.Error(), "nil returned") {
		err = NewInternalError(CacheErrCode, err)
	}
	return res, err
}

func (cache *Cache) MGetValue(keys []interface{}) ([]interface{}, error) {
	conn := cache.RedisPool().Get()
	defer conn.Close()
	res, err := redis.Values(conn.Do("MGET", keys...))
	if nil != err && !strings.Contains(err.Error(), "nil returned") {
		err = NewInternalError(CacheErrCode, err)
	}
	return res, err
}
func (cache *Cache) HSet(key interface{}, field string, value interface{}) (interface{}, error) {
	conn := cache.RedisPool().Get()
	defer conn.Close()
	res, err := conn.Do("HSET", key, field, value)

	if nil != err && !strings.Contains(err.Error(), "nil returned") {
		err = NewInternalError(CacheErrCode, err)
	}
	return res, err
}
func (cache *Cache) HMset(value []interface{}) (interface{}, error) {
	conn := cache.RedisPool().Get()
	defer conn.Close()
	res, err := conn.Do("HMSET", value...)

	if nil != err && !strings.Contains(err.Error(), "nil returned") {
		err = NewInternalError(CacheErrCode, err)
	}
	return res, err
}

func (cache *Cache) HGet(key interface{}, field string) (interface{}, error) {
	conn := cache.RedisPool().Get()
	defer conn.Close()
	res, err := conn.Do("HGET", key, field)

	if nil != err && !strings.Contains(err.Error(), "nil returned") {
		err = NewInternalError(CacheErrCode, err)
	}
	return res, err
}

func (cache *Cache) HIncrby(key, field string, value interface{}) (interface{}, error) {
	conn := cache.RedisPool().Get()
	defer conn.Close()
	res, err := conn.Do("HINCRBY", key, field, value)

	if nil != err && !strings.Contains(err.Error(), "nil returned") {
		err = NewInternalError(CacheErrCode, err)
	}
	return res, err
}

func (cache *Cache) Hmget(key string, fields []string) (interface{}, error) {
	conn := cache.RedisPool().Get()
	defer conn.Close()

	var args []interface{}
	args = append(args, key)
	for _, field := range fields {
		args = append(args, field)
	}

	res, err := conn.Do("HMGET", args...)

	if nil != err && !strings.Contains(err.Error(), "nil returned") {
		err = NewInternalError(CacheErrCode, err)
	}
	return res, err
}

func (cache *Cache) GetString(key string) (string, error) {
	conn := cache.RedisPool().Get()
	defer conn.Close()
	res, err := redis.String(conn.Do("GET", key))

	if nil != err && !strings.Contains(err.Error(), "nil returned") {
		err = NewInternalError(CacheErrCode, err)
	}
	return res, err
}

func (cache *Cache) GetInt64(key string) (int64, error) {
	conn := cache.RedisPool().Get()
	defer conn.Close()
	res, err := redis.Int64(conn.Do("GET", key))

	if nil != err && !strings.Contains(err.Error(), "nil returned") {
		err = NewInternalError(CacheErrCode, err)
	}
	return res, err
}

func (cache *Cache) GetStringMap(key string) (map[string]string, error) {
	conn := cache.RedisPool().Get()
	defer conn.Close()
	res, err := redis.StringMap(conn.Do("HGETALL", key))

	if nil != err && !strings.Contains(err.Error(), "nil returned") {
		err = NewInternalError(CacheErrCode, err)
	}
	return res, err
}

func (cache *Cache) HGetAll(key string) ([]byte, error) {
	conn := cache.RedisPool().Get()
	defer conn.Close()
	res, err := redis.Bytes(conn.Do("HGETALL", key))

	if nil != err && !strings.Contains(err.Error(), "nil returned") {
		err = NewInternalError(CacheErrCode, err)
	}
	return res, err
}

func (cache *Cache) GetInts(key string) ([]int, error) {
	conn := cache.RedisPool().Get()
	defer conn.Close()
	res, err := redis.Ints(conn.Do("GET", key))

	if nil != err && !strings.Contains(err.Error(), "nil returned") {
		err = NewInternalError(CacheErrCode, err)
	}
	return res, err
}

func (cache *Cache) Expire(key string, timeout int) error {
	conn := cache.RedisPool().Get()
	defer conn.Close()
	_, err := conn.Do("EXPIRE", key, timeout)

	if nil != err && !strings.Contains(err.Error(), "nil returned") {
		err = NewInternalError(CacheErrCode, err)
	}

	return err
}

func (cache *Cache) Set(key string, bytes interface{}, timeout int) error {
	var err error
	conn := cache.RedisPool().Get()
	defer conn.Close()
	if timeout == -1 {
		_, err = conn.Do("SET", key, bytes)
	} else {
		_, err = conn.Do("SET", key, bytes, "EX", timeout)
	}

	if nil != err && !strings.Contains(err.Error(), "nil returned") {
		err = NewInternalError(CacheErrCode, err)
	}
	return err
}

func (cache *Cache) Del(key string) error {
	conn := cache.RedisPool().Get()
	defer conn.Close()
	_, err := conn.Do("DEL", key)

	if nil != err && !strings.Contains(err.Error(), "nil returned") {
		err = NewInternalError(CacheErrCode, err)
	}
	return err
}

func (cache *Cache) Exists(key string) (bool, error) {
	conn := cache.RedisPool().Get()
	defer conn.Close()
	exists, err := redis.Bool(conn.Do("EXISTS", key))
	if err != nil {
		err = NewInternalError(CacheErrCode, err)
		Logger.Error(err.Error())
		return false, err
	}

	if exists {
		return true, nil
	} else {
		return false, nil
	}
}

func (cache *Cache) Zrange(key string, start, end int, withscores bool) ([]string, error) {
	conn := cache.RedisPool().Get()
	defer conn.Close()
	var res []string
	var err error
	if withscores {
		res, err = redis.Strings(conn.Do("ZRANGE", key, start, end, "withscores"))
	} else {
		res, err = redis.Strings(conn.Do("ZRANGE", key, start, end))
	}
	if nil != err && !strings.Contains(err.Error(), "nil returned") {
		err = NewInternalError(CacheErrCode, err)
	}
	return res, err
}

func (cache *Cache) ZrangeInts(key string, start, end int, withscores bool) ([]int, error) {
	conn := cache.RedisPool().Get()
	defer conn.Close()
	var res []int
	var err error
	if withscores {
		res, err = redis.Ints(conn.Do("ZRANGE", key, start, end, "withscores"))
	} else {
		res, err = redis.Ints(conn.Do("ZRANGE", key, start, end))
	}
	if nil != err && !strings.Contains(err.Error(), "nil returned") {
		err = NewInternalError(CacheErrCode, err)
	}
	return res, err
}

func (cache *Cache) Zrevrange(key string, start, end int, withscores bool) ([]int, error) {
	conn := cache.RedisPool().Get()
	defer conn.Close()
	var res []int
	var err error
	if withscores {
		res, err = redis.Ints(conn.Do("ZREVRANGE", key, start, end, "withscores"))
	} else {
		res, err = redis.Ints(conn.Do("ZREVRANGE", key, start, end))
	}

	if nil != err && !strings.Contains(err.Error(), "nil returned") {
		err = NewInternalError(CacheErrCode, err)
	}
	return res, err
}

func (cache *Cache) ZrevrangeStrings(key string, start, end int, withscores bool) ([]string, error) {
	conn := cache.RedisPool().Get()
	defer conn.Close()
	var res []string
	var err error
	if withscores {
		res, err = redis.Strings(conn.Do("ZREVRANGE", key, start, end, "withscores"))
	} else {
		res, err = redis.Strings(conn.Do("ZREVRANGE", key, start, end))
	}

	if nil != err && !strings.Contains(err.Error(), "nil returned") {
		err = NewInternalError(CacheErrCode, err)
	}
	return res, err
}

func (cache *Cache) ZrevrangeByScoreStrings(key string, max_num, min_num float64, withscores bool, offset, count int) ([]string, error) {
	conn := cache.RedisPool().Get()
	defer conn.Close()
	var res []string
	var err error
	if !withscores {
		res, err = redis.Strings(conn.Do("ZREVRANGEBYSCORE", key, max_num, min_num, "limit", offset, count))
	} else {
		res, err = redis.Strings(conn.Do("ZREVRANGEBYSCORE", key, max_num, min_num, "withscores", "limit", offset, count))
	}
	if nil != err && !strings.Contains(err.Error(), "nil returned") {
		err = NewInternalError(CacheErrCode, err)
	}
	return res, err
}

func (cache *Cache) ZrevrangeByScore(key string, max_num, min_num int, withscores bool, offset, count int) ([]int, error) {
	conn := cache.RedisPool().Get()
	defer conn.Close()
	var res []int
	var err error
	if !withscores {
		res, err = redis.Ints(conn.Do("ZREVRANGEBYSCORE", key, max_num, min_num, "limit", offset, count))
	} else {
		res, err = redis.Ints(conn.Do("ZREVRANGEBYSCORE", key, max_num, min_num, "withscores", "limit", offset, count))
	}
	if nil != err && !strings.Contains(err.Error(), "nil returned") {
		err = NewInternalError(CacheErrCode, err)
	}
	return res, err
}

func (cache *Cache) ZrangeByScore(key string, min_num, max_num int64, withscores bool, offset, count int) ([]int, error) {
	conn := cache.RedisPool().Get()
	defer conn.Close()
	var res []int
	var err error
	if withscores {
		res, err = redis.Ints(conn.Do("ZREVRANGEBYSCORE", key, max_num, min_num, "limit", offset, count))
	} else {
		res, err = redis.Ints(conn.Do("ZREVRANGEBYSCORE", key, max_num, min_num, "withscores", "limit", offset, count))
	}
	if nil != err && !strings.Contains(err.Error(), "nil returned") {
		err = NewInternalError(CacheErrCode, err)
	}
	return res, err
}

func (cache *Cache) Zscore(key string, member interface{}) (string, error) {
	conn := cache.RedisPool().Get()
	defer conn.Close()

	res, err := redis.String(conn.Do("ZSCORE", key, member))
	if nil != err && !strings.Contains(err.Error(), "nil returned") {
		err = NewInternalError(CacheErrCode, err)
	}
	return res, err
}

func (cache *Cache) Zadd(key string, value, member interface{}) (interface{}, error) {
	conn := cache.RedisPool().Get()
	defer conn.Close()

	res, err := conn.Do("ZADD", key, value, member)
	if nil != err && !strings.Contains(err.Error(), "nil returned") {
		err = NewInternalError(CacheErrCode, err)
	}
	return res, err
}

func (cache *Cache) Zincrby(key string, value, member interface{}) (interface{}, error) {
	conn := cache.RedisPool().Get()
	defer conn.Close()

	res, err := conn.Do("ZINCRBY", key, value, member)
	if nil != err && !strings.Contains(err.Error(), "nil returned") {
		err = NewInternalError(CacheErrCode, err)
	}
	return res, err
}

func (cache *Cache) ZincrbyFloat(key string, value float64, member interface{}) (float64, error) {
	conn := cache.RedisPool().Get()
	defer conn.Close()

	res, err := redis.Float64(conn.Do("ZINCRBY", key, value, member))
	if nil != err && !strings.Contains(err.Error(), "nil returned") {
		err = NewInternalError(CacheErrCode, err)
	}
	return res, err
}

func (cache *Cache) Zrank(key string, member interface{}) (int, error) {
	conn := cache.RedisPool().Get()
	defer conn.Close()

	var res int
	res, err := redis.Int(conn.Do("ZRANK", key, member))
	if nil != err && !strings.Contains(err.Error(), "nil returned") {
		err = NewInternalError(CacheErrCode, err)
	}
	return res, err
}

func (cache *Cache) Zrevrank(key string, member interface{}) (int, error) {
	conn := cache.RedisPool().Get()
	defer conn.Close()

	var res int
	res, err := redis.Int(conn.Do("ZREVRANK", key, member))
	if nil != err && !strings.Contains(err.Error(), "nil returned") {
		err = NewInternalError(CacheErrCode, err)
	}
	return res, err
}

func (cache *Cache) Zcard(key string) (int, error) {
	conn := cache.RedisPool().Get()
	defer conn.Close()

	var res int
	res, err := redis.Int(conn.Do("ZCARD", key))
	if nil != err && !strings.Contains(err.Error(), "nil returned") {
		err = NewInternalError(CacheErrCode, err)
	}
	return res, err
}

func (cache *Cache) Sadd(key string, items string) (int, error) {
	//var err error
	conn := cache.RedisPool().Get()
	defer conn.Close()
	res, err := redis.Int(conn.Do("SADD", key, items))
	if nil != err && !strings.Contains(err.Error(), "nil returned") {
		err = NewInternalError(CacheErrCode, err)
	}
	return res, err
}

func (cache *Cache) Rpush(key string, value interface{}) error {
	conn := cache.RedisPool().Get()
	defer conn.Close()
	_, err := conn.Do("RPUSH", key, value)
	if nil != err && !strings.Contains(err.Error(), "nil returned") {
		err = NewInternalError(CacheErrCode, err)
	}
	return err
}

func (cache *Cache) RpushBatch(keys []interface{}) error {
	conn := cache.RedisPool().Get()
	defer conn.Close()
	_, err := conn.Do("RPUSH", keys...)
	if nil != err && !strings.Contains(err.Error(), "nil returned") {
		err = NewInternalError(CacheErrCode, err)
	}
	return err
}

func (cache *Cache) Lrange(key string, start, end int) ([]interface{}, error) {
	conn := cache.RedisPool().Get()
	defer conn.Close()
	result, err := redis.Values(conn.Do("LRANGE", key, start, end))
	if nil != err && !strings.Contains(err.Error(), "nil returned") {
		err = NewInternalError(CacheErrCode, err)
	}
	return result, err
}

func (cache *Cache) Lrem(key string, value interface{}) error {
	conn := cache.RedisPool().Get()
	defer conn.Close()
	_, err := conn.Do("LREM", key, 0, value)
	if nil != err && !strings.Contains(err.Error(), "nil returned") {
		err = NewInternalError(CacheErrCode, err)
	}
	return err
}

func (cache *Cache) Lpop(key string) (int64, error) {
	conn := cache.RedisPool().Get()
	defer conn.Close()
	res, err := redis.Int64(conn.Do("LPOP", key))
	if nil != err && strings.Contains(err.Error(), "nil returned") {
		res = 0
		err = nil
	} else if nil != err {
		err = NewInternalError(CacheErrCode, err)
		Logger.Error("LPOP error :%v", err)
	}
	return res, err
}

func (cache *Cache) LpopString(key string) (string, error) {
	conn := cache.RedisPool().Get()
	defer conn.Close()
	res, err := redis.String(conn.Do("LPOP", key))
	if nil != err && strings.Contains(err.Error(), "nil returned") {
		res = ""
		err = nil
	} else if nil != err {
		err = NewInternalError(CacheErrCode, err)
		Logger.Error("LPOP error :%v", err)
	}
	return res, err
}

func (cache *Cache) LLEN(key string) (int64, error) {
	conn := cache.RedisPool().Get()
	defer conn.Close()
	res, err := redis.Int64(conn.Do("LLEN", key))
	if nil != err && strings.Contains(err.Error(), "nil returned") {
		res = 0
		err = nil
	} else if nil != err {
		err = NewInternalError(CacheErrCode, err)
		Logger.Error("LLEN error :%v", err)
	}
	return res, err
}

func (cache *Cache) Keys(pattern string) ([]string, error) {
	conn := cache.RedisPool().Get()
	defer conn.Close()
	res, err := redis.Strings(conn.Do("KEYS", pattern))
	if nil != err && strings.Contains(err.Error(), "nil returned") {
		res = nil
		err = nil
	} else if nil != err {
		err = NewInternalError(CacheErrCode, err)
		Logger.Error("KEYS error :%v", err)
	}
	return res, err
}
