package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-toast/toast"
	"github.com/go-vgo/robotgo"
	"imgupload/util"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"
)

type configuration struct {
	ServerUrl 		string
	ServerImgUrl    string
}

func main() {

	file, _ := os.Open("client.json")
	defer file.Close()
	decoder  := json.NewDecoder(file)
	conf := configuration{}
	err := decoder.Decode(&conf)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	serverUrl := conf.ServerUrl
	serverImgUrl := conf.ServerImgUrl

	showToast("程序启动成功")

	for  {
		keve := robotgo.AddEvents("ctrl","space")
		if keve {
			data, err := util.ReadClipboard()
			if err != nil {
				showToast("图片上传失败")
				continue
			}

			now := time.Now()
			s := now.Format("2006-01-02-15_04_05")
			fileName := s +".png"
			util.SaveAs(data, fileName)

			file, _ := os.Open(fileName)

			bodyBuf := &bytes.Buffer{}
			bodyWriter := multipart.NewWriter(bodyBuf)
			//关键的一步操作
			fileWriter, err := bodyWriter.CreateFormFile("file", fileName)
			if err != nil {
				showToast("图片上传失败")
				file.Close()
				continue
			}

			io.Copy(fileWriter, file)
			file.Close()

			contentType := bodyWriter.FormDataContentType()
			bodyWriter.Close()

			_, err = http.Post(serverUrl+"/file/upload",
				contentType, bodyBuf)
			if err != nil {
				showToast("图片上传失败")
				os.Remove(fileName)
				continue
			}

			split := strings.Split(fileName, "-")
			filePath := serverImgUrl + "/" + strings.Join(split, "/")

			robotgo.WriteAll(filePath)
			os.Remove(fileName)

			showToast("图片上传成功")
		}
	}
}

func showToast(message string) {
	notification := toast.Notification{
		Title:   "ImgGo",
		Message: message,
		//Icon:    "/logo.png", // 文件必须存在
	}
	notification.Push()
}

