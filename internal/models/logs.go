package models

import (
	"time"
)

type LogRecord struct {
	Time   time.Time      `json:"time"`
	Level  string         `json:"level"`
	Msg    string         `json:"msg"`
	Fields map[string]any `json:"fields"`
}
