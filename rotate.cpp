#include <math.h>
#include <stdint.h>
#include "rotate.h"

// This is a great site with illustrations of EXIF orientations:
// https://www.impulseadventure.com/photo/exif-orientation.html

#define MIN(a, b) ((a) < (b) ? (a) : (b))
#define MAX(a, b) ((a) > (b) ? (a) : (b))

// Rotate image 180 degrees
template <unsigned nchan>
void Rotate180(const uint8_t* src, unsigned width, unsigned height, int srcStride, uint8_t* dst, int dstStride) {
	const uint8_t* pSrc = src;
	uint8_t*       pDst = dst;
	for (unsigned y = 0; y < height; y++) {
		pSrc = (const uint8_t*) src + (height - 1 - y) * srcStride + (width - 1) * nchan;
		for (unsigned x = 0; x < width; x++) {
			for (unsigned i = 0; i < nchan; i++)
				*pDst++ = *pSrc++;
			pSrc -= nchan * 2; // pSrc has moved to the right by nchan, so move back 2*nchan bytes to move one to the left
		}
		pDst += dstStride - width * nchan;
	}
}

// Rotate image 90 degrees clockwise
template <unsigned nchan>
void Rotate90CW(const uint8_t* src, unsigned width, unsigned height, int srcStride, uint8_t* dst, int dstStride) {
	const uint8_t* pSrc = src;
	uint8_t*       pDst = dst;
	for (unsigned x = 0; x < width; x++) {
		pSrc = (const uint8_t*) src + (height - 1) * srcStride + x * nchan;
		for (unsigned y = 0; y < height; y++) {
			for (unsigned i = 0; i < nchan; i++)
				*pDst++ = *pSrc++;
			pSrc -= srcStride + nchan;
		}
		pDst += dstStride - height * nchan;
	}
}

// Rotate image 90 degrees counter-clockwise
template <unsigned nchan>
void Rotate90CCW(const uint8_t* src, unsigned width, unsigned height, int srcStride, uint8_t* dst, int dstStride) {
	const uint8_t* pSrc = src;
	uint8_t*       pDst = dst;
	for (unsigned x = 0; x < width; x++) {
		pSrc = (const uint8_t*) src + (width - 1 - x) * nchan;
		for (unsigned y = 0; y < height; y++) {
			for (unsigned i = 0; i < nchan; i++)
				*pDst++ = *pSrc++;
			pSrc += srcStride - nchan;
		}
		pDst += dstStride - height * nchan;
	}
}

// Inline fixed-point bilinear interpolation
template <unsigned nchan>
void Bilinear(
    const uint8_t* input,
    int            width,
    int            height,
    int            stride,
    double         x,
    double         y,
    uint8_t*       output) {
	// Compute integral parts
	int x_floor = (int) floor(x);
	int y_floor = (int) floor(y);

	// Check bounds for bilinear interpolation
	// We need x_floor, y_floor, x_floor+1, y_floor+1 to be valid indices
	if (x_floor < 0 || y_floor < 0 || x_floor >= width - 1 || y_floor >= height - 1) {
		// Out of bounds: clamp to edge
		x       = MIN(MAX(x, 0), width - 1.001);
		y       = MIN(MAX(y, 0), height - 1.001);
		x_floor = (int) floor(x);
		y_floor = (int) floor(y);
	}

	// Compute fractional parts in fixed-point Q16 (1.0 = 65536)
	// x_frac = fraction(x), y_frac = fraction(y)
	double x_frac_d = x - x_floor;
	double y_frac_d = y - y_floor;

	int32_t x_frac = (int32_t) (x_frac_d * 65536.0);
	int32_t y_frac = (int32_t) (y_frac_d * 65536.0);

	int32_t one_minus_x = 65536 - x_frac;
	int32_t one_minus_y = 65536 - y_frac;

	// Compute weights (Q16)
	// W00 = (1 - x_frac)*(1 - y_frac)
	// W10 = x_frac*(1 - y_frac)
	// W01 = (1 - x_frac)*y_frac
	// W11 = x_frac*y_frac
	// All results fit into 32-bit safely.
	int32_t W00 = (int32_t) (((int64_t) one_minus_x * one_minus_y) >> 16);
	int32_t W10 = (int32_t) (((int64_t) x_frac * one_minus_y) >> 16);
	int32_t W01 = (int32_t) (((int64_t) one_minus_x * y_frac) >> 16);
	int32_t W11 = (int32_t) (((int64_t) x_frac * y_frac) >> 16);

	const uint8_t* p00 = input + y_floor * stride + x_floor * nchan;
	const uint8_t* p10 = input + y_floor * stride + (x_floor + 1) * nchan;
	const uint8_t* p01 = input + (y_floor + 1) * stride + x_floor * nchan;
	const uint8_t* p11 = input + (y_floor + 1) * stride + (x_floor + 1) * nchan;

	// Interpolate each channel using fixed-point arithmetic.
	// Final = (p00*C00 + p10*C10 + p01*C01 + p11*C11) >> 16, with rounding.
	// We'll add half (32768) before shifting for rounding.
	for (unsigned i = 0; i < nchan; i++) {
		int32_t v = ((int32_t) p00[i] * W00) + ((int32_t) p10[i] * W10) +
		            ((int32_t) p01[i] * W01) + ((int32_t) p11[i] * W11);

		// Add 0x8000 for rounding and shift right by 16
		output[i] = (uint8_t) ((v + 32768) >> 16);
	}
}

extern "C" {

void RotateDiscrete(int angle, void* _src, int _width, int _height, int stride, int _nchan, void* _dst, int dstStride) {
	const uint8_t* src    = (const uint8_t*) _src;
	uint8_t*       dst    = (uint8_t*) _dst;
	unsigned       width  = _width;
	unsigned       height = _height;
	switch (angle) {
	case -180:
	case 180:
		switch (_nchan) {
		case 1: Rotate180<1>(src, width, height, stride, dst, dstStride); break;
		case 3: Rotate180<3>(src, width, height, stride, dst, dstStride); break;
		case 4: Rotate180<4>(src, width, height, stride, dst, dstStride); break;
		}
		break;
	case 90:
	case -270:
		switch (_nchan) {
		case 1: Rotate90CW<1>(src, width, height, stride, dst, dstStride); break;
		case 3: Rotate90CW<3>(src, width, height, stride, dst, dstStride); break;
		case 4: Rotate90CW<4>(src, width, height, stride, dst, dstStride); break;
		}
		break;
	case -90:
	case 270:
		switch (_nchan) {
		case 1: Rotate90CCW<1>(src, width, height, stride, dst, dstStride); break;
		case 3: Rotate90CCW<3>(src, width, height, stride, dst, dstStride); break;
		case 4: Rotate90CCW<4>(src, width, height, stride, dst, dstStride); break;
		}
		break;
	}
}

void RotateImageBilinear(
    const uint8_t* input,
    uint8_t*       output,
    int            nchan,
    int            input_width,
    int            input_height,
    int            input_stride,
    int            output_width,
    int            output_height,
    int            output_stride,
    double         angle_radians) {
	// Precompute cos and sin of angle
	double cos_angle = cos(angle_radians);
	double sin_angle = sin(angle_radians);

	// Precompute centers
	double cx_input  = (input_width - 1) / 2.0;
	double cy_input  = (input_height - 1) / 2.0;
	double cx_output = (output_width - 1) / 2.0;
	double cy_output = (output_height - 1) / 2.0;

	for (int y = 0; y < output_height; y++) {
		double y_rel = y - cy_output;
		for (int x = 0; x < output_width; x++) {
			double x_rel = x - cx_output;

			// Rotate back to source coordinates
			double src_x = x_rel * cos_angle + y_rel * sin_angle + cx_input;
			double src_y = -x_rel * sin_angle + y_rel * cos_angle + cy_input;

			if (nchan == 1) {
				uint8_t* dst = output + y * output_stride + x;
				Bilinear<1>(input, input_width, input_height, input_stride, src_x, src_y, dst);
			} else if (nchan == 2) {
				uint8_t* dst = output + y * output_stride + x * 2;
				Bilinear<2>(input, input_width, input_height, input_stride, src_x, src_y, dst);
			} else if (nchan == 3) {
				uint8_t* dst = output + y * output_stride + x * 3;
				Bilinear<3>(input, input_width, input_height, input_stride, src_x, src_y, dst);
			} else if (nchan == 4) {
				uint8_t* dst = output + y * output_stride + x * 4;
				Bilinear<4>(input, input_width, input_height, input_stride, src_x, src_y, dst);
			}
		}
	}
}

void UnrotateExif(int exifOrientation, void* _src, int _width, int _height, int stride, int _nchan, void* _dst, int dstStride) {
	int angle = 0;
	switch (exifOrientation) {
	case 1: angle = 0; break;
	case 3: angle = 180; break;
	case 6: angle = 90; break;
	case 8: angle = -90; break;
	}
	RotateDiscrete(angle, _src, _width, _height, stride, _nchan, _dst, dstStride);
}
}