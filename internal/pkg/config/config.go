package config

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"
)

const (
	// RecommendedHomeDir 定义放置 demo520 服务配置的默认目录
	RecommendedHomeDir = ".demo520"

	// DefaultConfigName 指定了 demo520 服务的默认配置文件名
	DefaultConfigName = "demo520.yaml"
)

// Config 定义了应用程序的完整配置结构
type Config struct {
	Server ServerConfig `mapstructure:"server" yaml:"server"`
	DB     DBConfig     `mapstructure:"db" yaml:"db"`
	Log    LogConfig    `mapstructure:"log" yaml:"log"`
	Image  ImageConfig  `mapstructure:"image" yaml:"image"`
	JWT    JWTConfig    `mapstructure:"jwt" yaml:"jwt"`
	App    AppConfig    `mapstructure:"app" yaml:"app"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Addr string `mapstructure:"addr" yaml:"addr"`
}

// DBConfig 数据库配置
type DBConfig struct {
	Host                  string        `mapstructure:"host" yaml:"host"`
	Username              string        `mapstructure:"username" yaml:"username"`
	Password              string        `mapstructure:"password" yaml:"password"`
	Database              string        `mapstructure:"database" yaml:"database"`
	MaxIdleConnections    int           `mapstructure:"max-idle-connections" yaml:"max-idle-connections"`
	MaxOpenConnections    int           `mapstructure:"max-open-connections" yaml:"max-open-connections"`
	MaxConnectionLifeTime time.Duration `mapstructure:"max-connection-life-time" yaml:"max-connection-life-time"`
	LogLevel              int           `mapstructure:"log-level" yaml:"log-level"`
}

// LogConfig 日志配置
type LogConfig struct {
	DisableCaller     bool     `mapstructure:"disable-caller" yaml:"disable-caller"`
	DisableStacktrace bool     `mapstructure:"disable-stacktrace" yaml:"disable-stacktrace"`
	Level             string   `mapstructure:"level" yaml:"level"`
	Encoding          string   `mapstructure:"encoding" yaml:"encoding"`
	OutputPaths       []string `mapstructure:"output-paths" yaml:"output-paths"`
}

// ImageConfig 图片处理配置
type ImageConfig struct {
	ImageMaxSize int64        `mapstructure:"ImageMaxSize" yaml:"ImageMaxSize"`
	ImageDir     string       `mapstructure:"image_dir" yaml:"image_dir"`
	Convert      ConvertConfig `mapstructure:"Convert" yaml:"Convert"`
}

// ConvertConfig 图片转换配置
type ConvertConfig struct {
	WebPQuality        int  `mapstructure:"WebPQuality" yaml:"WebPQuality"`
	WebReductionEffort int  `mapstructure:"WebReductionEffort" yaml:"WebReductionEffort"`
	AvifQuality        int  `mapstructure:"AvifQuality" yaml:"AvifQuality"`
	AvifEffort         int  `mapstructure:"AvifEffort" yaml:"AvifEffort"`
	ImageLossless      bool `mapstructure:"ImageLossless" yaml:"ImageLossless"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret string `mapstructure:"secret" yaml:"secret"`
}

// AppConfig 应用配置
type AppConfig struct {
	RunMode string `mapstructure:"runmode" yaml:"runmode"`
}

// globalConfig 全局配置实例
var globalConfig *Config

// Init 初始化配置
func Init(cfgFile string) error {
	if err := initViper(cfgFile); err != nil {
		return err
	}

	// 设置默认值
	setDefaults()

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		// 配置文件不存在时不报错，使用默认值
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
	}

	// 解析配置到结构体
	globalConfig = &Config{}
	if err := viper.Unmarshal(globalConfig); err != nil {
		return err
	}

	return nil
}

// initViper 初始化viper配置
func initViper(cfgFile string) error {
	if cfgFile != "" {
		// 从命令行选项指定的配置文件中读取
		viper.SetConfigFile(cfgFile)
	} else {
		// 查找用户主目录
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}

		// 将用 `$HOME/<RecommendedHomeDir>` 目录加入到配置文件的搜索路径中
		viper.AddConfigPath(filepath.Join(home, RecommendedHomeDir))

		// 把当前目录加入到配置文件的搜索路径中
		viper.AddConfigPath(".")

		// 设置配置文件格式为 YAML
		viper.SetConfigType("yaml")

		// 配置文件名称（没有文件扩展名）
		viper.SetConfigName(DefaultConfigName)
	}

	// 读取匹配的环境变量
	viper.AutomaticEnv()

	// 读取环境变量的前缀为 DEMO520
	viper.SetEnvPrefix("DEMO520")

	// 将 viper.Get(key) key 字符串中 '.' 和 '-' 替换为 '_'
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)

	return nil
}

// setDefaults 设置默认配置值
func setDefaults() {
	// 服务器配置默认值
	viper.SetDefault("server.addr", ":8080")

	// 数据库配置默认值
	viper.SetDefault("db.host", "127.0.0.1:3306")
	viper.SetDefault("db.username", "root")
	viper.SetDefault("db.password", "")
	viper.SetDefault("db.database", "demo520")
	viper.SetDefault("db.max-idle-connections", 10)
	viper.SetDefault("db.max-open-connections", 100)
	viper.SetDefault("db.max-connection-life-time", "10s")
	viper.SetDefault("db.log-level", 1)

	// 日志配置默认值
	viper.SetDefault("log.disable-caller", false)
	viper.SetDefault("log.disable-stacktrace", false)
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.encoding", "console")
	viper.SetDefault("log.output-paths", []string{"stdout"})

	// 图片配置默认值
	viper.SetDefault("image.ImageMaxSize", int64(20*1024*1024)) // 20 MB
	viper.SetDefault("image.image_dir", "temp_image")
	viper.SetDefault("image.Convert.WebPQuality", 80)
	viper.SetDefault("image.Convert.WebReductionEffort", 4)
	viper.SetDefault("image.Convert.AvifQuality", 60)
	viper.SetDefault("image.Convert.AvifEffort", 4)
	viper.SetDefault("image.Convert.ImageLossless", false)

	// JWT配置默认值
	viper.SetDefault("jwt.secret", "demo520-secret-key")

	// 应用配置默认值
	viper.SetDefault("app.runmode", "debug")
}

// GetConfig 获取全局配置
func GetConfig() *Config {
	return globalConfig
}

// GetServer 获取服务器配置
func GetServer() *ServerConfig {
	return &globalConfig.Server
}

// GetDB 获取数据库配置
func GetDB() *DBConfig {
	return &globalConfig.DB
}

// GetLog 获取日志配置
func GetLog() *LogConfig {
	return &globalConfig.Log
}

// GetImage 获取图片配置
func GetImage() *ImageConfig {
	return &globalConfig.Image
}

// GetJWT 获取JWT配置
func GetJWT() *JWTConfig {
	return &globalConfig.JWT
}

// GetApp 获取应用配置
func GetApp() *AppConfig {
	return &globalConfig.App
}

// GenerateDefaultConfig 生成默认配置文件
func GenerateDefaultConfig(filePath string) error {
	// 设置默认值
	setDefaults()

	// 创建配置实例
	defaultConfig := &Config{
		Server: ServerConfig{
			Addr: viper.GetString("server.addr"),
		},
		DB: DBConfig{
			Host:                  viper.GetString("db.host"),
			Username:              viper.GetString("db.username"),
			Password:              viper.GetString("db.password"),
			Database:              viper.GetString("db.database"),
			MaxIdleConnections:    viper.GetInt("db.max-idle-connections"),
			MaxOpenConnections:    viper.GetInt("db.max-open-connections"),
			MaxConnectionLifeTime: viper.GetDuration("db.max-connection-life-time"),
			LogLevel:              viper.GetInt("db.log-level"),
		},
		Log: LogConfig{
			DisableCaller:     viper.GetBool("log.disable-caller"),
			DisableStacktrace: viper.GetBool("log.disable-stacktrace"),
			Level:             viper.GetString("log.level"),
			Encoding:          viper.GetString("log.encoding"),
			OutputPaths:       viper.GetStringSlice("log.output-paths"),
		},
		Image: ImageConfig{
			ImageMaxSize: viper.GetInt64("image.ImageMaxSize"),
			ImageDir:     viper.GetString("image.image_dir"),
			Convert: ConvertConfig{
				WebPQuality:        viper.GetInt("image.Convert.WebPQuality"),
				WebReductionEffort: viper.GetInt("image.Convert.WebReductionEffort"),
				AvifQuality:        viper.GetInt("image.Convert.AvifQuality"),
				AvifEffort:         viper.GetInt("image.Convert.AvifEffort"),
				ImageLossless:      viper.GetBool("image.Convert.ImageLossless"),
			},
		},
		JWT: JWTConfig{
			Secret: viper.GetString("jwt.secret"),
		},
		App: AppConfig{
			RunMode: viper.GetString("app.runmode"),
		},
	}

	// 将配置写入viper
	viper.Set("server", defaultConfig.Server)
	viper.Set("db", defaultConfig.DB)
	viper.Set("log", defaultConfig.Log)
	viper.Set("image", defaultConfig.Image)
	viper.Set("jwt", defaultConfig.JWT)
	viper.Set("app", defaultConfig.App)

	// 写入文件
	return viper.WriteConfigAs(filePath)
}