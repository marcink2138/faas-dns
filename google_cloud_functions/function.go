package faas_dns

import (
	"context"
	"encoding/json"
	"example.com/faas-dns/internal"
	"fmt"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"net/http"
)

var dnsHandler *internal.DNSHandler
var mongoDBHandler *internal.MongoDbHandler

func init() {
	functions.HTTP("HandleDnsQuery", handleDnsQuery)
	dnsHandler = internal.NewDNSHandler()
	collectionConn := internal.CreateMongoCollectionConn()
	mongoDBHandler = internal.NewMongoDBHandler(collectionConn)
}

func handleDnsQuery(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Inside function!")
	switch r.Method {
	case "GET":
		fmt.Printf("Handling DNS query -> START")
		domainName, recordType := internal.ExtractQueryParams(r)
		res, err := internal.HandleDnsQueryLocally(context.TODO(), domainName, recordType, mongoDBHandler)
		if err == nil {
			fmt.Printf("Records retrived locally")
			createHttpSuccessResponse(w, res)
			return
		}
		fmt.Printf("Cannot retrive records locally. Sending query to external server")
		res, err = internal.HandleQueryViaExternalServer(context.TODO(), domainName, recordType, dnsHandler)
		if err == nil {
			fmt.Printf("Records retrived by external server")
			createHttpSuccessResponse(w, res)
			return
		}
		mess := fmt.Sprintf("Cannot retrive dns records failure reason = %e", err)
		createHttpFailureReason(w, mess)
	default:
		mess := fmt.Sprintf("HTTP method= %s is not supported. Try again with POST", r.Method)
		createHttpFailureReason(w, mess)
	}
}

func createHttpFailureReason(w http.ResponseWriter, mess string) {
	if err := json.NewEncoder(w).Encode(internal.ErrorResponse{Message: mess}); err != nil {
		w.WriteHeader(500)
		fmt.Println(err)
	}
}

func createHttpSuccessResponse(w http.ResponseWriter, domainNames []string) {
	if err := json.NewEncoder(w).Encode(internal.DNSResponse{
		IP:      domainNames,
		Message: "Success",
	}); err != nil {
		w.WriteHeader(500)
		fmt.Printf("Cannot send DNS response, err: %e\n", err)
	}
}
