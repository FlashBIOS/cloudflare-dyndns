package cloudflare

import (
	"bytes"
	"cloudflare-dyndns/config"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	cfg    *config.Config
	Client *http.Client
}

func New(cfg *config.Config) *Client {
	return &Client{
		cfg:    cfg,
		Client: &http.Client{},
	}
}

func joinUri(base, pathStr string) (string, error) {
	baseUrl, err := url.Parse(base)
	if err != nil {
		return "", errors.New("unable to parse base url")
	}

	// Ensure base URL ends with a slash
	if baseUrl.Path == "" || baseUrl.Path[len(baseUrl.Path)-1] != '/' {
		baseUrl.Path += "/"
	}

	// Concatenate base URL and path
	fullUrl := baseUrl.String() + pathStr

	return fullUrl, nil
}

func (c *Client) request(ctx context.Context, method, endpoint string, data interface{}) ([]byte, error) {
	apiUrl, err := joinUri(c.cfg.BaseURL, endpoint)
	if err != nil {
		return nil, err
	}

	// Encode the body if needed.
	var body *bytes.Reader
	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(jsonData)
	} else {
		body = bytes.NewReader(nil)
	}

	// Create the HTTP request.
	req, err := http.NewRequestWithContext(ctx, method, apiUrl, body)
	if err != nil {
		return nil, err
	}

	// Set headers.
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.cfg.APIToken))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("User-Agent", c.cfg.UserAgent)
	req.Header.Add("Accept", "*/*")

	// Make the request.
	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	// Read the response.
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return respBody, nil
	}

	return respBody, nil
}

func (c *Client) GetDnsRecords() ([]DnsRecord, []ResponseErrors, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := c.request(ctx, "GET", "/zones/"+c.cfg.ZoneID+"/dns_records", nil)
	if err != nil {
		return nil, nil, err
	}

	dnsRecordsResp, err := unmarshalDnsRecordsResponse(response)
	if err != nil {
		return nil, nil, err
	}

	if !dnsRecordsResp.Success {
		return nil, dnsRecordsResp.Errors, errors.New("")
	}

	return dnsRecordsResp.Result, nil, nil
}

func (c *Client) UpdateDnsRecord(record DnsRecord) ([]ResponseErrors, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := c.request(ctx, "PUT", fmt.Sprintf("/zones/%s/dns_records/%s", c.cfg.ZoneID, record.ID), record)
	if err != nil {
		return nil, err
	}

	dnsRecordsResp, err := unmarshalDnsRecordsResponse(response)
	if err != nil {
		return nil, err
	}

	if !dnsRecordsResp.Success {
		return dnsRecordsResp.Errors, errors.New("")
	}

	return nil, nil
}

func (c *Client) ListDnsRecords() ([]DnsRecord, []ResponseErrors, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := c.request(ctx, "GET", fmt.Sprintf("/zones/%s/dns_records", c.cfg.ZoneID), nil)
	if err != nil {
		return nil, nil, err
	}

	dnsRecordsResp, err := unmarshalDnsRecordsResponse(response)
	if err != nil {
		return nil, nil, err
	}

	if !dnsRecordsResp.Success {
		return nil, dnsRecordsResp.Errors, errors.New("")
	}

	return dnsRecordsResp.Result, nil, nil
}

func unmarshalDnsRecordsResponse(response []byte) (DnsRecordsResponse, error) {
	var dnsRecordsResp DnsRecordsResponse

	if err := json.Unmarshal(response, &dnsRecordsResp); err != nil {
		return DnsRecordsResponse{}, err
	}

	return dnsRecordsResp, nil
}
