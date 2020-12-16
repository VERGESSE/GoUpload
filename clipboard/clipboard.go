package clipboard

import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/lxn/win"
	"golang.org/x/image/bmp"
	"image/jpeg"
	"io"
	"log"
	"os"
	"strings"
	"syscall"
	"unsafe"
)

const (
	cfBitmap      = 2
	cfDib         = 8
	cfUnicodetext = 13
	cFHDROP       = 15
	cfDibV5       = 17
	gmemFixed     = 0x0000
)

type FileHeader struct {
	bfType      uint16
	bfSize      uint32
	bfReserved1 uint16
	bfReserved2 uint16
	bfOffBits   uint32
}

type infoHeader struct {
	iSize          uint32
	iWidth         uint32
	iHeight        uint32
	iPLanes        uint16
	iBitCount      uint16
	iCompression   uint32
	iSizeImage     uint32
	iXPelsPerMeter uint32
	iYPelsPerMeter uint32
	iClrUsed       uint32
	iClrImportant  uint32
}

// 存储要上传数据的结构体
type FileInfo struct {
	// 要上传的数据
	Data *bytes.Buffer
	// 文件名
	FileName string
	// 截图为true 文件为false
	Img bool
}

var (
	user32                     = syscall.MustLoadDLL("user32")
	openClipboard              = user32.MustFindProc("OpenClipboard")
	closeClipboard             = user32.MustFindProc("CloseClipboard")
	emptyClipboard             = user32.MustFindProc("EmptyClipboard")
	getClipboardData           = user32.MustFindProc("GetClipboardData")
	setClipboardData           = user32.MustFindProc("SetClipboardData")
	isClipboardFormatAvailable = user32.MustFindProc("IsClipboardFormatAvailable")

	kernel32     = syscall.NewLazyDLL("kernel32")
	globalAlloc  = kernel32.NewProc("GlobalAlloc")
	globalFree   = kernel32.NewProc("GlobalFree")
	globalLock   = kernel32.NewProc("GlobalLock")
	globalUnlock = kernel32.NewProc("GlobalUnlock")
	lstrcpy      = kernel32.NewProc("lstrcpyW")
	copyMemory   = kernel32.NewProc("CopyMemory")
)

func CopyInfoHdr(dst *byte, psrc *infoHeader) (string, error) {

	pdst := (*infoHeader)(unsafe.Pointer(dst))

	pdst.iSize = psrc.iSize
	pdst.iWidth = psrc.iWidth
	pdst.iHeight = psrc.iHeight
	pdst.iPLanes = psrc.iPLanes
	pdst.iBitCount = psrc.iBitCount
	pdst.iCompression = psrc.iCompression
	pdst.iSizeImage = psrc.iSizeImage
	pdst.iXPelsPerMeter = psrc.iXPelsPerMeter
	pdst.iYPelsPerMeter = psrc.iYPelsPerMeter
	pdst.iClrUsed = psrc.iClrUsed
	pdst.iClrImportant = psrc.iClrImportant

	return "copy infoHeader success", nil
}

func ReadUint16(b []byte) uint16 {
	return uint16(b[0]) | uint16(b[1])<<8
}

func ReadUint32(b []byte) uint32 {
	return uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
}

// 获取剪切版的截图数据流
func ReadClipboard() (*FileInfo, error) {
	const (
		fileHeaderLen = 14
		infoHeaderLen = 40
	)

	r, _, err := openClipboard.Call(0)
	if r == 0 {
		return nil, err
	}
	defer closeClipboard.Call()

	r, _, err = isClipboardFormatAvailable.Call(cfDib)
	if r == 0 {
		return ReadClipboardFile()
	}

	h, _, err := getClipboardData.Call(cfDib)
	if r == 0 {
		return nil, err
	}

	pdata, _, err := globalLock.Call(h)
	if pdata == 0 {
		return nil, err
	}
	defer globalUnlock.Call(h)

	h2 := (*infoHeader)(unsafe.Pointer(pdata))

	//	fmt.Println(h2)
	dataSize := h2.iSizeImage + fileHeaderLen + infoHeaderLen

	if h2.iSizeImage == 0 && h2.iCompression == 0 {
		iSizeImage := h2.iHeight * ((h2.iWidth*uint32(h2.iBitCount)/8 + 3) &^ 3)
		dataSize += iSizeImage
	}

	data := new(bytes.Buffer)
	binary.Write(data, binary.LittleEndian, uint16('B')|(uint16('M')<<8))
	binary.Write(data, binary.LittleEndian, uint32(dataSize))
	binary.Write(data, binary.LittleEndian, uint32(0))
	const sizeofColorbar = 0
	binary.Write(data, binary.LittleEndian, uint32(fileHeaderLen+infoHeaderLen+sizeofColorbar))
	j := 0
	for i := fileHeaderLen; i < int(dataSize); i++ {
		binary.Write(data, binary.BigEndian, *(*byte)(unsafe.Pointer(pdata + uintptr(j))))
		j++
	}

	// 构造返回的文件数据
	info := &FileInfo{Data: data, FileName: "img", Img: true}
	return info, nil
}

// 获取剪切板文件路径的数据流
func ReadClipboardFile() (*FileInfo, error) {
	r, _, err := isClipboardFormatAvailable.Call(cFHDROP)
	if r == 0 {
		log.Println(err)
		return nil, err
	}

	h, _, err := getClipboardData.Call(cFHDROP)
	if r == 0 {
		return nil, err
	}

	// 获取剪切板数据指针
	h2 := (win.HDROP)(unsafe.Pointer(h))
	// 获取文件路径
	var filePath0 = make([]uint16, 1000)

	// 不支持文件夹
	if fileNum := win.DragQueryFile(h2, 0xFFFFFFFF, nil, 0); fileNum > 1 {
		return nil, errors.New("当前不支持文件夹操作")
	}

	// 获取文件名长度
	fileLen := win.DragQueryFile(h2, 0, &(filePath0[0]), 1000)
	fileByte := Int16SliceToByte(filePath0[:fileLen])
	// 获取文件名
	filePath := string(fileByte)

	// 获取上传服务器的文件名
	fileNameSlice := strings.Split(filePath, "\\")
	newFileName := fileNameSlice[len(fileNameSlice)-1]

	// 打开文件
	file, e := os.OpenFile(filePath, os.O_RDONLY, 0755)
	if e != nil {
		log.Println("上传文件失败, 文件路径", filePath)
		return nil, e
	}
	// 关闭文件
	defer file.Close()
	data := new(bytes.Buffer)
	log.Println("加载文件成功：" + filePath)
	// 拷贝文件数据
	io.Copy(data, file)

	// 备用，便于以后上传文件夹使用
	//fileNum := win.DragQueryFile(h2, 0xFFFFFFFF, nil, 0)
	//var i uint = 0
	//for ; i < fileNum; i++ {
	//
	//}

	// 构造返回的文件数据
	info := &FileInfo{Data: data, FileName: newFileName, Img: false}
	return info, e
}

// 将 src 流中的图片数据解析到 det 输出流中
func ImgCopy(det *io.Writer, src *bytes.Buffer) error {

	originalImage, err := bmp.Decode(src)
	if err != nil {
		return err
	}
	//log.Println("decode success")

	err = jpeg.Encode(*det, originalImage, nil)
	if err != nil {
		return err
	}

	return nil
}

// 把一个 uint16 类型的数组转成 byte 数组
func Int16SliceToByte(intSlice []uint16) []byte {
	j := 0
	// 获取byte数组长度
	for i := 0; i < len(intSlice); i++ {
		bytes2 := int16ToBytes(intSlice[i])
		// 非 纯英文
		if bytes2[0] != 0 {
			j++
		}
		j++
	}
	var buf = make([]byte, j)
	j = 0
	for i := 0; i < len(intSlice); i++ {
		bytes2 := int16ToBytes(intSlice[i])
		// 非 纯英文
		if bytes2[0] != 0 {
			buf[j] = bytes2[0]
			j++
		}
		buf[j] = bytes2[1]
		j++
	}
	return buf
}

// 把一个 byte 类型的数组转成 rune 数组
func ByteSliceToRune(byteSlice []byte) []rune {
	return []rune(string(byteSlice))
}

// 把一个 uint16 类型的数组转成 rune 数组
func Int16SliceToRune(intSlice []uint16) []rune {
	slice := Int16SliceToByte(intSlice)
	return ByteSliceToRune(slice)
}

func int16ToBytes(i uint16) []byte {
	var buf = make([]byte, 2)
	binary.BigEndian.PutUint16(buf, uint16(i))
	return buf
}
