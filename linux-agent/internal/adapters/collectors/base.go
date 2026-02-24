package collectors

// BaseCollector базовая структура для коллекторов
type BaseCollector struct {
	HostID     string
	Hostname   string
	NameVal    string
	EnabledVal bool
}

// Name возвращает имя коллектора
func (b *BaseCollector) Name() string {
	return b.NameVal
}

// Enabled возвращает включен ли коллектор
func (b *BaseCollector) Enabled() bool {
	return b.EnabledVal
}
