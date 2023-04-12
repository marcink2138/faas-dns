package main

import (
	"context"
	"encoding/json"
	"example.com/faas-dns/internal"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

var dnsHandler *internal.DNSHandler
var mongoDBHandler *internal.MongoDbHandler

func init() {
	dnsHandler = internal.NewDNSHandler()
	collectionConn := internal.CreateMongoCollectionConn()
	mongoDBHandler = internal.NewMongoDBHandler(collectionConn)
}

func hello(r events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	switch r.HTTPMethod {
	case "GET":
		fmt.Printf("Handling DNS query -> START")
		domainName, recordType := extractParamsFromMap(r.QueryStringParameters)
		res, err := internal.HandleDnsQueryLocally(context.TODO(), domainName, recordType, mongoDBHandler)
		if err == nil {
			fmt.Printf("Records retrived locally")
			response, err := createHttpSuccessResponse(res)
			if err != nil {
				fmt.Printf("Error occured during successful response creation %e. Next step -> External call", err)
			} else {
				return response, nil
			}
		}
		fmt.Printf("Cannot retrive records locally. Sending query to external server")
		res, err = internal.HandleQueryViaExternalServer(context.TODO(), domainName, recordType, dnsHandler)
		if err == nil {
			fmt.Printf("Records retrived by external server")
			response, err := createHttpSuccessResponse(res)
			if err != nil {
				fmt.Printf("Error occured during successful response creation %e. Next step -> failure response", err)
				return createHttpFailureReason(err.Error()), err
			}
			return response, nil
		}
		mess := fmt.Sprintf("Cannot retrive dns records failure reason = %e", err)
		return createHttpFailureReason(mess), nil
	default:
		mess := fmt.Sprintf("HTTP method= %s is not supported. Try again with GET", r.HTTPMethod)
		return createHttpFailureReason(mess), nil
	}
}

func main() {
	// Make the handler available for Remote Procedure Call by AWS Lambda
	lambda.Start(hello)
}

func extractParamsFromMap(params map[string]string) (string, internal.RecordType) {
	return params[internal.DomainParam], internal.RecordTypeFromString(params[internal.RecordTypeParam])
}

func createHttpSuccessResponse(res []string) (*events.APIGatewayProxyResponse, error) {
	resp := events.APIGatewayProxyResponse{Headers: map[string]string{"Content-Type": "application/json"}}
	resp.StatusCode = 200

	stringBody, err := json.Marshal(internal.DNSResponse{
		IP:      res,
		Message: "success",
	})
	resp.Body = string(stringBody)
	return &resp, err
}

func createHttpFailureReason(mess string) *events.APIGatewayProxyResponse {
	resp := events.APIGatewayProxyResponse{}
	resp.StatusCode = 500
	stringBody, _ := json.Marshal(internal.ErrorResponse{Message: mess})
	resp.Body = string(stringBody)
	return &resp
}
