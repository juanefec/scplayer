package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math/rand"
	"time"

	"github.com/faiface/gui"
	"github.com/faiface/gui/win"
	"github.com/juanefec/scplayer/sc"
	"golang.org/x/image/math/fixed"
)

func Browser(env gui.Env, theme *Theme, cd <-chan string, view chan<- sc.Song, next <-chan int, pausebtn chan<- bool, reloadUser <-chan string) {
	username := ""

	reload := func(songs []sc.Song) ([]sc.Song, int, *image.RGBA) {

		var images []image.Image
		//fmt.Printf("Listed %v songs.\n", len(songs))
		for _, song := range songs {
			images = append(images, MakeTextImage(fmt.Sprintf("%v - %v", song.Artist, song.Title), theme.Face, theme.Text))
		}

		const inset = 4

		var width int
		for _, img := range images {
			if img.Bounds().Dx() > width {
				width = img.Bounds().Inset(-inset).Dx()
			}
		}

		metrics := theme.Face.Metrics()
		lineHeight := (metrics.Height + 2*fixed.I(inset)).Ceil()
		height := lineHeight * len(songs)

		namesImage := image.NewRGBA(image.Rect(0, 0, width+2*inset, height+2*inset))
		for i := range images {
			r := image.Rect(
				0, lineHeight*i,
				width, lineHeight*(i+1),
			)
			DrawLeftCentered(namesImage, r.Inset(inset), images[i], draw.Over)
		}

		return songs, lineHeight, namesImage
	}

	redraw := func(r image.Rectangle, selected int, position image.Point, lineHeight int, namesImage image.Image) func(draw.Image) image.Rectangle {
		return func(drw draw.Image) image.Rectangle {
			draw.Draw(drw, r, &image.Uniform{theme.Background}, image.Point{}, draw.Src)
			draw.Draw(drw, r, namesImage, position, draw.Over)
			if selected >= 0 {
				highlightR := image.Rect(
					namesImage.Bounds().Min.X,
					namesImage.Bounds().Min.Y+lineHeight*selected,
					namesImage.Bounds().Max.X,
					namesImage.Bounds().Min.Y+lineHeight*(selected+1),
				)
				highlightR = highlightR.Sub(position).Add(r.Min)
				draw.DrawMask(
					drw, highlightR.Intersect(r),
					&image.Uniform{theme.Highlight}, image.Point{},
					&image.Uniform{color.Alpha{64}}, image.Point{},
					draw.Over,
				)
			}
			return r
		}
	}

	songs := []sc.Song{
		{Artist: "          type ENTER and open the searchbar             < - - -", Title: "- >    quit the searchbar with ESCAPE       "},
		{Artist: "search for a username (eg. soundcloud.com/<usename>)  < - - - -", Title: "- - > or if its empty just press ENTER      "},
		{Artist: "     type ENTER and fetch all the user likes.       < - - - - -", Title: "- - - >                                     "},
	}

	songs, lineHeight, namesImage := reload(songs)

	var (
		err          error
		r            image.Rectangle
		position     = image.Point{}
		selected     = -1
		selectedOGID = -1
	)

	refresh := func() {
		songs, err = sc.GetLikes(username)
		if err != nil {
			songs = []sc.Song{
				{Artist: "- - - - - - -", Title: fmt.Sprintf("- - - - >  %v                 ", err.Error())},
			}
		}
		songs, lineHeight, namesImage = reload(songs)

		position = image.Point{}
		selected = -1
		for i, s := range songs {
			if s.OriginalID == selectedOGID {
				selected = i
			}
		}
	}

	for {
		select {
		case newuser := <-reloadUser:
			if newuser != "" {
				username = newuser
				refresh()
			}
			env.Draw() <- redraw(r, selected, position, lineHeight, namesImage)
		case action := <-cd:
			switch action {
			case "refresh":
				refresh()
				env.Draw() <- redraw(r, selected, position, lineHeight, namesImage)
			case "shuffle":
				rand.Seed(time.Now().Unix())
				rand.Shuffle(len(songs), func(i, j int) { songs[i], songs[j] = songs[j], songs[i] })
				songs, lineHeight, namesImage = reload(songs)

				position = image.Point{}
				selected = -1
				for i, s := range songs {
					if s.OriginalID == selectedOGID {
						selected = i
					}
				}

				env.Draw() <- redraw(r, selected, position, lineHeight, namesImage)
			}
		case v := <-next:
			if selected >= -1 && selected < len(songs)+1 {
				if selected == 0 && v == -1 {
					continue
				}
				selected = selected + v
				if selected > len(songs)-1 || selected < 0 || songs[selected].OriginalID == 0 {
					continue
				}
				selectedOGID = songs[selected].OriginalID
				view <- songs[selected]
				pausebtn <- true
				env.Draw() <- redraw(r, selected, position, lineHeight, namesImage)
			}
		case e, ok := <-env.Events():
			if !ok {
				close(env.Draw())
				return
			}

			switch e := e.(type) {
			case gui.Resize:
				r = e.Rectangle
				env.Draw() <- redraw(r, selected, position, lineHeight, namesImage)

			case win.MoDown:
				if !e.Point.In(r) {
					continue
				}
				click := e.Point.Sub(r.Min).Add(position)
				i := click.Y / lineHeight
				if i < 0 || i >= len(songs) {
					continue
				}

				selected = i
				if songs[selected].OriginalID == 0 {
					continue
				}

				selectedOGID = songs[selected].OriginalID
				view <- songs[selected]
				pausebtn <- true
				env.Draw() <- redraw(r, selected, position, lineHeight, namesImage)

			case win.MoScroll:
				newP := position.Sub(e.Point.Mul(160))
				if newP.X > namesImage.Bounds().Max.X-r.Dx() {
					newP.X = namesImage.Bounds().Max.X - r.Dx()
				}
				if newP.Y > namesImage.Bounds().Max.Y-r.Dy() {
					newP.Y = namesImage.Bounds().Max.Y - r.Dy()
				}
				if newP.X < namesImage.Bounds().Min.X {
					newP.X = namesImage.Bounds().Min.X
				}
				if newP.Y < namesImage.Bounds().Min.Y {
					newP.Y = namesImage.Bounds().Min.Y
				}
				if newP != position {
					position = newP
					env.Draw() <- redraw(r, selected, position, lineHeight, namesImage)
				}
			}
		}
	}
}
