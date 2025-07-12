package cimg

import (
	"bytes"
	"image/png"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func MakeRGBA(width, height int) *Image {
	img := NewImage(width, height, PixelFormatRGBA)
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
	img := NewImage(width, height, PixelFormatRGB)
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

func MakeGray(width, height int) *Image {
	img := NewImage(width, height, PixelFormatGRAY)
	buf := img.Pixels
	g := byte(0)
	p := 0
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			buf[p] = g
			g += 3
		}
		g += 30
	}
	return img
}

func MakeImage(nchan, width, height int) *Image {
	if nchan == 1 {
		return MakeGray(width, height)
	} else if nchan == 3 {
		return MakeRGB(width, height)
	} else if nchan == 4 {
		return MakeRGBA(width, height)
	} else {
		panic("Bad channel count")
	}
}

func AddAlphaNoise(img *Image) {
	for y := 0; y < img.Height; y++ {
		line := img.Stride * y
		a := uint8(y)
		for x := 0; x < img.Width; x++ {
			img.Pixels[line+3] = a
			a += 3
			line += 4
		}
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
	isGray := img1.NChan() == 1
	for y := 0; y < img1.Height; y++ {
		p1 := y * img1.Stride
		p2 := y * img2.Stride
		inc1 := img1.NChan()
		inc2 := img2.NChan()
		for x := 0; x < img1.Width; x++ {
			r := intAbs(int(img1.Pixels[p1]) - int(img2.Pixels[p2]))
			g := 0
			b := 0
			if !isGray {
				g = intAbs(int(img1.Pixels[p1+1]) - int(img2.Pixels[p2+1]))
				b = intAbs(int(img1.Pixels[p1+2]) - int(img2.Pixels[p2+2]))
			}
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
	enc, err := Compress(img, MakeCompressParams(Sampling444, 95, 0))
	require.Equal(t, err, nil)
	os.WriteFile(filename, enc, 0660)
}

func TestCompress(t *testing.T) {
	w := 300
	h := 200
	for nchan := 3; nchan <= 4; nchan++ {
		raw1 := MakeImage(nchan, w, h)
		params := MakeCompressParams(Sampling444, 90, 0)
		jpg, err := Compress(raw1, params)
		t.Logf("Encode return: %v, %v", len(jpg), err)
		raw2, err := Decompress(jpg)
		t.Logf("Decode return: %v x %v, %v, %v, %v", raw2.Width, raw2.Height, raw2.Stride, len(raw2.Pixels), err)
		require.Equal(t, &w, &raw2.Width, "Width same")
		require.Equal(t, &h, &raw2.Height, "Height same")
		require.Equal(t, w*3, raw2.Stride, "Stride")
		//ioutil.WriteFile("test.jpg", jpg, 0660)
	}
}

// This isn't much of a unit test - but at least the code gets run
func TestResize(t *testing.T) {
	w := 700
	h := 400
	for nchan := 3; nchan <= 4; nchan++ {
		org := MakeImage(nchan, w, h)
		small := ResizeNew(org, w/3, h/3, nil)
		big := ResizeNew(org, w*2, h*2, nil)
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
	chans := []int{1, 3, 4}
	for _, nchan := range chans {
		org := MakeImage(nchan, 200, 100)
		nat, err := org.ToImage()
		require.Nil(t, err)
		buf := bytes.Buffer{}
		err = png.Encode(&buf, nat)
		require.Nil(t, err)
		v2, err := Decompress(buf.Bytes())
		require.Nil(t, err)
		// require.Equal(t, org.NChan, v2.NChan) -- not true, because Go PNG lib opens RGB images as RGBA
		diff := AvgRGBDifference(org, v2)
		require.Equal(t, 0.0, diff)
	}
}

// This was a bug, but it went away when I added explicit PixelFormat into Image
//func TestPNGBad(t *testing.T) {
//	raw, _ := os.ReadFile("/home/ben/Downloads/ptguiviewer_icon.png")
//	img, err := Decompress(raw)
//	require.Nil(t, err)
//	img.Matte(255, 255, 255)
//	j, _ := Compress(img, MakeCompressParams(Sampling444, 95, 0))
//	os.WriteFile("test/foo.jpg", j, 0666)
//}

func TestMatte(t *testing.T) {
	rgba := MakeImage(4, 200, 200)
	AddAlphaNoise(rgba)
	SaveJPEG(t, rgba, "test/alpha1.jpg")

	white := rgba.Clone()
	white.Matte(255, 255, 255)
	SaveJPEG(t, white, "test/alphaWhiteMatte.jpg")

	black := rgba.Clone()
	black.Matte(0, 0, 0)
	SaveJPEG(t, black, "test/alphaBlackMatte.jpg")

	green := rgba.Clone()
	green.Matte(0, 255, 0)
	SaveJPEG(t, green, "test/alphaGreenMatteGoodPremul.jpg")

	green = rgba.Clone()
	green.Premultiplied = true // This is a lie
	green.Matte(0, 255, 0)
	SaveJPEG(t, green, "test/alphaGreenMatteBadPremul.jpg")

	// invalid on RGB
	rgb := MakeImage(3, 200, 200)
	rgb.Matte(0, 255, 0)
}

func TestToRGB(t *testing.T) {
	rgba := MakeImage(4, 200, 100)
	rgb := rgba.ToRGB()
	diff := AvgRGBDifference(rgba, rgb)
	require.Equal(t, 0.0, diff)
}

// Read EXIF data from a known good JPEG file
func TestReadExif(t *testing.T) {
	enc, err := ioutil.ReadFile("test/rotated270.jpg")
	require.Nil(t, err)
	exif, err := LoadExif(enc)
	require.Nil(t, err)
	require.Equal(t, exif.GetOrientation(), 8)
	t.Logf("Orientation: %v", exif.GetOrientation())
}

// Test the whole cycle of creating EXIF data from scratch
func TestReadModifyWriteExif(t *testing.T) {
	// Test a JPEG with no EXIF data
	raw1 := MakeRGBA(20, 20)
	params := MakeCompressParams(Sampling444, 90, 0)
	jpg, err := Compress(raw1, params)
	require.Nil(t, err)
	jpgExif, err := LoadExif(jpg)
	require.Nil(t, err)
	require.Equal(t, jpgExif.GetOrientation(), 0)

	// Add orientation to a JPEG without any EXIF data
	err = jpgExif.SetOrientation(6)
	require.Nil(t, err)
	// Unfortunately this doesn't work, because the reader and writer interfaces are diferent
	// require.Equal(t, exif.GetOrientation(), 6)

	// We need to save the file first...
	buf := bytes.Buffer{}
	err = jpgExif.Save(&buf)
	require.Nil(t, err)
	raw2 := buf.Bytes()

	// Finally, if we reload the file, then we get a good orientation tag
	jpgExif, err = LoadExif(raw2)
	require.Nil(t, err)
	require.Equal(t, jpgExif.GetOrientation(), 6)

	// Test modifying existing EXIF data
	err = jpgExif.SetOrientation(3)
	require.Nil(t, err)
	buf = bytes.Buffer{}
	err = jpgExif.Save(&buf)
	require.Nil(t, err)
	raw3 := buf.Bytes()

	// reload and verify
	jpgExif, err = LoadExif(raw3)
	require.Nil(t, err)
	require.Equal(t, jpgExif.GetOrientation(), 3)
}

func LoadJPEG(t *testing.T, filename string) (img *Image, exifOrientation int) {
	buf, err := os.ReadFile(filename)
	require.Nil(t, err)
	img, err = Decompress(buf)
	require.Nil(t, err)
	exif, err := LoadExif(buf)
	require.Nil(t, err)
	exifOrientation = exif.GetOrientation()
	return img, exifOrientation
}

func Unrotate(t *testing.T, orient int, img *Image) *Image {
	unrot, err := UnrotateExif(orient, img)
	require.Nil(t, err)
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
	require.Equal(t, 4, len(avg))
	require.EqualValues(t, 115, avg[0])
	require.EqualValues(t, 49, avg[1])
	require.EqualValues(t, 127, avg[2])
	require.EqualValues(t, 255, avg[3])
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
	require.EqualValues(t, 0, avg[0])
	require.EqualValues(t, 5, avg[1])
	require.EqualValues(t, 6, avg[2])
	require.EqualValues(t, 7, avg[3])
}

func TestCrop(t *testing.T) {
	i1 := MakeRGBA(101, 95)
	i2 := i1.ReferenceCrop(5, 6, 60, 70)
	require.NotEqual(t, byte(123), i1.Pixels[i1.PixelByte(5, 6)])
	// Modify a pixel on i2, and make sure that the change also occurs on i1
	i2.Pixels[0] = 123
	require.Equal(t, byte(123), i1.Pixels[i1.PixelByte(5, 6)])
	require.Equal(t, byte(123), i2.Pixels[i2.PixelByte(0, 0)])
	SaveJPEG(t, i1, "test/crop1.jpg")
	SaveJPEG(t, i2, "test/crop2.jpg")
}

func TestDrawRect(t *testing.T) {
	i1 := MakeRGB(97, 123)
	i1.DrawRectangle(10, 20, 50, 80, 255, 0, 0)        // Red rectangle
	i1.DrawRectangle(20, 30, 60, 90, 0, 255, 0)        // Green rectangle
	i1.DrawRectangle(30, 40, 70, 100, 0, 0, 255)       // Blue rectangle
	i1.DrawRectangle(-5, -5, 15, 18, 250, 250, 250)    // crop top-left
	i1.DrawRectangle(80, 100, 120, 150, 250, 250, 250) // crop bottom-right
	SaveJPEG(t, i1, "test/drawrect.jpg")
}

// On my Skylake 6700K, I get 305ms for resizing 5184x3456 to 1200x800
func BenchmarkResizeRGBA(b *testing.B) {
	w := 5184
	h := 3456
	org := MakeRGBA(w, h)
	for i := 0; i < b.N; i++ {
		ResizeNew(org, 1200, 800, nil)
	}
}

// On my Skylake 6700K, I get 191ms for resizing 5184x3456 to 1200x800
func BenchmarkResizeRGBDown(b *testing.B) {
	w := 5184
	h := 3456
	org := MakeRGB(w, h)
	for i := 0; i < b.N; i++ {
		ResizeNew(org, 1200, 800, nil)
	}
}

func BenchmarkResizeRGBUp(b *testing.B) {
	w := 320
	h := 240
	org := MakeRGB(w, h)
	params := ResizeParams{
		Filter:          ResizeFilterBox,
		CheapSRGBFilter: true,
	}
	for i := 0; i < b.N; i++ {
		ResizeNew(org, 640, 640, &params)
	}
}
