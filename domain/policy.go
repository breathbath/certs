package domain

import (
	"context"
	"errors"
	"sync"
)

type DynamicHostPolicy struct {
	mu      sync.RWMutex
	domains map[string]struct{}
}

func NewDynamicHostPolicy() *DynamicHostPolicy {
	dhp := &DynamicHostPolicy{domains: make(map[string]struct{})}
	return dhp
}

func (dhp *DynamicHostPolicy) AllowHost(ctx context.Context, host string) error {
	dhp.mu.RLock()
	defer dhp.mu.RUnlock()
	if _, ok := dhp.domains[host]; ok {
		return nil
	}

	return errors.New("host is not permitted")
}

func (dhp *DynamicHostPolicy) AddDomain(domain string) {
	dhp.mu.Lock()
	defer dhp.mu.Unlock()
	dhp.domains[domain] = struct{}{}
}

func (dhp *DynamicHostPolicy) RemoveDomain(domain string) {
	dhp.mu.Lock()
	defer dhp.mu.Unlock()
	delete(dhp.domains, domain)
}
