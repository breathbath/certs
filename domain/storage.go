package domain

import (
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
	"log"
	"os"
	"sync"
)

type StorageType string

const StorageTypeInMemory StorageType = "in-memory"
const StorageTypeRedis StorageType = "redis"
const RedisKeyPrefix = "domain_storage_"

type Storage interface {
	HasHost(host string) (bool, error)
	Add(host, target string) error
	Remove(domain string) error
	Get(host string) (string, error)
}

type InMemoryStorage struct {
	data sync.Map
}

func NewStorage() (Storage, error) {
	storageType := StorageType(os.Getenv("DOMAIN_STORAGE_TYPE"))
	log.Printf("domain storage type from DOMAIN_STORAGE_TYPE is %s\n", storageType)
	switch storageType {
	case StorageTypeRedis:
		return NewRedisStorage()
	case StorageTypeInMemory:
		return NewInMemoryStorage(), nil
	default:
		log.Println("falling back to in-memory storage type as DOMAIN_STORAGE_TYPE is emtpy")
		return NewInMemoryStorage(), nil
	}
}

type RedisStorage struct {
	baseClient *redis.Client
}

func NewRedisStorage() (*RedisStorage, error) {
	redisUrl := os.Getenv("REDIS_URL")
	if redisUrl == "" {
		return nil, errors.New("REDIS_URL environment variable not set")
	}

	cl := redis.NewClient(&redis.Options{
		Addr:     ":6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	err := cl.Ping(context.Background()).Err()
	if err != nil {
		log.Printf("redis ping to %s failed: %v\n", redisUrl, err)
		return nil, err
	}

	log.Println("redis ping success to " + redisUrl)

	return &RedisStorage{baseClient: cl}, nil
}

func (s *RedisStorage) buildKey(host string) string {
	return RedisKeyPrefix + host
}

func (s *RedisStorage) HasHost(host string) (bool, error) {
	key := s.buildKey(host)
	_, err := s.baseClient.Get(context.Background(), key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			log.Printf("redis key %s does not exist\n", key)
			return false, nil
		}

		log.Printf("redis get key %s failed: %v\n", key, err)

		return false, err
	}

	return true, nil
}

func (s *RedisStorage) Add(host, target string) error {
	key := s.buildKey(host)
	res := s.baseClient.Set(context.Background(), key, target, 0)
	if res.Err() != nil {
		log.Printf("redis set key %s to %sfailed: %v\n", key, target, res.Err())
		return res.Err()
	}

	log.Printf("redis set key %s to %s success\n", key, target)
	return nil
}

func (s *RedisStorage) Remove(host string) error {
	key := s.buildKey(host)
	res := s.baseClient.Del(context.Background(), key)
	if res.Err() != nil {
		log.Printf("redis delete key %s failed: %v\n", key, res.Err())
		return res.Err()
	}

	log.Printf("deleted key %s from redis\n", key)
	return nil
}

func (s *RedisStorage) Get(host string) (string, error) {
	key := s.buildKey(host)
	val, err := s.baseClient.Get(context.Background(), key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			log.Printf("redis key %s does not exist\n", key)
			return "", nil
		}

		log.Printf("redis read key %s failed: %v\n", key, err)

		return "", err
	}

	return val, nil
}

func NewInMemoryStorage() *InMemoryStorage {
	st := &InMemoryStorage{data: sync.Map{}}

	appDomain := os.Getenv("APP_DOMAIN")
	if appDomain != "" {
		log.Printf("added app domain to the list of supported cetrificate domains: %s", appDomain)
		st.Add(appDomain, "")
	}

	return st
}

func (s *InMemoryStorage) HasHost(host string) (bool, error) {
	_, ok := s.data.Load(host)
	return ok, nil
}

func (s *InMemoryStorage) Add(host, target string) error {
	s.data.Store(host, target)
	return nil
}

func (s *InMemoryStorage) Remove(domain string) error {
	s.data.Delete(domain)

	return nil
}

func (s *InMemoryStorage) Get(host string) (string, error) {
	target, ok := s.data.Load(host)
	if ok {
		return target.(string), nil
	}

	return "", nil
}
