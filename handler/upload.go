package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
)

type configuration struct {
	Port     string
	FilePath string
}

//处理文件上传
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	file, _ := os.Open("server.json")
	defer file.Close()
	decoder := json.NewDecoder(file)
	conf := configuration{}
	err := decoder.Decode(&conf)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	if r.Method == http.MethodPost {
		r.ParseForm()
		//接收文件流及存储到本地目录
		file, head, err := r.FormFile("file")
		if err != nil {
			fmt.Printf("Failed to get data,err:%s\n", err.Error())
			return
		}
		defer file.Close()

		fileName := head.Filename
		split := strings.Split(fileName, "-")
		filePath := conf.FilePath + "/" + strings.Join(split, "/")
		err = os.MkdirAll(path.Dir(filePath), 0755)
		if err != nil {
			fmt.Println(err)
			return
		}

		location := filePath

		newFile, err := os.Create(location)
		if err != nil {
			fmt.Printf("Failed to create file,err:%s\n", err.Error())
			return
		}
		defer newFile.Close()
		io.Copy(newFile, file)
	}
}
