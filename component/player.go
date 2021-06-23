package component

import (
	"fmt"
	"image"
	"image/draw"
	"time"

	"github.com/juanefec/gui"
	"github.com/juanefec/scplayer/sc"

	. "github.com/juanefec/scplayer/util"
)

func Player(env gui.Env, theme *Theme, newsong <-chan sc.Song, pausebtn <-chan bool, next chan<- int, updateTitle chan<- string, updateVol <-chan float64) {

	redraw := func(r image.Rectangle, rail image.Image, progress image.Image, imgProgress image.Image, imgProgressTop image.Image) func(draw.Image) image.Rectangle {
		return func(drw draw.Image) image.Rectangle {
			draw.Draw(drw, r, &image.Uniform{theme.Empty}, image.Point{}, draw.Src)

			DrawLeftCentered(drw, r.Inset(4), imgProgress, draw.Over)
			DrawRightCentered(drw, r.Inset(4), imgProgressTop, draw.Over)

			if rail != nil {
				DrawCentered(drw, r.Inset(4), rail, draw.Over)
			}
			if progress != nil {
				DrawCentered(drw, r.Inset(4), progress, draw.Over)
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
		doneDownload   = make(chan int)
		doneTimer      = make(chan struct{})
		playing        bool
		//doneDownload   = make(chan struct{})
		pvol float64
		//songsPlayed []sc.Song
		song *sc.Song
	)

	songTimer := func(s string) {
		for {
			select {
			case <-doneTimer:
				return

			case <-time.After(time.Millisecond * 100):

				imgProgress = MakeTextImage(song.Progress(), theme.Face, theme.Text)
				imgProgressTop = MakeTextImage(song.Duration(), theme.Face, theme.Text)

				if ra, pr, ok := MakeRailAndProgressImage(r, song.DurationMs(), song.ProgressMs(), theme.Rail); ok {
					rail, progress = ra, pr
				}

				env.Draw() <- redraw(r, rail, progress, imgProgress, imgProgressTop)
			}
		}
	}

	songStarter := func() {

		for {
			select {

			case dd := <-doneDownload:
				if dd == song.OriginalID {
					if playing {
						doneTimer <- struct{}{}
					}

					err := song.Play(pvol, done)
					if err != nil {
						fmt.Println(err)
						continue
					}

					go songTimer(song.Artist + " - " + song.Title)

					playing = true
					title := fmt.Sprintf("%v - %v", song.Artist, song.Title)

					updateTitle <- title
					env.Draw() <- redraw(r, rail, progress, imgProgress, imgProgressTop)
				}
			}
		}
	}
	go songStarter()
	for {
		select {
		case <-done:
			playing = false
			doneTimer <- struct{}{}
			next <- 1
		case bplaying := <-pausebtn:
			if bplaying {
				playing = false
				song.Pause()
			} else {

				err := song.Resume()
				if err != nil {
					next <- 1
				} else {
					playing = true
				}
			}

		case nnsong := <-newsong:
			//songsPlayed = append(songsPlayed, song)
			pvol = song.GetVolume()
			song = &nnsong

			go func(song *sc.Song, doneDownload chan int) {
				err := song.Download(doneDownload)
				if err != nil {
					fmt.Println(err)
				}
			}(song, doneDownload)

			sc.ClearSpeaker()
		case v := <-updateVol:
			song.SetVolume(v)
		case e, ok := <-env.Events():
			if !ok {
				close(env.Draw())
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
			imgTitle = MakeTextScaledImage(t, theme.Face, theme.Text, 1.5)
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
