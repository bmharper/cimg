#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

const int AvgColorMaxChannels = 8;

void AvgColor(void* _src, int _width, int _height, int stride, int _nchan, void* _outChannels);
void ToGray(void* _src, int _width, int height, int srcStride, int _nchan, int dstStride, void* _out);
void ToRGB(void* _src, int _width, int height, int srcStride, int _nchan, int dstStride, void* _out);
void ToRGBA(void* _src, int _width, int height, int srcStride, int _nchan, int dstStride, uint8_t alpha, void* _out);
void Matte(void* src, int width, int height, int srcStride, int format, int isPremultiplied, uint8_t matteR, uint8_t matteG, uint8_t matteB);
void Premultiply(void* src, int width, int height, int stride, int format);
void DrawRect(void* _src, int _width, int _height, int _stride, int _nchan, uint8_t c1, uint8_t c2, uint8_t c3, int x1, int y1, int x2, int y2);

#ifdef __cplusplus
}
#endif
