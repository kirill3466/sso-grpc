package memstore

import (
	"context"
	"sync"
	"time"

	"rtdb/internal/models"
	"rtdb/internal/storage"
)

type Storage struct {
	mu         sync.RWMutex
	tags       map[string]models.Tag
	baselines  []models.Tag
	now        func() time.Time
	staleAfter time.Duration
}

func New(ctx context.Context, tags []models.Tag, staleAfter time.Duration) (*Storage, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	byName := make(map[string]models.Tag, len(tags))
	baselines := make([]models.Tag, 0, len(tags))
	for _, tag := range tags {
		byName[tag.Name] = tag
		if !tag.Access.Writable() {
			baselines = append(baselines, tag)
		}
	}

	return &Storage{
		tags:       byName,
		baselines:  baselines,
		now:        time.Now,
		staleAfter: staleAfter,
	}, nil
}

func (s *Storage) Get(ctx context.Context, name string) (models.TagValue, error) {
	if err := ctx.Err(); err != nil {
		return models.TagValue{}, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	tag, ok := s.tags[name]
	if !ok {
		return models.TagValue{}, storage.ErrNotFound
	}

	return s.withQuality(tag), nil
}

func (s *Storage) Set(ctx context.Context, name string, value float64) (models.TagValue, error) {
	if err := ctx.Err(); err != nil {
		return models.TagValue{}, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	tag, ok := s.tags[name]
	if !ok {
		return models.TagValue{}, storage.ErrNotFound
	}
	if !tag.Access.Writable() {
		return models.TagValue{}, storage.ErrReadOnly
	}

	tag.Value = value
	tag.Quality = models.QualityGood
	tag.TsUnixMs = s.now().UnixMilli()
	s.tags[name] = tag

	return tag.TagValue, nil
}

func (s *Storage) withQuality(tag models.Tag) models.TagValue {
	value := tag.TagValue
	if tag.Access.Writable() {
		return value
	}
	if s.staleAfter <= 0 {
		return value
	}
	age := s.now().Sub(time.UnixMilli(value.TsUnixMs))
	if age > s.staleAfter {
		value.Quality = models.QualityBad
	}
	return value
}
