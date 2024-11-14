package cimg

import (
	"errors"
	"fmt"
	"image"
)

// Image is the concrete image type that is used by all functions inside cimg
type Image struct {
	Pixels        []byte
	Width         int
	Height        int
	Stride        int
	Format        PixelFormat
	Premultiplied bool
}

// NChan returns the number of channels of the pixel format
func NChan(pf PixelFormat) int {
	switch pf {
	case PixelFormatRGB:
		return 3
	case PixelFormatBGR:
		return 3
	case PixelFormatRGBX:
		return 4
	case PixelFormatBGRX:
		return 4
	case PixelFormatXBGR:
		return 4
	case PixelFormatXRGB:
		return 4
	case PixelFormatGRAY:
		return 1
	case PixelFormatRGBA:
		return 4
	case PixelFormatBGRA:
		return 4
	case PixelFormatABGR:
		return 4
	case PixelFormatARGB:
		return 4
	case PixelFormatCMYK:
		return 4
	}
	panic(fmt.Errorf("Unrecognized pixel format %v", pf))
}

// NewImage creates a new 8-bit image
func NewImage(width, height int, format PixelFormat) *Image {
	return &Image{
		Width:         width,
		Height:        height,
		Stride:        width * NChan(format),
		Format:        format,
		Pixels:        make([]byte, height*width*NChan(format)),
		Premultiplied: false,
	}
}

// Wrap an array of bytes into an Image object (do not copy pixels)
func WrapImage(width, height int, format PixelFormat, pixels []byte) *Image {
	return &Image{
		Width:         width,
		Height:        height,
		Stride:        width * NChan(format),
		Format:        format,
		Pixels:        pixels,
		Premultiplied: false,
	}
}

// Wrap an array of bytes into an Image object, with controllable stride (do not copy pixels)
func WrapImageStrided(width, height int, format PixelFormat, pixels []byte, stride int) *Image {
	return &Image{
		Width:         width,
		Height:        height,
		Stride:        stride,
		Format:        format,
		Pixels:        pixels,
		Premultiplied: false,
	}
}

// Convert a Go image.Image into a cimg.Image
// If allowDeepClone is true, and the source image is type GRAY, NRGBA, or RGBA,
// then the resulting Image points directly to the pixel buffer of the source image.
func FromImage(src image.Image, allowDeepClone bool) (*Image, error) {
	dst := &Image{
		Width:  src.Bounds().Dx(),
		Height: src.Bounds().Dy(),
	}
	switch v := src.(type) {
	case *image.Gray:
		dst.Format = PixelFormatGRAY
		dst.Stride = NChan(dst.Format) * dst.Width
		if allowDeepClone {
			dst.Pixels = v.Pix
		} else {
			dst.Pixels = make([]byte, dst.Stride*dst.Height)
			copy(dst.Pixels, v.Pix)
		}
		return dst, nil
	case *image.RGBA:
		dst.Format = PixelFormatRGBA
		dst.Premultiplied = true
		dst.Stride = NChan(dst.Format) * dst.Width
		if allowDeepClone {
			dst.Pixels = v.Pix
		} else {
			dst.Pixels = make([]byte, dst.Stride*dst.Height)
			copy(dst.Pixels, v.Pix)
		}
		return dst, nil
	case *image.NRGBA:
		dst.Format = PixelFormatRGBA
		dst.Premultiplied = false
		dst.Stride = NChan(dst.Format) * dst.Width
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
func (img *Image) ToImage() (image.Image, error) {
	if img.Format == PixelFormatGRAY {
		dst := image.NewGray(image.Rect(0, 0, img.Width, img.Height))
		srcBuf := img.Pixels
		dstBuf := dst.Pix
		for y := 0; y < img.Height; y++ {
			srcP := img.Stride * y
			dstP := dst.Stride * y
			copy(dstBuf[dstP:dstP+dst.Stride], srcBuf[srcP:srcP+img.Stride])
		}
		return dst, nil
	} else if img.Format == PixelFormatRGB || img.Format == PixelFormatBGR {
		dst := image.NewRGBA(image.Rect(0, 0, img.Width, img.Height))
		srcBuf := img.Pixels
		dstBuf := dst.Pix
		width := img.Width
		for y := 0; y < img.Height; y++ {
			srcP := img.Stride * y
			dstP := dst.Stride * y
			if img.Format == PixelFormatBGR {
				// BGR -> RGB
				for x := 0; x < width; x++ {
					dstBuf[dstP] = srcBuf[srcP+2]
					dstBuf[dstP+1] = srcBuf[srcP+1]
					dstBuf[dstP+2] = srcBuf[srcP]
					dstBuf[dstP+3] = 255
					srcP += 3
					dstP += 4
				}
			} else {
				for x := 0; x < width; x++ {
					dstBuf[dstP] = srcBuf[srcP]
					dstBuf[dstP+1] = srcBuf[srcP+1]
					dstBuf[dstP+2] = srcBuf[srcP+2]
					dstBuf[dstP+3] = 255
					srcP += 3
					dstP += 4
				}
			}
		}
		return dst, nil
	} else if img.Format == PixelFormatRGBA {
		var dst image.Image
		var dstStride int
		var dstBuf []uint8
		if img.Premultiplied {
			d := image.NewRGBA(image.Rect(0, 0, img.Width, img.Height))
			dstStride = d.Stride
			dstBuf = d.Pix
			dst = d
		} else {
			d := image.NewNRGBA(image.Rect(0, 0, img.Width, img.Height))
			dstStride = d.Stride
			dstBuf = d.Pix
			dst = d
		}
		srcBuf := img.Pixels
		for y := 0; y < img.Height; y++ {
			srcP := img.Stride * y
			dstP := dstStride * y
			copy(dstBuf[dstP:dstP+dstStride], srcBuf[srcP:srcP+img.Stride])
		}
		return dst, nil
	} else if img.Format == PixelFormatBGRA || img.Format == PixelFormatABGR || img.Format == PixelFormatARGB {
		var dst image.Image
		var dstStride int
		var dstBuf []uint8
		if img.Premultiplied {
			d := image.NewRGBA(image.Rect(0, 0, img.Width, img.Height))
			dstStride = d.Stride
			dstBuf = d.Pix
			dst = d
		} else {
			d := image.NewNRGBA(image.Rect(0, 0, img.Width, img.Height))
			dstStride = d.Stride
			dstBuf = d.Pix
			dst = d
		}
		srcBuf := img.Pixels
		width := img.Width
		for y := 0; y < img.Height; y++ {
			srcP := img.Stride * y
			dstP := dstStride * y
			switch img.Format {
			case PixelFormatBGRA:
				for x := 0; x < width; x++ {
					dstBuf[dstP] = srcBuf[srcP+2]
					dstBuf[dstP+1] = srcBuf[srcP+1]
					dstBuf[dstP+2] = srcBuf[srcP]
					dstBuf[dstP+3] = srcBuf[srcP+3]
					srcP += 4
					dstP += 4
				}
			case PixelFormatABGR:
				for x := 0; x < width; x++ {
					dstBuf[dstP] = srcBuf[srcP+3]
					dstBuf[dstP+1] = srcBuf[srcP+2]
					dstBuf[dstP+2] = srcBuf[srcP+1]
					dstBuf[dstP+3] = srcBuf[srcP]
					srcP += 4
					dstP += 4
				}
			case PixelFormatARGB:
				for x := 0; x < width; x++ {
					dstBuf[dstP] = srcBuf[srcP+1]
					dstBuf[dstP+1] = srcBuf[srcP+2]
					dstBuf[dstP+2] = srcBuf[srcP+3]
					dstBuf[dstP+3] = srcBuf[srcP]
					srcP += 4
					dstP += 4
				}
			}
		}
		return dst, nil
	} else {
		return nil, fmt.Errorf("Unsupported image type %v", img.Format)
	}
}

// Clone returns a deep clone of the image
func (img *Image) Clone() *Image {
	copy := NewImage(img.Width, img.Height, img.Format)
	copy.Premultiplied = img.Premultiplied
	copy.CopyImage(img, 0, 0)
	return copy
}

// NChan returns the number of channels of the pixel format of the image
func (img *Image) NChan() int {
	return NChan(img.Format)
}
