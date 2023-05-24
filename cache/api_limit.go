package cache

import (
	"errors"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"strings"
	"time"
)

const (
	CreateKeyListConfigCell string = "createKeyListConfigCell"
)

func (r *RedisCache) getSearchLimitKey(chainType common.ChainType, address, action string) string {
	key := fmt.Sprintf("search:limit:%s:%d:%s", action, chainType, address)
	return strings.ToLower(key)
}

var ErrDistributedLockPreemption = errors.New("distributed lock preemption")

func (r *RedisCache) LockWithRedis(chainType common.ChainType, address, action string, expiration time.Duration) (func() error, error) {
	log.Info("LockWithRedis:", chainType, address, action)
	key := r.getSearchLimitKey(chainType, address, action)
	ret := r.red.SetNX(key, "", expiration)
	if err := ret.Err(); err != nil {
		return nil, fmt.Errorf("redis set order nx-->%s", err.Error())
	}
	ok := ret.Val()
	log.Info("LockWithRedis:", ok)
	if !ok {
		return nil, ErrDistributedLockPreemption
	}
	del := func() error {
		return r.red.Del(key).Err()
	}
	return del, nil
}
