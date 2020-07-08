package main

import (
	"bytes"
	"github.com/go-toast/toast"
	"github.com/go-vgo/robotgo"
	"github.com/robotn/gohook"
	"imgupload/conf/client"
	"imgupload/util"
	"imgupload/clipboard"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"strings"
)

var conf = client.Conf

func main() {

	//设置log级别
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// 读取配置文件
	err := util.LoadConf("conf/client.json", conf)
	if err != nil {
		log.Println("Error: " , err)
		return
	}

	showToast("程序启动成功")
	// 根据配置的组别启动程序
	for _, groupInfo := range conf.Groups  {
		keys := strings.Split(groupInfo.ShortcutKey, "+")
		groupName := groupInfo.GroupName
		robotgo.EventHook(hook.KeyDown, keys, func(e hook.Event) {
			doUpload(groupName)
		})
	}
	log.Println("程序启动成功")
	s := robotgo.EventStart()
	<-robotgo.EventProcess(s)
	showToast("程序退出！")
}

func doUpload(group string) {
	fileData, err := clipboard.ReadClipboard()
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
	clipboard.ImgCopy(&fileWriter, fileData)
	bodyWriter.WriteField("group", group)

	contentType := bodyWriter.FormDataContentType()
	// 必须在这里显式关闭
	bodyWriter.Close()
	// 创建一个http客户端
	uploadClient := http.Client{}
	request, _ := http.NewRequest(http.MethodPost,
		conf.ServerUrl+"/imgGo/upload", bodyBuf)
	request.Header.Set("access", util.Sha2(conf.Auth))
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
	log.Println("上传成功" + respBytes.String())
	robotgo.WriteAll(conf.ServerUrl + respBytes.String())

	showToast(group + ": 图片上传成功")
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

