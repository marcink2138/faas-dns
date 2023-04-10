package internal

type DNSResponse struct {
	IP      []string `json:"ip"`
	Message string   `json:"message"`
}
type ErrorResponse struct {
	Message string `json:"message"`
}
