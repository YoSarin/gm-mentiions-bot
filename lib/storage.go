package lib

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/garyburd/redigo/redis"
	"github.com/martin-reznik/logger"
)

const RedisKeyPattern = "gm-mentions-bot:group:%v"

type Storage struct {
	redis *redis.Pool
	log   *logger.Log
}

func NewStorage(redis *redis.Pool, log *logger.Log) *Storage {
	return &Storage{redis, log}
}

func (h *Storage) getUsers(data CallbackData, alias string) ([]int, error) {
	c := h.redis.Get()
	defer c.Close()
	redisData, err := c.Do("HGET", redisKey(data.GroupID), alias)
	if err != nil {
		h.log.Warning("%v", err)
		return []int{}, errors.New("Redis Fetch Failed")
	}

	usersJoined, err := redis.String(redisData, nil)
	if err != nil {
		return []int{}, nil
	}

	userStrings := strings.Split(usersJoined, separator)
	out := make([]int, len(userStrings))

	for i, s := range userStrings {
		n, err := strconv.Atoi(s)
		if err == nil {
			out[i] = n
		}
	}

	return out, nil
}

func (h *Storage) saveUsers(data CallbackData, alias string, users []int) (err error) {
	c := h.redis.Get()
	if len(users) == 0 {
		_, err = c.Do("HDEL", redisKey(data.GroupID), alias)
	} else {
		userStrings := make([]string, len(users))
		for i, u := range users {
			userStrings[i] = strconv.Itoa(u)
		}
		usersJoined := strings.Join(userStrings, separator)
		defer c.Close()
		_, err = c.Do("HSET", redisKey(data.GroupID), alias, usersJoined)
	}
	return err
}

func (h *Storage) getAliases(data CallbackData) ([]string, error) {
	c := h.redis.Get()
	defer c.Close()
	redisData, err := c.Do("HKEYS", redisKey(data.GroupID))
	if err != nil {
		h.log.Warning("%v", err)
		return []string{}, errors.New("Redis Fetch Failed")
	}

	aliases, err := redis.Strings(redisData, nil)
	if err != nil {
		return []string{}, nil
	}

	return aliases, nil
}

func redisKey(group string) string {
	return fmt.Sprintf(RedisKeyPattern, group)
}

func mergeListsUnique(a []int, b []int) []int {
	unifyingMap := make(map[int]bool)
	for _, item := range a {
		unifyingMap[item] = true
	}
	for _, item := range b {
		unifyingMap[item] = true
	}

	merged := make([]int, len(unifyingMap))
	k := 0
	for item := range unifyingMap {
		merged[k] = item
		k++
	}

	return merged
}

func substractLists(a []int, b []int) []int {
	unifyingMap := make(map[int]bool)
	for _, item := range a {
		unifyingMap[item] = true
	}
	for _, item := range b {
		delete(unifyingMap, item)
	}

	merged := make([]int, len(unifyingMap))
	k := 0
	for item := range unifyingMap {
		merged[k] = item
		k++
	}

	return merged
}
