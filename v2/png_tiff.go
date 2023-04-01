package cimg

import (
	"bytes"
	"image/png"

	"golang.org/x/image/tiff"
)

func decompressPNG(encoded []byte) (*Image, error) {
	img, err := png.Decode(bytes.NewReader(encoded))
	if err != nil {
		return nil, err
	}
	return FromImage(img, true)
}

func decompressTIFF(encoded []byte) (*Image, error) {
	img, err := tiff.Decode(bytes.NewReader(encoded))
	if err != nil {
		return nil, err
	}
	return FromImage(img, true)
}
