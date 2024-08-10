package domain

import (
	"log"
	"os"
	"sync"
)

type Storage struct {
	data sync.Map
}

func NewStorage() *Storage {
	st := &Storage{data: sync.Map{}}

	appDomain := os.Getenv("APP_DOMAIN")
	if appDomain != "" {
		log.Printf("added app domain to the list of supported cetrificate domains: %s", appDomain)
		st.Add(appDomain, "")
	}

	return st
}

func (s *Storage) HasDomain(host string) bool {
	_, ok := s.data.Load(host)
	return ok
}

func (s *Storage) Add(host, target string) {
	s.data.Store(host, target)
}

func (s *Storage) Remove(domain string) {
	s.data.Delete(domain)
}

func (s *Storage) GetTarget(host string) string {
	target, ok := s.data.Load(host)
	if ok {
		return target.(string)
	}

	return ""
}
