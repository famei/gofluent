// Package blur implements a fast separable Gaussian blur for RGBA8888 buffers.
//
// It depends only on the Go standard library (no Qt, no cgo), so it can be
// unit-tested on any platform and reused outside the widget code.
package blur

import "math"

// BlurRGBA performs a separable Gaussian blur, in place, on an RGBA8888 buffer.
//
// buf is the raw pixel buffer and stride is the row stride in bytes
// (>= width*4). width and height describe the image. radius is the kernel
// radius in pixels and sigma controls the blur strength. Pixels outside the
// image are clamped to the nearest edge.
//
// A separable blur runs two 1-D passes (horizontal then vertical), which makes
// it O(width*height*(2*radius+1)) instead of the O((2*radius+1)^2) of a naive
// 2-D convolution. The intermediate pass is kept in float64 so 8-bit rounding
// does not compound between the two passes.
func BlurRGBA(buf []byte, stride, width, height, radius int, sigma float64) {
	if width <= 0 || height <= 0 || radius <= 0 || sigma <= 0 {
		return
	}
	if stride < width*4 || len(buf) < stride*height {
		return
	}

	// Build a normalised 1-D Gaussian kernel.
	kernel := make([]float64, 2*radius+1)
	var sum float64
	for i := -radius; i <= radius; i++ {
		w := math.Exp(-(float64(i * i)) / (2 * sigma * sigma))
		kernel[i+radius] = w
		sum += w
	}
	for i := range kernel {
		kernel[i] /= sum
	}

	// Intermediate buffer in float64 (width*height*4, tightly packed).
	tmp := make([]float64, width*height*4)

	// Horizontal pass: buf -> tmp.
	for y := 0; y < height; y++ {
		row := y * stride
		for x := 0; x < width; x++ {
			var r, g, b, a float64
			for i := -radius; i <= radius; i++ {
				sx := x + i
				if sx < 0 {
					sx = 0
				} else if sx >= width {
					sx = width - 1
				}
				w := kernel[i+radius]
				o := row + sx*4
				r += float64(buf[o]) * w
				g += float64(buf[o+1]) * w
				b += float64(buf[o+2]) * w
				a += float64(buf[o+3]) * w
			}
			o := (y*width + x) * 4
			tmp[o] = r
			tmp[o+1] = g
			tmp[o+2] = b
			tmp[o+3] = a
		}
	}

	// Vertical pass: tmp -> buf, rounding to the nearest 8-bit value.
	for y := 0; y < height; y++ {
		row := y * stride
		for x := 0; x < width; x++ {
			var r, g, b, a float64
			for j := -radius; j <= radius; j++ {
				sy := y + j
				if sy < 0 {
					sy = 0
				} else if sy >= height {
					sy = height - 1
				}
				w := kernel[j+radius]
				o := (sy*width + x) * 4
				r += tmp[o] * w
				g += tmp[o+1] * w
				b += tmp[o+2] * w
				a += tmp[o+3] * w
			}
			o := row + x*4
			buf[o] = clampByte(r)
			buf[o+1] = clampByte(g)
			buf[o+2] = clampByte(b)
			buf[o+3] = clampByte(a)
		}
	}
}

func clampByte(v float64) byte {
	v = math.Round(v)
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return byte(v)
}

// AddNoise adds a subtle, deterministic per-pixel noise to the RGB channels of
// an RGBA8888 buffer. It mimics the grain of the Windows "Acrylic" material and
// prevents colour banding after the heavy Gaussian blur. The noise is a
// position-based hash (not random), so repaints do not flicker.
func AddNoise(buf []byte, stride, width, height, amount int) {
	if width <= 0 || height <= 0 || amount <= 0 || stride < width*4 || len(buf) < stride*height {
		return
	}
	for y := 0; y < height; y++ {
		row := y * stride
		for x := 0; x < width; x++ {
			h := uint32(x*73856093 ^ y*19349663)
			h = (h >> 13) ^ h
			h = h*(h*h*15731+789221) + 1376312589
			delta := int(h%(uint32(amount*2+1))) - amount

			o := row + x*4
			for c := 0; c < 3; c++ {
				v := int(buf[o+c]) + delta
				if v < 0 {
					v = 0
				} else if v > 255 {
					v = 255
				}
				buf[o+c] = byte(v)
			}
		}
	}
}
