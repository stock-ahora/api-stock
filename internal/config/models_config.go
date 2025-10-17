package config

import (
	"os"

	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type SecretApp struct {
	Host        string `json:"host"`
	Port        int    `json:"port"`
	User        string `json:"username"`
	Pass        string `json:"password"`
	Name        string `json:"dbname"`
	SSL         string `json:"sslmode"`
	S3Bucket    string `json:"bucket"`
	S3Region    string `json:"region"`
	MQ_HOST     string `json:"MQ_HOST"`
	MQ_PASSWORD string `json:"MQ_PASSWORD"`
	MQ_PORT     int    `json:"MQ_PORT"`
	MQ_USER     string `json:"MQ_USER_STOCK"`
	MQ_VHOST    string `json:"MQ_VHOST"`
}

type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type MQConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	VHost    string
}

type UploadService struct {
	S3Client    *s3.Client
	Uploader    *manager.Uploader
	Bucket      string
	PublicBase  string
	MaxUploadMB int64
}

type S3Config struct {
	Region string
	Bucket string
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

func (s SecretApp) ToS3Config() S3Config {
	return S3Config{
		Region: s.S3Region,
		Bucket: s.S3Bucket,
	}
}

func (s SecretApp) ToMQConfig() MQConfig {
	return MQConfig{
		Host:     s.MQ_HOST,
		Port:     s.MQ_PORT,
		User:     s.MQ_USER,
		Password: s.MQ_PASSWORD,
		VHost:    s.MQ_VHOST,
	}

}

// functions for configs
func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
