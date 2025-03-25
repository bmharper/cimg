package cimg

// #include "rotate.h"
import "C"
import (
	"fmt"
	"math"
	"unsafe"
)

// Rotation filter
type RotateFilter int

const (
	RotateFilterBilinear RotateFilter = iota // Bilinear filtering
)

const RotateDefaultSnapThreshold = 0.01 * math.Pi / 180

// Rotation parameters
type RotateParams struct {
	Filter               RotateFilter
	SnapThresholdRadians float64 // If rotation angle is close enough to -90, 90 or 180, then snap to discrete rotation
}

// Return default rotation parameters
func NewRotateParams() *RotateParams {
	return &RotateParams{
		Filter:               RotateFilterBilinear,
		SnapThresholdRadians: RotateDefaultSnapThreshold,
	}
}

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
	C.UnrotateExif(C.int(exifOrientation), unsafe.Pointer(&src.Pixels[0]), C.int(src.Width), C.int(src.Height), C.int(src.Stride), C.int(src.NChan()), unsafe.Pointer(&dst.Pixels[0]), C.int(dst.Stride))
	return dst, nil
}

// Rotate src into dst, by angleRadians
// If params is nil, then default values are used.
// A positive angle produces a clockwise rotation.
func Rotate(src *Image, dst *Image, angleRadians float64, params *RotateParams) {
	if src.NChan() != dst.NChan() {
		panic("Rotate: src and dst must have the same number of channels")
	}

	snapThreshold := RotateDefaultSnapThreshold * 180 / math.Pi
	if params != nil {
		snapThreshold = params.SnapThresholdRadians * 180 / math.Pi
	}

	angleDegrees := angleRadians * 180 / math.Pi
	sizeMatch90 := src.Width == dst.Height && src.Height == dst.Width
	sizeMatch180 := src.Width == dst.Width && src.Height == dst.Height
	isDiscrete90 := sizeMatch90 && (math.Abs(angleDegrees-90) < snapThreshold || math.Abs(angleDegrees+90) < snapThreshold || math.Abs(angleDegrees-270) < snapThreshold || math.Abs(angleDegrees+270) < snapThreshold)
	isDiscrete180 := sizeMatch180 && (math.Abs(angleDegrees-180) < snapThreshold || math.Abs(angleDegrees+180) < snapThreshold)

	if angleRadians == 0 && src.Width == dst.Width && src.Height == dst.Height {
		dst.CopyImage(src, 0, 0)
	} else if isDiscrete90 || isDiscrete180 {
		C.RotateDiscrete(C.int(math.Round(angleDegrees)), unsafe.Pointer(&src.Pixels[0]), C.int(src.Width), C.int(src.Height), C.int(src.Stride), C.int(src.NChan()),
			unsafe.Pointer(&dst.Pixels[0]), C.int(dst.Stride))
	} else {
		C.RotateImageBilinear((*C.uint8_t)(&src.Pixels[0]), (*C.uint8_t)(&dst.Pixels[0]), C.int(src.NChan()),
			C.int(src.Width), C.int(src.Height), C.int(src.Stride),
			C.int(dst.Width), C.int(dst.Height), C.int(dst.Stride),
			C.double(angleRadians))
	}
}
