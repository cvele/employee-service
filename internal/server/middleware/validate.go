package middleware

import (
	"context"

	"buf.build/go/protovalidate"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"
	"google.golang.org/protobuf/proto"
)

// validator interface for messages that support validation
type validator interface {
	Validate() error
}

// ProtoValidate returns a Kratos middleware that validates requests using protovalidate
func ProtoValidate() middleware.Middleware {
	// Create protovalidate validator instance
	v, err := protovalidate.New()
	if err != nil {
		// Fall back to simple validation if protovalidate fails to initialize
		return func(handler middleware.Handler) middleware.Handler {
			return func(ctx context.Context, req interface{}) (interface{}, error) {
				if val, ok := req.(validator); ok {
					if err := val.Validate(); err != nil {
						return nil, errors.BadRequest("VALIDATOR", err.Error())
					}
				}
				return handler(ctx, req)
			}
		}
	}

	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			// Validate using protovalidate
			if msg, ok := req.(proto.Message); ok {
				if err := v.Validate(msg); err != nil {
					return nil, errors.BadRequest("VALIDATOR", err.Error())
				}
			}
			return handler(ctx, req)
		}
	}
}

