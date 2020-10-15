package cimg

/*
#cgo LDFLAGS: -lexif
#include <libexif/exif-data.h>

int GetExifShort(ExifData* ed, int tag, int* val) {
	ExifByteOrder byteOrder = exif_data_get_byte_order(ed);
	ExifEntry* ee = exif_data_get_entry(ed, tag);
	if (ee) {
		*val = exif_get_short(ee->data, byteOrder);
		return 1;
	}
	return 0;
}
*/
import "C"
import (
	"errors"
)

type ExifData struct {
	handle *C.ExifData
}

func LoadExif(jpeg []byte) (*ExifData, error) {
	ptr := C.exif_data_new_from_data((*C.uchar)(&jpeg[0]), C.uint(len(jpeg)))
	if ptr == nil {
		return nil, errors.New("Failed to load EXIF data")
	}
	return &ExifData{handle: ptr}, nil
}

func (x *ExifData) Close() {
	if x.handle != nil {
		C.exif_data_free(x.handle)
		x.handle = nil
	}
}

// See https://www.impulseadventure.com/photo/exif-orientation.html
// 0th Row         0th Column
// 1  top          left side
// 2  top          right side
// 3  bottom       right side
// 4  bottom       left side
// 5  left side    top
// 6  right side   top
// 7  right side   bottom
// 8  left side    bottom
func (x *ExifData) GetOrientation() int {
	val := C.int(0)
	C.GetExifShort(x.handle, C.EXIF_TAG_ORIENTATION, &val)
	return int(val)
}
