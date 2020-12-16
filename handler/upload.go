package handler

import (
	"imgupload/conf/server"
	"imgupload/util"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"
)

var conf = server.Conf

//处理截图上传
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	// 权限验证
	access := r.Header.Get("access")
	if access != util.Sha2(conf.Auth) {
		w.Write([]byte("Illegal request"))
		return
	}
	// 只接受POST请求
	if r.Method == http.MethodPost {
		r.ParseForm()
		// 解析输入文件信息
		file, _, err := r.FormFile("file")
		group := r.FormValue("group")
		if err != nil {
			log.Printf("Failed to get data,err:%s\n", err.Error())
			return
		}
		defer file.Close()
		now := time.Now()
		// 格式化当前时间日期
		createTime := now.Format("2006-0102")
		// 以时间为策略创建文件名
		dir := strings.Split(createTime, "-")
		fileName := group + "/" + strings.Join(dir, "/") + "/" +
			getId() + ".png"
		filePath := conf.FilePath + "/" + fileName
		// 创建文件路径
		err = os.MkdirAll(path.Dir(filePath), 0755)
		if err != nil {
			log.Println(err)
			return
		}
		// 创建并写入文件
		newFile, err := os.Create(filePath)
		if err != nil {
			log.Printf("Failed to create file,err:%s\n", err.Error())
			return
		}
		defer newFile.Close()
		// 向新文件, 写入输入
		_, err = io.Copy(newFile, file)
		if err != nil {
			log.Println(err)
			return
		}

		log.Println("上传成功: " + filePath)
		// 返回文件地址
		w.Write([]byte("/imgGo/" + fileName))
	} else {
		w.Write([]byte("Request Must Be Post"))
	}
}

//处理文件上传
func UploadFileHandler(w http.ResponseWriter, r *http.Request) {
	// 权限验证
	access := r.Header.Get("access")
	if access != util.Sha2(conf.Auth) {
		w.Write([]byte("Illegal request"))
		return
	}
	// 只接受POST请求
	if r.Method == http.MethodPost {
		r.ParseForm()
		// 解析输入文件信息
		file, fileinfo, err := r.FormFile("file")
		group := r.FormValue("group")
		if err != nil {
			log.Printf("Failed to get data,err:%s\n", err.Error())
			return
		}
		defer file.Close()
		now := time.Now()
		// 格式化当前时间日期
		createTime := now.Format("2006-0102")
		// 以时间为策略创建文件名
		dir := strings.Split(createTime, "-")
		// 从 文件信息中解析出文件名 并拼接文件全路径
		fileName := group + "/" + strings.Join(dir, "/") +
			"/" + fileinfo.Filename
		filePath := conf.FilePath + "/" + fileName
		// 创建文件路径
		err = os.MkdirAll(path.Dir(filePath), 0755)
		if err != nil {
			log.Println(err)
			return
		}
		// 创建并写入文件
		newFile, err := os.Create(filePath)
		if err != nil {
			log.Printf("Failed to create file,err:%s\n", err.Error())
			return
		}
		defer newFile.Close()
		// 向新文件, 写入输入
		_, err = io.Copy(newFile, file)
		if err != nil {
			log.Println(err)
			return
		}

		log.Println("上传成功: " + filePath)
		// 返回文件地址
		w.Write([]byte("/imgGo/" + fileName))
	} else {
		w.Write([]byte("Request Must Be Post"))
	}
}

// 处理文件删除
// 输入的图片路径需为可直接访问的完整地址
func DeleteImgHandler(w http.ResponseWriter, r *http.Request) {
	// 权限验证
	access := r.Header.Get("access")
	if access != util.Sha2(conf.Auth) {
		w.Write([]byte("Illegal request"))
		return
	}
	// 只接受POST请求
	if r.Method == http.MethodPost {
		r.ParseForm()
		// 获取要删除的图片
		imgPath := r.FormValue("imgPath")
		// 计算图片地址
		pathSlice := strings.Split(imgPath, "/")
		if len(pathSlice) < 4 {
			w.Write([]byte(NOTFOUNT))
			return
		}
		imgPath = conf.FilePath + "/" +
			strings.Join(pathSlice[len(pathSlice)-4:], "/")
		os.Remove(imgPath)
		log.Println("图片删除: " + imgPath)
	} else {
		w.Write([]byte("Request Must Be Post"))
	}
}

// 解决高并发下图片覆盖问题，每毫秒仅允许生成一个Id
var lock sync.Mutex

// 毫秒级别ID生成器
// 保证高并发时不产生图片覆盖行为
func getId() string {
	lock.Lock()
	defer lock.Unlock()
	id := strconv.Itoa(int(time.Now().UnixNano() / 1000))
	time.Sleep(time.Microsecond)
	return id
}
