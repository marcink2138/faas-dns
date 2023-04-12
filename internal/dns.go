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
	fmt.Println("Creating external dns query handler")
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

func (handler *DNSHandler) Query(ctx context.Context, domainName string, recordType RecordType) ([]string, error) {
	msg := new(dns.Msg)
	switch recordType {
	case IPV4:
		msg.SetQuestion(dns.Fqdn(domainName), dns.TypeA)
	case IPV6:
		msg.SetQuestion(dns.Fqdn(domainName), dns.TypeAAAA)
	default:
		msg.SetQuestion(dns.Fqdn(domainName), dns.TypeA)
	}
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
		fmt.Printf("Error recived during external query call %e\n", err)
		return nil, err
	}
	res := make([]string, 0)
	for _, rr := range response.Answer {
		res = append(res, rr.String())
	}
	return res, nil
}

func HandleQueryViaExternalServer(ctx context.Context, domainName string, recordType RecordType, handler *DNSHandler) ([]string, error) {
	return handler.Query(ctx, domainName, recordType)
}
