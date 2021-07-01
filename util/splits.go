package util

import (
	"image"
	"image/draw"

	"github.com/juanefec/gui"
)

type EnvPair struct {
	Ev <-chan gui.Event
	Dw chan<- func(draw.Image) image.Rectangle
}

func (ep *EnvPair) Events() <-chan gui.Event                      { return ep.Ev }
func (ep *EnvPair) Draw() chan<- func(draw.Image) image.Rectangle { return ep.Dw }

func FixedLeft(env gui.Env, maxX int) gui.Env {
	out, in := gui.MakeEventsChan()

	go func() {
		for e := range env.Events() {
			if resize, ok := e.(gui.Resize); ok {
				//resize.Max.X = maxX
				in <- resize
			} else {
				in <- e
			}
		}
		close(in)
	}()

	return &EnvPair{out, env.Draw()}
}

func FixedRight(env gui.Env, minX int) gui.Env {
	out, in := gui.MakeEventsChan()

	go func() {
		for e := range env.Events() {
			if resize, ok := e.(gui.Resize); ok {
				resize.Min.X = minX
				in <- resize
			} else {
				in <- e
			}
		}
		close(in)
	}()

	return &EnvPair{out, env.Draw()}
}

func FixedTop(env gui.Env, maxY int) gui.Env {
	out, in := gui.MakeEventsChan()

	go func() {
		for e := range env.Events() {
			if resize, ok := e.(gui.Resize); ok {
				resize.Max.Y = maxY
				in <- resize
			} else {
				in <- e
			}
		}
		close(in)
	}()

	return &EnvPair{out, env.Draw()}
}

func FixedBottom(env gui.Env, minY int) gui.Env {
	out, in := gui.MakeEventsChan()

	go func() {
		for e := range env.Events() {
			if resize, ok := e.(gui.Resize); ok {
				resize.Min.Y = minY
				in <- resize
			} else {
				in <- e
			}
		}
		close(in)
	}()

	return &EnvPair{out, env.Draw()}
}

func EvenHorizontalWithRectMinMax(env gui.Env, minI, maxI, n, minRX, maxRX int) gui.Env {
	out, in := gui.MakeEventsChan()

	go func() {
		for e := range env.Events() {
			if resize, ok := e.(gui.Resize); ok {
				resize.Max.X = resize.Max.X - maxRX
				resize.Min.X = resize.Min.X + minRX
				x0, x1 := resize.Min.X, resize.Max.X
				resize.Min.X, resize.Max.X = x0+(x1-x0)*minI/n, x0+(x1-x0)*maxI/n
				in <- resize
			} else {
				in <- e
			}
		}
		close(in)
	}()

	return &EnvPair{out, env.Draw()}
}

func EvenHorizontal(env gui.Env, minI, maxI, n int) gui.Env {
	out, in := gui.MakeEventsChan()

	go func() {
		for e := range env.Events() {
			if resize, ok := e.(gui.Resize); ok {
				x0, x1 := resize.Min.X, resize.Max.X
				resize.Min.X, resize.Max.X = x0+(x1-x0)*minI/n, x0+(x1-x0)*maxI/n
				in <- resize
			} else {
				in <- e
			}
		}
		close(in)
	}()

	return &EnvPair{out, env.Draw()}
}

func EvenVertical(env gui.Env, minI, maxI, n int) gui.Env {
	out, in := gui.MakeEventsChan()

	go func() {
		for e := range env.Events() {
			if resize, ok := e.(gui.Resize); ok {
				y0, y1 := resize.Min.Y, resize.Max.Y
				resize.Min.Y, resize.Max.Y = y0+(y1-y0)*minI/n, y0+(y1-y0)*maxI/n
				in <- resize
			} else {
				in <- e
			}
		}
		close(in)
	}()

	return &EnvPair{out, env.Draw()}
}

func EvenVerticalMinMaxY(env gui.Env, minI, maxI, n, minY, maxY int) gui.Env {
	out, in := gui.MakeEventsChan()

	go func() {
		for e := range env.Events() {
			if resize, ok := e.(gui.Resize); ok {
				out := false
				if resize.Min.Y < minY {
					resize.Min.Y = minY
					out = true
				}
				if resize.Max.Y > maxY {
					resize.Max.Y = maxY
					out = true
				}
				if out {
					in <- resize
					continue
				}
				y0, y1 := resize.Min.Y, resize.Max.Y
				resize.Min.Y, resize.Max.Y = y0+(y1-y0)*minI/n, y0+(y1-y0)*maxI/n
				in <- resize
			} else {
				in <- e
			}
		}
		close(in)
	}()

	return &EnvPair{out, env.Draw()}
}

func FixedFromBounds(env gui.Env, minX, min2X int) gui.Env {
	out, in := gui.MakeEventsChan()
	go func() {
		for e := range env.Events() {
			if resize, ok := e.(gui.Resize); ok {
				resize.Min.X = minX
				resize.Max.X = resize.Max.X - min2X
				in <- resize
			} else {
				in <- e
			}
		}
		close(in)
	}()
	return &EnvPair{out, env.Draw()}
}

func FixedFromRight(env gui.Env, minX, maxX int) gui.Env {
	out, in := gui.MakeEventsChan()

	go func() {
		for e := range env.Events() {
			if resize, ok := e.(gui.Resize); ok {
				x1 := resize.Max.X
				resize.Max.X = x1 - minX
				resize.Min.X = x1 - maxX
				in <- resize
			} else {
				in <- e
			}
		}
		close(in)
	}()

	return &EnvPair{out, env.Draw()}
}

func FixedFromTopLeft(env gui.Env, minX, maxX, minY, maxY int) gui.Env {
	out, in := gui.MakeEventsChan()

	go func() {
		for e := range env.Events() {
			if resize, ok := e.(gui.Resize); ok {
				resize.Max.X = maxX
				resize.Min.X = minX
				resize.Max.Y = maxY
				resize.Min.Y = minY
				in <- resize
			} else {
				in <- e
			}
		}
		close(in)
	}()

	return &EnvPair{out, env.Draw()}
}
func FixedFromTopXBounds(env gui.Env, lminX, rminX, minY, maxY int) gui.Env {
	out, in := gui.MakeEventsChan()

	go func() {
		for e := range env.Events() {
			if resize, ok := e.(gui.Resize); ok {
				resize.Min.X = lminX
				resize.Max.X = resize.Max.X - rminX
				resize.Max.Y = maxY
				resize.Min.Y = minY
				in <- resize
			} else {
				in <- e
			}
		}
		close(in)
	}()

	return &EnvPair{out, env.Draw()}
}

func FixedFromTop(env gui.Env, minY, maxY int) gui.Env {
	out, in := gui.MakeEventsChan()

	go func() {
		for e := range env.Events() {
			if resize, ok := e.(gui.Resize); ok {
				resize.Max.Y = maxY
				resize.Min.Y = minY
				in <- resize
			} else {
				in <- e
			}
		}
		close(in)
	}()

	return &EnvPair{out, env.Draw()}
}

func FixedFromLeft(env gui.Env, minX, maxX int) gui.Env {
	out, in := gui.MakeEventsChan()

	go func() {
		for e := range env.Events() {
			if resize, ok := e.(gui.Resize); ok {
				resize.Max.X = maxX
				resize.Min.X = minX
				in <- resize
			} else {
				in <- e
			}
		}
		close(in)
	}()

	return &EnvPair{out, env.Draw()}
}

func EvenHorizontalMinX(env gui.Env, minI, maxI, n, minX int) gui.Env {
	out, in := gui.MakeEventsChan()

	go func() {
		for e := range env.Events() {
			if resize, ok := e.(gui.Resize); ok {

				if resize.Min.X < minX {
					resize.Min.X = minX
					in <- resize
					continue
				}
				x0, x1 := resize.Min.X, resize.Max.X
				resize.Min.X, resize.Max.X = x0+(x1-x0)*minI/n, x0+(x1-x0)*maxI/n
				in <- resize
			} else {
				in <- e
			}
		}
		close(in)
	}()

	return &EnvPair{out, env.Draw()}
}

func EvenHorizontalMaxX(env gui.Env, minI, maxI, n, maxX int) gui.Env {
	out, in := gui.MakeEventsChan()

	go func() {
		for e := range env.Events() {
			if resize, ok := e.(gui.Resize); ok {
				if resize.Max.X > maxX {
					resize.Max.X = maxX
					in <- resize
					continue
				}
				x0, x1 := resize.Min.X, resize.Max.X
				resize.Min.X, resize.Max.X = x0+(x1-x0)*minI/n, x0+(x1-x0)*maxI/n
				in <- resize
			} else {
				in <- e
			}
		}
		close(in)
	}()

	return &EnvPair{out, env.Draw()}
}
