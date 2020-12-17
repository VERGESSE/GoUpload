package main

import (
	"bytes"
	"github.com/go-toast/toast"
	"github.com/go-vgo/robotgo"
	"github.com/robotn/gohook"
	"imgupload/clipboard"
	"imgupload/conf/client"
	"imgupload/util"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"strings"
)

var conf = client.Conf

// 存储已经失效的 groupName
var oldGroup = make(map[string]bool, 10)

func main() {

	//设置log级别
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// 配置文件路径
	confFile := "conf/client.json"
	// 读取配置文件
	err := util.LoadConf(confFile, conf)
	if err != nil {
		log.Println("Error: " , err)
		return
	}

	// 设置文件监听器函数，配置文件修改时立即重新加载配置
	go util.FileUpDateListener(confFile, func() {

			// 把旧的 Group 标记为失效
			for _, groupInfo := range conf.Groups  {
				oldGroup[groupInfo.GroupName + groupInfo.ShortcutKey] = true
			}

			// 重新加载配置文件
			err := util.LoadConf(confFile, conf)
			if err != nil {
				log.Println("Error: " , err)
				return
			}
			showToast("新配置加载完成")
			// 根据配置的组别重新启动程序
			for _, groupInfo := range conf.Groups  {
				// 获取快捷键
				keys := strings.Split(groupInfo.ShortcutKey, "+")
				// 解除之前的修改的配置标记失效
				oldGroup[groupInfo.GroupName + groupInfo.ShortcutKey] = false
				groupName := groupInfo.GroupName
				// 设置指定文件名和快捷键的监听
				robotgo.EventHook(hook.KeyDown, keys,
					func(e hook.Event) {
						// 启动文件上传程序
						doUpload(groupName, groupInfo.ShortcutKey)
					})
			}
		})

	showToast("程序启动成功")
	// 根据配置的组别启动程序
	for _, groupInfo := range conf.Groups  {
		// 获取快捷键
		keys := strings.Split(groupInfo.ShortcutKey, "+")
		groupName := groupInfo.GroupName
		// 设置指定文件名和快捷键的监听
		robotgo.EventHook(hook.KeyDown, keys,
			func(e hook.Event) {
				// 启动文件上传程序
				doUpload(groupName, groupInfo.ShortcutKey)
		})
	}
	log.Println("程序启动成功")
	s := robotgo.EventStart()
	// 阻塞程序, 使程序不主动退出
	<-robotgo.EventProcess(s)
	showToast("程序退出！")
}

func doUpload(group, groupKey string) {
	// 如果 Group 失效则不执行
	if oldGroup[group + groupKey] {
		return
	}

	// 获取剪切板中的图片数据
	fileData, err := clipboard.ReadClipboard()
	if err != nil {
		showToast("图片上传失败")
		log.Println(err)
		return
	}
	// 声明 http 上传的数据
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	// 服务器文件路径
	serverUrl := conf.ServerUrl
	// 定义上传至服务器的 Writer
	fileWriter := new(io.Writer)
	// 判断上传的是文件还是截图
	if fileData.Img {
		// 上传的是截图
		*fileWriter, err = bodyWriter.CreateFormFile("file", "file")
		serverUrl += "/imgGo/upload"
		// 将文件数据拷贝到fileWriter中
		clipboard.ImgCopy(fileWriter, fileData.Data)
	} else {
		// 上传的额是文件
		*fileWriter, err = bodyWriter.CreateFormFile("file", fileData.FileName)
		serverUrl += "/imgGo/uploadFile"
		io.Copy(*fileWriter, fileData.Data)
	}
	// 判断上面执行的代码是否有错误
	if err != nil {
		log.Println(err)
		showToast("图片上传失败")
		return
	}

	bodyWriter.WriteField("group", group)

	contentType := bodyWriter.FormDataContentType()
	// 必须在这里显式关闭
	bodyWriter.Close()
	// 创建一个http客户端
	uploadClient := http.Client{}
	//向 request 设置服务器地址
	request, _ := http.NewRequest(http.MethodPost, serverUrl, bodyBuf)
	// 写入通行证
	request.Header.Set("access", util.Sha2(conf.Auth))
	request.Header.Set("Content-Type", contentType)
	// 发起文件上传请求
	resp, err := uploadClient.Do(request)
	if err != nil {
		log.Println(err)
		showToast("图片上传失败")
		return
	}
	defer resp.Body.Close()
	var respBytes = &bytes.Buffer{}
	// 从返回数据中解析文件名
	io.Copy(respBytes,resp.Body)
	if err != nil {
		log.Println(err)
		showToast("图片上传失败")
		return
	}
	log.Println("上传成功" + respBytes.String())
	// 向剪切板写入可访问文件路径
	robotgo.WriteAll(conf.ServerUrl + respBytes.String())

	// 告知用户上传成功
	showToast(group + ": 图片上传成功")
}

// win10的右下角通知器
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

