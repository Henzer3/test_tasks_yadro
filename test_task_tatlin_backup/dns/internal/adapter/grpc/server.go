package grpc

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"test.task.dns/internal/entity"

	dnspb "test.task.pkg/generated/dns"
)

type dnsHandler interface {
	AddDNS(ctx context.Context, dns string) error
	DeleteDNS(ctx context.Context, dns string) error
	GetList(ctx context.Context) ([]string, error)
}

type server struct {
	dnspb.UnimplementedDnsServiceServer
	logger     *slog.Logger
	dnsHandler dnsHandler
}

func NewServer(log *slog.Logger, dnsHandler dnsHandler) *server {
	return &server{
		logger:     log,
		dnsHandler: dnsHandler,
	}
}

func (s *server) AddDns(ctx context.Context, in *dnspb.AddDnsRequest) (*dnspb.AddDnsResponse, error) {
	if err := validate(in.GetIp()); err != nil {
		return nil, status.Error(codes.InvalidArgument, "wrong dns")
	}
	if err := s.dnsHandler.AddDNS(ctx, in.GetIp()); err != nil {
		if errors.Is(err, entity.ErrAlreadyExist) {
			return nil, status.Error(codes.AlreadyExists, "already exist")
		}
		s.logger.Error("cant add dns in adapter", "err", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &dnspb.AddDnsResponse{}, nil
}

func (s *server) DeleteDns(ctx context.Context, in *dnspb.DeleteDnsRequest) (*dnspb.DeleteDnsResponse, error) {
	if err := validate(in.GetIp()); err != nil {
		return nil, status.Error(codes.InvalidArgument, "wrong dns")
	}

	if err := s.dnsHandler.DeleteDNS(ctx, in.GetIp()); err != nil {
		if errors.Is(err, entity.ErrNotFoundDNS) {
			return nil, status.Error(codes.NotFound, "cant found dns")
		}

		s.logger.Error("cant add dns in adapter", "err", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &dnspb.DeleteDnsResponse{}, nil
}

func (s *server) GetList(ctx context.Context, in *dnspb.GetListRequest) (*dnspb.GetListResponse, error) {
	dnses, err := s.dnsHandler.GetList(ctx)
	if err != nil {
		s.logger.Error("cant get list of dns in adapter", "err", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &dnspb.GetListResponse{Ips: dnses}, nil
}

var ErrWrongDns = errors.New("wrong dns")

func validate(ip string) error {
	ip = strings.TrimSpace(ip)

	if ip == "" {
		return ErrWrongDns
	}

	if net.ParseIP(ip) == nil {
		return ErrWrongDns
	}

	return nil
}
