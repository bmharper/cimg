#ifdef __cplusplus
extern "C" {
#endif

#include <stdint.h>

void UnrotateExif(int exifOrientation, void* _src, int _width, int _height, int stride, int _nchan, void* _dst, int dstStride);

// Rotate image by 90,180,270,-90,-180,-270 degrees (A few of these are duplicates: -90 = 270, -180 = 180, -270 = 90)
void RotateDiscrete(int angle, void* _src, int _width, int _height, int stride, int _nchan, void* _dst, int dstStride);

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
    double         angle_radians);

#ifdef __cplusplus
}
#endif
