package config

import (
	"log"
	"os"
	"time"

	"github.com/spf13/viper"
)

type YamlConfig struct {
	AppName  string
	Log      LogConfig
	Api      ApiConfig
	Redis    RedisConfig
	Jwt      JwtConfig
	Spew     SpewConfig
	Music    MusicConfig
	Security SecurityConfig
}

type LogConfig struct {
	Level string
	Path  string
}

type cookie struct {
	Filepath string
	Interval time.Duration
}

type ApiConfig struct {
	Debug     bool
	Timeout   time.Duration
	Retry     int
	Cookie    cookie
	RateLimit RateLimitConfig
}

type RateLimitConfig struct {
	RequestsPerMinute int
	BurstSize         int
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type JwtConfig struct {
	Secret  string
	Expires time.Duration
}

type SpewConfig struct {
	Indent                string
	MaxDepth              int
	DisableMethods        bool
	DisablePointerMethods bool
	ContinueOnMethod      bool
	SortKeys              bool
}

type MusicConfig struct {
	Bits        int   `mapstructure:"bits"`
	Concurrency int64 `mapstructure:"concurrency"`
}

type SecurityConfig struct {
	SessionSecret    string
	MaxFileSize      string
	AllowedFileTypes []string
	CORS             CORSConfig
}

type CORSConfig struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
}

var c YamlConfig

func init() {
	// Docker部署时不需要加载.env文件，因为通过环境变量传递配置,本地运行时需要添加
	// envPath := filepath.Join("..", ".env")
	// err := godotenv.Load(envPath)
	// if err != nil {
	// 	panic("fail to load .env file,err : " + err.Error())
	// }

	// 配置 viper
	viper.SetConfigName("conf")     // 配置文件名称（不带扩展名）
	viper.SetConfigType("yaml")     // 配置文件类型
	viper.AddConfigPath("./config") // 配置文件所在路径

	// 设置环境变量优先级
	viper.AutomaticEnv()

	// 绑定环境变量
	bindEnvVars()

	// 调试：打印环境变量
	debugEnvVars()

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			panic("配置文件未找到,err : " + err.Error())
		} else {
			panic("读取配置文件时出错,err : " + err.Error())
		}
	}

	viper.Unmarshal(&c)
}

// bindEnvVars 绑定环境变量
func bindEnvVars() {
	// Redis配置
	if err := viper.BindEnv("redis.host", "REDIS_HOST"); err != nil {
		log.Printf("Failed to bind REDIS_HOST: %v", err)
	}
	if err := viper.BindEnv("redis.port", "REDIS_PORT"); err != nil {
		log.Printf("Failed to bind REDIS_PORT: %v", err)
	}
	if err := viper.BindEnv("redis.password", "REDIS_PASSWORD"); err != nil {
		log.Printf("Failed to bind REDIS_PASSWORD: %v", err)
	}
	if err := viper.BindEnv("redis.db", "REDIS_DB"); err != nil {
		log.Printf("Failed to bind REDIS_DB: %v", err)
	}

	// 安全配置
	if err := viper.BindEnv("security.session_secret", "SESSION_SECRET"); err != nil {
		log.Printf("Failed to bind SESSION_SECRET: %v", err)
	}
	if err := viper.BindEnv("security.allowed_file_types", "ALLOWED_FILE_TYPES"); err != nil {
		log.Printf("Failed to bind ALLOWED_FILE_TYPES: %v", err)
	}
	if err := viper.BindEnv("security.cors.allowed_origins", "CORS_ALLOWED_ORIGINS"); err != nil {
		log.Printf("Failed to bind CORS_ALLOWED_ORIGINS: %v", err)
	}
	if err := viper.BindEnv("security.cors.allowed_methods", "CORS_ALLOWED_METHODS"); err != nil {
		log.Printf("Failed to bind CORS_ALLOWED_METHODS: %v", err)
	}
	if err := viper.BindEnv("security.cors.allowed_headers", "CORS_ALLOWED_HEADERS"); err != nil {
		log.Printf("Failed to bind CORS_ALLOWED_HEADERS: %v", err)
	}

	// 应用配置
	if err := viper.BindEnv("log.path", "LOG_PATH"); err != nil {
		log.Printf("Failed to bind LOG_PATH: %v", err)
	}

	// API配置
	if err := viper.BindEnv("api.cookie.filepath", "NETEASE_API_COOKIE_FILEPATH"); err != nil {
		log.Printf("Failed to bind NETEASE_API_COOKIE_FILEPATH: %v", err)
	}
	if err := viper.BindEnv("api.cookie.interval", "NETEASE_API_COOKIE_INTERVAL"); err != nil {
		log.Printf("Failed to bind NETEASE_API_COOKIE_INTERVAL: %v", err)
	}
}

// debugEnvVars 调试环境变量
func debugEnvVars() {
	log.Printf("=== Environment Variables Debug ===")
	log.Printf("CORS_ALLOWED_ORIGINS: %s", os.Getenv("CORS_ALLOWED_ORIGINS"))
	log.Printf("CORS_ALLOWED_METHODS: %s", os.Getenv("CORS_ALLOWED_METHODS"))
	log.Printf("CORS_ALLOWED_HEADERS: %s", os.Getenv("CORS_ALLOWED_HEADERS"))
	log.Printf("REDIS_HOST: %s", os.Getenv("REDIS_HOST"))
	log.Printf("REDIS_PASSWORD: %s", os.Getenv("REDIS_PASSWORD"))
	log.Printf("SESSION_SECRET: %s", os.Getenv("SESSION_SECRET"))
	log.Printf("=== End Debug ===")
}

// GetConfig 用于获取解析后的配置结构体
func GetConfig() YamlConfig {
	return c
}

// GetRedisPassword 获取Redis密码，支持环境变量
func GetRedisPassword() string {
	if password := os.Getenv("REDIS_PASSWORD"); password != "" {
		return password
	}
	return c.Redis.Password
}
