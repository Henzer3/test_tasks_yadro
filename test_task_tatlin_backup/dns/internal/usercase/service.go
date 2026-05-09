package usercase

import (
	"context"
	"errors"
	"log/slog"

	"test.task.dns/internal/entity"
)

type dnsManager interface {
	AddDNS(ip string) error
	RemoveDNS(ip string) error
	ListDNS() ([]string, error)
}

type dnsService struct {
	log *slog.Logger
	dns dnsManager
}

func NewService(log *slog.Logger, dns dnsManager) *dnsService {
	return &dnsService{
		log: log,
		dns: dns,
	}
}

func (s *dnsService) AddDNS(_ context.Context, dns string) error {
	if err := s.dns.AddDNS(dns); err != nil {
		if errors.Is(err, entity.ErrAlreadyExist) {
			return err
		}

		s.log.Error("cant add dns in usercase", "err", err)
		return err
	}

	return nil
}

func (s *dnsService) DeleteDNS(_ context.Context, dns string) error {
	if err := s.dns.RemoveDNS(dns); err != nil {
		if errors.Is(err, entity.ErrNotFoundDNS) {
			return err
		}

		s.log.Error("cant delete dns in usercase", "err", err)
		return err
	}

	return nil
}

func (s *dnsService) GetList(_ context.Context) ([]string, error) {
	res, err := s.dns.ListDNS()
	if err != nil {
		s.log.Error("cant get list of dns servers in usercase", "err", err)
		return nil, err
	}
	return res, nil
}
