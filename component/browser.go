package component

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math/rand"
	"time"

	"github.com/juanefec/gui"
	"github.com/juanefec/gui/win"
	"github.com/juanefec/scplayer/sc"
	. "github.com/juanefec/scplayer/util"
	"golang.org/x/image/math/fixed"
)

func Browser(env gui.Env, theme *Theme, cd <-chan string, view chan<- sc.Song, move <-chan int, pausebtn chan<- bool, reloadUser <-chan string, newInfo chan<- string, gotop <-chan struct{}, gotosong <-chan struct{}, listenBrowserSlider <-chan int, newBrowserSlider chan<- int, updateBrowserSlider chan<- int, playingPos chan<- int) {
	username := "kr3a71ve"

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

	type pnext struct {
		i   int
		ogi int
	}

	redraw := func(r image.Rectangle, selected int, position image.Point, lineHeight int, namesImage image.Image, playnexts []pnext) func(draw.Image) image.Rectangle {
		return func(drw draw.Image) image.Rectangle {
			draw.Draw(drw, r, &image.Uniform{theme.Background}, image.Point{}, draw.Src)
			draw.Draw(drw, r, namesImage, position, draw.Over)
			if selected >= 0 {
				highlightR := image.Rect(
					namesImage.Bounds().Min.X,
					namesImage.Bounds().Min.Y+lineHeight*selected,
					r.Max.X,
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
			if len(playnexts) > 0 {
				for _, pnx := range playnexts {
					highlightR := image.Rect(
						namesImage.Bounds().Min.X,
						namesImage.Bounds().Min.Y+lineHeight*pnx.i,
						r.Max.X,
						namesImage.Bounds().Min.Y+lineHeight*(pnx.i+1),
					)
					highlightR = highlightR.Sub(position).Add(r.Min)
					draw.DrawMask(
						drw, highlightR.Intersect(r),
						&image.Uniform{theme.NextHighlight}, image.Point{},
						&image.Uniform{color.Alpha{64}}, image.Point{},
						draw.Over,
					)
				}
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

		playnexts []pnext
	)

	refresh := func() {
		songs, err = sc.GetLikes(username)
		if err != nil {
			songs = []sc.Song{
				{Artist: "- - - - - - -", Title: fmt.Sprintf("- - - - >  %v                 ", err.Error())},
			}
		}
		songs, lineHeight, namesImage = reload(songs)
		newInfo <- fmt.Sprintf("listed %d likes from user: %s", len(songs), username)
		position = image.Point{}
		selected = -1
		for i, s := range songs {
			if s.OriginalID == selectedOGID {
				selected = i
			}
		}
		newBrowserSlider <- namesImage.Rect.Dy()
	}
	// go func() {
	// 	for {
	// 		select {
	// 		case _, ok := <-env.Events():
	// 			if !ok {
	// 				return
	// 			}
	// 		case <-time.After(time.Millisecond * 100):
	// 			env.Draw() <- redraw(r, selected, position, lineHeight, namesImage, playnexts)
	// 		}
	// 	}
	// }()
	for {
		select {
		case y := <-listenBrowserSlider:
			position = image.Point{0, y}
			env.Draw() <- redraw(r, selected, position, lineHeight, namesImage, playnexts)
		case <-gotop:
			position = image.Point{}
			updateBrowserSlider <- position.Y
			env.Draw() <- redraw(r, selected, position, lineHeight, namesImage, playnexts)
		case <-gotosong:
			yPos := selected * lineHeight
			if selected != -1 {
				position = image.Point{0, yPos}
			}
			env.Draw() <- redraw(r, selected, position, lineHeight, namesImage, playnexts)
			updateBrowserSlider <- position.Y
		case newuser := <-reloadUser:
			if newuser != "" {
				username = newuser
				refresh()
			}
			env.Draw() <- redraw(r, selected, position, lineHeight, namesImage, playnexts)
		case action := <-cd:
			switch action {
			case "refresh":
				refresh()
				env.Draw() <- redraw(r, selected, position, lineHeight, namesImage, playnexts)
				updateBrowserSlider <- position.Y
				playingPos <- lineHeight * selected

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
					for pi, pnx := range playnexts {
						if s.OriginalID == pnx.ogi {
							playnexts[pi].i = i
						}
					}
				}
				yPos := selected * lineHeight
				if selected != -1 {
					position = position.Add(image.Point{0, yPos})
				}

				env.Draw() <- redraw(r, selected, position, lineHeight, namesImage, playnexts)
				updateBrowserSlider <- position.Y
				playingPos <- lineHeight * selected

			}
		case v := <-move:
			if selected >= -1 && selected < len(songs)+1 {
				if selected == 0 && v == -1 {
					continue
				}
				if v > 0 {
					if len(playnexts) > 0 {
						selected = playnexts[0].i
						selectedOGID = playnexts[0].ogi
						playnexts = playnexts[1:]

					} else {
						selected = selected + v
					}
				} else {
					selected = selected + v
				}
				if selected > len(songs)-1 {
					selected = 0
				}
				if selected < 0 || songs[selected].OriginalID == 0 {
					continue
				}
				selectedOGID = songs[selected].OriginalID
				view <- songs[selected]
				pausebtn <- true
				playingPos <- lineHeight * selected
				env.Draw() <- redraw(r, selected, position, lineHeight, namesImage, playnexts)
			}
		case e, ok := <-env.Events():
			if !ok {
				close(env.Draw())
				return
			}

			switch e := e.(type) {
			case gui.Resize:
				r = e.Rectangle
				env.Draw() <- redraw(r, selected, position, lineHeight, namesImage, playnexts)

			case win.MoDown:
				if !e.Point.In(r) {
					continue
				}
				click := e.Point.Sub(r.Min).Add(position)
				i := click.Y / lineHeight
				if i < 0 || i >= len(songs) {
					continue
				}
				switch e.Button {
				case win.ButtonLeft:
					if selected != i {
						selected = i
						if songs[selected].OriginalID == 0 {
							continue
						}

						selectedOGID = songs[selected].OriginalID
						view <- songs[selected]
						pausebtn <- true
						env.Draw() <- redraw(r, selected, position, lineHeight, namesImage, playnexts)
						playingPos <- lineHeight * selected
					}
				case win.ButtonRight:
					if songs[i].OriginalID == 0 {
						continue
					}

					delete := -1
					for ipn := 0; ipn < len(playnexts); ipn++ {
						if songs[i].OriginalID == playnexts[ipn].ogi {
							delete = ipn
						}
					}

					if delete == -1 {
						playnexts = append(playnexts, pnext{i: i, ogi: songs[i].OriginalID})
					} else {
						playnexts = append(playnexts[:delete], playnexts[delete+1:]...)
					}

					env.Draw() <- redraw(r, selected, position, lineHeight, namesImage, playnexts)

				}

			case win.MoScroll:
				newP := position.Sub(e.Point.Mul(130))
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
					env.Draw() <- redraw(r, selected, position, lineHeight, namesImage, playnexts)
					updateBrowserSlider <- position.Y
				}
			}
		}
	}
}
