package faas_dns

import (
	"context"
	"encoding/json"
	"example.com/faas-dns/internal"
	"fmt"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
	"time"
)

var mongoDbConnectionString = "mongodb+srv://dnsFaas:dnsFaas@faas-dns-db-cluster.bilqfgp.mongodb.net/?retryWrites=true&w=majority"
var dnsHandler *internal.DNSHandler
var mongoDBHandler *internal.MongoDbHandler

func init() {
	functions.HTTP("HandleDnsQuery", handleDnsQuery)
	dnsHandler = internal.NewDNSHandler()
	mongoClient := createMongoDBClient()
	dnsCollection := mongoClient.Database("faas-dns").Collection("dnsRecords")
	mongoDBHandler = internal.NewMongoDBHandler(dnsCollection)
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

func createMongoDBClient() *mongo.Client {
	serverAPIOptions := options.ServerAPI(options.ServerAPIVersion1)
	clientOptions := options.Client().
		ApplyURI(mongoDbConnectionString).
		SetServerAPIOptions(serverAPIOptions)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	return client
}
