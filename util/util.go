package util

import (
	"image"
	"image/color"

	"sync"

	"github.com/golang/freetype/truetype"
	"github.com/juanefec/scplayer/icons"

	"golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

type concurrentFace struct {
	mu   sync.Mutex
	face font.Face
}

func (cf *concurrentFace) Close() error {
	cf.mu.Lock()
	defer cf.mu.Unlock()
	return cf.face.Close()
}

func (cf *concurrentFace) Glyph(dot fixed.Point26_6, r rune) (dr image.Rectangle, mask image.Image, maskp image.Point, advance fixed.Int26_6, ok bool) {
	cf.mu.Lock()
	defer cf.mu.Unlock()
	return cf.face.Glyph(dot, r)
}

func (cf *concurrentFace) GlyphBounds(r rune) (bounds fixed.Rectangle26_6, advance fixed.Int26_6, ok bool) {
	cf.mu.Lock()
	defer cf.mu.Unlock()
	return cf.face.GlyphBounds(r)
}

func (cf *concurrentFace) GlyphAdvance(r rune) (advance fixed.Int26_6, ok bool) {
	cf.mu.Lock()
	defer cf.mu.Unlock()
	return cf.face.GlyphAdvance(r)
}

func (cf *concurrentFace) Kern(r0, r1 rune) fixed.Int26_6 {
	cf.mu.Lock()
	defer cf.mu.Unlock()
	return cf.face.Kern(r0, r1)
}

func (cf *concurrentFace) Metrics() font.Metrics {
	cf.mu.Lock()
	defer cf.mu.Unlock()
	return cf.face.Metrics()
}

func TTFToFace(ttf []byte, size float64) (font.Face, error) {
	font, err := truetype.Parse(ttf)
	if err != nil {
		return nil, err
	}
	return &concurrentFace{face: truetype.NewFace(font, &truetype.Options{
		Size: size,
	})}, nil
}

func MakeTextImage(text string, face font.Face, clr color.Color) image.Image {
	drawer := &font.Drawer{
		Src:  &image.Uniform{clr},
		Face: face,
		Dot:  fixed.P(0, 0),
	}
	b26_6, _ := drawer.BoundString(text)
	//fmt.Printf("%v -> \n\tx0: %v\n\ty0: %v\n", text, b26_6.Min.X.Floor(), b26_6.Min.Y.Floor())
	bounds := image.Rect(
		b26_6.Min.X.Floor(),
		b26_6.Min.Y.Floor(),
		//b26_6.Min.Y.Floor()-6,
		b26_6.Max.X.Ceil(),
		b26_6.Max.Y.Ceil()+4,
		//pixfont.MeasureString(text),
		//pixfont.DefaultFont.GetHeight()/2,
	)
	drawer.Dst = image.NewRGBA(bounds)
	drawer.DrawString(text)
	//pixfont.DrawString(drawer.Dst, b26_6.Min.X.Floor(), b26_6.Min.Y.Floor(), text, clr)
	return drawer.Dst
}

func MakeTextScaledImage(text string, face font.Face, clr color.Color, scale float32) image.Image {
	drawer := &font.Drawer{
		Src:  &image.Uniform{clr},
		Face: face,
		Dot:  fixed.P(0, 0),
	}
	b26_6, _ := drawer.BoundString(text)
	//fmt.Printf("%v -> \n\tx0: %v\n\ty0: %v\n", text, b26_6.Min.X.Floor(), b26_6.Min.Y.Floor())
	bounds := image.Rect(
		b26_6.Min.X.Floor(),
		b26_6.Min.Y.Floor(),
		//b26_6.Min.Y.Floor()-6,
		b26_6.Max.X.Ceil(),
		b26_6.Max.Y.Ceil()+4,
		//pixfont.MeasureString(text),
		//pixfont.DefaultFont.GetHeight()/2,
	)
	drawer.Dst = image.NewRGBA(bounds)
	drawer.DrawString(text)

	scaled := image.Rect(
		drawer.Dst.Bounds().Min.X,
		drawer.Dst.Bounds().Min.Y,
		int(float32(drawer.Dst.Bounds().Max.X)*scale),
		int(float32(drawer.Dst.Bounds().Max.Y)*scale),
	)

	dst := image.NewRGBA(scaled)

	draw.CatmullRom.Scale(dst, dst.Rect, drawer.Dst, drawer.Dst.Bounds(), draw.Over, nil)

	//pixfont.DrawString(drawer.Dst, b26_6.Min.X.Floor(), b26_6.Min.Y.Floor(), text, clr)
	return dst
}

func MakeTextScaledBestQualityImage(text string, face font.Face, clr color.Color, scale float32) image.Image {
	drawer := &font.Drawer{
		Src:  &image.Uniform{clr},
		Face: face,
		Dot:  fixed.P(0, 0),
	}
	b26_6, _ := drawer.BoundString(text)
	//fmt.Printf("%v -> \n\tx0: %v\n\ty0: %v\n", text, b26_6.Min.X.Floor(), b26_6.Min.Y.Floor())
	bounds := image.Rect(
		b26_6.Min.X.Floor(),
		b26_6.Min.Y.Floor(),
		//b26_6.Min.Y.Floor()-6,
		b26_6.Max.X.Ceil(),
		b26_6.Max.Y.Ceil()+4,
		//pixfont.MeasureString(text),
		//pixfont.DefaultFont.GetHeight()/2,
	)
	drawer.Dst = image.NewRGBA(bounds)
	drawer.DrawString(text)

	scaled := image.Rect(
		drawer.Dst.Bounds().Min.X,
		drawer.Dst.Bounds().Min.Y,
		int(float32(drawer.Dst.Bounds().Max.X)*scale),
		int(float32(drawer.Dst.Bounds().Max.Y)*scale),
	)

	dst := image.NewRGBA(scaled)

	draw.CatmullRom.Scale(dst, dst.Rect, drawer.Dst, drawer.Dst.Bounds(), draw.Over, nil)

	//pixfont.DrawString(drawer.Dst, b26_6.Min.X.Floor(), b26_6.Min.Y.Floor(), text, clr)
	return dst
}

func MakeRailAndProgressImage(r image.Rectangle, duration, progress int, col color.Color) (image.Image, image.Image, bool) {
	if r.Dx() >= 0 && r.Dy() >= 0 {
		//off := r.Dx() / 12
		pixs, pixe := r.Min.X, r.Max.X
		rail := image.NewRGBA(r)
		hline(rail, pixs, 60, pixe, col)

		ptop := Map(progress, 0, duration, pixs, pixe)
		progress := image.NewRGBA(r)
		hlineBold(progress, pixs, 60, ptop, col)
		return rail, progress, true
	}
	return nil, nil, false
}

func Map(vi, s1i, st1i, s2i, st2i int) int {
	v, s1, st1, s2, st2 := float64(vi), float64(s1i), float64(st1i), float64(s2i), float64(st2i)
	newval := (v-s1)/(st1-s1)*(st2-s2) + s2
	if s2 < st2 {
		if newval < s2 {
			return int(s2)
		}
		if newval > st2 {
			return int(st2)
		}
	} else {
		if newval > s2 {
			return int(s2)
		}
		if newval < st2 {
			return int(st2)
		}
	}
	return int(newval)
}

func MapIntFloat(vi, s1i, st1i, s2i, st2i int) float64 {
	v, s1, st1, s2, st2 := float64(vi), float64(s1i), float64(st1i), float64(s2i), float64(st2i)
	newval := (v-s1)/(st1-s1)*(st2-s2) + s2
	if s2 < st2 {
		if newval < s2 {
			return s2
		}
		if newval > st2 {
			return st2
		}
	} else {
		if newval > s2 {
			return s2
		}
		if newval < st2 {
			return st2
		}
	}
	return newval
}

func MapFloat(v, s1, st1, s2, st2 float64) float64 {
	newval := (v-s1)/(st1-s1)*(st2-s2) + s2
	if s2 < st2 {
		if newval < s2 {
			return s2
		}
		if newval > st2 {
			return st2
		}
	} else {
		if newval > s2 {
			return s2
		}
		if newval < st2 {
			return st2
		}
	}
	return newval
}

// hline draws a horizontal line
func hline(img *image.RGBA, x1, y, x2 int, col color.Color) {
	for ; x1 <= x2; x1++ {
		img.Set(x1, y, col)
	}
}
func vline(img *image.RGBA, y1, x, y2 int, col color.Color) {
	for ; y1 <= y2; y1++ {
		img.Set(x, y1, col)
	}
}
func hlineBold(img *image.RGBA, x1, y, x2 int, col color.Color) {
	for ; x1 <= x2; x1++ {
		img.Set(x1, y-2, col)
		img.Set(x1, y-1, col)
		img.Set(x1, y, col)
		img.Set(x1, y+1, col)
		img.Set(x1, y+2, col)
	}
}

func MakeCursorImage(r image.Rectangle, col color.Color) image.Image {
	if r.Dx() >= 0 && r.Dy() >= 0 {
		//off := r.Dx() / 12
		pixs, pixe := r.Min.Y+4, r.Max.Y-6
		cursor := image.NewRGBA(r)
		vline(cursor, pixs, 29, pixe, col) // jajja, ochentiuno

		return cursor
	}
	return nil
}

func MakeIconImage(icon string) image.Image {
	return icons.GetIcon(icon)
}

func DrawCentered(dst draw.Image, r image.Rectangle, src image.Image, op draw.Op) {
	if src == nil {
		return
	}
	bounds := src.Bounds()
	center := bounds.Min.Add(bounds.Max).Div(2)
	target := r.Min.Add(r.Max).Div(2)
	delta := target.Sub(center)
	draw.Draw(dst, bounds.Add(delta).Intersect(r), src, bounds.Min, op)
}

func DrawLeftCentered(dst draw.Image, r image.Rectangle, src image.Image, op draw.Op) {
	if src == nil {
		return
	}
	bounds := src.Bounds()
	leftCenter := image.Pt(bounds.Min.X, (bounds.Min.Y+bounds.Max.Y)/2)
	target := image.Pt(r.Min.X, (r.Min.Y+r.Max.Y)/2)
	delta := target.Sub(leftCenter)
	draw.Draw(dst, bounds.Add(delta).Intersect(r), src, bounds.Min, op)
}

func DrawRightCentered(dst draw.Image, r image.Rectangle, src image.Image, op draw.Op) {
	if src == nil {
		return
	}
	bounds := src.Bounds()
	rightCenter := image.Pt(bounds.Max.X, (bounds.Min.Y+bounds.Max.Y)/2)
	target := image.Pt(r.Max.X, (r.Min.Y+r.Max.Y)/2)
	delta := target.Sub(rightCenter)
	draw.Draw(dst, bounds.Add(delta).Intersect(r), src, bounds.Min, op)
}
