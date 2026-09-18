package memstore

import (
	"context"
	"math"
	"time"

	"rtdb/internal/models"
)

func (s *Storage) RunSimulator(ctx context.Context, every time.Duration) {
	if every <= 0 {
		every = 100 * time.Millisecond
	}

	s.tick(time.Now())

	ticker := time.NewTicker(every)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			s.tick(now)
		}
	}
}

func (s *Storage) tick(now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ts := now.UnixMilli()
	for i, base := range s.baselines {
		tag, ok := s.tags[base.Name]
		if !ok {
			continue
		}
		tag.Value = animate(base, now, i)
		tag.Quality = models.QualityGood
		tag.TsUnixMs = ts
		s.tags[base.Name] = tag
	}
}

func animate(tag models.Tag, now time.Time, i int) float64 {
	if tag.Value == 0 || tag.Value == 1 {
		if (now.Unix()/5+int64(i))%2 == 0 {
			return 1
		}
		return 0
	}
	return tag.Value * (1 + 0.04*math.Sin(float64(now.UnixMilli())/800+float64(i)))
}
