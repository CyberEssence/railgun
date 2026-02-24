package config

type ElasticConfig struct {
	URL      string   `yaml:"url"`
	URLs     []string `yaml:"urls"`     // Список URL для кластера
	Username string   `yaml:"username"` // Добавляем поле
	Password string   `yaml:"password"` // Добавляем поле
	Index    string   `yaml:"index"`
	Enabled  bool     `yaml:"enabled"`
}
