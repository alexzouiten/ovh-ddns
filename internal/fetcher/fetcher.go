package fetcher

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"

	"github.com/alexzouiten/ovh-ddns/internal/model"
	"golang.org/x/sync/errgroup"
)

type client struct {
	http.Client
	transport string
}

type Fetcher struct {
	providerUrl string
	v4Client    *client
	v6Client    *client
}

func NewFetcher() *Fetcher {
	return &Fetcher{
		v4Client: newTransportSpecificClient("tcp4"),
		v6Client: newTransportSpecificClient("tcp6"),
	}
}

func (i *Fetcher) FetchIps(parentCtx context.Context, providerUrl string) (model.Ips, error) {
	var ips model.Ips

	eg, ctx := errgroup.WithContext(parentCtx)
	eg.Go(func() error { return i.v4Client.fetchIp(&ips.V4, ctx, providerUrl) })
	eg.Go(func() error { return i.v6Client.fetchIp(&ips.V6, ctx, providerUrl) })
	if err := eg.Wait(); err != nil {
		return model.Ips{}, err
	}

	return ips, nil
}

func newTransportSpecificClient(transport string) *client {
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.DialContext = func(ctx context.Context, network string, addr string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, transport, addr)
	}
	c := *http.DefaultClient
	c.Transport = t
	return &client{
		Client:    c,
		transport: transport,
	}
}

func (c *client) fetchIp(ip *net.IP, ctx context.Context, providerUrl string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", providerUrl, nil)
	if err != nil {
		return fmt.Errorf("ip fetch(%s): failed to create request: %w", c.transport, err)
	}

	res, err := c.Do(req)
	if err != nil {
		return fmt.Errorf("ip fetch(%s): failed to get ip: %w", c.transport, err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("ip fetch(%s): failed to read body: %w", c.transport, err)
	}

	if res.StatusCode != 200 {
		return fmt.Errorf("ip fetch(%s): got status code %d. body: '%s'", c.transport, res.StatusCode, body)
	}

	*ip = net.ParseIP(string(body))
	if *ip == nil {
		return fmt.Errorf("ip fetch(%s): '%s' is not a valid IP", c.transport, *ip)
	}

	return nil
}
