package logfe

import (
	"context"
	"errors"
	"fmt"

	"github.com/ssgreg/logf/logfc"
)

// New returns a new error with fields inside it got from logger from the provided context.
func New(ctx context.Context, text string) error {
	return logfc.Get(ctx).WrapError(errors.New(text))
}

// Errorf returns a new error with fields inside it got from logger from the provided context.
func Errorf(ctx context.Context, text string, args ...interface{}) error {
	return logfc.Get(ctx).WrapError(fmt.Errorf(text, args...))
}
