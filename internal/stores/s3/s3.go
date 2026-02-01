package s3

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type S3Store struct {
	client   *minio.Client
	location string
	bucket   string
}

type S3StoreInput struct {
	Endpoint   string
	BucketName string
	Location   string
	AccessKey  string
	SecretKey  string

	Secure *bool
}

func NewS3Store(input S3StoreInput) (*S3Store, error) {

	if input.Secure == nil {
		input.Secure = new(bool)
		*input.Secure = true
	}

	transport := &http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		MaxIdleConns:        200,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 5 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false,
		},
	}

	minioClient, err := minio.New(input.Endpoint, &minio.Options{
		Creds:     credentials.NewStaticV4(input.AccessKey, input.SecretKey, ""),
		Secure:    *input.Secure,
		Transport: transport,
	})

	if err != nil {
		return nil, err
	}

	return &S3Store{
		client:   minioClient,
		location: input.Location,
		bucket:   input.BucketName,
	}, nil
}

/*
func (s *S3Store) PutFromFile(ctx context.Context, objectName, filePath, content_type string) (*minio.UploadInfo, error) {
	info, err := s.client.FPutObject(ctx, s.bucket, objectName, filePath, minio.PutObjectOptions{
		ContentType: content_type,
	})

	if err != nil {
		return nil, err
	}

	return &info, nil
}

func (s *S3Store) PutFromStream(ctx context.Context, objectName string, data io.Reader) (*minio.UploadInfo, error) {
	info, err := s.client.PutObject(ctx, s.bucket, objectName, data, -1, minio.PutObjectOptions{
		ContentType: "image",
	})

	if err != nil {
		return nil, err
	}

	return &info, nil
}
*/

func (s *S3Store) PutURL(ctx context.Context, objectName string, expires time.Duration) (u *url.URL, err error) {
	u, err = s.client.PresignedPutObject(ctx, s.bucket, objectName, expires)
	if err != nil {
		return nil, fmt.Errorf("s3 store at %q (bucket %q): %w", s.client.EndpointURL().Host, s.bucket, err)
	}
	return u, nil
}

func (s *S3Store) GetURL(ctx context.Context, objectName string, expires time.Duration) (u *url.URL, err error) {
	u, err = s.client.PresignedGetObject(ctx, s.bucket, objectName, expires, url.Values{})
	if err != nil {
		return nil, fmt.Errorf("s3 store at %q (bucket %q): %w", s.client.EndpointURL().Host, s.bucket, err)
	}
	return u, nil
}

func (s *S3Store) DeleteObject(ctx context.Context, objectName string) error {
	return s.client.RemoveObject(ctx, s.bucket, objectName, minio.RemoveObjectOptions{})
}

func (s *S3Store) Close() {

}

func (s *S3Store) Exists(objectName string) bool {
	_, err := s.client.StatObject(context.TODO(), s.bucket, objectName, minio.GetObjectOptions{})
	if err != nil {
		return false
	}

	return true
}
