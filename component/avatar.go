package component

import (
	"image"
	"image/draw"

	"github.com/faiface/gui"
	"github.com/juanefec/scplayer/sc"
	. "github.com/juanefec/scplayer/util"
)

func Avatar(env gui.Env, theme *Theme, newUser <-chan sc.User, loading chan struct{}) {
	redraw := func(r image.Rectangle, imgAvatar image.Image) func(draw.Image) image.Rectangle {
		return func(drw draw.Image) image.Rectangle {
			draw.Draw(drw, r, &image.Uniform{theme.Title}, image.Point{}, draw.Src)
			DrawCentered(drw, r, imgAvatar, draw.Over)
			return r
		}
	}

	var (
		r        image.Rectangle
		avatar   image.Image
		emptyImg = image.NewRGBA(r)
	)

	for {
		select {
		case <-loading:
			avatar = emptyImg
			env.Draw() <- redraw(r, avatar)
		case nu := <-newUser:
			if nu.Username != "" {
				avatar = nu.Av
				env.Draw() <- redraw(r, avatar)
			}
		case e, ok := <-env.Events():
			if !ok {
				close(env.Draw())
				return
			}
			if resize, ok := e.(gui.Resize); ok {
				r = resize.Rectangle
				env.Draw() <- redraw(r, avatar)
			}
		}
	}
}
