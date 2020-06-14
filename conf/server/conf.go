package server

var Conf = &Config{}

// 配置文件结构体
type Config struct {
	Port     string
	FilePath string
	Auth     string
}
