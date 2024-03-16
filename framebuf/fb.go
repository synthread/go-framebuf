package framebuf

import (
	"image"
	"image/color"
	"image/draw"
	"syscall"
	"time"

	"github.com/fogleman/gg"
	"github.com/pkg/errors"
	"github.com/synthread/framebuffer"
)

// TickFunction is a function that is called every tick to build or change the
// buffered image
type TickFunction func(*Framebuffer)

// Framebuffer is an instance of a particular framebuffer, helpfully wrapping
// a double-buffering and hashing mechanism to help improve the experience of
// using a raw framebuffer
type Framebuffer struct {
	raw     draw.Image
	canvas  *framebuffer.Canvas
	config  *Config
	drawing *gg.Context

	lastBufferHash uint64

	tick TickFunction
}

// Config defines any specific configuration for this framebuffer
type Config struct {
	// Rotate will rotate the final image by the specified
	// number of degrees before writing to the framebuffer
	// nb: get the best performance by using a multiple of 90
	Rotate int

	// TickFunc is the function you want to (optionally) call every time a frame
	// is rendered
	TickFunc TickFunction
	TickTime time.Duration
}

// NewFramebuffer will create a new Framebuffer with the given configuration
func NewFramebuffer(cfg *Config) (*Framebuffer, error) {
	var err error

	if cfg == nil {
		cfg = &Config{}
	}

	fb := &Framebuffer{
		config: cfg,
		tick:   cfg.TickFunc,
	}

	// open the framebuffer device
	fb.canvas, err = framebuffer.Open(nil, nil)
	if err != nil {
		return nil, errors.Wrap(err, "could not open framebuffer")
	}

	// convert it to a go image
	fb.raw, err = fb.canvas.Image()
	if err != nil {
		return nil, errors.Wrap(err, "could not init framebuffer")
	}

	// setup the double buffer
	//fb.buffer = image.NewRGBA(fb.raw.Bounds())
	fb.drawing = gg.NewContextForImage(image.NewRGBA(fb.Bounds()))

	return fb, nil
}

// Randomize will fill the framebuffer with random data
func (fb *Framebuffer) Randomize() error {
	fb.canvas.File()

	for {
		_, err := fb.canvas.File().Write(randBytes())
		if err != nil {
			if errors.Is(err, syscall.ENOSPC) {
				return nil
			}
			return errors.Wrap(err, "could not write random bytes")
		}
	}
}

// Start will run the framebuffer render in the background
// and will return a chan that can be written to to stop the render
// loop
func (fb *Framebuffer) Start(fps int) chan bool {
	done := make(chan bool, 1)

	frameDur := time.Duration(1.0/float64(fps)) * time.Second

	go func() {
		var lastTick time.Time

		for {
			select {
			case <-done:
				return
			default:
				fStart := time.Now()

				// we can tick slower than the frames are drawn to improve
				// performance (TickTime defaults to zero)
				if time.Since(lastTick) > fb.config.TickTime {
					// optionally run the tick function
					if fb.config.TickFunc != nil {
						fb.config.TickFunc(fb)
					}
					lastTick = fStart
				}

				// do the draw
				fb.writeBuffer()

				dur := time.Since(fStart)
				if dur < frameDur {
					time.Sleep(frameDur - dur)
				} else {
					// we want to sleep a minimum of 10ms or we will annihilate
					// performance
					time.Sleep(10 * time.Millisecond)
				}
			}
		}
	}()

	return done
}

// Bounds returns the size of the framebuffer
func (fb *Framebuffer) Bounds() image.Rectangle {
	return fb.raw.Bounds()
}

// Clear will clear the framebuffer with the specified colour
func (fb *Framebuffer) Clear(c color.Color) {
	fb.drawing.SetColor(c)
	fb.drawing.Clear()
}

// ForceDraw will force a redraw on the next frame
func (fb *Framebuffer) ForceDraw() {
	fb.lastBufferHash = 0
}

// Draw will return a context for drawing on the framebuffer
func (fb *Framebuffer) Draw() *gg.Context {
	return fb.drawing
}
