package project

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path"
	"time"

	"github.com/johannesboyne/gofakes3"
	"github.com/johannesboyne/gofakes3/backend/s3afero"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/pupload/pupload/internal/models"
	"github.com/pupload/pupload/internal/stores/s3"
	"github.com/spf13/afero"
)

const ARTIFACT_PATH = "artifact"

func NewDummyS3(stores []models.StoreInput, returnChan chan []models.StoreInput) error {

	root, err := GetProjectRoot()
	if err != nil {
		returnChan <- []models.StoreInput{}
		return err
	}

	mount_path := path.Join(root, ARTIFACT_PATH)

	fs := afero.NewOsFs()
	fs = afero.NewBasePathFs(fs, mount_path)

	backend, err := s3afero.MultiBucket(fs)
	if err != nil {
		returnChan <- []models.StoreInput{}
		return err
	}

	faker := gofakes3.New(backend)

	ts := httptest.NewServer(faker.Server())

	u, err := url.Parse(ts.URL)
	if err != nil {
		returnChan <- []models.StoreInput{}
		return err
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

	minioClient, err := minio.New(u.Host, &minio.Options{
		Creds:     credentials.NewStaticV4("ACCESS_KEY", "SECRET_KEY", ""),
		Secure:    false,
		Transport: transport,
	})

	if err != nil {
		returnChan <- []models.StoreInput{}
		return err
	}

	res := make([]models.StoreInput, 0)

	for _, s := range stores {
		minioClient.MakeBucket(context.TODO(), s.Name, minio.MakeBucketOptions{})

		secure := false

		params := s3.S3StoreInput{
			Endpoint:   u.Host,
			BucketName: s.Name,
			AccessKey:  "ACCESS_KEY",
			SecretKey:  "SECRET_KEY",
			Secure:     &secure,
		}

		marshalParams, err := json.Marshal(params)
		if err != nil {
			returnChan <- []models.StoreInput{}
			return err
		}

		newInput := models.StoreInput{
			Name:   s.Name,
			Type:   "s3",
			Params: marshalParams,
		}

		res = append(res, newInput)
	}

	returnChan <- res
	return nil
}
