package validate

import (
	"context"

	"buf.build/go/protovalidate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

func UnaryServerInterceptor(validator protovalidate.Validator) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		_ *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		msg, ok := req.(proto.Message)
		if ok {
			if err := validator.Validate(msg); err != nil {
				return nil, status.Error(codes.InvalidArgument, err.Error())
			}
		}

		return handler(ctx, req)
	}
}

func StreamServerInterceptor(validator protovalidate.Validator) grpc.StreamServerInterceptor {
	return func(
		srv any,
		ss grpc.ServerStream,
		_ *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		return handler(srv, &validatingStream{
			ServerStream: ss,
			validator:    validator,
		})
	}
}

type validatingStream struct {
	grpc.ServerStream
	validator protovalidate.Validator
	checked   bool
}

func (s *validatingStream) RecvMsg(m any) error {
	if err := s.ServerStream.RecvMsg(m); err != nil {
		return err
	}
	if s.checked {
		return nil
	}
	s.checked = true

	msg, ok := m.(proto.Message)
	if !ok {
		return nil
	}
	if err := s.validator.Validate(msg); err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	return nil
}
