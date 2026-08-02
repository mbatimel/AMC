package objectstorage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"path"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Config describes an S3-compatible bucket. Endpoint is used by backend
// services, while PublicEndpoint is used to build browser-facing URLs.
type Config struct {
	Endpoint       string
	PublicEndpoint string
	AccessKey      string
	SecretKey      string
	Bucket         string
	Region         string
	UseSSL         bool
}

// Client is the shared S3-compatible implementation used by AMC services.
type Client struct {
	client         *minio.Client
	bucket         string
	publicEndpoint *url.URL
}

func New(cfg Config) (*Client, error) {
	endpoint, secure, err := parseEndpoint(cfg.Endpoint, cfg.UseSSL)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(cfg.Bucket) == "" {
		return nil, fmt.Errorf("s3 bucket is empty")
	}
	publicEndpoint, err := url.Parse(strings.TrimRight(cfg.PublicEndpoint, "/"))
	if err != nil || publicEndpoint.Scheme == "" || publicEndpoint.Host == "" {
		return nil, fmt.Errorf("invalid S3 public endpoint %q", cfg.PublicEndpoint)
	}
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: secure,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("create S3 client: %w", err)
	}
	return &Client{client: minioClient, bucket: cfg.Bucket, publicEndpoint: publicEndpoint}, nil
}

func parseEndpoint(raw string, useSSL bool) (string, bool, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", false, fmt.Errorf("S3 endpoint is empty")
	}
	if !strings.Contains(raw, "://") {
		return strings.TrimRight(raw, "/"), useSSL, nil
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return "", false, fmt.Errorf("invalid S3 endpoint %q", raw)
	}
	if parsed.Path != "" && parsed.Path != "/" {
		return "", false, fmt.Errorf("S3 endpoint must not contain a path")
	}
	return parsed.Host, parsed.Scheme == "https", nil
}

func (c *Client) Upload(ctx context.Context, objectKey string, body io.Reader, size int64, contentType string) error {
	_, err := c.client.PutObject(ctx, c.bucket, objectKey, body, size, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return fmt.Errorf("upload S3 object: %w", err)
	}
	return nil
}

// Delete is idempotent for S3-compatible storage: removing a missing key is
// treated as success by the protocol and by MinIO.
func (c *Client) Delete(ctx context.Context, objectKey string) error {
	if err := c.client.RemoveObject(ctx, c.bucket, objectKey, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("delete S3 object: %w", err)
	}
	return nil
}

func (c *Client) URL(objectKey string) string {
	result := *c.publicEndpoint
	result.Path = path.Join(result.Path, c.bucket, objectKey)
	return result.String()
}
