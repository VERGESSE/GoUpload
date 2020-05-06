package main

import (
	"encoding/json"
	"fmt"
	"imgupload/handler"
	"net/http"
	"os"
)

type configuration struct {
	Port 		string
	FilePath    string
}

func main() {
	file, _ := os.Open("server.json")
	defer file.Close()
	decoder  := json.NewDecoder(file)
	conf := configuration{}
	err := decoder.Decode(&conf)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	// 文件操作接口
	http.HandleFunc("/file/upload", handler.UploadHandler)
	err = http.ListenAndServe(conf.Port, nil)
	if err != nil {
		fmt.Printf("Failed to start server,err:%s", err.Error())
	}
}