package internal

import (
	"net/http"
)

const (
	DomainParam     = "domain"
	RecordTypeParam = "recordType"
)

type RecordType string

const (
	IPV4 RecordType = "ipv4"
	IPV6 RecordType = "ipv6"
)

func String(recordType RecordType) string {
	switch recordType {
	case IPV4:
		return "ipv4"
	case IPV6:
		return "ipv6"
	default:
		return "ipv4"
	}
}

func RecordTypeFromString(recordTypeStr string) RecordType {
	switch recordTypeStr {
	case "ipv4":
		return IPV4
	case "ipv6":
		return IPV6
	default:
		return IPV4
	}
}

func ExtractQueryParams(r *http.Request) (string, RecordType) {
	domain := r.URL.Query().Get(DomainParam)
	recordType := r.URL.Query().Get(RecordTypeParam)
	return domain, RecordTypeFromString(recordType)
}
