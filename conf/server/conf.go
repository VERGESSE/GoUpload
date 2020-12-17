package server

var Conf = &Config{}

// 配置文件结构体
type Config struct {
	// 服务器 端口号
	Port string
	// 存储上传文件的地址
	FilePath string
	// 权限密码
	Auth string
}
