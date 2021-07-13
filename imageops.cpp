#include <stdint.h>
#include <malloc.h>
#include "imageops.h"

extern "C" {

void AvgColor(void* _src, int _width, int _height, int stride, int _nchan, void* _outChannels) {
	const uint8_t* src                      = (const uint8_t*) _src;
	uint64_t       sum[AvgColorMaxChannels] = {0};
	if (_nchan < 1 || _nchan > sizeof(sum) / sizeof(sum[0]))
		return;
	for (int y = 0; y < _height; y++) {
		const uint8_t* p = src;
		for (int x = 0; x < _width; x++) {
			for (int c = 0; c < _nchan; c++) {
				sum[c] += *p++;
			}
		}
		src += stride;
	}
	uint64_t nPixels     = (uint64_t) _width * (uint64_t) _height;
	uint8_t* outChannels = (uint8_t*) _outChannels;
	for (int c = 0; c < _nchan; c++) {
		outChannels[c] = sum[c] / nPixels;
	}
}

void ToRGB(void* _src, int _width, int height, int srcStride, int _nchan, int dstStride, void* _out) {
	auto src = (const uint8_t*) _src;
	auto dst = (uint8_t*) _out;
	for (int y = 0; y < height; y++) {
		size_t width = _width;
		size_t nchan = _nchan;
		auto   srcP  = src;
		auto   dstP  = dst;
		for (size_t x = 0; x < width; x++) {
			dstP[0] = srcP[0];
			dstP[1] = srcP[1];
			dstP[2] = srcP[2];
			srcP += nchan;
			dstP += 3;
		}
		src += srcStride;
		dst += dstStride;
	}
}
}
