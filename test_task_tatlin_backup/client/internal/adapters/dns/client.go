package dns

import (
	"context"
	"errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"test.task.dnsmanager/internal/entity"
	dnspb "test.task.pkg/generated/dns"
)

type Client struct {
	conn   *grpc.ClientConn
	client dnspb.DnsServiceClient
}

func NewClient(address string) (*Client, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &Client{
		conn:   conn,
		client: dnspb.NewDnsServiceClient(conn),
	}, nil

}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c Client) AddDNS(ctx context.Context, ip string) error {
	if _, err := c.client.AddDns(ctx, &dnspb.AddDnsRequest{Ip: ip}); err != nil {
		switch status.Code(err) {
		case codes.InvalidArgument:
			return entity.ErrInvalidArgument
		case codes.AlreadyExists:
			return entity.ErrAlreadyexist
		case codes.Unavailable:
			return entity.ErrUnAvailableServer
		case codes.Internal:
			return entity.ErrInternalError
		default:
			return errors.New("unknown error")
		}
	}
	return nil
}

func (c Client) DeleteDNS(ctx context.Context, ip string) error {
	if _, err := c.client.DeleteDns(ctx, &dnspb.DeleteDnsRequest{Ip: ip}); err != nil {
		switch status.Code(err) {
		case codes.InvalidArgument:
			return entity.ErrInvalidArgument
		case codes.NotFound:
			return entity.ErrNotFoundDNS
		case codes.Unavailable:
			return entity.ErrUnAvailableServer
		case codes.Internal:
			return entity.ErrInternalError
		default:
			return errors.New("unknown error")
		}
	}
	return nil
}

func (c Client) GetList(ctx context.Context) ([]string, error) {
	res, err := c.client.GetList(ctx, &dnspb.GetListRequest{})
	if err != nil {
		switch status.Code(err) {
		case codes.Unavailable:
			return nil, entity.ErrUnAvailableServer
		case codes.Internal:
			return nil, entity.ErrInternalError
		default:
			return nil, errors.New("unknown error")
		}
	}
	return res.Ips, nil
}
