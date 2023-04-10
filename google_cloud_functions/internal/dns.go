package internal

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/miekg/dns"
)

type DNSHandler struct {
	upstreams []string
	client    *dns.Client
}

func NewDNSHandler() *DNSHandler {
	timeout := 4

	handler := &DNSHandler{}

	if handler.upstreams == nil {
		handler.upstreams = []string{
			// cloudflare
			"1.0.0.1:53",
			"1.1.1.1:53",

			// google
			"8.8.4.4:53",
			"8.8.8.8:53",
		}
	}

	clientTimeout := time.Duration(timeout) * time.Second

	handler.client = &dns.Client{
		Net:     "udp",
		Timeout: clientTimeout,
		Dialer: &net.Dialer{
			Timeout:   clientTimeout,
			LocalAddr: nil,
		},
	}

	return handler
}

func (handler *DNSHandler) Query(ctx context.Context, binary string) ([]string, error) {
	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(binary), dns.TypeA)
	fmt.Printf("Inside query!")
	var response *dns.Msg
	var err error

	for _, upstream := range handler.upstreams {
		response, _, err = handler.client.Exchange(msg, upstream)
		if err != nil {
			fmt.Printf("Err %e\n", err)
			continue
		}

		if response.Answer != nil {
			break
		}
	}

	if err != nil {
		fmt.Printf("Err %e\n", err)
		return nil, err
	}
	res := make([]string, 0)
	for _, rr := range response.Answer {
		res = append(res, rr.String())
	}
	return res, nil
}
