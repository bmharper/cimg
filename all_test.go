package cimg

import (
	"bytes"
	"image/png"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func MakeRGBA(width, height int) *Image {
	img := NewImage(width, height, 4)
	buf := img.Pixels
	g := byte(0)
	b := byte(0)
	p := 0
	a := byte(255)
	for y := 0; y < height; y++ {
		r := byte(0)
		for x := 0; x < width; x++ {
			buf[p] = r
			buf[p+1] = g
			buf[p+2] = b
			buf[p+3] = a
			r += 3
			b += 5
			p += 4
		}
		g += 1
	}
	return img
}

func MakeRGB(width, height int) *Image {
	img := NewImage(width, height, 3)
	buf := img.Pixels
	g := byte(0)
	b := byte(0)
	p := 0
	for y := 0; y < height; y++ {
		r := byte(0)
		for x := 0; x < width; x++ {
			buf[p] = r
			buf[p+1] = g
			buf[p+2] = b
			r += 3
			b += 5
			p += 3
		}
		g += 1
	}
	return img
}

func MakeImage(nchan, width, height int) *Image {
	if nchan == 3 {
		return MakeRGB(width, height)
	} else if nchan == 4 {
		return MakeRGBA(width, height)
	} else {
		panic("Bad channel count")
	}
}

func intAbs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func AvgRGBDifference(img1, img2 *Image) float64 {
	if img1.Width != img2.Width || img1.Height != img2.Height {
		return 1e9
	}
	dr := 0
	dg := 0
	db := 0
	for y := 0; y < img1.Height; y++ {
		p1 := y * img1.Stride
		p2 := y * img2.Stride
		inc1 := img1.NChan
		inc2 := img2.NChan
		for x := 0; x < img1.Width; x++ {
			r := intAbs(int(img1.Pixels[p1]) - int(img2.Pixels[p2]))
			g := intAbs(int(img1.Pixels[p1+1]) - int(img2.Pixels[p2+1]))
			b := intAbs(int(img1.Pixels[p1+2]) - int(img2.Pixels[p2+2]))
			dr += r
			dg += g
			db += b
			p1 += inc1
			p2 += inc2
		}
	}
	npix := img1.Width * img1.Height
	return float64(dr+dg+db) / float64(npix)
}

func SaveJPEG(t *testing.T, img *Image, filename string) {
	format := PixelFormatRGBA
	if img.NChan == 3 {
		format = PixelFormatRGB
	}
	enc, err := Compress(img, MakeCompressParams(format, Sampling444, 95, 0))
	assert.Equal(t, err, nil)
	ioutil.WriteFile(filename, enc, 0660)
}

func TestCompress(t *testing.T) {
	w := 300
	h := 200
	for nchan := 3; nchan <= 4; nchan++ {
		raw1 := MakeImage(nchan, w, h)
		pixelFormat := PixelFormatRGBA
		if nchan == 3 {
			pixelFormat = PixelFormatRGB
		}
		params := MakeCompressParams(pixelFormat, Sampling444, 90, 0)
		jpg, err := Compress(raw1, params)
		t.Logf("Encode return: %v, %v", len(jpg), err)
		raw2, err := Decompress(jpg)
		t.Logf("Decode return: %v x %v, %v, %v, %v", raw2.Width, raw2.Height, raw2.Stride, len(raw2.Pixels), err)
		assert.Equal(t, &w, &raw2.Width, "Width same")
		assert.Equal(t, &h, &raw2.Height, "Height same")
		assert.Equal(t, w*3, raw2.Stride, "Stride")
		//ioutil.WriteFile("test.jpg", jpg, 0660)
	}
}

// This isn't much of a unit test - but at least the code gets run
func TestResize(t *testing.T) {
	w := 700
	h := 400
	for nchan := 3; nchan <= 4; nchan++ {
		org := MakeImage(nchan, w, h)
		small := ResizeNew(org, w/3, h/3)
		big := ResizeNew(org, w*2, h*2)
		SaveJPEG(t, org, "test/resize-org.jpg")
		SaveJPEG(t, small, "test/resize-small.jpg")
		SaveJPEG(t, big, "test/resize-big.jpg")
	}
}

func TestCopyImage(t *testing.T) {
	w := 700
	h := 400
	nchan := 3
	org := MakeImage(nchan, w, h)
	new := MakeImage(nchan, 800, 500)
	org.CopyImageRect(new, 320, 200, 480, 230, 10, 20)
	SaveJPEG(t, org, "test/copyimage.jpg")
}

func TestPNGLoad(t *testing.T) {
	for nchan := 3; nchan <= 4; nchan++ {
		org := MakeImage(nchan, 200, 100)
		nat := org.ToImage()
		buf := bytes.Buffer{}
		err := png.Encode(&buf, nat)
		assert.Nil(t, err)
		v2, err := Decompress(buf.Bytes())
		assert.Nil(t, err)
		diff := AvgRGBDifference(org, v2)
		assert.Equal(t, 0.0, diff)
	}
}

func TestToRGB(t *testing.T) {
	rgba := MakeImage(4, 200, 100)
	rgb := rgba.ToRGB()
	diff := AvgRGBDifference(rgba, rgb)
	assert.Equal(t, 0.0, diff)
}

// Read EXIF data from a known good JPEG file
func TestReadExif(t *testing.T) {
	enc, err := ioutil.ReadFile("test/rotated270.jpg")
	assert.Nil(t, err)
	exif, err := LoadExif(enc)
	assert.Nil(t, err)
	assert.Equal(t, exif.GetOrientation(), 8)
	t.Logf("Orientation: %v", exif.GetOrientation())
}

// Test the whole cycle of creating EXIF data from scratch
func TestReadModifyWriteExif(t *testing.T) {
	// Test a JPEG with no EXIF data
	raw1 := MakeRGBA(20, 20)
	params := MakeCompressParams(PixelFormatRGBA, Sampling444, 90, 0)
	jpg, err := Compress(raw1, params)
	assert.Nil(t, err)
	jpgExif, err := LoadExif(jpg)
	assert.Nil(t, err)
	assert.Equal(t, jpgExif.GetOrientation(), 0)

	// Add orientation to a JPEG without any EXIF data
	err = jpgExif.SetOrientation(6)
	assert.Nil(t, err)
	// Unfortunately this doesn't work, because the reader and writer interfaces are diferent
	// assert.Equal(t, exif.GetOrientation(), 6)

	// We need to save the file first...
	buf := bytes.Buffer{}
	err = jpgExif.Save(&buf)
	assert.Nil(t, err)
	raw2 := buf.Bytes()

	// Finally, if we reload the file, then we get a good orientation tag
	jpgExif, err = LoadExif(raw2)
	assert.Nil(t, err)
	assert.Equal(t, jpgExif.GetOrientation(), 6)

	// Test modifying existing EXIF data
	err = jpgExif.SetOrientation(3)
	assert.Nil(t, err)
	buf = bytes.Buffer{}
	err = jpgExif.Save(&buf)
	assert.Nil(t, err)
	raw3 := buf.Bytes()

	// reload and verify
	jpgExif, err = LoadExif(raw3)
	assert.Nil(t, err)
	assert.Equal(t, jpgExif.GetOrientation(), 3)
}

func LoadJPEG(t *testing.T, filename string) (img *Image, exifOrientation int) {
	buf, err := os.ReadFile(filename)
	assert.Nil(t, err)
	img, err = Decompress(buf)
	assert.Nil(t, err)
	exif, err := LoadExif(buf)
	assert.Nil(t, err)
	exifOrientation = exif.GetOrientation()
	return img, exifOrientation
}

func Unrotate(t *testing.T, orient int, img *Image) *Image {
	unrot, err := UnrotateExif(orient, img)
	assert.Nil(t, err)
	return unrot
}

func UnrotateFile(t *testing.T, filename string) {
	img, orient := LoadJPEG(t, filename)
	SaveJPEG(t, Unrotate(t, orient, img), filename+"-unrotated.jpg")
}

func TestUnrotate(t *testing.T) {
	w := 50
	h := 20
	org := MakeRGBA(w, h)
	SaveJPEG(t, org, "test/unrotated-0.jpg")
	SaveJPEG(t, Unrotate(t, 3, org), "test/unrotated-3.jpg")
	SaveJPEG(t, Unrotate(t, 6, org), "test/unrotated-6.jpg")
	SaveJPEG(t, Unrotate(t, 8, org), "test/unrotated-8.jpg")
	// I don't want to commit these files because they're all 20k because
	// of their bulky EXIF data, and I can't figure out an easy way to
	// remove all the EXIF data except for the orientation.
	//UnrotateFile(t, "test/onceoff-1.jpg")
	//UnrotateFile(t, "test/onceoff-2.jpg")
	//UnrotateFile(t, "test/onceoff-3.jpg")
}

func TestAvgColor(t *testing.T) {
	img1 := MakeRGBA(200, 100)
	avg := img1.AvgColor()
	assert.Equal(t, 4, len(avg))
	assert.EqualValues(t, 115, avg[0])
	assert.EqualValues(t, 49, avg[1])
	assert.EqualValues(t, 127, avg[2])
	assert.EqualValues(t, 255, avg[3])
	t.Logf("AvgColor img1: %v", avg)
	for y := 0; y < img1.Height; y++ {
		for x := 0; x < img1.Width; x++ {
			img1.Pixels[y*img1.Stride+x*4] = 0
			img1.Pixels[y*img1.Stride+x*4+1] = 5
			img1.Pixels[y*img1.Stride+x*4+2] = 6
			img1.Pixels[y*img1.Stride+x*4+3] = 7
		}
	}
	avg = img1.AvgColor()
	t.Logf("AvgColor img1: %v", avg)
	assert.EqualValues(t, 0, avg[0])
	assert.EqualValues(t, 5, avg[1])
	assert.EqualValues(t, 6, avg[2])
	assert.EqualValues(t, 7, avg[3])
}

// On my Skylake 6700K, I get 305ms for resizing 5184x3456 to 1200x800
func BenchmarkResizeRGBA(b *testing.B) {
	w := 5184
	h := 3456
	org := MakeRGBA(w, h)
	for i := 0; i < b.N; i++ {
		ResizeNew(org, 1200, 800)
	}
}

// On my Skylake 6700K, I get 191ms for resizing 5184x3456 to 1200x800
func BenchmarkResizeRGB(b *testing.B) {
	w := 5184
	h := 3456
	org := MakeRGB(w, h)
	for i := 0; i < b.N; i++ {
		ResizeNew(org, 1200, 800)
	}
}
