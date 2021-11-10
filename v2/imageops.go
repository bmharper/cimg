package cimg

/*
#include "imageops.h"
*/
import "C"
import (
	"fmt"
	"unsafe"
)

// AvgColor computes the average color of the entire image, per channel
// The averaging is performed in sRGB space (i.e. not linear light)
// If the image has more than 8 channels, then the function will panic
func (img *Image) AvgColor() []uint8 {
	if C.int(img.NChan()) > C.AvgColorMaxChannels {
		panic("Image for AvgColor has more than 8 channels")
	}
	channels := [8]uint8{}
	C.AvgColor(unsafe.Pointer(&img.Pixels[0]), C.int(img.Width), C.int(img.Height), C.int(img.Stride), C.int(img.NChan()), unsafe.Pointer(&channels[0]))
	return channels[:img.NChan()]
}

// CopyImage copies src into dst at the location dstX1, dstY1
func (dst *Image) CopyImage(src *Image, dstX1, dstY1 int) error {
	return dst.CopyImageRect(src, 0, 0, src.Width, src.Height, dstX1, dstY1)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a < b {
		return b
	}
	return a
}

func clamp(v, min, max int) int {
	if v < min {
		return min
	} else if v > max {
		return max
	}
	return v
}

// CopyImageRect copies src into dst, at dstX1,dstY1. The source imagery is read from the rectangle
// specified by the 4 source location parameters. All coordinates are clipped prior to drawing.
// The only error condition is when the two images have a different number of channels.
// Note that you will get swapped RGB channels if you do something like copy from an RGB image
// into a BGR image (i.e. this function does not swizzle the channels, it just does a dumb memcpy of the rows).
func (dst *Image) CopyImageRect(src *Image, srcX1, srcY1, srcX2, srcY2 int, dstX1, dstY1 int) error {
	if src.NChan() != dst.NChan() {
		return fmt.Errorf("Source image channels: %v, target image channels: %v", src.NChan(), dst.NChan())
	}
	srcX1 = max(srcX1, 0)
	srcY1 = max(srcY1, 0)
	srcX2 = min(srcX2, src.Width)
	srcY2 = min(srcY2, src.Height)
	dstX1 = clamp(dstX1, 0, dst.Width)
	dstY1 = clamp(dstY1, 0, dst.Height)
	w := min(srcX2-srcX1, dst.Width-dstX1)
	h := min(srcY2-srcY1, dst.Height-dstY1)
	if w <= 0 || h <= 0 {
		return nil
	}
	bytesPerLine := w * dst.NChan()
	srcOffset := srcX1 * src.NChan()
	dstOffset := dstX1 * dst.NChan()
	for y := 0; y < h; y++ {
		srcP := src.Stride*(srcY1+y) + srcOffset
		dstP := dst.Stride*(dstY1+y) + dstOffset
		copy(dst.Pixels[dstP:dstP+bytesPerLine], src.Pixels[srcP:srcP+bytesPerLine])
	}
	return nil
}

// ToRGB returns a 3 channel image.
// This is used to remove the alpha channel from an image that was loaded from a PNG,
// or to turn a gray image into an RGB image.
// If the image is already a 3 channel image, then a clone is returned
func (img *Image) ToRGB() *Image {
	if img.NChan() == 3 {
		return img.Clone()
	}
	dst := NewImage(img.Width, img.Height, PixelFormatRGB)
	C.ToRGB(unsafe.Pointer(&img.Pixels[0]), C.int(img.Width), C.int(img.Height), C.int(img.Stride), C.int(img.NChan()), C.int(dst.Stride), unsafe.Pointer(&dst.Pixels[0]))
	return dst
}

// For an RGBA image, blend it on top of the given color, so that transparent regions of the image
// will be filled with the given color.
// Returns an error if the image is not RGBA format
func (img *Image) Matte(r, g, b uint8) error {
	premul := 0
	if img.Premultiplied {
		premul = 1
	}
	C.Matte(unsafe.Pointer(&img.Pixels[0]), C.int(img.Width), C.int(img.Height), C.int(img.Stride), C.int(img.Format), C.int(premul), C.uint8_t(r), C.uint8_t(g), C.uint8_t(b))
	return nil
}

// Premultiply RGB by A.
// If the image does not have an alpha channel, or if Premultiplied=true then this is a no-op
func (img *Image) Premultiply() {
	if img.Premultiplied || img.NChan() != 4 {
		return
	}
	C.Premultiply(unsafe.Pointer(&img.Pixels[0]), C.int(img.Width), C.int(img.Height), C.int(img.Stride), C.int(img.Format))
	img.Premultiplied = true
}
