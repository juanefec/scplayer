package component

import (
	"fmt"
	"image"
	"image/draw"
	"log"
	"time"

	"github.com/faiface/gui"
	"github.com/juanefec/scplayer/sc"

	. "github.com/juanefec/scplayer/util"
)

func Player(env gui.Env, theme *Theme, newsong <-chan sc.Song, pausebtn <-chan bool, next chan<- int, updateTitle chan<- string) {

	redraw := func(r image.Rectangle, rail image.Image, progress image.Image, imgProgress image.Image, imgProgressTop image.Image) func(draw.Image) image.Rectangle {
		return func(drw draw.Image) image.Rectangle {
			draw.Draw(drw, r, &image.Uniform{theme.Empty}, image.Point{}, draw.Src)

			DrawLeftCentered(drw, r.Inset(4), imgProgress, draw.Over)
			DrawRightCentered(drw, r.Inset(4), imgProgressTop, draw.Over)

			if rail != nil {
				DrawCentered(drw, r, rail, draw.Over)
			}
			if progress != nil {
				DrawCentered(drw, r, progress, draw.Over)
			}
			return r
		}
	}

	//invalid := MakeTextImage("Invalid image", theme.Face, theme.Text)

	var (
		r              image.Rectangle
		imgProgress    image.Image
		imgProgressTop image.Image
		progress       image.Image
		rail           image.Image
		done           = make(chan struct{})
		doneTimer      = make(chan struct{})
	)

	song := sc.Song{}

	songTimer := func() {
		for {
			select {
			case <-doneTimer:
				return

			case <-time.After(time.Millisecond * 100):
				if r, p, ok := MakeRailAndProgressImage(r, song); ok {
					rail, progress = r, p
				}
				imgProgress = MakeTextImage(song.Progress(), theme.Face, theme.Text)
				imgProgressTop = MakeTextImage(song.Duration(), theme.Face, theme.Text)
				env.Draw() <- redraw(r, rail, progress, imgProgress, imgProgressTop)

			}
		}
	}

	for {
		select {
		case <-done:
			doneTimer <- struct{}{}
			next <- 1

		case playing := <-pausebtn:
			if playing {
				song.Pause()
			} else {
				err := song.Resume()
				if err != nil {
					next <- 1
				}
			}

		case song = <-newsong:
			song.Stop()

			err := song.Play(done)
			if err != nil {
				log.Fatal(err.Error())
			}

			go songTimer()

			title := fmt.Sprintf("%v - %v", song.Artist, song.Title)

			updateTitle <- title

			env.Draw() <- redraw(r, rail, progress, imgProgress, imgProgressTop)

		case e, ok := <-env.Events():
			if !ok {
				return
			}

			if resize, ok := e.(gui.Resize); ok {
				r = resize.Rectangle
				env.Draw() <- redraw(r, rail, progress, imgProgress, imgProgressTop)
			}
		}
	}
}

func Title(env gui.Env, theme *Theme, newTitle <-chan string) {
	redraw := func(r image.Rectangle, imgTitle image.Image) func(draw.Image) image.Rectangle {
		return func(drw draw.Image) image.Rectangle {
			draw.Draw(drw, r, &image.Uniform{theme.Title}, image.Point{}, draw.Src)
			DrawCentered(drw, r, imgTitle, draw.Over)
			return r
		}
	}

	var (
		r        image.Rectangle
		imgTitle image.Image
	)

	for {
		select {
		case t := <-newTitle:
			imgTitle = MakeTextImage(t, theme.Face, theme.Text)
			env.Draw() <- redraw(r, imgTitle)
		case e, ok := <-env.Events():
			if !ok {
				close(env.Draw())
				return
			}
			if resize, ok := e.(gui.Resize); ok {
				r = resize.Rectangle
				env.Draw() <- redraw(r, imgTitle)
			}
		}
	}
}
