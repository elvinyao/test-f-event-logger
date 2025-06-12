package config

import (
	"strings"

	"github.com/spf13/viper"
)

// Config 结构体定义了应用的所有配置项
type Config struct {
	Server struct {
		Address string `mapstructure:"address"`
	} `mapstructure:"server"`
	Auth struct {
		Token string `mapstructure:"token"`
	} `mapstructure:"auth"`
	Log struct {
		Level    string `mapstructure:"level"`
		FilePath string `mapstructure:"filePath"`
	} `mapstructure:"log"`
}

// AppConfig 是一个全局的配置实例
var AppConfig *Config

// Load 加载配置
// 会从 config.yaml 文件和环境变量中读取配置
func Load() (*Config, error) {
	v := viper.New()

	// 设置配置文件名和路径
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".") // 在当前目录查找

	// 设置环境变量前缀和自动绑定
	v.SetEnvPrefix("FOCALBOARD_MONITOR")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// 设置默认值
	v.SetDefault("server.address", ":8080")
	v.SetDefault("auth.token", "default-secret-token")
	v.SetDefault("log.level", "INFO")
	v.SetDefault("log.filePath", "")

	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		// 如果找不到配置文件，可以忽略错误，因为我们有默认值和环境变量
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	AppConfig = &cfg
	return AppConfig, nil
}
