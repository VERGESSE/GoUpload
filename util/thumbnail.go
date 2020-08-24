// Copyright © 2016 Alan A. A. Donovan & Brian W. Kernighan.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

// See page 234.

// The thumbnail package produces thumbnail-size images from
// larger images.  Only JPEG images are currently supported.
package util

import (
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// 取自《Go程序设计语言》第八章
// Image returns a thumbnail-size version of src.
func Image(src image.Image, pixel int) image.Image {
	// Compute thumbnail size, preserving aspect ratio.
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

	// a very crude scaling algorithm
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			srcx := int(float64(x) * xscale)
			srcy := int(float64(y) * yscale)
			dst.Set(x, y, src.At(srcx, srcy))
		}
	}
	return dst
}

// ImageStream reads an image from r and
// writes a thumbnail-size version of it to w.
func ImageStream(w io.Writer, r io.Reader, pixel int) error {
	src, _, err := image.Decode(r)
	if err != nil {
		log.Println(err)
		return err
	}
	dst := Image(src, pixel)
	return jpeg.Encode(w, dst, nil)
}

// ImageFile2 reads an image from infile and writes
// a thumbnail-size version of it to outfile.
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

// ImageFile reads an image from infile and writes
// a thumbnail-size version of it in the same directory.
// It returns the generated file name, e.g. "foo.thumb.jpeg".
func ThumbImage(infile string, pixel int) (string, error) {
	pixelsStr := strconv.Itoa(pixel)
	suffix := "_" + pixelsStr + "x" + pixelsStr
	ext := filepath.Ext(infile) // e.g., ".jpg", ".JPEG"
	outfile := strings.TrimSuffix(infile, ext) + suffix + ext
	return outfile, ImageFile(outfile, infile, pixel)
}
