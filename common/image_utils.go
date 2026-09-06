package common

import (
	"image"
	"math"
	"os"
	"sort"
	"strings"
	"unsafe"

	// Register PNG/JPEG decoders used by DominantColor.
	_ "image/jpeg"
	_ "image/png"

	qt "github.com/mappu/miqt/qt"
)

// GaussianBlur returns a blurred copy of the image, mirroring the Python
// scipy-based gaussianBlur: the pixmap is (optionally) shrunk to blurPicSize,
// its RGB channels are blurred with a gaussian kernel of sigma=blurRadius and
// scaled by brightFactor, and the alpha channel is preserved. This provides the
// real "background texture" used by the acrylic/material widgets.
func GaussianBlur(image *qt.QPixmap, blurRadius int, brightFactor float64, blurPicSize []int) *qt.QPixmap {
	if image == nil || image.IsNull() {
		return image
	}
	if blurRadius <= 0 {
		blurRadius = 1
	}
	if brightFactor <= 0 {
		brightFactor = 1
	}

	img := image.ToImage()
	if img == nil {
		return image
	}

	// Shrink first to bound the blur cost (never enlarge), matching the Python
	// blurPicSize behaviour.
	if len(blurPicSize) >= 2 && blurPicSize[0] > 0 && blurPicSize[1] > 0 {
		w, h := img.Width(), img.Height()
		if w > 0 && h > 0 {
			ratio := math.Min(float64(blurPicSize[0])/float64(w), float64(blurPicSize[1])/float64(h))
			if ratio < 1 {
				nw, nh := int(float64(w)*ratio), int(float64(h)*ratio)
				if nw < 1 {
					nw = 1
				}
				if nh < 1 {
					nh = 1
				}
				scaled := img.Scaled3(nw, nh, qt.IgnoreAspectRatio, qt.SmoothTransformation)
				if scaled == nil {
					return image
				}
				img = scaled
			}
		}
	}
	// RGBA8888 gives a predictable 4-byte-per-pixel, endian-independent layout.
	rgba := img.ConvertToFormat(qt.QImage__Format_RGBA8888)
	if rgba == nil {
		return image
	}

	w, h := rgba.Width(), rgba.Height()
	if w <= 0 || h <= 0 {
		return image
	}

	src := copyImageBytes(rgba.Bits(), w, h, rgba.BytesPerLine())
	if src == nil {
		return image
	}

	kernel := gaussianKernel(float64(blurRadius))
	for ch := 0; ch < 3; ch++ { // blur R, G, B; keep alpha
		blurChannel(src, w, h, ch, kernel, brightFactor)
	}

	out := qt.NewQImage3(w, h, qt.QImage__Format_RGBA8888)
	if out == nil {
		return image
	}
	defer out.Delete()
	writeImageBytes(out.Bits(), w, h, out.BytesPerLine(), src)

	result := qt.NewQPixmap()
	result.ConvertFromImage(out)
	return result
}

// BlendTint blends a tint color over the pixmap's RGB channels (the pixmap's
// alpha channel is left unchanged). The tint is applied in pure Go so the result
// is independent of QPainter compositing behaviour:
//   - a <= 0      → return pm untouched (no tint)
//   - 0 < a < 255 → per-channel blend: dst = dst*(255-a)/255 + tint*a/255
//   - a >= 255    → every pixel's RGB becomes the tint RGB (a fully opaque tint
//     is a solid fill, not the original image)
func BlendTint(pm *qt.QPixmap, color *qt.QColor) *qt.QPixmap {
	if pm == nil || pm.IsNull() || color == nil {
		return pm
	}
	a := color.Alpha()
	if a <= 0 {
		return pm
	}

	img := pm.ToImage()
	if img == nil {
		return pm
	}
	rgba := img.ConvertToFormat(qt.QImage__Format_RGBA8888)
	if rgba == nil {
		return pm
	}
	w, h := rgba.Width(), rgba.Height()
	if w <= 0 || h <= 0 {
		return pm
	}
	data := copyImageBytes(rgba.Bits(), w, h, rgba.BytesPerLine())
	if data == nil {
		return pm
	}

	tr, tg, tb := color.Red(), color.Green(), color.Blue()
	inv := 255 - a
	for i := 0; i < w*h; i++ {
		o := i * 4
		data[o] = byte((int(data[o])*inv + tr*a) / 255)
		data[o+1] = byte((int(data[o+1])*inv + tg*a) / 255)
		data[o+2] = byte((int(data[o+2])*inv + tb*a) / 255)
	}

	out := qt.NewQImage3(w, h, qt.QImage__Format_RGBA8888)
	if out == nil {
		return pm
	}
	writeImageBytes(out.Bits(), w, h, out.BytesPerLine(), data)
	result := qt.NewQPixmap()
	result.ConvertFromImage(out)
	out.Delete()
	return result
}

// copyImageBytes reads a QImage's raw RGBA bytes into a compact Go slice with
// stride 4*width (dropping any per-row padding of the source).
func copyImageBytes(bits *byte, w, h, stride int) []byte {
	if bits == nil || w <= 0 || h <= 0 || stride < w*4 {
		return nil
	}
	out := make([]byte, w*h*4)
	src := unsafe.Slice(bits, stride*h)
	for y := 0; y < h; y++ {
		copy(out[y*w*4:(y+1)*w*4], src[y*stride:y*stride+w*4])
	}
	return out
}

// writeImageBytes writes compact RGBA bytes into a QImage's raw buffer,
// respecting the destination row stride.
func writeImageBytes(bits *byte, w, h, stride int, data []byte) {
	if bits == nil || data == nil || w <= 0 || h <= 0 || stride < w*4 {
		return
	}
	dst := unsafe.Slice(bits, stride*h)
	for y := 0; y < h; y++ {
		copy(dst[y*stride:y*stride+w*4], data[y*w*4:(y+1)*w*4])
	}
}

// gaussianKernel builds a normalized, truncated 1-D gaussian kernel.
func gaussianKernel(sigma float64) []float64 {
	radius := int(math.Ceil(3 * sigma))
	if radius < 1 {
		radius = 1
	}
	k := make([]float64, 2*radius+1)
	twoSigmaSq := 2 * sigma * sigma
	sum := 0.0
	for i := -radius; i <= radius; i++ {
		v := math.Exp(-float64(i*i) / twoSigmaSq)
		k[i+radius] = v
		sum += v
	}
	for i := range k {
		k[i] /= sum
	}
	return k
}

// blurChannel performs a separable gaussian blur on one byte channel of a
// compact RGBA buffer (stride 4*width) and scales the result by brightFactor.
// Edges are clamped.
func blurChannel(data []byte, w, h, ch int, kernel []float64, brightFactor float64) {
	radius := (len(kernel) - 1) / 2

	// Horizontal pass.
	tmp := make([]float64, w*h)
	for y := 0; y < h; y++ {
		row := y * w * 4
		for x := 0; x < w; x++ {
			sum := 0.0
			for k, kv := range kernel {
				xx := x + k - radius
				if xx < 0 {
					xx = 0
				} else if xx >= w {
					xx = w - 1
				}
				sum += float64(data[row+xx*4+ch]) * kv
			}
			tmp[y*w+x] = sum
		}
	}

	// Vertical pass.
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			sum := 0.0
			for k, kv := range kernel {
				yy := y + k - radius
				if yy < 0 {
					yy = 0
				} else if yy >= h {
					yy = h - 1
				}
				sum += tmp[yy*w+x] * kv
			}
			v := sum * brightFactor
			if v < 0 {
				v = 0
			} else if v > 255 {
				v = 255
			}
			data[y*w*4+x*4+ch] = byte(v)
		}
	}
}

// RGB2HSV converts an RGB triple to HSV (h in [0,360), s/v in [0,1]).
func RGB2HSV(r, g, b int) (float64, float64, float64) {
	rf, gf, bf := float64(r)/255, float64(g)/255, float64(b)/255
	mx := math.Max(rf, math.Max(gf, bf))
	mn := math.Min(rf, math.Min(gf, bf))
	df := mx - mn

	var h float64
	switch {
	case mx == mn:
		h = 0
	case mx == rf:
		h = math.Mod(60*((gf-bf)/df)+360, 360)
	case mx == gf:
		h = 60*((bf-rf)/df) + 120
	case mx == bf:
		h = 60*((rf-gf)/df) + 240
	}

	s := 0.0
	if mx != 0 {
		s = df / mx
	}
	return h, s, mx
}

// HSV2RGB converts HSV (h in [0,360), s/v in [0,1]) to RGB.
func HSV2RGB(h, s, v float64) (int, int, int) {
	h60 := h / 60.0
	h60f := math.Floor(h60)
	hi := int(h60f) % 6
	f := h60 - h60f
	p := v * (1 - s)
	q := v * (1 - f*s)
	t := v * (1 - (1-f)*s)

	var rf, gf, bf float64
	switch hi {
	case 0:
		rf, gf, bf = v, t, p
	case 1:
		rf, gf, bf = q, v, p
	case 2:
		rf, gf, bf = p, v, t
	case 3:
		rf, gf, bf = p, q, v
	case 4:
		rf, gf, bf = t, p, v
	case 5:
		rf, gf, bf = v, p, q
	}
	return int(rf*255 + 0.5), int(gf*255 + 0.5), int(bf*255 + 0.5)
}

// Colorfulness computes the Hasler–Süsstrunk colorfulness metric for a single
// RGB pixel (matching the effective numpy behaviour of the Python port).
func Colorfulness(r, g, b int) float64 {
	rg := math.Abs(float64(r - g))
	yb := math.Abs(0.5*float64(r+g) - float64(b))
	return 0.3 * math.Hypot(rg, yb)
}

// DominantColor extracts the dominant color of an image.
type DominantColor struct{}

// GetDominantColor extracts the dominant (r, g, b) color from an image path.
// Qt resource paths (":...") return (24, 24, 24) like the Python implementation.
func (DominantColor) GetDominantColor(imagePath string) (int, int, int) {
	if strings.HasPrefix(imagePath, ":") {
		return 24, 24, 24
	}

	f, err := os.Open(imagePath)
	if err != nil {
		return 24, 24, 24
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return 24, 24, 24
	}

	// Scale down to at most 400px on the longest side.
	if max(img.Bounds().Dx(), img.Bounds().Dy()) > 400 {
		img = resizeMax(img, 400)
	}

	palette := extractPalette(img, 9)
	palette = adjustPaletteValue(palette)

	// Remove near-gray colors (h < 0.02); keep at least two.
	filtered := palette[:0]
	for _, c := range palette {
		h, _, _ := RGB2HSV(c[0], c[1], c[2])
		if h < 0.02 {
			if len(filtered) <= 2 {
				filtered = append(filtered, c)
				continue
			}
			continue
		}
		filtered = append(filtered, c)
	}
	palette = filtered

	if len(palette) > 5 {
		palette = palette[:5]
	}
	sort.SliceStable(palette, func(i, j int) bool {
		return Colorfulness(palette[i][0], palette[i][1], palette[i][2]) >
			Colorfulness(palette[j][0], palette[j][1], palette[j][2])
	})

	if len(palette) == 0 {
		return 24, 24, 24
	}
	return palette[0][0], palette[0][1], palette[0][2]
}

// adjustPaletteValue dims very bright palette colors (ports __adjustPaletteValue).
func adjustPaletteValue(palette [][3]int) [][3]int {
	out := make([][3]int, 0, len(palette))
	for _, c := range palette {
		h, s, v := RGB2HSV(c[0], c[1], c[2])
		var factor float64
		switch {
		case v > 0.9:
			factor = 0.8
		case v > 0.8:
			factor = 0.9
		case v > 0.7:
			factor = 0.95
		default:
			factor = 1
		}
		r, g, b := HSV2RGB(h, s, v*factor)
		out = append(out, [3]int{r, g, b})
	}
	return out
}

// extractPalette quantizes the image colors and returns the most frequent.
func extractPalette(img image.Image, quality int) [][3]int {
	bounds := img.Bounds()
	type bucket struct {
		r, g, b int
		count   int
	}
	counts := map[[3]int]int{}
	shift := uint(4) // quantize to 16 levels per channel
	if quality <= 0 {
		quality = 1
	}
	step := quality

	for y := bounds.Min.Y; y < bounds.Max.Y; y += step {
		for x := bounds.Min.X; x < bounds.Max.X; x += step {
			r, g, b, a := img.At(x, y).RGBA()
			if a>>8 < 128 {
				continue // skip transparent pixels
			}
			key := [3]int{int(r>>(8+shift)) << shift, int(g>>(8+shift)) << shift, int(b>>(8+shift)) << shift}
			counts[key]++
		}
	}

	buckets := make([]bucket, 0, len(counts))
	for k, c := range counts {
		buckets = append(buckets, bucket{r: k[0], g: k[1], b: k[2], count: c})
	}
	sort.SliceStable(buckets, func(i, j int) bool { return buckets[i].count > buckets[j].count })

	const maxColors = 10
	if len(buckets) > maxColors {
		buckets = buckets[:maxColors]
	}
	palette := make([][3]int, 0, len(buckets))
	for _, b := range buckets {
		palette = append(palette, [3]int{b.r, b.g, b.b})
	}
	return palette
}

// resizeMax scales an image so its longest side does not exceed maxDim
// (nearest-neighbour, sufficient for palette extraction).
func resizeMax(src image.Image, maxDim int) image.Image {
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	longest := max(w, h)
	if longest <= maxDim {
		return src
	}
	scale := float64(maxDim) / float64(longest)
	nw, nh := int(float64(w)*scale), int(float64(h)*scale)
	if nw < 1 {
		nw = 1
	}
	if nh < 1 {
		nh = 1
	}
	dst := image.NewRGBA(image.Rect(0, 0, nw, nh))
	for y := 0; y < nh; y++ {
		for x := 0; x < nw; x++ {
			sx := b.Min.X + int(float64(x)/float64(nw)*float64(w))
			sy := b.Min.Y + int(float64(y)/float64(nh)*float64(h))
			dst.Set(x, y, src.At(sx, sy))
		}
	}
	return dst
}
