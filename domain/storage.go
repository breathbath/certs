package domain

import (
	"errors"
	"go.mills.io/bitcask/v2"
	"log"
	"os"
)

type StorageType string

type Storage interface {
	HasHost(host string) (bool, error)
	Add(host, target string) error
	Remove(domain string) error
	Get(host string) (string, error)
	Close() error
}

type KVStorage struct {
	db *bitcask.Bitcask
}

func NewStorage() (Storage, error) {
	storagePath := os.Getenv("STORAGE_PATH")
	if storagePath == "" {
		storagePath = ".build"
	}
	db, err := bitcask.Open(storagePath)
	if err != nil {
		return nil, err
	}

	return &KVStorage{db: db}, nil
}

func (s *KVStorage) Close() error {
	return s.db.Close()
}

func (s *KVStorage) HasHost(host string) (bool, error) {
	_, err := s.db.Get([]byte(host))
	if err != nil {
		if errors.Is(err, bitcask.ErrKeyNotFound) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func (s *KVStorage) Add(host, target string) error {
	if err := s.db.Put([]byte(host), []byte(target)); err != nil {
		return err
	}

	log.Printf("put key %s to %s success\n", host, target)
	return nil
}

func (s *KVStorage) Remove(host string) error {
	if err := s.db.Delete([]byte(host)); err != nil {
		return err
	}

	log.Printf("deleted key %s from db\n", host)

	return nil
}

func (s *KVStorage) Get(host string) (string, error) {
	val, err := s.db.Get([]byte(host))
	if err != nil {
		if errors.Is(err, bitcask.ErrKeyNotFound) {
			return "", nil
		} else {
			return "", err
		}
	}

	return string(val), nil
}
