package util

import (
	"encoding/json"
	"os"
)

// 读取配置文件方法
func LoadConf(path string, conf interface{}) error {
	file, _ := os.Open(path)
	defer file.Close()
	decoder := json.NewDecoder(file)
	err := decoder.Decode(conf)
	if err != nil {
		return err
	}
	return nil
}
