package framebuf

import (
	"crypto/rand"
	"fmt"
	"image/color"
	"strings"
)

func randBytes() []byte {
	buf := make([]byte, 1024)
	rand.Read(buf)
	return buf
}

// Colour will turn RGB values into a Go-compatible colour
func Colour(r, g, b uint8) color.Color {
	return color.RGBA{
		R: r, G: g, B: b, A: 255,
	}
}

// HexColour will parse a HTML-style hexcode into a Go-compatible colour
func HexColour(s string) (c color.RGBA) {
	s = strings.TrimPrefix(s, "#")
	c.A = 0xff
	switch len(s) {
	case 6:
		_, _ = fmt.Sscanf(s, "%02x%02x%02x", &c.R, &c.G, &c.B)
	case 3:
		_, _ = fmt.Sscanf(s, "%1x%1x%1x", &c.R, &c.G, &c.B)
		// Double the hex digits:
		c.R *= 17
		c.G *= 17
		c.B *= 17
	default:
		panic("invalid length")
	}
	return
}
