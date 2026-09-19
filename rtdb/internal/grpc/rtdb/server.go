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
	GetTag(ctx context.Context, name string) (models.Tag, error)
	SetTag(ctx context.Context, name string, value float64) (models.TagValue, error)
	ListTags(ctx context.Context) ([]models.Tag, error)
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
) (*rtdbv1.Tag, error) {
	tag, err := s.rtdb.GetTag(ctx, req.GetName())
	if err != nil {
		return nil, s.grpcError("GetTag", err)
	}

	return toProtoTag(tag), nil
}

func (s *serverAPI) SetTag(
	ctx context.Context,
	req *rtdbv1.SetTagRequest,
) (*rtdbv1.TagValue, error) {
	value, err := s.rtdb.SetTag(ctx, req.GetName(), req.GetValue())
	if err != nil {
		return nil, s.grpcError("SetTag", err)
	}

	return toProtoValue(value), nil
}

func (s *serverAPI) ListTags(
	ctx context.Context,
	_ *rtdbv1.ListTagsRequest,
) (*rtdbv1.ListTagsResponse, error) {
	tags, err := s.rtdb.ListTags(ctx)
	if err != nil {
		return nil, s.grpcError("ListTags", err)
	}

	out := make([]*rtdbv1.Tag, 0, len(tags))
	for _, tag := range tags {
		out = append(out, toProtoTag(tag))
	}

	return &rtdbv1.ListTagsResponse{Tags: out}, nil
}

func (s *serverAPI) Subscribe(
	req *rtdbv1.SubscribeRequest,
	stream rtdbv1.RTDB_SubscribeServer,
) error {
	err := s.rtdb.Subscribe(stream.Context(), req.GetNames(), func(value models.TagValue) error {
		return stream.Send(toProtoValue(value))
	})
	if err != nil {
		return s.grpcError("Subscribe", err)
	}

	return nil
}

func toProtoTag(tag models.Tag) *rtdbv1.Tag {
	return &rtdbv1.Tag{
		Def:   toProtoDef(tag.Def),
		Value: toProtoValue(tag.Value),
	}
}

func toProtoDef(def models.TagDef) *rtdbv1.TagDef {
	return &rtdbv1.TagDef{
		Name:        def.Name,
		DisplayName: def.DisplayName,
		Description: def.Description,
		Kind:        rtdbv1.TagKind(def.Kind),
		DataType:    rtdbv1.DataType(def.DataType),
		Access:      rtdbv1.Access(def.Access),
		EngUnit:     def.EngUnit,
		EngLow:      def.EngLow,
		EngHigh:     def.EngHigh,
		SourceType:  rtdbv1.SourceType(def.SourceType),
		Address:     def.Address,
		ScanMs:      def.ScanMs,
		Deadband:    def.Deadband,
		LoLo:        def.LoLo,
		Lo:          def.Lo,
		Hi:          def.Hi,
		HiHi:        def.HiHi,
		Archive:     def.Archive,
	}
}

func toProtoValue(value models.TagValue) *rtdbv1.TagValue {
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
