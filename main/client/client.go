package main

import (
	"bytes"
	"github.com/go-toast/toast"
	"github.com/go-vgo/robotgo"
	"imgupload/conf/client"
	"imgupload/util"
	"io"
	"log"
	"mime/multipart"
	"net/http"
)

var conf = client.Conf

func main() {

	//设置log级别6
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// 读取配置文件
	err := util.LoadConf("conf/client.json", conf)
	if err != nil {
		log.Println("Error: " , err)
		return
	}

	showToast("程序启动成功")

	for {
		ok := robotgo.AddEvents("ctrl","space")
		if ok {
			doUpload()
		}
	}
}

func doUpload() {
	fileData, err := util.ReadClipboard()
	if err != nil {
		showToast("图片上传失败")
		return
	}

	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)
	//关键的一步操作
	fileWriter, err := bodyWriter.CreateFormFile("file", "file")
	if err != nil {
		log.Println(err)
		showToast("图片上传失败")
		return
	}
	// 将文件数据拷贝到fileWriter中
	io.Copy(fileWriter, fileData)

	contentType := bodyWriter.FormDataContentType()
	// 必须在这里显式关闭
	bodyWriter.Close()
	// 创建一个http客户端
	uploadClient := http.Client{}
	request, _ := http.NewRequest(http.MethodPost, conf.ServerUrl+"/figureBed/upload", bodyBuf)
	request.Header.Set("access", conf.Auth)
	request.Header.Set("Content-Type", contentType)
	resp, err := uploadClient.Do(request)
	if err != nil {
		log.Println(err)
		showToast("图片上传失败")
		return
	}
	defer resp.Body.Close()
	var respBytes = &bytes.Buffer{}
	io.Copy(respBytes,resp.Body)
	if err != nil {
		log.Println(err)
		showToast("图片上传失败")
		return
	}
	log.Println(respBytes.String())
	robotgo.WriteAll(conf.ServerUrl + respBytes.String())

	showToast("图片上传成功")
}

func showToast(message string) {
	notification := toast.Notification{
		Title:   "ImgGo",
		Message: message,
	}
	err := notification.Push()
	if err != nil {
		log.Println(err)
	}
}

