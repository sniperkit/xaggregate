package service

import (
	"context"
	"errors"
)

// NotFound is used when the specified service is not found
type NotFound struct{}

// Cache is not implemented
func (nf *NotFound) Cache(ctx context.Context, provider interface{}) error {
	return errors.New("Cache method is not implemented yet.")
}

// Login is not implemented
func (nf *NotFound) Login(ctx context.Context) error {
	return errors.New("Login method is notimplemented yet.")
}
