package collector

// Collector интерфейс для всех коллекторов
type Collector interface {
	Name() string
	Enabled() bool
	Collect() (interface{}, error)
}

// BaseCollector базовая структура для всех коллекторов
type BaseCollector struct {
	HostID   string
	Hostname string
}
