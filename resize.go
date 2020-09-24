package cimg

/*
#include "stb_image_resize.h"
*/
import "C"
import "unsafe"

// ResizeEasy allocates the output image for you and returns it
// Assumes sRGB image
func ResizeEasy(src *Image, dstWidth, dstHeight int) *Image {
	/*
		STBIRDEF int stbir_resize(         const void *input_pixels , int input_w , int input_h , int input_stride_in_bytes,
		                                         void *output_pixels, int output_w, int output_h, int output_stride_in_bytes,
		                                   stbir_datatype datatype,
		                                   int num_channels, int alpha_channel, int flags,
		                                   stbir_edge edge_mode_horizontal, stbir_edge edge_mode_vertical,
		                                   stbir_filter filter_horizontal,  stbir_filter filter_vertical,
		                                   stbir_colorspace space, void *alloc_context);

	*/
	dst := NewImage(dstWidth, dstHeight, src.NChan)

	C.stbir_resize(
		unsafe.Pointer(&src.Pixels[0]), C.int(src.Width), C.int(src.Height), C.int(src.Stride),
		unsafe.Pointer(&dst.Pixels[0]), C.int(dst.Width), C.int(dst.Height), C.int(dst.Stride),
		C.STBIR_TYPE_UINT8,
		C.int(src.NChan), C.STBIR_ALPHA_CHANNEL_NONE, 0,
		C.STBIR_EDGE_CLAMP, C.STBIR_EDGE_CLAMP,
		C.STBIR_FILTER_MITCHELL, C.STBIR_FILTER_MITCHELL,
		C.STBIR_COLORSPACE_SRGB, C.NULL)

	return dst
}
