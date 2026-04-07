package s3

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client is a minimal S3-compatible storage client.
// It works with AWS S3 and S3-compatible services like MinIO.
type Client struct {
	endpoint  string
	bucket    string
	accessKey string
	secretKey string
	region    string
	httpClient *http.Client
}

func New(endpoint, bucket, accessKey, secretKey, region string) *Client {
	return &Client{
		endpoint:   strings.TrimRight(endpoint, "/"),
		bucket:     bucket,
		accessKey:  accessKey,
		secretKey:  secretKey,
		region:     region,
		httpClient: &http.Client{Timeout: 120 * time.Second},
	}
}

func (c *Client) objectURL(key string) string {
	return fmt.Sprintf("%s/%s/%s", c.endpoint, c.bucket, key)
}

func (c *Client) Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, c.objectURL(key), r)
	if err != nil {
		return err
	}
	req.ContentLength = size
	req.Header.Set("Content-Type", contentType)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("s3 put error: status %d", resp.StatusCode)
	}
	return nil
}

func (c *Client) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.objectURL(key), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		resp.Body.Close()
		return nil, fmt.Errorf("s3 get error: status %d", resp.StatusCode)
	}
	return resp.Body, nil
}

func (c *Client) Delete(ctx context.Context, key string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.objectURL(key), nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("s3 delete error: status %d", resp.StatusCode)
	}
	return nil
}

func (c *Client) URL(ctx context.Context, key string) (string, error) {
	u := fmt.Sprintf("%s/%s/%s", c.endpoint, c.bucket, url.PathEscape(key))
	return u, nil
}
