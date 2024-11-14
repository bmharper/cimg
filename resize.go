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

type ResizeFilter int

const (
	ResizeFilterDefault     ResizeFilter = C.STBIR_FILTER_DEFAULT      // use same filter type that easy-to-use API chooses
	ResizeFilterBox         ResizeFilter = C.STBIR_FILTER_BOX          // A trapezoid w/1-pixel wide ramps, same result as box for integer scale ratios
	ResizeFilterTriangle    ResizeFilter = C.STBIR_FILTER_TRIANGLE     // On upsampling, produces same results as bilinear texture filtering
	ResizeFilterCubicSpline ResizeFilter = C.STBIR_FILTER_CUBICBSPLINE // The cubic b-spline (aka Mitchell-Netrevalli with B=1,C=0), gaussian-esque
	ResizeFilterCatmullRom  ResizeFilter = C.STBIR_FILTER_CATMULLROM   // An interpolating cubic spline
	ResizeFilterMitchell    ResizeFilter = C.STBIR_FILTER_MITCHELL     // Mitchell-Netrevalli filter with B=1/3, C=1/3
	ResizeFilterPointSample ResizeFilter = C.STBIR_FILTER_POINT_SAMPLE // Simple point sampling
)

type ResizeEdge int

const (
	ResizeEdgeClamp   ResizeEdge = C.STBIR_EDGE_CLAMP
	ResizeEdgeReflect ResizeEdge = C.STBIR_EDGE_REFLECT
	ResizeEdgeWrap    ResizeEdge = C.STBIR_EDGE_WRAP // this edge mode is slower and uses more memory
	ResizeEdgeZero    ResizeEdge = C.STBIR_EDGE_ZERO
)

type ResizeParams struct {
	Edge            ResizeEdge   // 0 = STBIR_EDGE_CLAMP, which is convenient
	Filter          ResizeFilter // 0 = STBIR_FILTER_DEFAULT, which is convenient
	CheapSRGBFilter bool         // If data is sRGB (for 8-bit it almost always is), then use cheap non-linear filtering directly in the sRGB space
}

// ResizeNew allocates the output image for you and returns it
// Assumes sRGB image
func ResizeNew(src *Image, dstWidth, dstHeight int, params *ResizeParams) *Image {
	dst := NewImage(dstWidth, dstHeight, src.Format)
	Resize(src, dst, params)
	return dst
}

// Resize resizes an image into a destination buffer that you provide
// Assumes sRGB image
func Resize(src, dst *Image, params *ResizeParams) error {
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

	var dataType C.stbir_datatype
	dataType = C.STBIR_TYPE_UINT8_SRGB
	if params != nil && params.CheapSRGBFilter {
		dataType = C.STBIR_TYPE_UINT8
	}

	var filter C.stbir_filter
	filter = C.STBIR_FILTER_DEFAULT
	if params != nil {
		filter = C.stbir_filter(params.Filter)
	}

	var edge C.stbir_edge
	edge = C.STBIR_EDGE_CLAMP
	if params != nil {
		edge = C.stbir_edge(params.Edge)
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
		layout, dataType,
		edge, filter)

	return nil
}
