// Copyright (c) 2025 Youzill
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package config

import (
	"log"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type YamlConfig struct {
	AppName  string         `mapstructure:"appName"`
	Log      LogConfig      `mapstructure:"log"`
	Api      ApiConfig      `mapstructure:"api"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Jwt      JwtConfig      `mapstructure:"jwt"`
	Spew     SpewConfig     `mapstructure:"spew"`
	Music    MusicConfig    `mapstructure:"music"`
	Security SecurityConfig `mapstructure:"security"`
	Ai       AIConfig       `mapstructure:"Ai"`
}

type LogConfig struct {
	Level string `mapstructure:"level"`
	Path  string `mapstructure:"path"`
}

type cookie struct {
	Filepath string        `mapstructure:"filepath"`
	Interval time.Duration `mapstructure:"interval"`
}

type ApiConfig struct {
	Debug     bool            `mapstructure:"debug"`
	Timeout   time.Duration   `mapstructure:"timeout"`
	Retry     int             `mapstructure:"retry"`
	Cookie    cookie          `mapstructure:"cookie"`
	RateLimit RateLimitConfig `mapstructure:"rateLimit"`
}

type RateLimitConfig struct {
	RequestsPerMinute int `mapstructure:"requestsPerMinute"`
	BurstSize         int `mapstructure:"burstSize"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type JwtConfig struct {
	Secret  string        `mapstructure:"secret"`
	Expires time.Duration `mapstructure:"expires"`
}

type SpewConfig struct {
	Indent                string `mapstructure:"indent"`
	MaxDepth              int    `mapstructure:"maxdepth"`
	DisableMethods        bool   `mapstructure:"disablemethods"`
	DisablePointerMethods bool   `mapstructure:"disablepointermethods"`
	ContinueOnMethod      bool   `mapstructure:"continueonmethod"`
	SortKeys              bool   `mapstructure:"sortkeys"`
}

type MusicConfig struct {
	Bits        int   `mapstructure:"bits"`
	Concurrency int64 `mapstructure:"concurrency"`
}

type SecurityConfig struct {
	SessionSecret    string     `mapstructure:"session_secret"`
	MaxFileSize      string     `mapstructure:"max_file_size"`
	AllowedFileTypes []string   `mapstructure:"allowed_file_types"`
	CORS             CORSConfig `mapstructure:"cors"`
}

type CORSConfig struct {
	AllowedOrigins []string `mapstructure:"allowed_origins"`
	AllowedMethods []string `mapstructure:"allowed_methods"`
	AllowedHeaders []string `mapstructure:"allowed_headers"`
}

type AIConfig struct {
	Provider       string        `mapstructure:"provider"`
	BaseURL        string        `mapstructure:"base_url"`
	Model          string        `mapstructure:"model"`
	Timeout        time.Duration `mapstructure:"timeout"`
	MaxTitleLength int           `mapstructure:"max_title_length"`
	CacheTTL       time.Duration `mapstructure:"cache_ttl"`
	Concurrency    int64         `mapstructure:"concurrency"`
}

var c YamlConfig

func init() {
	// Docker部署时不需要加载.env文件，因为通过环境变量传递配置,本地运行时需要添加
	envPath := filepath.Join("..", ".env")
	err := godotenv.Load(envPath)
	if err != nil {
		panic("fail to load .env file,err : " + err.Error())
	}

	// 配置 viper
	viper.SetConfigName("conf")     // 配置文件名称（不带扩展名）
	viper.SetConfigType("yaml")     // 配置文件类型
	viper.AddConfigPath("./config") // 配置文件所在路径

	// 设置环境变量优先级
	viper.AutomaticEnv()

	// 绑定环境变量
	bindEnvVars()

	// 调试：打印环境变量 - 生产环境已禁用
	// debugEnvVars()

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

	// AI配置
	if err := viper.BindEnv("ai.provider", "AI_PROVIDER"); err != nil {
		log.Printf("Failed to bind AI_PROVIDER: %v", err)
	}
	if err := viper.BindEnv("ai.base_url", "AI_BASE_URL"); err != nil {
		log.Printf("Failed to bind AI_BASE_URL: %v", err)
	}
	if err := viper.BindEnv("ai.model", "AI_MODEL"); err != nil {
		log.Printf("Failed to bind AI_MODEL: %v", err)
	}
	if err := viper.BindEnv("ai.timeout", "AI_TIMEOUT_SECONDS"); err != nil {
		log.Printf("Failed to bind AI_TIMEOUT_SECONDS: %v", err)
	}
	if err := viper.BindEnv("ai.max_title_length", "AI_MAX_TITLE_LENGTH"); err != nil {
		log.Printf("Failed to bind AI_MAX_TITLE_LENGTH: %v", err)
	}
	if err := viper.BindEnv("ai.cache_ttl", "AI_CACHE_TTL"); err != nil {
		log.Printf("Failed to bind AI_CACHE_TTL: %v", err)
	}
	if err := viper.BindEnv("ai.concurrency", "AI_CONCURRENCY"); err != nil {
		log.Printf("Failed to bind AI_CONCURRENCY: %v", err)
	}
}

// debugEnvVars 调试环境变量
// func debugEnvVars() {
// 	log.Printf("=== Environment Variables Debug ===")
//  	log.Printf("CORS_ALLOWED_ORIGINS: %s", os.Getenv("CORS_ALLOWED_ORIGINS"))
//  	log.Printf("CORS_ALLOWED_METHODS: %s", os.Getenv("CORS_ALLOWED_METHODS"))
//  	log.Printf("CORS_ALLOWED_HEADERS: %s", os.Getenv("CORS_ALLOWED_HEADERS"))
//  	log.Printf("REDIS_HOST: %s", os.Getenv("REDIS_HOST"))
//  	log.Printf("REDIS_PASSWORD: %s", os.Getenv("REDIS_PASSWORD"))
//  	log.Printf("SESSION_SECRET: %s", os.Getenv("SESSION_SECRET"))
//  	log.Printf("=== End Debug ===")
// }

// GetConfig 用于获取解析后的配置结构体
func GetConfig() YamlConfig {
	return c
}
