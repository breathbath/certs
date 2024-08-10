package domain

import (
	"context"
	"errors"
	"log"
)

type DynamicHostPolicy struct {
	storage *Storage
}

func NewDynamicHostPolicy(s *Storage) *DynamicHostPolicy {
	dhp := &DynamicHostPolicy{storage: s}
	return dhp
}

func (dhp *DynamicHostPolicy) AllowHost(ctx context.Context, host string) error {
	if dhp.storage.HasDomain(host) {
		return nil
	}
	log.Printf("failed to find target for host %s to issue a certificate\n", host)

	return errors.New("host is not permitted")
}
