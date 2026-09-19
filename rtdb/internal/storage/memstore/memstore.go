package memstore

import (
	"context"
	"sort"
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
		byName[tag.Def.Name] = tag
		if !tag.Def.Access.Writable() {
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

func (s *Storage) Get(ctx context.Context, name string) (models.Tag, error) {
	if err := ctx.Err(); err != nil {
		return models.Tag{}, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	tag, ok := s.tags[name]
	if !ok {
		return models.Tag{}, storage.ErrNotFound
	}

	return s.view(tag), nil
}

func (s *Storage) List(ctx context.Context) ([]models.Tag, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]models.Tag, 0, len(s.tags))
	for _, tag := range s.tags {
		out = append(out, s.view(tag))
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Def.Name < out[j].Def.Name
	})
	return out, nil
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
	if !tag.Def.Access.Writable() {
		return models.TagValue{}, storage.ErrReadOnly
	}

	tag.Value.Value = tag.Def.Coerce(value)
	tag.Value.Quality = models.QualityGood
	tag.Value.TsUnixMs = s.now().UnixMilli()
	s.tags[name] = tag

	return s.view(tag).Value, nil
}

func (s *Storage) view(tag models.Tag) models.Tag {
	return tag.WithLiveQuality(s.now(), s.staleAfter)
}
