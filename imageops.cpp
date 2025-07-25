#include <stdint.h>
#include <string.h>
#include <turbojpeg.h>
#include <algorithm>
#include "imageops.h"

// Jim Blinn's perfect unsigned byte multiply
template <typename T>
T ByteMul(T a, T b) {
	T i = a * b + 128;
	return (i + (i >> 8)) >> 8;
}

// A cheaper unsigned byte multiplier, which only guarantees that 1 * x = x, and 0 * x = 0
template <typename T>
T ByteMulCheap(T a, T b) {
	return ((a + 1) * b) >> 8;
}

struct OrderRGBA {
	static const int R = 0;
	static const int G = 1;
	static const int B = 2;
	static const int A = 3;
};

struct OrderBGRA {
	static const int R = 2;
	static const int G = 1;
	static const int B = 0;
	static const int A = 3;
};

struct OrderARGB {
	static const int R = 1;
	static const int G = 2;
	static const int B = 3;
	static const int A = 0;
};

struct OrderABGR {
	static const int R = 3;
	static const int G = 2;
	static const int B = 1;
	static const int A = 0;
};

//constexpr uint8_t OrderRGBA[4] = {0, 1, 2, 3};
//constexpr uint8_t OrderBGRA[4] = {2, 1, 0, 3};
//constexpr uint8_t OrderARGB[4] = {1, 2, 3, 0};
//constexpr uint8_t OrderABGR[4] = {3, 2, 1, 0};

// This is all done in sRGB space, which is not correct (should be in linear space)
template <bool premultiply, int cR, int cG, int cB, int cA>
void MatteLine(uint8_t* line, size_t width, uint8_t matteR, uint8_t matteG, uint8_t matteB) {
	for (; width != 0; width--) {
		uint8_t r = line[cR];
		uint8_t g = line[cG];
		uint8_t b = line[cB];
		uint8_t a = line[cA];
		if (premultiply) {
			r = ByteMul<int32_t>(r, a);
			g = ByteMul<int32_t>(g, a);
			b = ByteMul<int32_t>(b, a);
		}
		r        = ByteMul<int32_t>(matteR, 255 - a) + r;
		g        = ByteMul<int32_t>(matteG, 255 - a) + g;
		b        = ByteMul<int32_t>(matteB, 255 - a) + b;
		a        = ByteMul<int32_t>(matteR, 255 - a) + a;
		line[cR] = r;
		line[cG] = g;
		line[cB] = b;
		line[cA] = a;
		line += 4;
	}
}

template <bool premultiply, typename order>
void MatteLine2(uint8_t* line, size_t width, uint8_t matteR, uint8_t matteG, uint8_t matteB) {
	MatteLine<premultiply, order::R, order::G, order::B, order::A>(line, width, matteR, matteG, matteB);
}

template <int cR, int cG, int cB, int cA>
void PremultiplyLine(uint8_t* line, size_t width) {
	for (; width != 0; width--) {
		uint8_t a = line[cA];
		line[cR]  = ByteMul<int32_t>(line[cR], a);
		line[cG]  = ByteMul<int32_t>(line[cG], a);
		line[cB]  = ByteMul<int32_t>(line[cB], a);
		line += 4;
	}
}

template <typename order>
void PremultiplyLine2(uint8_t* line, size_t width) {
	PremultiplyLine<order::R, order::G, order::B, order::A>(line, width);
}

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

// _src is either RGB or RGBA
void ToGray(void* _src, int _width, int height, int srcStride, int _nchan, int dstStride, void* _out) {
	auto src = (const uint8_t*) _src;
	auto dst = (uint8_t*) _out;
	for (int y = 0; y < height; y++) {
		size_t width = _width;
		size_t nchan = _nchan;
		auto   srcP  = src;
		auto   dstP  = dst;
		if (nchan >= 3) {
			for (size_t x = 0; x < width; x++) {
				auto r  = srcP[0];
				auto g  = srcP[1];
				auto b  = srcP[2];
				dstP[0] = r * 77 + g * 150 + b * 29 >> 8;
				srcP += nchan;
				dstP++;
			}
		}
		src += srcStride;
		dst += dstStride;
	}
}

// _src is either RGBA or G
void ToRGB(void* _src, int _width, int height, int srcStride, int _nchan, int dstStride, void* _out) {
	auto src = (const uint8_t*) _src;
	auto dst = (uint8_t*) _out;
	for (int y = 0; y < height; y++) {
		size_t width = _width;
		size_t nchan = _nchan;
		auto   srcP  = src;
		auto   dstP  = dst;
		if (nchan == 1) {
			for (size_t x = 0; x < width; x++) {
				auto g  = srcP[0];
				dstP[0] = g;
				dstP[1] = g;
				dstP[2] = g;
				srcP++;
				dstP += 3;
			}
		} else {
			for (size_t x = 0; x < width; x++) {
				dstP[0] = srcP[0];
				dstP[1] = srcP[1];
				dstP[2] = srcP[2];
				srcP += nchan;
				dstP += 3;
			}
		}
		src += srcStride;
		dst += dstStride;
	}
}

// _src is either RGB or G
void ToRGBA(void* _src, int _width, int height, int srcStride, int _nchan, int dstStride, uint8_t alpha, void* _out) {
	auto src = (const uint8_t*) _src;
	auto dst = (uint8_t*) _out;
	for (int y = 0; y < height; y++) {
		size_t width = _width;
		size_t nchan = _nchan;
		auto   srcP  = src;
		auto   dstP  = dst;
		if (nchan == 1) {
			for (size_t x = 0; x < width; x++) {
				auto g  = srcP[0];
				dstP[0] = g;
				dstP[1] = g;
				dstP[2] = g;
				dstP[3] = alpha;
				srcP++;
				dstP += 4;
			}
		} else if (nchan == 3) {
			for (size_t x = 0; x < width; x++) {
				dstP[0] = srcP[0];
				dstP[1] = srcP[1];
				dstP[2] = srcP[2];
				dstP[3] = alpha;
				srcP += nchan;
				dstP += 4;
			}
		}
		src += srcStride;
		dst += dstStride;
	}
}

void Matte(void* src, int width, int height, int srcStride, int format, int isPremultiplied, uint8_t matteR, uint8_t matteG, uint8_t matteB) {
	for (int y = 0; y < height; y++) {
		void* line = (uint8_t*) src + srcStride * y;
		switch (format) {
		case TJPF_RGBA:
			if (isPremultiplied)
				MatteLine2<false, OrderRGBA>((uint8_t*) line, (size_t) width, matteR, matteG, matteB);
			else
				MatteLine2<true, OrderRGBA>((uint8_t*) line, (size_t) width, matteR, matteG, matteB);
			break;
		case TJPF_BGRA:
			if (isPremultiplied)
				MatteLine2<false, OrderBGRA>((uint8_t*) line, (size_t) width, matteR, matteG, matteB);
			else
				MatteLine2<true, OrderBGRA>((uint8_t*) line, (size_t) width, matteR, matteG, matteB);
			break;
		case TJPF_ABGR:
			if (isPremultiplied)
				MatteLine2<false, OrderABGR>((uint8_t*) line, (size_t) width, matteR, matteG, matteB);
			else
				MatteLine2<true, OrderABGR>((uint8_t*) line, (size_t) width, matteR, matteG, matteB);
			break;
		case TJPF_ARGB:
			if (isPremultiplied)
				MatteLine2<false, OrderARGB>((uint8_t*) line, (size_t) width, matteR, matteG, matteB);
			else
				MatteLine2<true, OrderARGB>((uint8_t*) line, (size_t) width, matteR, matteG, matteB);
			break;
		default:
			return;
		}
	}
}

void Premultiply(void* src, int width, int height, int stride, int format) {
	for (int y = 0; y < height; y++) {
		void* line = (uint8_t*) src + stride * y;
		switch (format) {
		case TJPF_RGBA:
			PremultiplyLine2<OrderRGBA>((uint8_t*) line, (size_t) width);
			break;
		case TJPF_BGRA:
			PremultiplyLine2<OrderBGRA>((uint8_t*) line, (size_t) width);
			break;
		case TJPF_ABGR:
			PremultiplyLine2<OrderABGR>((uint8_t*) line, (size_t) width);
			break;
		case TJPF_ARGB:
			PremultiplyLine2<OrderARGB>((uint8_t*) line, (size_t) width);
			break;
		default:
			return;
		}
	}
}

void DrawHorizontalLine(uint8_t* src, int stride, int nchan, uint8_t c1, uint8_t c2, uint8_t c3, int y, int xs, int xe) {
	uint8_t* row = src + static_cast<size_t>(y) * stride;
	if (nchan == 1) {
		uint8_t* p = row + static_cast<size_t>(xs) * 1;
		for (int x = xs; x < xe; ++x) {
			*p = c1;
			p += 1;
		}
	} else if (nchan == 3) {
		uint8_t* p = row + static_cast<size_t>(xs) * 3;
		for (int x = xs; x < xe; ++x) {
			p[0] = c1;
			p[1] = c2;
			p[2] = c3;
			p += 3;
		}
	} else if (nchan == 4) {
		uint8_t* p = row + static_cast<size_t>(xs) * 4;
		for (int x = xs; x < xe; ++x) {
			p[0] = c1;
			p[1] = c2;
			p[2] = c3;
			p[3] = 255;
			p += 4;
		}
	}
}

void DrawVerticalLine(uint8_t* src, int stride, int nchan, uint8_t c1, uint8_t c2, uint8_t c3, int x, int ys, int ye) {
	uint8_t* col = src + static_cast<size_t>(x) * nchan;
	if (nchan == 1) {
		uint8_t* p = col + static_cast<size_t>(ys) * stride;
		for (int y = ys; y < ye; ++y) {
			*p = c1;
			p += stride;
		}
	} else if (nchan == 3) {
		uint8_t* p = col + static_cast<size_t>(ys) * stride;
		for (int y = ys; y < ye; ++y) {
			p[0] = c1;
			p[1] = c2;
			p[2] = c3;
			p += stride;
		}
	} else if (nchan == 4) {
		uint8_t* p = col + static_cast<size_t>(ys) * stride;
		for (int y = ys; y < ye; ++y) {
			p[0] = c1;
			p[1] = c2;
			p[2] = c3;
			p[3] = 255;
			p += stride;
		}
	}
}

void DrawRect(void* _src, int _width, int _height, int _stride, int _nchan, uint8_t c1, uint8_t c2, uint8_t c3, int x1, int y1, int x2, int y2) {
	if (x1 >= x2 || y1 >= y2)
		return;
	auto src = (uint8_t*) _src;

	// Top line
	int yy = y1;
	if (yy >= 0 && yy < _height) {
		int xs = std::max(x1, 0);
		int xe = std::min(x2, _width);
		if (xs < xe) {
			DrawHorizontalLine(src, _stride, _nchan, c1, c2, c3, yy, xs, xe);
		}
	}

	// Bottom line
	yy = y2 - 1;
	if (yy >= 0 && yy < _height) {
		int xs = std::max(x1, 0);
		int xe = std::min(x2, _width);
		if (xs < xe) {
			DrawHorizontalLine(src, _stride, _nchan, c1, c2, c3, yy, xs, xe);
		}
	}

	// Left line
	int xx = x1;
	if (xx >= 0 && xx < _width) {
		int ys = std::max(y1, 0);
		int ye = std::min(y2, _height);
		if (ys < ye) {
			DrawVerticalLine(src, _stride, _nchan, c1, c2, c3, xx, ys, ye);
		}
	}

	// Right line
	xx = x2 - 1;
	if (xx >= 0 && xx < _width) {
		int ys = std::max(y1, 0);
		int ye = std::min(y2, _height);
		if (ys < ye) {
			DrawVerticalLine(src, _stride, _nchan, c1, c2, c3, xx, ys, ye);
		}
	}
}
}
