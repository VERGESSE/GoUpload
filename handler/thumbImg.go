package handler

import (
	"imgupload/util"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const (
	NOTFOUNT    = "1" // 请求路径图片不存在
	THUMBERR    = "3" // 图片剪裁失败
	PIXELTOOBIG = "5" // 图片剪裁失败
)

//处图片剪裁请求
func ThumbImgHandler(w http.ResponseWriter, r *http.Request) {
	access := r.Header.Get("access")
	if access != util.Sha2(conf.Auth) {
		w.Write([]byte("Illegal request"))
		return
	}
	// 只接受POST请求
	if r.Method == http.MethodPost {
		r.ParseForm()
		imgPath := r.FormValue("img")
		pixel, _ := strconv.Atoi(r.FormValue("pixel"))
		// 剪裁尺寸不宜过大
		if pixel > 10000 {
			w.Write([]byte(PIXELTOOBIG))
			return
		}

		pathSlice := strings.Split(imgPath, "/")
		if len(pathSlice) < 4 {
			w.Write([]byte(NOTFOUNT))
			return
		}
		imgPath = conf.FilePath + "/" +
			strings.Join(pathSlice[len(pathSlice)-4:], "/")
		// 检验图片是否存在
		img, err := os.Open(imgPath)
		if err != nil {
			log.Println(err)
			w.Write([]byte(NOTFOUNT))
			return
		}
		img.Close()
		// 开始剪裁
		thumbImgPath, err := util.ThumbImage(imgPath, pixel)
		if err != nil {
			log.Println(err)
			w.Write([]byte(THUMBERR))
			return
		}
		log.Println("图片剪裁完成: " + thumbImgPath)
		thumbImgPath = "/imgGo" + strings.TrimPrefix(thumbImgPath, conf.FilePath)
		w.Write([]byte(thumbImgPath))
	} else {
		w.Write([]byte("Request Must Be Post"))
	}
}
