package config

type SecretApp struct {
	Host string `json:"host"`
	Port int    `json:"port"`
	User string `json:"username"`
	Pass string `json:"password"`
	Name string `json:"dbname"`
	SSL  string `json:"sslmode"`
}

type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

/// mapping objects

func (s SecretApp) ToDBConfig() DBConfig {
	return DBConfig{
		Host:     s.Host,
		Port:     s.Port,
		User:     s.User,
		Password: s.Pass,
		DBName:   s.Name,
		SSLMode:  s.SSL,
	}
}
