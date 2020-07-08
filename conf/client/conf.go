package client

var Conf = &Config{}

// 客户端配置
type Config struct {
	// 后端服务器地址
	ServerUrl string
	// 权限
	Auth string
	// 分组
	Groups []struct {
		GroupName, ShortcutKey string
	}
}
