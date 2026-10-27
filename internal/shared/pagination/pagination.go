package pagination

import (
	"fmt"
	"net/url"
	"strconv"
)

const (
	defaultPageSize = 20
	maxPageSize     = 100
)

type Request struct {
	Page     int
	PageSize int
}

type Result[T any] struct {
	Items    []T   `json:"items"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
	Total    int64 `json:"total"`
}

func Parse(values url.Values) (Request, error) {
	page, err := positiveInt(values.Get("page"), 1)
	if err != nil {
		return Request{}, fmt.Errorf("page: %w", err)
	}
	pageSize, err := positiveInt(values.Get("page_size"), defaultPageSize)
	if err != nil {
		return Request{}, fmt.Errorf("page_size: %w", err)
	}
	if pageSize > maxPageSize {
		return Request{}, fmt.Errorf("page_size must not exceed %d", maxPageSize)
	}
	return Request{Page: page, PageSize: pageSize}, nil
}

func (r Request) Offset() int {
	return (r.Page - 1) * r.PageSize
}

func positiveInt(raw string, fallback int) (int, error) {
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 {
		return 0, fmt.Errorf("must be a positive integer")
	}
	return value, nil
}
