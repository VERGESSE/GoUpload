package util

import (
	"encoding/json"
	"github.com/fsnotify/fsnotify"
	"io/ioutil"
	"log"
	"path"
	"runtime"
	"sync"
)

var lock = sync.Mutex{}

// 读取配置文件方法
func LoadConf(path string, conf interface{}) error {
	// 同步方法 同时只允许一个线程修改配置信息
	lock.Lock()
	defer lock.Unlock()

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

// 监听文件修改的
// filename 为要监听的文件名称
// hook为要执行的回调函数
// 重要  需要异步调用此函数，否则会阻塞主程序
func FileUpDateListener(filename string, hook func()) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal("文件监听器启动失败: ", err)
	}
	defer watcher.Close()

	// 对所传文件的绝对路径进行监听
	err = watcher.Add(GetCurrPath() + "/" + filename)
	if err != nil {
		log.Fatal("文件监听器启动失败:", err)
	}

	done := make(chan struct{})

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				// 只对文件修改感兴趣
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("监听到文件修改，文件名: ", filename)
					// 执行钩子
					hook()
				}
			case err := <-watcher.Errors:
				log.Println("error:", err)
				// 退出文件监听
				done <- struct{}{}
				return
			}
		}
		// 退出文件监听
		done <- struct{}{}
	}()

	// 阻塞程序
	<-done
}

// 获取当前项目的绝对路径
func GetCurrPath() string {
	var abPath string
	_, filename, _, ok := runtime.Caller(1)
	if ok {
		abPath = path.Dir(path.Dir(filename))
	}
	return abPath
}
