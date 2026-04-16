package config

import (
	"log"
	"os"

	"github.com/spf13/viper"
)

type ServerConfig struct {
	Port string `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
	Charset  string `mapstructure:"charset"`
}

type JWTConfig struct {
	Secret      string `mapstructure:"secret"`
	ExpireHours int    `mapstructure:"expire_hours"`
}

type EmailConfig struct {
	SMTPHost string `mapstructure:"smtp_host"`
	SMTPPort int    `mapstructure:"smtp_port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	From     string `mapstructure:"from"`
}

type SecurityConfig struct {
	VerificationCodeExpireMinutes  int `mapstructure:"verification_code_expire_minutes"`
	VerificationCodeMaxErrors      int `mapstructure:"verification_code_max_errors"`
	VerificationCodeIntervalSecond int `mapstructure:"verification_code_interval_seconds"`
}

type CertConfig struct {
	PriceSingle   float64 `mapstructure:"price_single"`
	PriceWildcard float64 `mapstructure:"price_wildcard"`
	DefaultCA     string  `mapstructure:"default_ca"`
	StorageDir    string  `mapstructure:"storage_dir"`
}

type ACMEConfig struct {
	DirectoryURL string `mapstructure:"directory_url"`
	Email        string  `mapstructure:"email"`
}

type AESConfig struct {
	Key string `mapstructure:"key"`
}

type Config struct {
	// Env 为当前选用的数据库环境名（如 prod、dev），来自 config.yaml 的 env 或 CERHUB_ENV。
	Env      string         `mapstructure:"-"`
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"-"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Email    EmailConfig    `mapstructure:"email"`
	Security SecurityConfig `mapstructure:"security"`
	Cert     CertConfig     `mapstructure:"cert"`
	ACME     ACMEConfig     `mapstructure:"acme"`
	AES      AESConfig      `mapstructure:"aes"`
}

type fileConfig struct {
	Env      string                        `mapstructure:"env"`
	Server   ServerConfig                  `mapstructure:"server"`
	Database map[string]DatabaseConfig     `mapstructure:"database"`
	JWT      JWTConfig                     `mapstructure:"jwt"`
	Email    EmailConfig                   `mapstructure:"email"`
	Security SecurityConfig                `mapstructure:"security"`
	Cert     CertConfig                    `mapstructure:"cert"`
	ACME     ACMEConfig                    `mapstructure:"acme"`
	AES      AESConfig                     `mapstructure:"aes"`
}

var C Config

// Init loads config/config.yaml。database 下按 env（或 CERHUB_ENV）选择 prod / dev 等分组。
func Init() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("../config")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("failed to read config: %v", err)
	}

	var raw fileConfig
	if err := viper.Unmarshal(&raw); err != nil {
		log.Fatalf("failed to unmarshal config: %v", err)
	}

	env := os.Getenv("CERHUB_ENV")
	if env == "" {
		env = raw.Env
	}
	if env == "" {
		env = "prod"
	}

	db, ok := raw.Database[env]
	if !ok {
		log.Fatalf("config: database.%s 未定义，请在 config.yaml 的 database 下增加该环境", env)
	}

	C.Env = env
	C.Server = raw.Server
	C.Database = db
	C.JWT = raw.JWT
	C.Email = raw.Email
	C.Security = raw.Security
	C.Cert = raw.Cert
	C.ACME = raw.ACME
	C.AES = raw.AES
}
