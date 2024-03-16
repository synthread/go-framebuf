package framebuf

import (
	"image"
	"image/draw"

	"github.com/cespare/xxhash"
	"github.com/disintegration/imaging"
)

func (fb *Framebuffer) writeBuffer() {
	var transImage *image.NRGBA

	buffer := fb.drawing.Image()

	switch fb.config.Rotate {
	case 0:
		transImage = imaging.Clone(buffer)
	case 90:
		transImage = imaging.Rotate90(buffer)
	case 180:
		transImage = imaging.Rotate180(buffer)
	case 270:
		transImage = imaging.Rotate270(buffer)
	default:
		transImage = imaging.Rotate(buffer, float64(fb.config.Rotate), Colour(0, 0, 0))
	}

	// avoid doing the double buffered draw unless the buffer has actually
	// changed or we haven't draw yet
	newHash := hashImage(transImage)
	if fb.lastBufferHash != 0 && newHash == fb.lastBufferHash {
		return
	}
	fb.lastBufferHash = newHash

	draw.Draw(fb.raw, fb.raw.Bounds(), transImage, image.Point{}, draw.Src)
}

func hashImage(im *image.NRGBA) uint64 {
	xh := xxhash.New()
	for _, px := range im.Pix {
		xh.Write([]byte{px})
	}
	return xh.Sum64()
}
