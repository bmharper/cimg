#ifdef __cplusplus
extern "C" {
#endif

const int AvgColorMaxChannels = 8;

void AvgColor(void* _src, int _width, int _height, int stride, int _nchan, void* _outChannels);

#ifdef __cplusplus
}
#endif
