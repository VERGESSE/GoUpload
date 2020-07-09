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

	// 图片新增接口
	http.HandleFunc("/imgGo/upload", handler.UploadHandler)
	// 剪裁图片接口
	http.HandleFunc("/imgGo/thumb", handler.ThumbImgHandler)
	// 图片删除接口
	http.HandleFunc("/imgGo/delete", handler.DeleteImgHandler)

	//静态资源访问
	http.Handle("/imgGo/", http.StripPrefix("/imgGo/", http.FileServer(http.Dir(conf.FilePath))))

	log.Println("ImgGo 服务端程序启动成功！")
	log.Fatal(http.ListenAndServe(conf.Port, nil))
}

