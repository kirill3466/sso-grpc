package catalog

import (
	"time"

	"rtdb/internal/models"
)

func Tags() []models.Tag {
	now := time.Now().UnixMilli()
	return []models.Tag{
		ai("Цех1.Печь.Температура", "Температура печи", "Температура рабочей зоны печи",
			"°C", 0, 1200, models.SourceTypePLC, "DB1.DBD0", 200, 0.3,
			700, 750, 850, 900, true, 812.4, now),
		ai("Цех1.Печь.Давление", "Давление печи", "Давление в камере печи",
			"бар", 0, 2.5, models.SourceTypePLC, "DB1.DBD4", 200, 0.01,
			0.5, 0.8, 1.8, 2.2, true, 1.21, now),
		ai("Цех1.Линия.Расход", "Расход линии", "Объёмный расход на линии",
			"м³/ч", 0, 30, models.SourceTypeModbus, "HR40001", 200, 0.1,
			2, 5, 20, 25, true, 12.4, now),
		ai("Цех1.Бак.Уровень", "Уровень бака", "Уровень продукта в расходном баке",
			"%", 0, 100, models.SourceTypeOPCUA, "ns=2;s=Tank.Level", 500, 0.2,
			10, 20, 85, 95, true, 67.5, now),
		ai("Цех1.Двигатель.Ток", "Ток двигателя", "Ток двигателя линии",
			"А", 0, 120, models.SourceTypePLC, "DB1.DBD8", 200, 0.2,
			5, 15, 80, 100, true, 48.2, now),
		ai("Цех1.Помещение.Температура", "Температура помещения", "Температура воздуха в цехе",
			"°C", -20, 60, models.SourceTypeModbus, "HR40003", 1000, 0.2,
			5, 12, 35, 45, false, 24.1, now),

		di("Цех1.Печь.ГорелкаВкл", "Горелка включена", "Состояние горелки печи",
			models.SourceTypePLC, "M10.0", 200, true, 1, now),
		di("Цех1.Насос.Работа", "Насос в работе", "Состояние насоса линии",
			models.SourceTypePLC, "M10.1", 200, true, 1, now),
		di("Цех1.Бак.АварияУровня", "Авария уровня", "Дискрет аварии уровня бака",
			models.SourceTypeOPCUA, "ns=2;s=Tank.LevelAlarm", 200, true, 0, now),
		di("Цех1.Дверь.Открыта", "Дверь открыта", "Концевик двери печи",
			models.SourceTypePLC, "M10.2", 200, false, 0, now),

		memInt("Цех1.Счётчик.Партии", "Счётчик партий", "Количество выпущенных партий",
			"шт", 0, 100000, models.SourceTypePLC, "DB2.DBD0", 1000, 1,
			0, 0, 10000, 50000, true, 142, now),
		calc("Цех1.Печь.Отклонение", "Отклонение температуры", "Уставка минус текущая температура",
			"°C", -50, 50, "calc://furnace.dev", 500, 0.2,
			-30, -20, 20, 30, true, 12.4, now),
		calc("Цех1.Печь.КПД", "КПД печи", "Расчётный КПД по теплу",
			"", 0, 1, "calc://furnace.eff", 500, 0.005,
			0.5, 0.7, 0.98, 1, true, 0.87, now),

		ao("Цех1.Печь.Уставка", "Уставка печи", "Задание температуры печи",
			"°C", 0, 1200, models.SourceTypeManual, "DB1.DBD20", 200, 0.1,
			700, 750, 850, 900, true, 800, now),
		ao("Цех1.Линия.УставкаРасхода", "Уставка расхода", "Задание расхода линии",
			"м³/ч", 0, 30, models.SourceTypeManual, "HR40101", 200, 0.1,
			2, 5, 20, 25, true, 12, now),
		ao("Цех1.Клапан.Положение", "Положение клапана", "Задание положения регулирующего клапана",
			"%", 0, 100, models.SourceTypeManual, "DB3.DBD0", 200, 0.5,
			5, 10, 90, 95, false, 45, now),

		do("Цех1.Насос.КомандаПуск", "Пуск насоса", "Команда пуска насоса",
			models.SourceTypeManual, "M20.0", 100, false, 0, now),
		do("Цех1.Клапан.КомандаОткрыть", "Открыть клапан", "Команда открытия дискретного клапана",
			models.SourceTypeManual, "M20.1", 100, false, 1, now),
		do("Цех1.Авария.Квитирование", "Квитирование", "Квитирование аварий цеха",
			models.SourceTypeManual, "M20.2", 100, false, 0, now),
		memBool("Цех1.Режим.Авто", "Режим Авто", "Автоматический режим линии",
			models.SourceTypeManual, "M30.0", 200, false, 1, now),
	}
}

func ai(name, display, desc, unit string, engLow, engHigh float64, src models.SourceType, addr string, scan int32, deadband, loLo, lo, hi, hiHi float64, archive bool, value float64, ts int64) models.Tag {
	return tag(name, display, desc, models.TagKindAI, models.DataTypeFloat, models.AccessRead, unit, engLow, engHigh, src, addr, scan, deadband, loLo, lo, hi, hiHi, archive, value, ts)
}

func ao(name, display, desc, unit string, engLow, engHigh float64, src models.SourceType, addr string, scan int32, deadband, loLo, lo, hi, hiHi float64, archive bool, value float64, ts int64) models.Tag {
	return tag(name, display, desc, models.TagKindAO, models.DataTypeFloat, models.AccessReadWrite, unit, engLow, engHigh, src, addr, scan, deadband, loLo, lo, hi, hiHi, archive, value, ts)
}

func di(name, display, desc string, src models.SourceType, addr string, scan int32, archive bool, value float64, ts int64) models.Tag {
	return tag(name, display, desc, models.TagKindDI, models.DataTypeBool, models.AccessRead, "", 0, 1, src, addr, scan, 0, 0, 0, 1, 1, archive, value, ts)
}

func do(name, display, desc string, src models.SourceType, addr string, scan int32, archive bool, value float64, ts int64) models.Tag {
	return tag(name, display, desc, models.TagKindDO, models.DataTypeBool, models.AccessReadWrite, "", 0, 1, src, addr, scan, 0, 0, 0, 1, 1, archive, value, ts)
}

func calc(name, display, desc, unit string, engLow, engHigh float64, addr string, scan int32, deadband, loLo, lo, hi, hiHi float64, archive bool, value float64, ts int64) models.Tag {
	return tag(name, display, desc, models.TagKindCALC, models.DataTypeFloat, models.AccessRead, unit, engLow, engHigh, models.SourceTypeCalc, addr, scan, deadband, loLo, lo, hi, hiHi, archive, value, ts)
}

func memInt(name, display, desc, unit string, engLow, engHigh float64, src models.SourceType, addr string, scan int32, deadband, loLo, lo, hi, hiHi float64, archive bool, value float64, ts int64) models.Tag {
	return tag(name, display, desc, models.TagKindMEM, models.DataTypeInt, models.AccessRead, unit, engLow, engHigh, src, addr, scan, deadband, loLo, lo, hi, hiHi, archive, value, ts)
}

func memBool(name, display, desc string, src models.SourceType, addr string, scan int32, archive bool, value float64, ts int64) models.Tag {
	return tag(name, display, desc, models.TagKindMEM, models.DataTypeBool, models.AccessReadWrite, "", 0, 1, src, addr, scan, 0, 0, 0, 1, 1, archive, value, ts)
}

func tag(
	name, display, desc string,
	kind models.TagKind,
	dtype models.DataType,
	access models.Access,
	unit string,
	engLow, engHigh float64,
	src models.SourceType,
	addr string,
	scan int32,
	deadband, loLo, lo, hi, hiHi float64,
	archive bool,
	value float64,
	ts int64,
) models.Tag {
	def := models.TagDef{
		Name:        name,
		DisplayName: display,
		Description: desc,
		Kind:        kind,
		DataType:    dtype,
		Access:      access,
		EngUnit:     unit,
		EngLow:      engLow,
		EngHigh:     engHigh,
		SourceType:  src,
		Address:     addr,
		ScanMs:      scan,
		Deadband:    deadband,
		LoLo:        loLo,
		Lo:          lo,
		Hi:          hi,
		HiHi:        hiHi,
		Archive:     archive,
	}
	return models.Tag{
		Def: def,
		Value: models.TagValue{
			Name:     name,
			Value:    def.Coerce(value),
			Quality:  models.QualityGood,
			TsUnixMs: ts,
		},
	}
}
