package faas_dns

import (
	"encoding/json"
	"example.com/faas-dns/internal"
	"fmt"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"net/http"
)

var dnsHandler *internal.DNSHandler

func init() {
	functions.HTTP("HandleDnsQuery", handleDnsQuery)
	dnsHandler = internal.NewDNSHandler()
}

func handleDnsQuery(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Inside function!")
	switch r.Method {
	case "POST":
		fmt.Printf("Inside POST!")
		domainNames, err := handleGetDnsQuery(r)
		if err != nil {
			w.WriteHeader(500)
			fmt.Printf("Cannot handle DNS query, err: %e\n", err)
			return
		}
		if err := json.NewEncoder(w).Encode(internal.DNSResponse{
			IP:      domainNames,
			Message: "Success",
		}); err != nil {
			w.WriteHeader(500)
			fmt.Printf("Cannot send DNS response, err: %e\n", err)
			return
		}
	default:
		mess := fmt.Sprintf("HTTP method= %s is not supported. Try again with POST", r.Method)
		if err := json.NewEncoder(w).Encode(internal.ErrorResponse{Message: mess}); err != nil {
			w.WriteHeader(500)
			fmt.Println(err)
			return
		}
	}
}

func handleGetDnsQuery(r *http.Request) ([]string, error) {
	fmt.Printf("Inside handle dns query!")
	var dnsRequest internal.DNSRequest
	err := json.NewDecoder(r.Body).Decode(&dnsRequest)
	if err != nil {
		fmt.Printf("Cannot decode DNS request, err: %e\n", err)
		return nil, err
	}
	fmt.Printf("Domain name = %s \n", dnsRequest.DomainName)
	body, err := dnsHandler.Query(nil, dnsRequest.DomainName)
	if err != nil {
		fmt.Printf("Cannot receive DNS response, err: %e\n", err)
		return nil, err
	}
	fmt.Println(body)
	return body, nil
}
