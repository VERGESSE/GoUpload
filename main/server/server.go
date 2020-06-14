package main

import (
	"log"
	"net/http"

	"imgupload/conf/server"
	"imgupload/handler"
	"imgupload/util"
)

func main() {

	//设置log级别
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// 读取配置文件
	conf := server.Conf
	err := util.LoadConf("conf/server.json", conf)
	if err != nil {
		log.Println("Error: " , err)
		return
	}

	//静态资源访问
	http.Handle("/img/", http.StripPrefix("/img/", http.FileServer(http.Dir(conf.FilePath))))
	// 文件操作接口
	http.HandleFunc("/figureBed/upload", handler.UploadHandler)
	log.Fatal(http.ListenAndServe(conf.Port, nil))
}

