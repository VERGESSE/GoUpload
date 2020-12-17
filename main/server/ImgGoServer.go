package main

import (
	"imgupload/conf/server"
	"imgupload/handler"
	"imgupload/util"
	"log"
	"net/http"
)

func main() {

	//设置log级别
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// 读取配置文件
	conf := server.Conf
	const confFile = "conf/server.json"
	err := util.LoadConf(confFile, conf)
	if err != nil {
		log.Println("Error: " , err)
		return
	}

	// 设置文件监听器函数，配置文件修改时立即重新加载配置
	go util.FileUpDateListener(confFile, func() {
			// 重新加载配置文件
			err := util.LoadConf(confFile, conf)
			if err != nil {
				log.Println("Error: " , err)
				return
			}
			log.Println("服务端配置文件修改: ", confFile)
		})

	// 截图新增接口
	http.HandleFunc("/imgGo/upload", handler.UploadHandler)
	// 文件夹新增接口
	http.HandleFunc("/imgGo/uploadFile", handler.UploadFileHandler)
	// 剪裁图片接口
	http.HandleFunc("/imgGo/thumb", handler.ThumbImgHandler)
	// 图片删除接口
	http.HandleFunc("/imgGo/delete", handler.DeleteImgHandler)

	//静态资源访问
	http.Handle("/imgGo/", http.StripPrefix("/imgGo/", http.FileServer(http.Dir(conf.FilePath))))

	log.Println("ImgGo 服务端程序启动成功！")
	log.Fatal(http.ListenAndServe(conf.Port, nil))
}

