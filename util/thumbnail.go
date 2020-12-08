package util

import (
	"fmt"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// 将给定的输入流按照 pixel 进行剪裁 并返回新的输入流
func Image(src image.Image, pixel int) image.Image {
	// 获取按 pixel 剪裁后新图片的长宽
	xs := src.Bounds().Size().X
	ys := src.Bounds().Size().Y
	width, height := pixel, pixel
	if aspect := float64(xs) / float64(ys); aspect < 1.0 {
		width = int(float64(pixel) * aspect) // portrait
	} else {
		height = int(float64(pixel) / aspect) // landscape
	}
	xscale := float64(xs) / float64(width)
	yscale := float64(ys) / float64(height)

	dst := image.NewRGBA(image.Rect(0, 0, width, height))

	// 读取每个像素进行压缩
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			// 获取压缩后图片 x y 位置的原图片的对应像素,进行填充
			srcx := int(float64(x) * xscale)
			srcy := int(float64(y) * yscale)
			dst.Set(x, y, src.At(srcx, srcy))
		}
	}
	return dst
}

// 从 r 中读取图片数据,并按 pixel 剪裁后 写入到 w 输出流中
func ImageStream(w io.Writer, r io.Reader, pixel int) error {
	// 将输入流按照图片进行编码 支持格式为 jpg png gif
	src, _, err := image.Decode(r)
	if err != nil {
		log.Println(err)
		return err
	}
	// 根据输入生成剪裁后的图片流
	dst := Image(src, pixel)
	// 将图片流写入到输出流
	return jpeg.Encode(w, dst, nil)
}

// 创建 outfile 文件名的文件, 将 infile
// 按 pixel 剪裁后写入到 outfile 中
func ImageFile(outfile, infile string, pixel int) (err error) {
	in, err := os.Open(infile)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(outfile)
	if err != nil {
		return err
	}

	if err := ImageStream(out, in, pixel); err != nil {
		out.Close()
		return fmt.Errorf("scaling %s to %s: %s", infile, outfile, err)
	}
	return out.Close()
}

// 剪裁图片
// infile 要剪裁的图片的本机地址
// pixel 剪裁后的图片像素
func ThumbImage(infile string, pixel int) (string, error) {
	pixelsStr := strconv.Itoa(pixel)
	suffix := "_" + pixelsStr + "x" + pixelsStr
	// 获取文件扩展名
	ext := filepath.Ext(infile)
	// 拼接剪裁后的文件名
	outfile := strings.TrimSuffix(infile, ext) + suffix + ext
	return outfile, ImageFile(outfile, infile, pixel)
}
