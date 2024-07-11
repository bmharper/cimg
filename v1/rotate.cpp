#include <stdint.h>
#include "rotate.h"

// This is a great site with illustrations:
// https://www.impulseadventure.com/photo/exif-orientation.html

// 180 degrees
template <unsigned nchan>
void Unrotate3(const uint8_t* src, unsigned width, unsigned height, int stride, uint8_t* dst) {
	const uint8_t* pSrc = src;
	uint8_t*       pDst = dst;
	for (unsigned y = 0; y < height; y++) {
		pSrc = (const uint8_t*) src + (height - 1 - y) * stride + (width - 1) * nchan;
		for (unsigned x = 0; x < width; x++) {
			for (unsigned i = 0; i < nchan; i++)
				*pDst++ = *pSrc++;
			pSrc -= nchan * 2;
		}
	}
}

// 90 degrees clockwise
template <unsigned nchan>
void Unrotate6(const uint8_t* src, unsigned width, unsigned height, int stride, uint8_t* dst) {
	const uint8_t* pSrc = src;
	uint8_t*       pDst = dst;
	for (unsigned x = 0; x < width; x++) {
		pSrc = (const uint8_t*) src + (height - 1) * stride + x * nchan;
		for (unsigned y = 0; y < height; y++) {
			for (unsigned i = 0; i < nchan; i++)
				*pDst++ = *pSrc++;
			pSrc -= stride + nchan;
		}
	}
}

// 90 degrees counter-clockwise
template <unsigned nchan>
void Unrotate8(const uint8_t* src, unsigned width, unsigned height, int stride, uint8_t* dst) {
	const uint8_t* pSrc = src;
	uint8_t*       pDst = dst;
	for (unsigned x = 0; x < width; x++) {
		pSrc = (const uint8_t*) src + (width - 1 - x) * nchan;
		for (unsigned y = 0; y < height; y++) {
			for (unsigned i = 0; i < nchan; i++)
				*pDst++ = *pSrc++;
			pSrc += stride - nchan;
		}
	}
}

extern "C" {

void Unrotate(int exifOrientation, void* _src, int _width, int _height, int stride, int _nchan, void* _dst) {
	const uint8_t* src    = (const uint8_t*) _src;
	uint8_t*       dst    = (uint8_t*) _dst;
	unsigned       width  = _width;
	unsigned       height = _height;
	switch (exifOrientation) {
	case 3:
		switch (_nchan) {
		case 1: Unrotate3<1>(src, width, height, stride, dst); break;
		case 3: Unrotate3<3>(src, width, height, stride, dst); break;
		case 4: Unrotate3<4>(src, width, height, stride, dst); break;
		}
		break;
	case 6:
		switch (_nchan) {
		case 1: Unrotate6<1>(src, width, height, stride, dst); break;
		case 3: Unrotate6<3>(src, width, height, stride, dst); break;
		case 4: Unrotate6<4>(src, width, height, stride, dst); break;
		}
		break;
	case 8:
		switch (_nchan) {
		case 1: Unrotate8<1>(src, width, height, stride, dst); break;
		case 3: Unrotate8<3>(src, width, height, stride, dst); break;
		case 4: Unrotate8<4>(src, width, height, stride, dst); break;
		}
		break;
	}
}
}