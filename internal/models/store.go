package models

import (
	"context"
	"encoding/json"
	"net/url"
	"time"
)

type StoreInput struct {
	Name string
	Type string

	Params json.RawMessage
}

func (s *StoreInput) Normalize() {

}

type Store interface {
	PutURL(ctx context.Context, objectName string, expires time.Duration) (*url.URL, error)
	GetURL(ctx context.Context, objectName string, expires time.Duration) (*url.URL, error)
	DeleteObject(ctx context.Context, objectName string) error
	Exists(objectName string) bool
}
