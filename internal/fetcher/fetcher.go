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

const PROVIDER = "https://ifconfig.me/ip"

type client struct {
	http.Client
	transport string
}

type Fetcher struct {
	v4Client *client
	v6Client *client
}

func New() *Fetcher {
	return &Fetcher{
		v4Client: newTransportSpecificClient("tcp4"),
		v6Client: newTransportSpecificClient("tcp6"),
	}
}

func (i *Fetcher) FetchIps(parentCtx context.Context) (model.Ips, error) {
	var ips model.Ips

	eg, ctx := errgroup.WithContext(parentCtx)
	eg.Go(func() error { return fetchIp(&ips.V4, ctx, i.v4Client) })
	eg.Go(func() error { return fetchIp(&ips.V6, ctx, i.v6Client) })
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

func fetchIp(ip *net.IP, ctx context.Context, c *client) error {
	req, err := http.NewRequestWithContext(ctx, "GET", PROVIDER, nil)
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
