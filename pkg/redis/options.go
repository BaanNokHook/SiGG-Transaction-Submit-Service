package redis

import (
	"github.com/go-redis/redis/v8"
)

// Option -.
type Option func(*redis.Options)

// MaxPoolSize -.
func Addr(addr string) Option {
	return func(c *redis.Options) {
		c.Addr = addr
	}
}

// Database -.
func Db(db int) Option {
	return func(c *redis.Options) {
		c.DB = db
	}
}

// Expiration -.
func Password(password string) Option {
	return func(c *redis.Options) {
		c.Password = password
	}
}
