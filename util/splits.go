package util

import (
	"image"
	"image/draw"

	"github.com/faiface/gui"
)

type envPair struct {
	events <-chan gui.Event
	draw   chan<- func(draw.Image) image.Rectangle
}

func (ep *envPair) Events() <-chan gui.Event                      { return ep.events }
func (ep *envPair) Draw() chan<- func(draw.Image) image.Rectangle { return ep.draw }

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

	return &envPair{out, env.Draw()}
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

	return &envPair{out, env.Draw()}
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

	return &envPair{out, env.Draw()}
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

	return &envPair{out, env.Draw()}
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

	return &envPair{out, env.Draw()}
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

	return &envPair{out, env.Draw()}
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

	return &envPair{out, env.Draw()}
}

func EvenHorizontalMinMaxX(env gui.Env, minI, maxI, n, minX, maxX int) gui.Env {
	out, in := gui.MakeEventsChan()

	go func() {
		for e := range env.Events() {
			if resize, ok := e.(gui.Resize); ok {
				out := false
				if resize.Min.X < minX {
					resize.Min.X = minX
					out = true
				}
				if resize.Max.X > maxX {
					resize.Max.X = maxX
					out = true
				}
				if out {
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

	return &envPair{out, env.Draw()}
}

func EvenHorizontalRightMinMaxX(env gui.Env, minI, maxI, n, minX, maxX int) gui.Env {
	out, in := gui.MakeEventsChan()

	go func() {
		for e := range env.Events() {
			if resize, ok := e.(gui.Resize); ok {
				out := false

				dxr := resize.Max.X - minX
				x1 := resize.Max.X
				if minX < x1-dxr {
					resize.Max.X = dxr - minX
					out = true
				}
				if minX > x1-dxr {
					resize.Max.X = dxr - minX
					out = true
				}

				dxrmax := resize.Max.X - maxX
				if maxX < x1-dxrmax {
					resize.Max.X = dxrmax - maxX
					out = true
				}
				if maxX > x1-dxrmax {
					resize.Max.X = dxrmax - maxX
					out = true
				}

				if out {
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

	return &envPair{out, env.Draw()}
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

	return &envPair{out, env.Draw()}
}

func EvenHorizontalRightMinX(env gui.Env, minI, maxI, n, minX int) gui.Env {
	out, in := gui.MakeEventsChan()

	go func() {
		for e := range env.Events() {
			if resize, ok := e.(gui.Resize); ok {
				//out := false
				// dxr := resize.Max.X - minX
				// x1 := resize.Max.X
				// //resize.Max.X = dxr

				// if minX < x1-dxr {
				// 	fmt.Println("rtl", x1, " - ", dxr, " = ", x1-dxr)
				// 	resize.Max.X = dxr - minX
				// 	out = true
				// }
				// if minX > x1-dxr {
				// 	fmt.Println("ltr", x1, " - ", dxr, " = ", x1-dxr)
				// 	resize.Max.X = dxr - minX
				// 	out = true
				// }

				// if out {
				// 	fmt.Println(resize.Min.X, resize.Max.X, minX, dxr)
				// 	in <- resize
				// 	continue
				// }

				resize.Max.X = resize.Max.X - minX
				// x1 := resize.Max.X
				// x0 := resize.Min.X
				// resize.Min.X, resize.Max.X = x0+(x1-x0)*minI/n, x0+(x1-x0)*maxI/n
				in <- resize
			} else {
				in <- e
			}
		}
		close(in)
	}()

	return &envPair{out, env.Draw()}
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

	return &envPair{out, env.Draw()}
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

	return &envPair{out, env.Draw()}
}
