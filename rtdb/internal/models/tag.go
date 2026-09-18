package models

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

type TagValue struct {
	Name     string
	Value    float64
	Quality  Quality
	TsUnixMs int64
}

type Tag struct {
	TagValue
	Access Access
}
