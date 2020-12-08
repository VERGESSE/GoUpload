package clipboard

import (
	"bytes"
	"encoding/binary"
	"io"

	"golang.org/x/image/bmp"
	"image/jpeg"
	"syscall"
	"unsafe"
)

const (
	cfBitmap      = 2
	cfDib         = 8
	cfUnicodetext = 13
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

func ReadClipboard() (*bytes.Buffer, error) {
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
		return nil, err
	}

	h, _, err := getClipboardData.Call(cfDib)
	if r == 0 {
		return nil, err
	}

	pdata, _, err := globalLock.Call(h)
	if pdata == 0 {
		return nil, err
	}

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
	const sizeof_colorbar = 0
	binary.Write(data, binary.LittleEndian, uint32(fileHeaderLen+infoHeaderLen+sizeof_colorbar))
	j := 0
	for i := fileHeaderLen; i < int(dataSize); i++ {
		binary.Write(data, binary.BigEndian, *(*byte)(unsafe.Pointer(pdata + uintptr(j))))
		j++
	}

	return data, nil
}

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
