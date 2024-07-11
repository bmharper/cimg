#ifdef __cplusplus
extern "C" {
#endif

const int AvgColorMaxChannels = 8;

void AvgColor(void* _src, int _width, int _height, int stride, int _nchan, void* _outChannels);
void ToRGB(void* _src, int _width, int height, int srcStride, int _nchan, int dstStride, void* _out);

#ifdef __cplusplus
}
#endif
