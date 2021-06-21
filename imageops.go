package cimg

/*
#include "imageops.h"
*/
import "C"
import "unsafe"

// AvgColor computes the average color of the entire image, per channel
// The averaging is performed in sRGB space (i.e. not physically correct)
// If the image has more than 8 channels, then the function will panic
func (img *Image) AvgColor() []uint8 {
	if C.int(img.NChan) > C.AvgColorMaxChannels {
		panic("Image for AvgColor has more than 8 channels")
	}
	channels := [8]uint8{}
	C.AvgColor(unsafe.Pointer(&img.Pixels[0]), C.int(img.Width), C.int(img.Height), C.int(img.Stride), C.int(img.NChan), unsafe.Pointer(&channels[0]))
	return channels[:img.NChan]
}
