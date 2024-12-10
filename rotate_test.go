package cimg

import (
	"fmt"
	"math"
	"testing"
)

const Deg2Rad = math.Pi / 180

func TestRotate(t *testing.T) {
	img := MakeImage(3, 600, 400)

	for _, pf := range []PixelFormat{PixelFormatRGB, PixelFormatRGBA, PixelFormatGRAY} {
		var asformat *Image
		if pf == PixelFormatGRAY {
			asformat = img.ToGray()
		} else if pf == PixelFormatRGB {
			asformat = img.ToRGB()
		} else if pf == PixelFormatRGBA {
			asformat = img.ToRGBA(255)
		}
		fn := fmt.Sprintf("test/rotated-%v_.jpg", pf)
		asformat.WriteJPEG(fn, MakeCompressParams(Sampling444, 99, 0), 0644)
		for _, angle := range []float64{0, 1, 5} {
			rotated := NewImage(asformat.Width, asformat.Height, asformat.Format)
			Rotate(asformat, rotated, angle*Deg2Rad, nil)
			fn := fmt.Sprintf("test/rotated-%v_%v.jpg", pf, int(angle))
			rotated.WriteJPEG(fn, MakeCompressParams(Sampling444, 99, 0), 0644)
		}
	}
}
