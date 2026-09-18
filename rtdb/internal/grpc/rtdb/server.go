package rtdb

import (
	"context"
	"errors"
	"log/slog"

	rtdbv1 "github.com/kirill3466/protos/gen/go/rtdb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"rtdb/internal/models"
	rtdbservice "rtdb/internal/rtdb"
	"rtdb/internal/slogx"
)

type RTDB interface {
	GetTag(ctx context.Context, name string) (models.TagValue, error)
	SetTag(ctx context.Context, name string, value float64) (models.TagValue, error)
	Subscribe(ctx context.Context, names []string, emit func(models.TagValue) error) error
}

type serverAPI struct {
	rtdbv1.UnimplementedRTDBServer
	log  *slog.Logger
	rtdb RTDB
}

func Register(gRPC *grpc.Server, log *slog.Logger, rtdb RTDB) {
	rtdbv1.RegisterRTDBServer(gRPC, &serverAPI{
		log:  log,
		rtdb: rtdb,
	})
}

func (s *serverAPI) GetTag(
	ctx context.Context,
	req *rtdbv1.GetTagRequest,
) (*rtdbv1.TagValue, error) {
	value, err := s.rtdb.GetTag(ctx, req.GetName())
	if err != nil {
		return nil, s.grpcError("GetTag", err)
	}

	return toProto(value), nil
}

func (s *serverAPI) SetTag(
	ctx context.Context,
	req *rtdbv1.SetTagRequest,
) (*rtdbv1.TagValue, error) {
	value, err := s.rtdb.SetTag(ctx, req.GetName(), req.GetValue())
	if err != nil {
		return nil, s.grpcError("SetTag", err)
	}

	return toProto(value), nil
}

func (s *serverAPI) Subscribe(
	req *rtdbv1.SubscribeRequest,
	stream rtdbv1.RTDB_SubscribeServer,
) error {
	err := s.rtdb.Subscribe(stream.Context(), req.GetNames(), func(value models.TagValue) error {
		return stream.Send(toProto(value))
	})
	if err != nil {
		return s.grpcError("Subscribe", err)
	}

	return nil
}

func toProto(value models.TagValue) *rtdbv1.TagValue {
	return &rtdbv1.TagValue{
		Name:     value.Name,
		Value:    value.Value,
		Quality:  rtdbv1.Quality(value.Quality),
		TsUnixMs: value.TsUnixMs,
	}
}

func (s *serverAPI) grpcError(method string, err error) error {
	switch {
	case errors.Is(err, rtdbservice.ErrNotFound):
		return status.Error(codes.NotFound, "tag not found")
	case errors.Is(err, rtdbservice.ErrReadOnly):
		return status.Error(codes.InvalidArgument, "tag is not writable")
	case errors.Is(err, context.DeadlineExceeded):
		return status.Error(codes.DeadlineExceeded, "deadline exceeded")
	case errors.Is(err, context.Canceled):
		return status.Error(codes.Canceled, "canceled")
	default:
		s.log.Error("handler failed", slog.String("method", method), slogx.Err(err))
		return status.Error(codes.Internal, "internal error")
	}
}
