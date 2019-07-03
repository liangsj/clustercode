package redis

import (
	"fmt"
	"time"

	"github.com/garyburd/redigo/redis"
)

type Sentinel struct {
	Name  string
	Addrs []string
}

var S Sentinel = Sentinel{
	Name:  "default",
	Addrs: []string{"redis_sentinel_1:26379", "redis_sentinel_2:26379"},
}

func (s *Sentinel) getRedisConn() (redis.Conn, error) {

	for _, addr := range s.Addrs {
		sentinelConn, err := redis.DialTimeout("tcp", addr, 0, 1*time.Second, 1*time.Second)
		if err != nil {
			continue
		}
		defer sentinelConn.Close()
		res, err := redis.Strings(sentinelConn.Do("SENTINEL", "get-master-addr-by-name", "master"))
		fmt.Printf("redis_master_addr:%v", res)
		if err != nil {
			return nil, err
		}

		redisConn, err := redis.DialTimeout("tcp", fmt.Sprintf("%s:%s", res[0], res[1]), 0, 1*time.Second, 1*time.Second)
		if err != nil {
			continue
		}
		return redisConn, nil

	}
	return nil, fmt.Errorf("sentinel err")
}
