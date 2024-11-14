package cimg

/*
#include "rotate.h"
*/
import "C"
import (
	"fmt"
	"unsafe"
)

// UnrotateExif rewrites the bytes of an image so that the EXIF orientation information can be discarded.
// In other words, after running UnrotateExif, the encoded image orientation is the same as the natural
// display image orientation.
// exifOrientation must be either 3, 6, or 8.
func UnrotateExif(exifOrientation int, src *Image) (*Image, error) {
	if !(exifOrientation == 3 || exifOrientation == 6 || exifOrientation == 8) {
		return nil, fmt.Errorf("UnrotateExif can't unrotate orientation %v. Only 3, 6, 8 are supported", exifOrientation)
	}
	dstWidth, dstHeight := src.Width, src.Height
	if exifOrientation == 6 || exifOrientation == 8 {
		dstWidth, dstHeight = src.Height, src.Width
	}
	dst := NewImage(dstWidth, dstHeight, src.Format)
	C.Unrotate(C.int(exifOrientation), unsafe.Pointer(&src.Pixels[0]), C.int(src.Width), C.int(src.Height), C.int(src.Stride), C.int(src.NChan()), unsafe.Pointer(&dst.Pixels[0]))
	return dst, nil
}
