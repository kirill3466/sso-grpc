package rtdb

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"rtdb/internal/models"
	"rtdb/internal/slogx"
	"rtdb/internal/storage"
)

const defaultSubscribeEvery = 200 * time.Millisecond

type Store interface {
	Get(ctx context.Context, name string) (models.Tag, error)
	Set(ctx context.Context, name string, value float64) (models.TagValue, error)
	List(ctx context.Context) ([]models.Tag, error)
}

type Service struct {
	log            *slog.Logger
	store          Store
	subscribeEvery time.Duration
}

func New(log *slog.Logger, store Store) *Service {
	return &Service{
		log:            log,
		store:          store,
		subscribeEvery: defaultSubscribeEvery,
	}
}

func (s *Service) GetTag(ctx context.Context, name string) (models.Tag, error) {
	const op = "rtdb.GetTag"

	log := s.log.With(slog.String("op", op), slog.String("name", name))
	log.Debug("getting tag")

	tag, err := s.store.Get(ctx, name)
	if err != nil {
		return models.Tag{}, mapStoreError(op, log, err)
	}

	return tag, nil
}

func (s *Service) ListTags(ctx context.Context) ([]models.Tag, error) {
	const op = "rtdb.ListTags"

	log := s.log.With(slog.String("op", op))
	log.Debug("listing tags")

	tags, err := s.store.List(ctx)
	if err != nil {
		return nil, mapStoreError(op, log, err)
	}

	return tags, nil
}

func (s *Service) SetTag(ctx context.Context, name string, value float64) (models.TagValue, error) {
	const op = "rtdb.SetTag"

	log := s.log.With(slog.String("op", op), slog.String("name", name))
	log.Debug("setting tag")

	updated, err := s.store.Set(ctx, name, value)
	if err != nil {
		return models.TagValue{}, mapStoreError(op, log, err)
	}

	return updated, nil
}

func (s *Service) Subscribe(ctx context.Context, names []string, emit func(models.TagValue) error) error {
	const op = "rtdb.Subscribe"

	s.log.With(slog.String("op", op)).
		Debug("subscribing", slog.Int("tags", len(names)))

	names = uniqueNames(names)
	if err := s.snapshot(ctx, names, nil, nil); err != nil {
		return err
	}

	last := make(map[string]models.TagValue, len(names))
	if err := s.snapshot(ctx, names, last, emit); err != nil {
		return err
	}

	ticker := time.NewTicker(s.subscribeEvery)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := s.snapshot(ctx, names, last, emit); err != nil {
				return err
			}
		}
	}
}

func (s *Service) snapshot(
	ctx context.Context,
	names []string,
	last map[string]models.TagValue,
	emit func(models.TagValue) error,
) error {
	const op = "rtdb.Subscribe"

	log := s.log.With(slog.String("op", op))

	for _, name := range names {
		tag, err := s.store.Get(ctx, name)
		if err != nil {
			return mapStoreError(op, log, err)
		}
		if emit == nil {
			continue
		}

		prev, ok := last[name]
		var prevPtr *models.TagValue
		if ok {
			prevPtr = &prev
		}
		if !models.ShouldEmit(prevPtr, tag.Value, tag.Def.Deadband) {
			continue
		}
		if err := emit(tag.Value); err != nil {
			return err
		}
		last[name] = tag.Value
	}

	return nil
}

func uniqueNames(names []string) []string {
	seen := make(map[string]struct{}, len(names))
	out := make([]string, 0, len(names))
	for _, name := range names {
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		out = append(out, name)
	}
	return out
}

func mapStoreError(op string, log *slog.Logger, err error) error {
	switch {
	case errors.Is(err, storage.ErrNotFound):
		return fmt.Errorf("%s: %w", op, ErrNotFound)
	case errors.Is(err, storage.ErrReadOnly):
		return fmt.Errorf("%s: %w", op, ErrReadOnly)
	default:
		log.Error("store failed", slogx.Err(err))
		return fmt.Errorf("%s: %w", op, err)
	}
}
