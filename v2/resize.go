package cimg

/*
#include "stb_image_resize2.h"
*/
import "C"
import (
	"errors"
	"fmt"
	"unsafe"
)

// ResizeNew allocates the output image for you and returns it
// Assumes sRGB image
func ResizeNew(src *Image, dstWidth, dstHeight int) *Image {
	dst := NewImage(dstWidth, dstHeight, src.Format)
	Resize(src, dst)
	return dst
}

// Resize resizes an image into a destination buffer that you provide
// Assumes sRGB image
func Resize(src, dst *Image) error {
	if dst.Width == 0 || dst.Height == 0 {
		return errors.New("Image target dimensions must be non-zero")
	}

	layout := C.stbir_pixel_layout(0)
	switch src.Format {
	case PixelFormatRGB:
		layout = C.STBIR_RGB
	case PixelFormatBGR:
		layout = C.STBIR_BGR
	case PixelFormatRGBX:
		layout = C.STBIR_RGBA_PM
	case PixelFormatBGRX:
		layout = C.STBIR_BGRA_PM
	case PixelFormatXBGR:
		layout = C.STBIR_ABGR_PM
	case PixelFormatXRGB:
		layout = C.STBIR_ARGB_PM
	case PixelFormatGRAY:
		layout = C.STBIR_1CHANNEL
	case PixelFormatRGBA:
		if src.Premultiplied {
			layout = C.STBIR_RGBA_PM
		} else {
			layout = C.STBIR_RGBA
		}
	case PixelFormatBGRA:
		if src.Premultiplied {
			layout = C.STBIR_BGRA_PM
		} else {
			layout = C.STBIR_BGRA
		}
	case PixelFormatABGR:
		if src.Premultiplied {
			layout = C.STBIR_ABGR_PM
		} else {
			layout = C.STBIR_ABGR
		}
	case PixelFormatARGB:
		if src.Premultiplied {
			layout = C.STBIR_ARGB_PM
		} else {
			layout = C.STBIR_ARGB
		}
	case PixelFormatCMYK:
		layout = C.STBIR_4CHANNEL
	default:
		return fmt.Errorf("Unsupported pixel format for resize: %v", src.Format)
	}

	if src.NChan() != dst.NChan() {
		return fmt.Errorf("Source channel count %v differs from target channel count %v", src.NChan(), dst.NChan())
	}

	/*
		STBIRDEF void *  stbir_resize( const void *input_pixels , int input_w , int input_h, int input_stride_in_bytes,
		                                     void *output_pixels, int output_w, int output_h, int output_stride_in_bytes,
		                               stbir_pixel_layout pixel_layout, stbir_datatype data_type,
		                               stbir_edge edge, stbir_filter filter );
	*/
	C.stbir_resize(
		unsafe.Pointer(&src.Pixels[0]), C.int(src.Width), C.int(src.Height), C.int(src.Stride),
		unsafe.Pointer(&dst.Pixels[0]), C.int(dst.Width), C.int(dst.Height), C.int(dst.Stride),
		layout, C.STBIR_TYPE_UINT8_SRGB,
		C.STBIR_EDGE_CLAMP, C.STBIR_FILTER_DEFAULT)

	return nil
}
