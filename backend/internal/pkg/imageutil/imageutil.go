// Package imageutil provides decoding, resizing, rotation and JPEG encoding
// for jpeg/png/webp images using only the standard library and x/image.
package imageutil

import (
	"bytes"
	"encoding/binary"
	"errors"
	"image"
	"image/color"
	stdDraw "image/draw"
	"image/jpeg"
	_ "image/png"
	"io"
	"math"

	xdraw "golang.org/x/image/draw"
	"golang.org/x/image/math/f64"
	_ "golang.org/x/image/webp"
)

// maxDecodedPixels bounds the memory a single decoded image may use.
const maxDecodedPixels = 64 << 20

var (
	ErrImageTooLarge   = errors.New("image dimensions exceed limit")
	ErrImageDimensions = errors.New("image has invalid dimensions")
)

// Decode decodes jpeg/png/webp data after checking its declared
// dimensions, then applies the EXIF orientation (JPEG only).
func Decode(data []byte) (image.Image, error) {
	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	if cfg.Width <= 0 || cfg.Height <= 0 {
		return nil, ErrImageDimensions
	}
	if int64(cfg.Width)*int64(cfg.Height) > maxDecodedPixels {
		return nil, ErrImageTooLarge
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	return applyOrientation(img, jpegOrientation(data)), nil
}

// NewRGBA returns a copy of src as a mutable RGBA image with origin (0,0).
func NewRGBA(src image.Image) *image.RGBA {
	b := src.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	stdDraw.Draw(dst, dst.Bounds(), src, b.Min, stdDraw.Src)
	return dst
}

// Resize scales src to w x h using Catmull-Rom interpolation.
func Resize(src image.Image, w, h int) *image.RGBA {
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	xdraw.CatmullRom.Scale(dst, dst.Bounds(), src, src.Bounds(), xdraw.Src, nil)
	return dst
}

// Rotate rotates src counter-clockwise by angle degrees, growing the
// canvas to fit and filling the corners with transparent pixels.
func Rotate(src image.Image, angle float64) *image.RGBA {
	b := src.Bounds()
	sw, sh := float64(b.Dx()), float64(b.Dy())
	rad := angle * math.Pi / 180
	sin, cos := math.Sin(rad), math.Cos(rad)

	dw := max(1, int(math.Ceil(math.Abs(sw*cos)+math.Abs(sh*sin)-1e-9)))
	dh := max(1, int(math.Ceil(math.Abs(sw*sin)+math.Abs(sh*cos)-1e-9)))
	dst := image.NewRGBA(image.Rect(0, 0, dw, dh))

	scx, scy := sw/2+float64(b.Min.X), sh/2+float64(b.Min.Y)
	dcx, dcy := float64(dw)/2, float64(dh)/2

	m := f64.Aff3{
		cos, sin, dcx - cos*scx - sin*scy,
		-sin, cos, dcy + sin*scx - cos*scy,
	}
	xdraw.BiLinear.Transform(dst, m, src, b, xdraw.Over, nil)
	return dst
}

// EncodeJPEG writes img as a JPEG with the given quality.
func EncodeJPEG(w io.Writer, img image.Image, quality int) error {
	return jpeg.Encode(w, img, &jpeg.Options{Quality: quality})
}

// applyOrientation transforms img according to an EXIF orientation (1-8).
func applyOrientation(img image.Image, orientation int) image.Image {
	if orientation < 2 || orientation > 8 {
		return img
	}

	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	dw, dh := w, h
	if orientation >= 5 {
		dw, dh = h, w
	}

	dst := image.NewRGBA(image.Rect(0, 0, dw, dh))
	for y := 0; y < dh; y++ {
		for x := 0; x < dw; x++ {
			var sx, sy int
			switch orientation {
			case 2:
				sx, sy = w-1-x, y
			case 3:
				sx, sy = w-1-x, h-1-y
			case 4:
				sx, sy = x, h-1-y
			case 5:
				sx, sy = y, x
			case 6:
				sx, sy = y, h-1-x
			case 7:
				sx, sy = w-1-y, h-1-x
			case 8:
				sx, sy = w-1-y, x
			}
			dst.Set(x, y, color.RGBAModel.Convert(img.At(b.Min.X+sx, b.Min.Y+sy)))
		}
	}
	return dst
}

// jpegOrientation returns the EXIF orientation of JPEG data, or 1 if the
// data is not a JPEG or carries no valid orientation.
func jpegOrientation(data []byte) int {
	if len(data) < 4 || data[0] != 0xFF || data[1] != 0xD8 {
		return 1
	}

	pos := 2
	for pos+4 <= len(data) {
		if data[pos] != 0xFF {
			return 1
		}
		marker := data[pos+1]
		if marker == 0xFF { // fill byte
			pos++
			continue
		}
		if marker == 0xDA || marker == 0xD9 { // start of scan / EOI
			return 1
		}
		size := int(binary.BigEndian.Uint16(data[pos+2:]))
		if size < 2 || pos+2+size > len(data) {
			return 1
		}
		seg := data[pos+4 : pos+2+size]
		if marker == 0xE1 && bytes.HasPrefix(seg, []byte("Exif\x00\x00")) {
			return tiffOrientation(seg[6:])
		}
		pos += 2 + size
	}
	return 1
}

func tiffOrientation(t []byte) int {
	if len(t) < 8 {
		return 1
	}
	var bo binary.ByteOrder
	switch string(t[:2]) {
	case "II":
		bo = binary.LittleEndian
	case "MM":
		bo = binary.BigEndian
	default:
		return 1
	}

	ifd := int(bo.Uint32(t[4:]))
	if ifd < 8 || ifd+2 > len(t) {
		return 1
	}
	count := int(bo.Uint16(t[ifd:]))
	for i := range count {
		e := ifd + 2 + i*12
		if e+12 > len(t) {
			return 1
		}
		if bo.Uint16(t[e:]) == 0x0112 {
			if v := int(bo.Uint16(t[e+8:])); v >= 1 && v <= 8 {
				return v
			}
			return 1
		}
	}
	return 1
}
