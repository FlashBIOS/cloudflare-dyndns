package cloudflare

type ResponseErrors struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
