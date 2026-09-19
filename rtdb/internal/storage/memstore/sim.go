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
		tag, ok := s.tags[base.Def.Name]
		if !ok {
			continue
		}
		if tag.Def.ScanMs > 0 {
			age := now.Sub(time.UnixMilli(tag.Value.TsUnixMs))
			if age < time.Duration(tag.Def.ScanMs)*time.Millisecond {
				continue
			}
		}
		tag.Value.Value = base.Def.Coerce(animate(base, now, i))
		tag.Value.Quality = models.QualityGood
		tag.Value.TsUnixMs = ts
		s.tags[base.Def.Name] = tag
	}
}

func animate(tag models.Tag, now time.Time, i int) float64 {
	if tag.Def.Discrete() {
		if (now.Unix()/5+int64(i))%2 == 0 {
			return 1
		}
		return 0
	}
	return tag.Value.Value * (1 + 0.04*math.Sin(float64(now.UnixMilli())/800+float64(i)))
}
