package cimg

// These functions were built to use https://github.com/dsoprea/go-exif, to make it
// easy to write EXIF data.

import (
	"errors"
	"io"

	"github.com/dsoprea/go-exif/v3"
	jis "github.com/dsoprea/go-jpeg-image-structure/v2"
)

const (
	ExifTagOrientation = 0x112 // Photo orientation
)

// ExifData is a wrapper around github.com/dsoprea/go-exif,
// with only the tags exposed that I've needed to manipulate
type ExifData struct {
	writerRoot *exif.IfdBuilder // Only created when we first start to modify EXIF data
	writerIfd0 *exif.IfdBuilder // Only created when we first start to modify EXIF data
	reader     *exif.Ifd        // Only exists if there is an EXIF block in the file
	segments   *jis.SegmentList // Segments of the JPEG file
}

// Load a JPEG file, and parse it into it's JFIF segments.
// You can then read the existing EXIF data, or alter it.
// Note that if you modify the EXIF data, then reading that same data back from
// this data structure will not reflect your changes. You changes will only be
// reflected if you call Save(), and then reload that file using LoadJpegExif again.
func LoadExif(jpeg []byte) (*ExifData, error) {
	jmp := jis.NewJpegMediaParser()
	ec, err := jmp.ParseBytes(jpeg)
	if err != nil {
		return nil, err
	}
	exifData := &ExifData{
		segments: ec.(*jis.SegmentList),
	}
	reader, _, err := ec.Exif()
	if err != nil {
		// when last checking (v3), the exif package errors didn't obey the errors.Is rules, so we need to resort to string matching
		if errors.Is(err, exif.ErrNoExif) || err.Error() == "no exif data" {
			return exifData, nil
		}
		return nil, err
	}
	exifData.reader = reader
	return exifData, nil
}

// Save writes the image file
func (x *ExifData) Save(w io.Writer) error {
	if x.writerRoot != nil {
		if err := x.segments.SetExif(x.writerRoot); err != nil {
			return err
		}
	}
	return x.segments.Write(w)
}

func (x *ExifData) startModify() error {
	if x.writerRoot != nil {
		return nil
	}
	//x.segments.Print()
	root, err := x.segments.ConstructExifBuilder()
	//x.segments.Print()
	if err != nil {
		return err
	}
	x.writerRoot = root
	// IFD0 = main image
	ifd0, err := exif.GetOrCreateIbFromRootIb(root, "IFD0")
	if err != nil {
		return err
	}
	x.writerIfd0 = ifd0
	// This doesn't work... it looks like the writer and reader are different data structures
	//if x.reader == nil {
	//	reader, _, _ := x.segments.Exif()
	//	x.reader = reader
	//}
	return nil
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
	if x.reader == nil {
		return 0
	}
	res, err := x.reader.FindTagWithId(ExifTagOrientation)
	if err != nil || len(res) == 0 {
		return 0
	}
	val, _ := res[0].Value()
	if v, ok := val.([]uint16); ok && len(v) == 1 {
		return int(v[0])
	}
	return 0
}

// Set photo orientation (See GetOrientation for meaning of the codes)
func (x *ExifData) SetOrientation(orient int) error {
	if err := x.startModify(); err != nil {
		return err
	}
	v := []uint16{uint16(orient)}
	return x.writerIfd0.SetStandard(ExifTagOrientation, v)
}
