package utils

import (
	"github.com/PharbersDeveloper/bp-go-lib/redis"
)

func AddKey2Redis(key string) (count int64, err error) {
	c, err := redis.GetRedisClient()
	if err != nil {
		return
	}
	count, err = c.Incr(key).Result()
	return
}

func CheckKeyExistInRedis(key string) (exist bool, err error) {

	c, e := redis.GetRedisClient()
	if e != nil {
		err = e
		return
	}

	i, e := c.Exists(key).Result()
	if e != nil {
		err = e
		return
	}

	switch i {
	case 1:
		exist = true
	}

	return
}
