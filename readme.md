# C++ image library bindings for Go

`cimg` is a Go wrapper for various C/C++ image libraries, including:

- libjpeg-turbo
- stb_image_resize2
- Unrotate image so that natural encoding orientation is same as display orientation
- Reading and writing EXIF orientation (provided via native Go code)

Why?

There are a host of high-performance C/C++ libraries out there for
image manipulation, and it's near impossible to write the same kind
of code in Go.

### Example: Compress/Decompress with TurboJPEG

```go
import "github.com/bmharper/cimg"

func compressImage(width, height int, rgb []byte) {
	raw := cimg.Image{
		Width: width,
		Height: height,
		Stride: width * 3,
		Pixels: rgb,
	}
	params := cimg.MakeCompressParams(cimg.PixelFormatRGB, cimg.Sampling420, 35, 0)
	jpg, err := cimg.Compress(&raw, params)
}

func decompressImage(jpg []byte) (*Image, error) {
	return cimg.Decompress(jpg)
}
```

### Example: Read and Modify EXIF Orientation

```go
import "github.com/bmharper/cimg"

func inspectOrientation(jpgRaw []byte) {
	// Parse JPEG/JFIF segments, and read EXIF Orientation tag
	jpgExif, err = cimg.LoadExif(jpgRaw)
	fmt.Printf("Orientation: %v\n", jpgExif.GetOrientation())

	// Modify EXIF rotation.
	// If the file contains no EXIF data, then this will create an
	// EXIF "segment".
	err = jpgExif.SetOrientation(3)
	out := bytes.Buffer{}
	err = jpgExif.Save(&out)
}
```

### Example: Resize with stb_image_resize2

```go
import "github.com/bmharper/cimg"

// Resize from bytes
func resizeImage(srcWidth, srcHeight int, rgba []byte, dstWidth, dstHeight int) *cimg.Image {
	src := cimg.WrapImage(srcWidth, srcHeight, 4, rgba)
	return cimg.ResizeNew(src, dstWidth, dstHeight)
}
```

### C/C++ compiler optimizations

I was initially worried that I needed to add the directive `#cgo CXXFLAGS: -O2`, but it looks like
`cgo` compiles with optimizations on by default. You can verify this by adding `#cgo CXXFLAGS: -O0`
to `resize.go`, and run `go test -bench=.`. Compare with `-O0` and `-O2` and there should be
a big difference. Removing the comment entirely should give similar performance to `-O2`.

### System requirements

I have only tested this on Ubuntu 20.04 `amd64`.

To install the necessary packages:

```
apt install libturbojpeg0-dev
```

### Testing

Warning! Many of the Go unit tests don't actually validate their results. Instead, they
simply write out a JPEG file into the `test` directory. It's your job to visually see that they look correct.
