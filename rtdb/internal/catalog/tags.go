package catalog

import (
	"time"

	"rtdb/internal/models"
)

func Tags() []models.Tag {
	now := time.Now().UnixMilli()
	return []models.Tag{
		ro("Цех1.Печь.Температура", 812.4, now),
		ro("Цех1.Печь.Давление", 1.21, now),
		ro("Цех1.Линия.Расход", 12.4, now),
		ro("Цех1.Бак.Уровень", 67.5, now),
		ro("Цех1.Двигатель.Ток", 48.2, now),
		ro("Цех1.Помещение.Температура", 24.1, now),
		ro("Цех1.Печь.ГорелкаВкл", 1, now),
		ro("Цех1.Насос.Работа", 1, now),
		ro("Цех1.Бак.АварияУровня", 0, now),
		ro("Цех1.Дверь.Открыта", 0, now),
		ro("Цех1.Счётчик.Партии", 142, now),
		ro("Цех1.Печь.Отклонение", 12.4, now),
		ro("Цех1.Печь.КПД", 0.87, now),

		rw("Цех1.Печь.Уставка", 800, now),
		rw("Цех1.Линия.УставкаРасхода", 12, now),
		rw("Цех1.Клапан.Положение", 45, now),
		rw("Цех1.Насос.КомандаПуск", 0, now),
		rw("Цех1.Клапан.КомандаОткрыть", 1, now),
		rw("Цех1.Авария.Квитирование", 0, now),
		rw("Цех1.Режим.Авто", 1, now),
	}
}

func ro(name string, value float64, ts int64) models.Tag {
	return tag(name, value, ts, models.AccessRead)
}

func rw(name string, value float64, ts int64) models.Tag {
	return tag(name, value, ts, models.AccessReadWrite)
}

func tag(name string, value float64, ts int64, access models.Access) models.Tag {
	return models.Tag{
		TagValue: models.TagValue{
			Name:     name,
			Value:    value,
			Quality:  models.QualityGood,
			TsUnixMs: ts,
		},
		Access: access,
	}
}
