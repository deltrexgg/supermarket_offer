// Package storage keeps uploaded files (offer photos, later posters) in Supabase
// Storage through its S3-compatible API. Handlers depend on the Store interface.
package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const MaxImageBytes = 5 << 20 // 5 MB

var (
	ErrTooLarge    = errors.New("image is larger than 5 MB")
	ErrUnsupported = errors.New("image must be JPEG, PNG or WebP")
)

type Store interface {
	Put(ctx context.Context, key, contentType string, data []byte) error
	Delete(ctx context.Context, key string) error
	// URL is the public address of a stored key.
	URL(key string) string
}

// Image is an upload that passed validation.
type Image struct {
	Data        []byte
	ContentType string
	Ext         string
}

var allowed = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

// ReadImage reads an uploaded file and checks its real type from the file
// signature (the first bytes), never from the file name or the client's header.
func ReadImage(fh *multipart.FileHeader) (*Image, error) {
	if fh.Size > MaxImageBytes {
		return nil, ErrTooLarge
	}
	f, err := fh.Open()
	if err != nil {
		return nil, err
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, MaxImageBytes+1))
	if err != nil {
		return nil, err
	}
	if len(data) > MaxImageBytes {
		return nil, ErrTooLarge
	}
	ct := http.DetectContentType(data)
	ext, ok := allowed[ct]
	if !ok {
		return nil, ErrUnsupported
	}
	return &Image{Data: data, ContentType: ct, Ext: ext}, nil
}

type S3 struct {
	client  *s3.Client
	bucket  string
	baseURL string
}

func NewS3(endpoint, region, bucket, accessKey, secretKey, publicBaseURL string) (*S3, error) {
	if endpoint == "" || region == "" || bucket == "" || accessKey == "" || secretKey == "" || publicBaseURL == "" {
		return nil, errors.New("S3_ENDPOINT, S3_REGION, S3_BUCKET, S3_ACCESS_KEY_ID, S3_SECRET_ACCESS_KEY and PUBLIC_FILE_BASE_URL must all be set")
	}
	client := s3.New(s3.Options{
		Region:       region,
		BaseEndpoint: aws.String(endpoint),
		Credentials:  credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""),
		UsePathStyle: true,
		// Supabase does not accept the newer default checksum headers.
		RequestChecksumCalculation: aws.RequestChecksumCalculationWhenRequired,
		ResponseChecksumValidation: aws.ResponseChecksumValidationWhenRequired,
	})
	return &S3{client: client, bucket: bucket, baseURL: strings.TrimRight(publicBaseURL, "/")}, nil
}

func (s *S3) Put(ctx context.Context, key, contentType string, data []byte) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(key),
		Body:          bytes.NewReader(data),
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(int64(len(data))),
		CacheControl:  aws.String("public, max-age=31536000, immutable"),
	})
	if err != nil {
		return fmt.Errorf("storage put %s: %w", key, err)
	}
	return nil
}

func (s *S3) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("storage delete %s: %w", key, err)
	}
	return nil
}

func (s *S3) URL(key string) string { return s.baseURL + "/" + key }
