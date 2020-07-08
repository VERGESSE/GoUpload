package util

import (
	"encoding/json"
	"io/ioutil"
)

// 读取配置文件方法
func LoadConf(path string, conf interface{}) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, conf)
	if err != nil {
		return err
	}
	return nil
}
