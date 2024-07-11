package cimg

/*
#cgo LDFLAGS: -lturbojpeg
#include <turbojpeg.h>
*/
import "C"

import (
	"bytes"
	"fmt"
	"image/png"
	"unsafe"
)

type Sampling C.int

const (
	Sampling444  Sampling = C.TJSAMP_444
	Sampling422  Sampling = C.TJSAMP_422
	Sampling420  Sampling = C.TJSAMP_420
	SamplingGray Sampling = C.TJSAMP_GRAY
)

type PixelFormat C.int

const (
	PixelFormatRGB     PixelFormat = C.TJPF_RGB
	PixelFormatBGR     PixelFormat = C.TJPF_BGR
	PixelFormatRGBX    PixelFormat = C.TJPF_RGBX
	PixelFormatBGRX    PixelFormat = C.TJPF_BGRX
	PixelFormatXBGR    PixelFormat = C.TJPF_XBGR
	PixelFormatXRGB    PixelFormat = C.TJPF_XRGB
	PixelFormatGRAY    PixelFormat = C.TJPF_GRAY
	PixelFormatRGBA    PixelFormat = C.TJPF_RGBA
	PixelFormatBGRA    PixelFormat = C.TJPF_BGRA
	PixelFormatABGR    PixelFormat = C.TJPF_ABGR
	PixelFormatARGB    PixelFormat = C.TJPF_ARGB
	PixelFormatCMYK    PixelFormat = C.TJPF_CMYK
	PixelFormatUNKNOWN PixelFormat = C.TJPF_UNKNOWN
)

type Flags C.int

const (
	FlagAccurateDCT   Flags = C.TJFLAG_ACCURATEDCT
	FlagBottomUp      Flags = C.TJFLAG_BOTTOMUP
	FlagFastDCT       Flags = C.TJFLAG_FASTDCT
	FlagFastUpsample  Flags = C.TJFLAG_FASTUPSAMPLE
	FlagNoRealloc     Flags = C.TJFLAG_NOREALLOC
	FlagProgressive   Flags = C.TJFLAG_PROGRESSIVE
	FlagStopOnWarning Flags = C.TJFLAG_STOPONWARNING
)

func makeError(handler C.tjhandle, returnVal C.int) error {
	if returnVal == 0 {
		return nil
	}
	str := C.GoString(C.tjGetErrorStr2(handler))
	return fmt.Errorf("turbojpeg error: %v", str)
}

// CompressParams are the TurboJPEG compression parameters
type CompressParams struct {
	PixelFormat PixelFormat
	Sampling    Sampling
	Quality     int // 1 .. 100
	Flags       Flags
}

// MakeCompressParams returns a fully populated CompressParams struct
func MakeCompressParams(pixelFormat PixelFormat, sampling Sampling, quality int, flags Flags) CompressParams {
	return CompressParams{
		PixelFormat: pixelFormat,
		Sampling:    sampling,
		Quality:     quality,
		Flags:       flags,
	}
}

// Compress compresses an image using TurboJPEG
func Compress(img *Image, params CompressParams) ([]byte, error) {
	encoder := C.tjInitCompress()
	defer C.tjDestroy(encoder)

	var outBuf *C.uchar
	var outBufSize C.ulong

	if params.PixelFormat == PixelFormatGRAY {
		// This is the only valid sampling, so just fix it up if the user forgot to set it
		params.Sampling = SamplingGray
	}

	// int tjCompress2(tjhandle handle, const unsigned char *srcBuf, int width, int pitch, int height, int pixelFormat,
	// unsigned char **jpegBuf, unsigned long *jpegSize, int jpegSubsamp, int jpegQual, int flags);
	res := C.tjCompress2(encoder, (*C.uchar)(&img.Pixels[0]), C.int(img.Width), C.int(img.Stride), C.int(img.Height), C.int(params.PixelFormat),
		&outBuf, &outBufSize, C.int(params.Sampling), C.int(params.Quality), C.int(params.Flags))

	var enc []byte
	err := makeError(encoder, res)
	if outBuf != nil {
		enc = C.GoBytes(unsafe.Pointer(outBuf), C.int(outBufSize))
		C.tjFree(outBuf)
	}

	if err != nil {
		return nil, err
	}
	return enc, nil
}

// Decompress decompresses a JPEG image using TurboJPEG, or PNG image using Go's native PNG library
// The resulting image is RGB for JPEGs, or RGBA/Gray for PNG
func Decompress(encoded []byte) (*Image, error) {
	if len(encoded) > 8 && bytes.Compare(encoded[:8], []byte("\x89\x50\x4e\x47\x0d\x0a\x1a\x0a")) == 0 {
		return decompressPNG(encoded)
	}

	decoder := C.tjInitDecompress()
	defer C.tjDestroy(decoder)

	width := C.int(0)
	height := C.int(0)
	sampling := C.int(0)
	colorspace := C.int(0)

	err := makeError(decoder, C.tjDecompressHeader3(decoder, (*C.uchar)(&encoded[0]), C.ulong(len(encoded)), &width, &height, &sampling, &colorspace))
	if err != nil {
		return nil, err
	}

	outBuf := make([]byte, width*height*3)
	stride := C.int(width * 3)

	// int tjDecompress2(tjhandle handle, const unsigned char *jpegBuf, unsigned long jpegSize, unsigned char *dstBuf,
	// int width, int pitch, int height, int pixelFormat, int flags);
	err = makeError(decoder, C.tjDecompress2(decoder, (*C.uchar)(&encoded[0]), C.ulong(len(encoded)), (*C.uchar)(&outBuf[0]), width, stride, height, C.int(PixelFormatRGB), 0))
	if err != nil {
		return nil, err
	}

	img := &Image{
		Width:  int(width),
		Height: int(height),
		Stride: int(stride),
		NChan:  3,
		Pixels: outBuf,
	}
	return img, nil
}

func decompressPNG(encoded []byte) (*Image, error) {
	img, err := png.Decode(bytes.NewReader(encoded))
	if err != nil {
		return nil, err
	}
	return FromImage(img, true)
}
