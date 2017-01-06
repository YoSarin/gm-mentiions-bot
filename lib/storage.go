package lib

import (
	"errors"
	"fmt"
	"strings"

	"github.com/garyburd/redigo/redis"
	"github.com/martin-reznik/logger"
)

const RedisKeyPattern = "gm-mentions-bot:group:%v:alias:%v"

type Storage struct {
	redis redis.Conn
	log   *logger.Log
}

func NewStorage(redis redis.Conn, log *logger.Log) *Storage {
	return &Storage{redis, log}
}

func (h *Storage) getUsers(data *CallbackData, alias string) ([]string, error) {
	redisData, err := h.redis.Do("GET", redisKey(data.GroupID, alias))
	if err != nil {
		h.log.Warning("%v", err)
		return []string{}, errors.New("Redis Fetch Failed")
	}

	usersJoined, err := redis.String(redisData, nil)
	if err != nil {
		return []string{}, nil
	}

	return strings.Split(usersJoined, separator), nil
}

func (h *Storage) saveUsers(data *CallbackData, alias string, users []string) error {
	usersJoined := strings.Join(users, separator)
	_, err := h.redis.Do("SET", redisKey(data.GroupID, alias), usersJoined)
	return err
}

func (h *Storage) getToken(data *CallbackData, alias string) ([]string, error) {
	redisData, err := h.redis.Do("GET", redisKey(data.GroupID, alias))
	if err != nil {
		h.log.Warning("%v", err)
		return []string{}, errors.New("Redis Fetch Failed")
	}

	usersJoined, err := redis.String(redisData, nil)
	if err != nil {
		return []string{}, nil
	}

	return strings.Split(usersJoined, separator), nil
}

func redisKey(group string, alias string) string {
	return fmt.Sprintf(RedisKeyPattern, group, alias)
}

func mergeListsUnique(a []string, b []string) []string {
	unifyingMap := make(map[string]bool)
	for _, item := range a {
		unifyingMap[item] = true
	}
	for _, item := range b {
		unifyingMap[item] = true
	}

	merged := make([]string, len(unifyingMap))
	k := 0
	for item := range unifyingMap {
		merged[k] = item
		k++
	}

	return merged
}

func substractLists(a []string, b []string) []string {
	unifyingMap := make(map[string]bool)
	for _, item := range a {
		unifyingMap[item] = true
	}
	for _, item := range b {
		delete(unifyingMap, item)
	}

	merged := make([]string, len(unifyingMap))
	k := 0
	for item := range unifyingMap {
		merged[k] = item
		k++
	}

	return merged
}
