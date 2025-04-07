package ipify

import (
	"cloudflare-dyndns/config"
	"cloudflare-dyndns/constants"
	"context"
	"errors"
	"github.com/jpillora/backoff"
	"io"
	"net/http"
	"strconv"
	"time"
)

type Client struct {
	config config.Config
}

func New(config *config.Config) *Client {
	return &Client{
		config: *config,
	}
}

func (ip *Client) GetPublicIP() (string, error) {
	result, err := ip.makeRequest(ip.config.IpifyURL)
	if err != nil {
		return "", nil
	}

	return result, nil
}

func (ip *Client) makeRequest(url string) (string, error) {
	b := &backoff.Backoff{
		Jitter: true,
	}

	// Create an HTTP client without a global timeout since we'll set
	// per-request timeouts via context.
	client := &http.Client{}

	for tries := 0; tries < constants.MaxTries; tries++ {
		// Create a context with a timeout for each request attempt.
		ctx, cancel := context.WithTimeout(context.Background(), constants.MaxTries*time.Second)
		// Ensure that the cancel function is called to free resources.
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			cancel()
			return "", errors.New("unable to make a new request to get ip address")
		}
		req.Header.Add("User-Agent", ip.config.UserAgent)

		resp, err := client.Do(req)
		// Cancel the context once the request has completed.
		cancel()
		if err != nil {
			// Sleep for a backoff duration before the next attempt.
			time.Sleep(b.Duration())
			continue
		}

		result, err := func() (string, error) {
			defer func(Body io.ReadCloser) {
				_ = Body.Close()
			}(resp.Body)
			data, err := io.ReadAll(resp.Body)
			if err != nil {
				return "", errors.New("unable to read response to get ip address")
			}

			if resp.StatusCode != http.StatusOK {
				return "", errors.New("received an invalid status code when getting ip address: " + strconv.Itoa(resp.StatusCode))
			}

			return string(data), nil
		}()
		if err == nil {
			return result, nil
		}

		// Sleep before the next retry.
		time.Sleep(b.Duration())
	}

	return "", errors.New("unable to get ip address")
}
