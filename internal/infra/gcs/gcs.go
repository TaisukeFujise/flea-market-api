package gcs

import (
	"context"
	"fmt"
	"io"
	"os"

	"cloud.google.com/go/storage"
)

type Client struct {
	raw        *storage.Client
	bucket     *storage.BucketHandle
	bucketName string
}

func NewClient(ctx context.Context) (*Client, error) {
	c, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	name := os.Getenv("GCS_BUCKET_NAME")
	return &Client{raw: c, bucket: c.Bucket(name), bucketName: name}, nil
}

func (c *Client) Close() error {
	return c.raw.Close()
}

func (c *Client) Delete(ctx context.Context, name string) error {
	return c.bucket.Object(name).Delete(ctx)
}

func (c *Client) Upload(ctx context.Context, name string, r io.Reader, contentType string) (string, error) {
	w := c.bucket.Object(name).NewWriter(ctx)
	w.ContentType = contentType
	if _, err := io.Copy(w, r); err != nil {
		_ = w.Close()
		return "", err
	}
	if err := w.Close(); err != nil {
		return "", err
	}
	return fmt.Sprintf("https://storage.googleapis.com/%s/%s", c.bucketName, name), nil
}
