package ipify

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"cloudflare-dyndns/config"
	"cloudflare-dyndns/constants"
)

func TestGetPublicIP(t *testing.T) {
	tests := []struct {
		name       string
		serverFunc http.HandlerFunc
		wantIP     string
		wantErr    bool
	}{
		{
			name: "success case",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("123.123.123.123"))
			},
			wantIP:  "123.123.123.123",
			wantErr: false,
		},
		{
			name: "server returns non-200 status",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			wantIP:  "",
			wantErr: true,
		},
		{
			name: "server returns malformed response",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("not-an-ip"))
			},
			wantIP:  "not-an-ip",
			wantErr: false,
		},
		{
			name: "network error",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(constants.MaxTries * 3 * time.Second)
			},
			wantIP:  "",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(tc.serverFunc)
			defer server.Close()

			client := &Client{
				config: config.Config{
					UserAgent: "TestAgent",
				},
			}

			ip, err := client.makeRequest(server.URL)
			gotErr := err != nil

			if gotErr != tc.wantErr {
				t.Errorf("expected error: %v, got: %v", tc.wantErr, gotErr)
			}
			if ip != tc.wantIP {
				t.Errorf("expected IP: %v, got: %v", tc.wantIP, ip)
			}
		})
	}
}
