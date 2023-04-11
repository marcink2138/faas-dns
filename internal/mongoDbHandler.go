package internal

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

var mongoDbConnectionString = "mongodb+srv://dnsFaas:dnsFaas@dns-faas-cluster.jhcnux0.mongodb.net/?retryWrites=true&w=majority"

type MongoDbHandler struct {
	dnsCollection *mongo.Collection
}

func NewMongoDBHandler(collection *mongo.Collection) *MongoDbHandler {
	return &MongoDbHandler{dnsCollection: collection}
}

func (dbHandler *MongoDbHandler) handleQuery(ctx context.Context, domainName string, fetch RecordType) ([]string, error) {
	filter, projection := getMongoDBQuery(domainName, fetch)
	_ = options.Find().SetProjection(projection)
	cur, err := dbHandler.dnsCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	var results []MongoDnsRecord
	err = cur.All(ctx, &results)
	if len(results) == 0 {
		return nil, mongo.ErrNoDocuments
	}
	finalRes := make([]string, 0)
	for _, result := range results {
		switch fetch {
		case IPV4:
			finalRes = append(finalRes, result.IPv4...)
		case IPV6:
			finalRes = append(finalRes, result.IPv6...)
		default:
			finalRes = append(finalRes, result.IPv4...)
		}
	}
	return finalRes, err
}

func getMongoDBQuery(domainName string, ipType RecordType) (bson.D, bson.D) {
	includeV4, includeV6 := determinateQueryFields(ipType)
	projectBson := bson.D{{"$project", bson.D{
		{"ipv4", includeV4},
		{"ipv6", includeV6},
		{"domain_name", 0},
		{"_id", 0},
	}}}
	filterBson := bson.D{{"domain_name", domainName}}
	return filterBson, projectBson
}

func determinateQueryFields(recordType RecordType) (int8, int8) {
	switch recordType {
	case IPV4:
		return 1, 0
	case IPV6:
		return 0, 1
	default:
		return 1, 0
	}
}

type MongoDnsRecord struct {
	IPv4 []string `bson:"ipv4"`
	IPv6 []string `bson:"ipv6"`
}

func HandleDnsQueryLocally(ctx context.Context, domainName string, recordType RecordType, handler *MongoDbHandler) ([]string, error) {
	result, err := handler.handleQuery(ctx, domainName, recordType)
	return result, err
}

func CreateMongoCollectionConn() *mongo.Collection {
	client := createMongoDBClient()
	collection := client.Database("faas-dns").Collection("dnsRecords")
	return collection
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
