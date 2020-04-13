package utils

import (
	"log"
	"sync"
	"time"

	"github.com/garyburd/redigo/redis"
)

//RedisSessionManager - manager sessions for Redis
type RedisSessionManager struct {
	ModuleName  string
	ConnStr     string
	Pool        *redis.Pool
	mutex       *sync.Mutex
	MaxIdle     int
	MaxActive   int
	IdleTimeout int
}

//NewRedisSessionManager - create a new RedisSessionManager
func NewRedisSessionManager(moduleName, connectionStr string, maxidle, maxactive, idletimeout int) *RedisSessionManager {
	return &RedisSessionManager{
		ModuleName:  moduleName,
		ConnStr:     connectionStr,
		mutex:       &sync.Mutex{},
		MaxIdle:     maxidle,
		MaxActive:   maxactive,
		IdleTimeout: idletimeout,
	}
}

//Get clone a mongo session from main session
func (r *RedisSessionManager) Get() (redis.Conn, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if r.Pool == nil {
		r.Pool = &redis.Pool{
			MaxIdle:     r.MaxIdle,
			MaxActive:   r.MaxActive,
			IdleTimeout: time.Duration(r.IdleTimeout) * time.Second,
			Dial: func() (redis.Conn, error) {
				c, err := redis.Dial("tcp", r.ConnStr)
				if err != nil {
					return nil, err
				}
				return c, nil
			},
		}
	}
	var conn = r.Pool.Get()
	return conn, conn.Err()
}

//GetPool - Get Redis pool
func (r *RedisSessionManager) GetPool() *redis.Pool {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if r.Pool == nil {
		r.Pool = &redis.Pool{
			MaxIdle:     r.MaxIdle,
			MaxActive:   r.MaxActive,
			IdleTimeout: time.Duration(r.IdleTimeout) * time.Second,
			Dial: func() (redis.Conn, error) {
				c, err := redis.Dial("tcp", r.ConnStr)
				if err != nil {
					return nil, err
				}
				return c, nil
			},
		}
	}
	return r.Pool
}

//Dispose - called when disposing the RedisSessionManager
func (r *RedisSessionManager) Dispose() {
	if r.Pool != nil {
		r.Pool.Close()
		log.Printf("[INFO] - Module:[%s] -> Redis Session Closed.\r\n", r.ModuleName)
	}
}
