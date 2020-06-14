package handler

import (
	"imgupload/conf/server"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

var conf = server.Conf

//处理文件上传
func UploadHandler(w http.ResponseWriter, r *http.Request) {

	access := r.Header.Get("access")
	if access != conf.Auth {
		w.Write([]byte("Illegal request"))
		return
	}
	// 只接受POST请求
	if r.Method == http.MethodPost {
		r.ParseForm()

		file, _, err := r.FormFile("file")
		if err != nil {
			log.Printf("Failed to get data,err:%s\n", err.Error())
			return
		}
		defer file.Close()

		now := time.Now()
		createTime := now.Format("2006-01-02")

		// 以时间为策略创建文件名
		dir := strings.Split(createTime, "-")
		fileName := strings.Join(dir, "/") + "/" + strconv.Itoa(int(time.Now().Unix())) + ".png"
		filePath := conf.FilePath + "/" + fileName
		err = os.MkdirAll(path.Dir(filePath), 0755)
		if err != nil {
			log.Println(err)
			return
		}

		newFile, err := os.Create(filePath)
		if err != nil {
			log.Printf("Failed to create file,err:%s\n", err.Error())
			return
		}
		defer newFile.Close()
		_, err = io.Copy(newFile, file)
		if err != nil {
			log.Println(err)
			return
		}

		log.Println("上传成功: " + filePath)
		w.Write([]byte("/img/" + fileName))
	} else {
		w.Write([]byte("Illegal request"))
	}
}
