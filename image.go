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
// If allowDeepClone is true, and the source image is type GRAY, NRGBA, or RGBA,
// then the resulting Image points directly to the pixel buffer of the source image.
func FromImage(src image.Image, allowDeepClone bool) (*Image, error) {
	dst := &Image{
		Width:  src.Bounds().Dx(),
		Height: src.Bounds().Dy(),
	}
	switch v := src.(type) {
	case *image.Gray:
		dst.NChan = 1
		dst.Stride = dst.NChan * dst.Width
		if allowDeepClone {
			dst.Pixels = v.Pix
		} else {
			dst.Pixels = make([]byte, dst.Stride*dst.Height)
			copy(dst.Pixels, v.Pix)
		}
		return dst, nil
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
}

// ToImage returns an image from the Go standard library 'image' package
func (img *Image) ToImage() image.Image {
	if img.NChan == 1 {
		dst := image.NewGray(image.Rect(0, 0, img.Width, img.Height))
		srcBuf := img.Pixels
		dstBuf := dst.Pix
		for y := 0; y < img.Height; y++ {
			srcP := img.Stride * y
			dstP := dst.Stride * y
			copy(dstBuf[dstP:dstP+dst.Stride], srcBuf[srcP:srcP+img.Stride])
		}
		return dst
	} else if img.NChan == 3 {
		dst := image.NewRGBA(image.Rect(0, 0, img.Width, img.Height))
		srcBuf := img.Pixels
		dstBuf := dst.Pix
		width := img.Width
		for y := 0; y < img.Height; y++ {
			srcP := img.Stride * y
			dstP := dst.Stride * y
			for x := 0; x < width; x++ {
				dstBuf[dstP] = srcBuf[srcP]
				dstBuf[dstP+1] = srcBuf[srcP+1]
				dstBuf[dstP+2] = srcBuf[srcP+2]
				dstBuf[dstP+3] = 255
				srcP += 3
				dstP += 4
			}
		}
		return dst
	} else if img.NChan == 4 {
		dst := image.NewNRGBA(image.Rect(0, 0, img.Width, img.Height))
		srcBuf := img.Pixels
		dstBuf := dst.Pix
		for y := 0; y < img.Height; y++ {
			srcP := img.Stride * y
			dstP := dst.Stride * y
			copy(dstBuf[dstP:dstP+dst.Stride], srcBuf[srcP:srcP+img.Stride])
		}
		return dst
	} else {
		return nil
	}
}

// Clone returns a deep clone of the image
func (img *Image) Clone() *Image {
	copy := NewImage(img.Width, img.Height, img.NChan)
	copy.CopyImage(img, 0, 0)
	return copy
}
