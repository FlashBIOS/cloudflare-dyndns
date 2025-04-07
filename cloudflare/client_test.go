package cloudflare

import (
	"bytes"
	"cloudflare-dyndns/config"
	"context"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"
)

func TestClient_GetDnsRecords(t *testing.T) {
	tests := []struct {
		name             string
		mockResponse     string
		mockStatusCode   int
		expectedError    bool
		expectedRecords  []DnsRecord
		expectedApiError []ResponseErrors
	}{
		{
			name: "successfulResponse",
			mockResponse: `{
				"success": true,
				"errors": [],
				"result": [
					{"id": "record1", "name": "example.com", "type": "A", "content": "1.2.3.4"},
					{"id": "record2", "name": "test.example.com", "type": "CNAME", "content": "example.com"}
				]
			}`,
			mockStatusCode:   http.StatusOK,
			expectedError:    false,
			expectedRecords:  []DnsRecord{{ID: "record1", Name: "example.com", Type: "A", IP: "1.2.3.4"}, {ID: "record2", Name: "test.example.com", Type: "CNAME", IP: "example.com"}},
			expectedApiError: nil,
		},
		{
			name: "apiErrorResponse",
			mockResponse: `{
				"success": false,
				"errors": [{"code": 1001, "message": "Invalid zone ID"}],
				"result": []
			}`,
			mockStatusCode:   http.StatusBadRequest,
			expectedError:    true,
			expectedRecords:  nil,
			expectedApiError: []ResponseErrors{{Code: 1001, Message: "Invalid zone ID"}},
		},
		{
			name:           "invalidJsonResponse",
			mockResponse:   `invalid-json`,
			mockStatusCode: http.StatusOK,
			expectedError:  true,
		},
		{
			name:           "networkError",
			mockResponse:   "",
			mockStatusCode: 0, // Simulates a network error
			expectedError:  true,
		},
		{
			name:             "emptyResponseSuccess",
			mockResponse:     `{"success": true, "errors": [], "result": []}`,
			mockStatusCode:   http.StatusOK,
			expectedError:    false,
			expectedRecords:  []DnsRecord{},
			expectedApiError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &http.Client{
				Transport: RoundTripFunc(func(req *http.Request) *http.Response {
					if tt.mockStatusCode == 0 {
						return nil // Simulates a network error
					}
					return &http.Response{
						StatusCode: tt.mockStatusCode,
						Body:       io.NopCloser(bytes.NewBufferString(tt.mockResponse)),
						Header:     make(http.Header),
					}
				}),
			}

			cfg := &config.Config{
				APIToken:  "mockToken",
				BaseURL:   "https://mockserver.com",
				UserAgent: "mockUserAgent",
				ZoneID:    "mockZoneID",
			}

			client := &Client{
				cfg:    cfg,
				Client: mockClient,
			}

			records, apiErrors, err := client.GetDnsRecords()

			if tt.expectedError {
				if err == nil {
					t.Errorf("expected an error, but got none")
				}
			} else {
				if err != nil {
					t.Errorf("did not expect an error, but got: %v", err)
				}
				if !compareDnsRecords(records, tt.expectedRecords) {
					t.Errorf("expected records %v, but got %v", tt.expectedRecords, records)
				}
				if !compareApiErrors(apiErrors, tt.expectedApiError) {
					t.Errorf("expected API errors %v, but got %v", tt.expectedApiError, apiErrors)
				}
			}
		})
	}
}

// Helper to compare slices of DnsRecords
func compareDnsRecords(a, b []DnsRecord) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// Helper to compare slices of ResponseErrors
func compareApiErrors(a, b []ResponseErrors) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestClient_Request(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		endpoint       string
		data           interface{}
		mockResponse   string
		mockStatusCode int
		expectedError  bool
		expectedOutput string
	}{
		{
			name:           "successfulGET",
			method:         http.MethodGet,
			endpoint:       "/valid-endpoint",
			data:           nil,
			mockResponse:   `{"key":"value"}`,
			mockStatusCode: http.StatusOK,
			expectedError:  false,
			expectedOutput: `{"key":"value"}`,
		},
		{
			name:           "successfulPOST",
			method:         http.MethodPost,
			endpoint:       "/valid-endpoint",
			data:           map[string]string{"name": "test"},
			mockResponse:   `{"success":true}`,
			mockStatusCode: http.StatusOK,
			expectedError:  false,
			expectedOutput: `{"success":true}`,
		},
		{
			name:           "httpError",
			method:         http.MethodGet,
			endpoint:       "/error-endpoint",
			data:           nil,
			mockResponse:   `{"error":"not_found"}`,
			mockStatusCode: http.StatusNotFound,
			expectedError:  false,
			expectedOutput: `{"error":"not_found"}`,
		},
		{
			name:           "invalidEndpoint",
			method:         http.MethodGet,
			endpoint:       "://invalid-endpoint",
			data:           nil,
			mockResponse:   "",
			mockStatusCode: 0,
			expectedError:  true,
			expectedOutput: "",
		},
		{
			name:           "jsonMarshalError",
			method:         http.MethodPost,
			endpoint:       "/valid-endpoint",
			data:           func() {}, // invalid data for JSON marshaling
			mockResponse:   "",
			mockStatusCode: 0,
			expectedError:  true,
			expectedOutput: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &http.Client{
				Transport: RoundTripFunc(func(req *http.Request) *http.Response {
					if tt.mockStatusCode == 0 {
						return nil // Simulates an invalid endpoint or other client issues
					}
					return &http.Response{
						StatusCode: tt.mockStatusCode,
						Body:       io.NopCloser(bytes.NewBufferString(tt.mockResponse)),
						Header:     make(http.Header),
					}
				}),
			}

			cfg := &config.Config{
				APIToken:  "mockToken",
				BaseURL:   "https://mockserver.com",
				UserAgent: "mockUserAgent",
			}

			client := &Client{
				cfg:    cfg,
				Client: mockClient,
			}

			output, err := client.request(context.Background(), tt.method, tt.endpoint, tt.data)

			if tt.expectedError {
				if err == nil {
					t.Errorf("expected an error, but got none")
				}
			} else {
				if err != nil {
					t.Errorf("did not expect an error, but got: %v", err)
				}
				if string(output) != tt.expectedOutput {
					t.Errorf("expected output %v, but got %v", tt.expectedOutput, string(output))
				}
			}
		})
	}
}

// RoundTripFunc is a utility to mock HTTP transport for testing.
type RoundTripFunc func(req *http.Request) *http.Response

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	resp := f(req)
	if resp == nil {
		return nil, errors.New("mock transport: no response provided")
	}
	return resp, nil
}

func TestJoinUri(t *testing.T) {
	tests := []struct {
		name    string
		base    string
		pathStr string
		want    string
		wantErr bool
	}{
		{
			name:    "standardBaseAndPath",
			base:    "https://example.com/",
			pathStr: "api/v1/resource",
			want:    "https://example.com/api/v1/resource",
			wantErr: false,
		},
		{
			name:    "baseWithoutTrailingSlash",
			base:    "https://example.com",
			pathStr: "api/v1/resource",
			want:    "https://example.com/api/v1/resource",
			wantErr: false,
		},
		{
			name:    "emptyPath",
			base:    "https://example.com/",
			pathStr: "",
			want:    "https://example.com/",
			wantErr: false,
		},
		{
			name:    "emptyBase",
			base:    "",
			pathStr: "api/v1/resource",
			want:    "/api/v1/resource",
			wantErr: false,
		},
		{
			name:    "invalidBase",
			base:    "://invalid_url",
			pathStr: "api/v1/resource",
			want:    "",
			wantErr: true,
		},
		{
			name:    "numericPath",
			base:    "https://example.com/",
			pathStr: "12345",
			want:    "https://example.com/12345",
			wantErr: false,
		},
		{
			name:    "queryParamsInBase",
			base:    "https://example.com/?key=value",
			pathStr: "api/v1/resource",
			want:    "https://example.com/?key=valueapi/v1/resource",
			wantErr: false,
		},
		{
			name:    "pathWithLeadingSlash",
			base:    "https://example.com/",
			pathStr: "/api/v1/resource",
			want:    "https://example.com//api/v1/resource",
			wantErr: false,
		},
		{
			name:    "whitespaceInBaseAndPath",
			base:    "  https://example.com  ",
			pathStr: "  api/v1/resource  ",
			want:    "",
			wantErr: true,
		},
		{
			name:    "longBaseAndPath",
			base:    "https://this.is.a.very.long.and.complex.url.with.many.components/",
			pathStr: "this/is/a/very/long/path/with/many/segments",
			want:    "https://this.is.a.very.long.and.complex.url.with.many.components/this/is/a/very/long/path/with/many/segments",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := joinUri(tt.base, tt.pathStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("joinUri() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("joinUri() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_UpdateDnsRecord(t *testing.T) {
	tests := []struct {
		name           string
		mockResponse   string
		mockStatusCode int
		record         DnsRecord
		expectedError  bool
		expectedApiErr []ResponseErrors
	}{
		{
			name: "successfulUpdate",
			mockResponse: `{
				"success": true,
				"errors": [],
				"result": {}
			}`,
			mockStatusCode: http.StatusOK,
			record: DnsRecord{
				ID:   "record1",
				Name: "example.com",
				Type: "A",
				IP:   "1.2.3.4",
			},
			expectedError:  false,
			expectedApiErr: nil,
		},
		{
			name: "apiErrorResponse",
			mockResponse: `{
				"success": false,
				"errors": [{"code": 1002, "message": "Unauthorized request"}],
				"result": {}
			}`,
			mockStatusCode: http.StatusUnauthorized,
			record: DnsRecord{
				ID:   "record2",
				Name: "example.com",
				Type: "A",
				IP:   "1.2.3.4",
			},
			expectedError:  true,
			expectedApiErr: []ResponseErrors{{Code: 1002, Message: "Unauthorized request"}},
		},
		{
			name:           "invalidJsonResponse",
			mockResponse:   `invalid-json`,
			mockStatusCode: http.StatusOK,
			record: DnsRecord{
				ID:   "record3",
				Name: "example.com",
				Type: "A",
				IP:   "1.2.3.4",
			},
			expectedError: true,
		},
		{
			name:           "httpError",
			mockResponse:   ``,
			mockStatusCode: http.StatusInternalServerError,
			record: DnsRecord{
				ID:   "record4",
				Name: "example.com",
				Type: "A",
				IP:   "1.2.3.4",
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &http.Client{
				Transport: RoundTripFunc(func(req *http.Request) *http.Response {
					return &http.Response{
						StatusCode: tt.mockStatusCode,
						Body:       io.NopCloser(bytes.NewBufferString(tt.mockResponse)),
						Header:     make(http.Header),
					}
				}),
			}

			cfg := &config.Config{
				APIToken:  "mockToken",
				BaseURL:   "https://mockserver.com",
				UserAgent: "mockUserAgent",
				ZoneID:    "mockZoneID",
			}

			client := &Client{
				cfg:    cfg,
				Client: mockClient,
			}

			if tt.name == "contextTimeout" {
				client.Client = &http.Client{
					Transport: RoundTripFunc(func(req *http.Request) *http.Response {
						time.Sleep(11 * time.Second)
						return nil
					}),
				}
			}

			apiErr, err := client.UpdateDnsRecord(tt.record)

			if tt.expectedError {
				if err == nil {
					t.Errorf("expected an error, but got none")
				}
			} else {
				if err != nil {
					t.Errorf("did not expect an error, but got: %v", err)
				}
			}

			if !compareApiErrors(apiErr, tt.expectedApiErr) {
				t.Errorf("expected API errors %v, but got %v", tt.expectedApiErr, apiErr)
			}
		})
	}
}

func TestClient_ListDnsRecords(t *testing.T) {
	tests := []struct {
		name             string
		mockResponse     string
		mockStatusCode   int
		expectedError    bool
		expectedRecords  []DnsRecord
		expectedApiError []ResponseErrors
	}{
		{
			name: "successfulResponse",
			mockResponse: `{
				"success": true,
				"errors": [],
				"result": [
					{"id": "record1", "name": "example.com", "type": "A", "content": "1.2.3.4"},
					{"id": "record2", "name": "test.example.com", "type": "CNAME", "content": "example.com"}
				]
			}`,
			mockStatusCode:   http.StatusOK,
			expectedError:    false,
			expectedRecords:  []DnsRecord{{ID: "record1", Name: "example.com", Type: "A", IP: "1.2.3.4"}, {ID: "record2", Name: "test.example.com", Type: "CNAME", IP: "example.com"}},
			expectedApiError: nil,
		},
		{
			name: "apiErrorResponse",
			mockResponse: `{
				"success": false,
				"errors": [{"code": 1001, "message": "Invalid zone ID"}],
				"result": []
			}`,
			mockStatusCode:   http.StatusBadRequest,
			expectedError:    true,
			expectedRecords:  nil,
			expectedApiError: []ResponseErrors{{Code: 1001, Message: "Invalid zone ID"}},
		},
		{
			name:           "invalidJsonResponse",
			mockResponse:   `invalid-json`,
			mockStatusCode: http.StatusOK,
			expectedError:  true,
		},
		{
			name:           "networkError",
			mockResponse:   "",
			mockStatusCode: 0, // Simulates a network error
			expectedError:  true,
		},
		{
			name:             "emptyResponseSuccess",
			mockResponse:     `{"success": true, "errors": [], "result": []}`,
			mockStatusCode:   http.StatusOK,
			expectedError:    false,
			expectedRecords:  []DnsRecord{},
			expectedApiError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &http.Client{
				Transport: RoundTripFunc(func(req *http.Request) *http.Response {
					if tt.mockStatusCode == 0 {
						return nil // Simulates a network error
					}
					return &http.Response{
						StatusCode: tt.mockStatusCode,
						Body:       io.NopCloser(bytes.NewBufferString(tt.mockResponse)),
						Header:     make(http.Header),
					}
				}),
			}

			cfg := &config.Config{
				APIToken:  "mockToken",
				BaseURL:   "https://mockserver.com",
				UserAgent: "mockUserAgent",
				ZoneID:    "mockZoneID",
			}

			client := &Client{
				cfg:    cfg,
				Client: mockClient,
			}

			records, apiErrors, err := client.ListDnsRecords()

			if tt.expectedError {
				if err == nil {
					t.Errorf("expected an error, but got none")
				}
			} else {
				if err != nil {
					t.Errorf("did not expect an error, but got: %v", err)
				}
				if !compareDnsRecords(records, tt.expectedRecords) {
					t.Errorf("expected records %v, but got %v", tt.expectedRecords, records)
				}
				if !compareApiErrors(apiErrors, tt.expectedApiError) {
					t.Errorf("expected API errors %v, but got %v", tt.expectedApiError, apiErrors)
				}
			}
		})
	}
}
