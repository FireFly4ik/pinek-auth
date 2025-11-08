package config

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Address        string
	Port           string
	ProductionType string

	Consul Consul
	DB     Database
	JWT    JWT
}

type Consul struct {
	Address       string
	Name          string
	CheckId       string
	RegisterTTL   string
	RefreshTTL    string
	DeregisterTTL string
}

type Database struct {
	Host            string
	Port            string
	User            string
	Name            string
	Password        string
	SSLMode         string
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime int
}

type JWT struct {
	AccessTTL  string
	RefreshTTL string
}

func getIPAddress() string {
	resp, err := http.Get("https://ifconfig.me/ip")
	if err != nil {
		panic(fmt.Sprintf("failed to get public IP address: %v", err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		panic(fmt.Sprintf("failed to get public IP address: received status code %d", resp.StatusCode))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(fmt.Sprintf("failed to read response body: %v", err))
	}

	return string(body)
}

func NewEnvConfig() *Config {
	maxIdleConnsStr := os.Getenv("DATABASE_MAX_IDLE_CONNS")
	maxIdleConns, err := strconv.Atoi(maxIdleConnsStr)
	if err != nil {
		panic(fmt.Errorf("NewEnvConfig: error converting maxIdleConnsStr: %w", err))
	}

	maxOpenConnsStr := os.Getenv("DATABASE_MAX_OPEN_CONNS")
	maxOpenConns, err := strconv.Atoi(maxOpenConnsStr)
	if err != nil {
		panic(fmt.Errorf("NewEnvConfig: error converting maxOpenConnsStr: %w", err))
	}

	connMaxLifetimeStr := os.Getenv("DATABASE_CONN_MAX_LIFETIME_IN_SECONDS")
	connMaxLifetime, err := strconv.Atoi(connMaxLifetimeStr)
	if err != nil {
		panic(fmt.Errorf("NewEnvConfig: error converting connMaxLifetimeStr: %w", err))
	}

	return &Config{
		Address:        getIPAddress(),
		Port:           os.Getenv("PORT"),
		ProductionType: os.Getenv("PRODUCTION_TYPE"),

		Consul: Consul{
			Address:       os.Getenv("CONSUL_ADDRESS"),
			Name:          os.Getenv("CONSUL_SERVICE_NAME"),
			CheckId:       os.Getenv("CONSUL_CHECK_ID"),
			RegisterTTL:   os.Getenv("CONSUL_REGISTER_TTL"),
			RefreshTTL:    os.Getenv("CONSUL_REFRESH_TTL"),
			DeregisterTTL: os.Getenv("CONSUL_DEREGISTER_TTL"),
		},

		DB: Database{
			Host:            os.Getenv("DATABASE_HOST"),
			Port:            os.Getenv("DATABASE_PORT"),
			User:            os.Getenv("DATABASE_USER"),
			Name:            os.Getenv("DATABASE_NAME"),
			Password:        os.Getenv("DATABASE_PASSWORD"),
			SSLMode:         os.Getenv("DATABASE_SSL_MODE"),
			MaxIdleConns:    maxIdleConns,
			MaxOpenConns:    maxOpenConns,
			ConnMaxLifetime: connMaxLifetime,
		},

		JWT: JWT{
			AccessTTL:  os.Getenv("JWT_ACCESS_TOKEN_TTL"),
			RefreshTTL: os.Getenv("JWT_REFRESH_TOKEN_TTL"),
		},
	}
}

func PrintConfigWithHiddenSecrets(config *Config) {
	// Функция для маскировки секретов
	mask := func(s string) string {
		if s == "" {
			return ""
		}
		return strings.Repeat("*", len(s))
	}

	fmt.Println("========== Configuration ==========\n")

	fmt.Println("App Configuration:")
	fmt.Printf("\tAddress: %s\n", config.Address)
	fmt.Printf("\tPort: %s\n", config.Port)
	fmt.Printf("\tProductionType: %s\n", config.ProductionType)

	fmt.Println("\nConsul Configuration:")
	fmt.Printf("\tAddress: %s\n", config.Consul.Address)
	fmt.Printf("\tName: %s\n", config.Consul.Name)
	fmt.Printf("\tCheckId: %s\n", config.Consul.CheckId)
	fmt.Printf("\tRegisterTTL: %s\n", config.Consul.RegisterTTL)
	fmt.Printf("\tRefreshTTL: %s\n", config.Consul.RefreshTTL)
	fmt.Printf("\tDeregisterTTL: %s\n", config.Consul.DeregisterTTL)

	fmt.Println("\nDatabase Configuration:")
	fmt.Printf("\tHost: %s\n", config.DB.Host)
	fmt.Printf("\tPort: %s\n", config.DB.Port)
	fmt.Printf("\tUser: %s\n", config.DB.User)
	fmt.Printf("\tName: %s\n", config.DB.Name)
	fmt.Printf("\tPassword: %s\n", mask(config.DB.Password))
	fmt.Printf("\tSSLMode: %s\n", config.DB.SSLMode)
	fmt.Printf("\tMaxIdleConns: %d\n", config.DB.MaxIdleConns)
	fmt.Printf("\tMaxOpenConns: %d\n", config.DB.MaxOpenConns)
	fmt.Printf("\tConnMaxLifetime: %d\n", config.DB.ConnMaxLifetime)

	fmt.Println("\nJWT Configuration:")
	fmt.Printf("\tAccessTTL: %s\n", config.JWT.AccessTTL)
	fmt.Printf("\tRefreshTTL: %s\n", config.JWT.RefreshTTL)

	fmt.Println("\n===================================")
}
