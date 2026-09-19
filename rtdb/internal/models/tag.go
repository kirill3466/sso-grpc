package models

import (
	"math"
	"time"
)

type Quality int32

const (
	QualityUnspecified Quality = 0
	QualityGood        Quality = 1
	QualityUncertain   Quality = 2
	QualityBad         Quality = 3
)

type Access int32

const (
	AccessUnspecified Access = 0
	AccessRead        Access = 1
	AccessWrite       Access = 2
	AccessReadWrite   Access = 3
)

func (a Access) Writable() bool {
	return a == AccessWrite || a == AccessReadWrite
}

type TagKind int32

const (
	TagKindUnspecified TagKind = 0
	TagKindAI          TagKind = 1
	TagKindAO          TagKind = 2
	TagKindDI          TagKind = 3
	TagKindDO          TagKind = 4
	TagKindMEM         TagKind = 5
	TagKindCALC        TagKind = 6
)

type DataType int32

const (
	DataTypeUnspecified DataType = 0
	DataTypeFloat       DataType = 1
	DataTypeInt         DataType = 2
	DataTypeBool        DataType = 3
	DataTypeString      DataType = 4
)

type SourceType int32

const (
	SourceTypeUnspecified SourceType = 0
	SourceTypePLC         SourceType = 1
	SourceTypeOPCUA       SourceType = 2
	SourceTypeModbus      SourceType = 3
	SourceTypeCalc        SourceType = 4
	SourceTypeManual      SourceType = 5
)

type TagDef struct {
	Name        string
	DisplayName string
	Description string
	Kind        TagKind
	DataType    DataType
	Access      Access
	EngUnit     string
	EngLow      float64
	EngHigh     float64
	SourceType  SourceType
	Address     string
	ScanMs      int32
	Deadband    float64
	LoLo        float64
	Lo          float64
	Hi          float64
	HiHi        float64
	Archive     bool
}

func (d TagDef) Discrete() bool {
	return d.DataType == DataTypeBool || d.Kind == TagKindDI || d.Kind == TagKindDO
}

func (d TagDef) Coerce(value float64) float64 {
	switch {
	case d.Discrete():
		if value >= 0.5 {
			return 1
		}
		return 0
	case d.DataType == DataTypeInt:
		return math.Round(value)
	default:
		return value
	}
}

type TagValue struct {
	Name     string
	Value    float64
	Quality  Quality
	TsUnixMs int64
}

type Tag struct {
	Def   TagDef
	Value TagValue
}

func (t Tag) WithLiveQuality(now time.Time, staleAfter time.Duration) Tag {
	t.Value.Quality = t.liveQuality(now, staleAfter)
	return t
}

func (t Tag) liveQuality(now time.Time, staleAfter time.Duration) Quality {
	if !t.Def.Access.Writable() {
		age := now.Sub(time.UnixMilli(t.Value.TsUnixMs))
		limit := staleAfter
		if t.Def.ScanMs > 0 {
			if scan := time.Duration(t.Def.ScanMs) * 3 * time.Millisecond; scan > limit {
				limit = scan
			}
		}
		if limit > 0 && age > limit {
			return QualityBad
		}
	}

	v := t.Value.Value
	if t.Def.Discrete() {
		if v != 0 && v != 1 {
			return QualityUncertain
		}
		return QualityGood
	}

	if v < t.Def.EngLow || v > t.Def.EngHigh || v <= t.Def.LoLo || v >= t.Def.HiHi {
		return QualityBad
	}
	if v <= t.Def.Lo || v >= t.Def.Hi {
		return QualityUncertain
	}
	return QualityGood
}

func ShouldEmit(prev *TagValue, next TagValue, deadband float64) bool {
	if prev == nil {
		return true
	}
	if prev.Quality != next.Quality {
		return true
	}
	delta := math.Abs(next.Value - prev.Value)
	if deadband <= 0 {
		return delta != 0
	}
	return delta >= deadband
}
