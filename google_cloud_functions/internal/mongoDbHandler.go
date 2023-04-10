package internal

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IpTypeToFetch string

const (
	IPV4 IpTypeToFetch = "ipv4"
	IPV6 IpTypeToFetch = "ipv6"
	ALL  IpTypeToFetch = "all"
)

type MongoDbHandler struct {
	dnsCollection *mongo.Collection
}

func NewMongoDBHandler(collection *mongo.Collection) *MongoDbHandler {
	return &MongoDbHandler{dnsCollection: collection}
}

func (dbHandler *MongoDbHandler) handleQuery(domainNames []string, fetch IpTypeToFetch) ([]MongoDnsRecord, error) {
	filter, projection := getMongoDBQuery(domainNames, fetch)
	_ = options.Find().SetProjection(projection)
	cur, err := dbHandler.dnsCollection.Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}
	var results []MongoDnsRecord
	err = cur.All(context.TODO(), &results)
	if len(results) == 0 {
		return nil, mongo.ErrNoDocuments
	}
	return results, err
}

func getMongoDBQuery(domainNames []string, ipType IpTypeToFetch) (bson.D, bson.D) {
	includeV4, includeV6 := DeterminateQueryFields(ipType)
	projectBson := bson.D{{"$project", bson.D{
		{"ipv4", includeV4},
		{"ipv6", includeV6},
	}}}
	filterBson := bson.D{{"domain_name", bson.D{{"$in", domainNames}}}}
	return filterBson, projectBson
}

func DeterminateQueryFields(fetch IpTypeToFetch) (int8, int8) {
	switch fetch {
	case IPV4:
		return 1, 0
	case IPV6:
		return 0, 1
	case ALL:
		return 1, 1
	default:
		return 1, 1
	}
}

type MongoDnsRecord struct {
	ID         primitive.ObjectID `bson:"_id"`
	DomainName string             `bson:"domain_name"`
	IPv4       []string           `bson:"ipv4"`
	IPv6       []string           `bson:"ipv6"`
}
