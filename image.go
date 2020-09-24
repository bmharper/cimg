package cimg

import (
	"errors"
	"image"
)

// Image is the concrete image type that is used by all functions
// inside cimg.
type Image struct {
	Width  int
	Height int
	Stride int
	NChan  int
	Pixels []byte
}

// NewImage creates a new 8-bit image
func NewImage(width, height, nchan int) *Image {
	return &Image{
		Width:  width,
		Height: height,
		Stride: width * nchan,
		NChan:  nchan,
		Pixels: make([]byte, height*width*nchan),
	}
}

// Wrap an array of bytes into an Image object (do not copy pixels)
func WrapImage(width, height, nchan int, pixels []byte) *Image {
	return &Image{
		Width:  width,
		Height: height,
		Stride: width * nchan,
		NChan:  nchan,
		Pixels: pixels,
	}
}

// Convert a Go image.Image into a turbo.Image
// If allowDeepClone is true, and the source image is type NRGBA or RGBA,
// then the resulting Image points directly to the pixel buffer of the source image.
func FromImage(src image.Image, allowDeepClone bool) (*Image, error) {
	dst := &Image{
		Width:  src.Bounds().Dx(),
		Height: src.Bounds().Dy(),
		Stride: src.Bounds().Dx() * 4,
	}
	switch v := src.(type) {
	case *image.RGBA:
		dst.NChan = 4
		dst.Stride = dst.NChan * dst.Width
		if allowDeepClone {
			dst.Pixels = v.Pix
		} else {
			dst.Pixels = make([]byte, dst.Stride*dst.Height)
			copy(dst.Pixels, v.Pix)
		}
		return dst, nil
	case *image.NRGBA:
		dst.NChan = 4
		dst.Stride = dst.NChan * dst.Width
		if allowDeepClone {
			dst.Pixels = v.Pix
		} else {
			dst.Pixels = make([]byte, dst.Stride*dst.Height)
			copy(dst.Pixels, v.Pix)
		}
		return dst, nil
	}
	return nil, errors.New("Unsupported source image type")

	/*
		// This must be super slow - I haven't tested
		dst.NChan = 4
		dst.Pixels = make([]byte, dst.Stride*dst.Height)
		p := 0
		for y := 0; y < dst.Height; y++ {
			for x := 0; x < dst.Width; x++ {
				r, g, b, a := src.At(x, y).RGBA()
				dst.Pixels[p] = byte(r >> 8)
				dst.Pixels[p+1] = byte(g >> 8)
				dst.Pixels[p+2] = byte(b >> 8)
				dst.Pixels[p+3] = byte(a >> 8)
				p += 4
			}
		}

		return dst
	*/
}
