package blur

import "testing"

func TestBlurRGBA_PreservesSolidColor(t *testing.T) {
	const w, h = 8, 8
	stride := w * 4
	buf := make([]byte, stride*h)
	for i := range buf {
		buf[i] = 128
	}
	BlurRGBA(buf, stride, w, h, 2, 1.0)
	for _, b := range buf {
		if b != 128 {
			t.Fatalf("solid color changed: got %d want 128", b)
		}
	}
}

func TestBlurRGBA_SpreadsImpulse(t *testing.T) {
	const w, h = 21, 21
	stride := w * 4
	buf := make([]byte, stride*h)

	// Single white pixel in the centre.
	cx, cy := w/2, h/2
	o := cy*stride + cx*4
	buf[o], buf[o+1], buf[o+2], buf[o+3] = 255, 255, 255, 255

	BlurRGBA(buf, stride, w, h, 3, 1.5)

	sum := func(x, y int) int {
		o := y*stride + x*4
		return int(buf[o]) + int(buf[o+1]) + int(buf[o+2])
	}

	center := sum(cx, cy)
	neighbor := sum(cx+1, cy)
	far := sum(0, 0)

	if center <= neighbor {
		t.Fatalf("center (%d) should be brighter than neighbour (%d)", center, neighbor)
	}
	if neighbor <= 0 {
		t.Fatalf("neighbour should have received some energy, got %d", neighbor)
	}
	if far >= neighbor {
		t.Fatalf("far corner (%d) should be darker than neighbour (%d)", far, neighbor)
	}
}

func TestBlurRGBA_ApproxConservesEnergy(t *testing.T) {
	const w, h = 31, 31
	stride := w * 4
	buf := make([]byte, stride*h)

	// A large bright square in the middle. Unlike a single impulse, a filled
	// region has enough energy that 8-bit rounding is a negligible fraction.
	const half = 7
	for y := h/2 - half; y <= h/2+half; y++ {
		for x := w/2 - half; x <= w/2+half; x++ {
			o := y*stride + x*4
			buf[o], buf[o+1], buf[o+2], buf[o+3] = 200, 200, 200, 255
		}
	}

	before := 0
	for i := 0; i < len(buf); i += 4 {
		before += int(buf[i]) + int(buf[i+1]) + int(buf[i+2])
	}

	// Radius >= 3*sigma so the truncated kernel captures ~99.7% of the Gaussian
	// mass; a much smaller cutoff would genuinely lose tail energy.
	BlurRGBA(buf, stride, w, h, 6, 2.0)

	after := 0
	for i := 0; i < len(buf); i += 4 {
		after += int(buf[i]) + int(buf[i+1]) + int(buf[i+2])
	}

	// The Gaussian is normalised, so total intensity should be ~unchanged.
	if after < before*99/100 || after > before*101/100 {
		t.Fatalf("energy not conserved: before=%d after=%d", before, after)
	}
}
