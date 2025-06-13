package config

import (
	"time"

	"github.com/spf13/viper"
)

type YamlConfig struct {
	AppName string
	Log     LogConfig
	Api     ApiConfig
	Redis   RedisConfig
	Jwt     JwtConfig
	Spew    SpewConfig
	Music   MusicConfig
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
	Debug   bool
	Timeout time.Duration
	Retry   int
	Cookie  cookie
}

type RedisConfig struct {
	Addr     string
	Post     string
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

var c YamlConfig

func init() {
	// 配置 viper
	viper.SetConfigName("conf")     // 配置文件名称（不带扩展名）
	viper.SetConfigType("yaml")     // 配置文件类型
	viper.AddConfigPath("./config") // 配置文件所在路径

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

// GetConfig 用于获取解析后的配置结构体
func GetConfig() YamlConfig {
	return c
}
