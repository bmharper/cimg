package cimg

import (
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

func SaveJPEG(t *testing.T, img *Image, filename string) {
	enc, err := Compress(img, MakeCompressParams(PixelFormatRGBA, Sampling444, 95, 0))
	assert.Equal(t, err, nil)
	ioutil.WriteFile(filename, enc, 0660)
}

func TestCompress(t *testing.T) {
	w := 300
	h := 200
	raw1 := MakeRGBA(w, h)
	params := MakeCompressParams(PixelFormatRGBA, Sampling444, 90, 0)
	jpg, err := Compress(raw1, params)
	t.Logf("Encode return: %v, %v", len(jpg), err)
	raw2, err := Decompress(jpg)
	t.Logf("Decode return: %v x %v, %v, %v, %v", raw2.Width, raw2.Height, raw2.Stride, len(raw2.Pixels), err)
	assert.Equal(t, &w, &raw2.Width, "Width same")
	assert.Equal(t, &h, &raw2.Height, "Height same")
	assert.Equal(t, &raw1.Stride, &raw2.Stride, "Stride same")
	//ioutil.WriteFile("test.jpg", jpg, 0660)
}

// This isn't much of a unit test - but at least the code gets run
func TestResize(t *testing.T) {
	w := 700
	h := 400
	org := MakeRGBA(w, h)
	small := ResizeNew(org, w/3, h/3)
	big := ResizeNew(org, w*2, h*2)
	SaveJPEG(t, org, "test/resize-org.jpg")
	SaveJPEG(t, small, "test/resize-small.jpg")
	SaveJPEG(t, big, "test/resize-big.jpg")
}

func TestExif(t *testing.T) {
	enc, err := ioutil.ReadFile("test/rotated270.jpg")
	assert.Nil(t, err)
	exif, err := LoadExif(enc)
	assert.Nil(t, err)
	defer exif.Close()
	assert.Equal(t, exif.GetOrientation(), 8)
	t.Logf("Orientation: %v", exif.GetOrientation())
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

// On my Skylake 6700K, I get 242ms for resizing 5184x3456 to 1200x800
func BenchmarkResize(b *testing.B) {
	w := 5184
	h := 3456
	org := MakeRGBA(w, h)
	for i := 0; i < b.N; i++ {
		ResizeNew(org, 1200, 800)
	}
}
