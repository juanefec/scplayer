package component

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math/rand"
	"time"

	"github.com/faiface/gui"
	"github.com/faiface/gui/win"
	"github.com/juanefec/scplayer/sc"
	. "github.com/juanefec/scplayer/util"
	"golang.org/x/image/math/fixed"
)

type pnext struct {
	sc.Song
	i int
}

func Browser(env gui.Env, theme *Theme,
	action <-chan string,
	song2player chan<- sc.Song,
	move <-chan int,
	pausebtn chan<- bool,
	reloadUser <-chan string,
	listenBrowserSlider <-chan int,
	newBrowserSlider chan<- int,
	updateBrowserSlider chan<- int,
	playingPos chan<- int,
	clearPlaylist <-chan struct{},
	addUserHistory chan<- sc.User,
	reloadUserAvatar chan<- sc.User,
	swapOption chan sc.User) {
	const inset = 4
	metrics := theme.Face.Metrics()
	lineHeight := (metrics.Height + 2*fixed.I(inset)).Ceil()

	createImage := func(images []image.Image) *image.RGBA {
		var width int
		for _, img := range images {
			if img.Bounds().Dx() > width {
				width = img.Bounds().Inset(-inset).Dx()
			}
		}

		height := lineHeight * len(images)

		namesImage := image.NewRGBA(image.Rect(0, 0, width+2*inset, height+2*inset))
		for i := range images {
			r := image.Rect(
				0, lineHeight*i,
				width, lineHeight*(i+1),
			)
			DrawLeftCentered(namesImage, r.Inset(inset), images[i], draw.Over)
		}

		return namesImage
	}

	reload := func(songs []sc.Song) *image.RGBA {
		var images []image.Image

		for _, song := range songs {
			images = append(images, MakeTextImage(fmt.Sprintf("[%v]    %v - %v", song.Username, song.Artist, song.Title), theme.Face, theme.Text))
		}
		return createImage(images)
	}

	reloadPlaynext := func(pnexts []pnext) *image.RGBA {
		var images []image.Image

		for _, ns := range pnexts {
			if ns.OriginalID != 0 {
				images = append(images, MakeTextImage(fmt.Sprintf("[%v]    %v - %v", ns.Username, ns.Artist, ns.Title), theme.Face, theme.Text))
			}
		}
		return createImage(images)
	}

	redraw := func(r image.Rectangle, showing string, selected int, position image.Point, lineHeight int, namesImage image.Image, playnextsView []pnext) func(draw.Image) image.Rectangle {
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

			if len(playnextsView) > 0 {
				maskOn := func(on int) {
					highlightR := image.Rect(
						namesImage.Bounds().Min.X,
						namesImage.Bounds().Min.Y+lineHeight*on,
						r.Max.X,
						namesImage.Bounds().Min.Y+lineHeight*(on+1),
					)
					highlightR = highlightR.Sub(position).Add(r.Min)
					draw.DrawMask(
						drw, highlightR.Intersect(r),
						&image.Uniform{theme.NextHighlight}, image.Point{},
						&image.Uniform{color.Alpha{64}}, image.Point{},
						draw.Over,
					)
				}
				if showing != "playlist" {
					for _, pnx := range playnextsView {
						maskOn(pnx.i)
					}
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

	type UserProfile struct {
		Username string
		Pos      map[string]image.Point
		Profile  sc.User
		Tracks,
		Likes,
		Locals []sc.Song
	}

	var (
		user  = UserProfile{Profile: sc.User{Err: errors.New("first user empty")}}
		users = map[string]UserProfile{}
		//err          error
		r            image.Rectangle
		selected     = -1
		selectedOGID = -1

		likesImage, tracksImage, playlistImage, localsImage *image.RGBA
		namesImage                                          = image.NewRGBA(r)

		showing string = "tracks" // "likes" || "tracks" || "playlist"

		playnexts            []pnext
		playnextsViewMatches []pnext
		mouseRbPressed       bool
	)

	//tracksImage = reload(tracks)
	//likesImage = reload(likes)
	playlistImage = reloadPlaynext(playnextsViewMatches)

	getSongs := func(user string) (sc.User, []sc.Song, []sc.Song) {
		scuser, tracks, likes, err := sc.GetAll(user)
		if err != nil {
			fmt.Println(err.Error())
			tracks = []sc.Song{
				{Artist: "- - - - - - -", Title: fmt.Sprintf("- - - - >  %v                 ", err.Error())},
			}
			likes = tracks
		}
		return scuser, tracks, likes
	}

	clearScreen := func() {
		namesImage = image.NewRGBA(r)
		playnextsViewMatches = []pnext{}
		playlistImage = reloadPlaynext(playnextsViewMatches)
		env.Draw() <- redraw(r, showing, -1, image.Pt(0, 0), lineHeight, namesImage, playnextsViewMatches)
	}

	refreshTracks := func() {
		tracksImage = reload(user.Tracks)
		selected = -1
		for i, s := range user.Tracks {
			if s.OriginalID == selectedOGID {
				selected = i
			}
			for _, pnx := range playnexts {
				if pnx.OriginalID == s.OriginalID {
					pnx.i = i
					playnextsViewMatches = append(playnextsViewMatches, pnx)
				}
			}
		}
		songs = user.Tracks
		namesImage = tracksImage

	}

	refreshPlaylist := func() {
		selected = -1
		playnextsViewMatches = playnexts
		playlistImage = reloadPlaynext(playnextsViewMatches)
		songs = []sc.Song{}
		for _, pnx := range playnextsViewMatches {
			songs = append(songs, pnx.Song)
		}

		namesImage = playlistImage

	}

	refreshLocals := func() {
		user.Locals = sc.LoadLocals()
		localsImage = reload(user.Locals)
		selected = -1
		for i, s := range user.Locals {
			if s.OriginalID == selectedOGID {
				selected = i
			}
			for _, pnx := range playnexts {
				if pnx.OriginalID == s.OriginalID {
					pnx.i = i
					playnextsViewMatches = append(playnextsViewMatches, pnx)
				}
			}
		}
		songs = user.Locals
		namesImage = localsImage

	}

	refreshLikes := func() {
		likesImage = reload(user.Likes)
		selected = -1
		for i, s := range user.Likes {
			if s.OriginalID == selectedOGID {
				selected = i
			}
			for _, pnx := range playnexts {
				if pnx.OriginalID == s.OriginalID {
					pnx.i = i
					playnextsViewMatches = append(playnextsViewMatches, pnx)
				}
			}
		}
		songs = user.Likes
		namesImage = likesImage

	}

	refresh := func(fetch bool) {
		playnextsViewMatches = []pnext{}
		if fetch {
			clearScreen()
			var err error
			user.Profile, user.Tracks, user.Likes = getSongs(user.Username)
			if err == nil {
				reloadUserAvatar <- user.Profile
			}
		}
		if showing == "tracks" {
			refreshTracks()
		} else if showing == "playlist" {
			refreshPlaylist()
		} else if showing == "local" {
			refreshLocals()
		} else {
			refreshLikes()
		}
		newBrowserSlider <- namesImage.Rect.Dy()
		updateBrowserSlider <- user.Pos[showing].Y
		playingPos <- lineHeight * selected
		env.Draw() <- redraw(r, showing, selected, user.Pos[showing], lineHeight, namesImage, playnextsViewMatches)
	}

	changeUserProfile := func(nusername string) {
		if u, ok := users[nusername]; ok {
			swapOption <- user.Profile
			users[user.Username] = user
			user = u
			refresh(false)
			return
		}
		if nusername != "" {

			// update username for fetching

			clearScreen()
			if user.Profile.Err == nil {
				users[user.Username] = user
				addUserHistory <- user.Profile
			}

			user = UserProfile{
				Username: nusername,
				Pos: map[string]image.Point{
					"tracks":   {},
					"likes":    {},
					"playlist": {},
				},
				Profile: sc.User{Err: errors.New("not loaded yet")},
			}

			user.Profile, user.Tracks, user.Likes = getSongs(user.Username)

			reloadUserAvatar <- user.Profile
		}
		if showing == "tracks" {
			refreshTracks()
		} else if showing == "playlist" {
			refreshPlaylist()
		} else {
			refreshLikes()
		}
		newBrowserSlider <- namesImage.Rect.Dy()
		updateBrowserSlider <- user.Pos[showing].Y
		playingPos <- lineHeight * selected
		env.Draw() <- redraw(r, showing, selected, user.Pos[showing], lineHeight, namesImage, playnextsViewMatches)
	}

	scrollView := func(e win.MoScroll) {
		newP := user.Pos[showing].Sub(e.Point.Mul(62))
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
		if newP != user.Pos[showing] {
			user.Pos[showing] = newP
			env.Draw() <- redraw(r, showing, selected, user.Pos[showing], lineHeight, namesImage, playnextsViewMatches)
			updateBrowserSlider <- user.Pos[showing].Y
		}
	}

	clickEvent := func(e win.MoUp, showing string) {
		if !e.Point.In(r) {
			return
		}

		if showing != "playlist" {
			click := e.Point.Sub(r.Min).Add(user.Pos[showing])
			i := click.Y / lineHeight
			if i < 0 || i >= len(songs) {
				return
			}
			switch e.Button {
			case win.ButtonLeft:
				if selected != i {
					selected = i
					if songs[selected].OriginalID == 0 {
						return
					}

					selectedOGID = songs[selected].OriginalID
					song2player <- songs[selected]
					pausebtn <- true
					env.Draw() <- redraw(r, showing, selected, user.Pos[showing], lineHeight, namesImage, playnextsViewMatches)
					playingPos <- lineHeight * selected
				}
			case win.ButtonRight:
				if songs[i].OriginalID == 0 {
					return
				}

				delete := -1
				for ipn := 0; ipn < len(playnexts); ipn++ {
					if songs[i].OriginalID == playnexts[ipn].OriginalID {
						delete = ipn
					}
				}

				if delete == -1 {
					pnx := pnext{songs[i], i}

					playnexts = append(playnexts, pnx)
					playnextsViewMatches = append(playnextsViewMatches, pnx)

				} else {
					playnexts = append(playnexts[:delete], playnexts[delete+1:]...)
					playnextsViewMatches = playnexts
					refresh(false)

				}

				env.Draw() <- redraw(r, showing, selected, user.Pos[showing], lineHeight, namesImage, playnextsViewMatches)

			}
		} else {
			click := e.Point.Sub(r.Min).Add(user.Pos[showing])
			i := click.Y / lineHeight
			if i < 0 || i >= len(playnexts) {
				return
			}
			switch e.Button {
			case win.ButtonLeft:
				pnx := playnexts[i]
				selectedOGID = pnx.OriginalID
				song2player <- pnx.Song
				pausebtn <- true
				env.Draw() <- redraw(r, showing, selected, user.Pos[showing], lineHeight, namesImage, playnextsViewMatches)
				//playingPos <- lineHeight * selected
			case win.ButtonRight:
				if playnexts[i].OriginalID == 0 {
					return
				}

				playnexts = append(playnexts[:i], playnexts[i+1:]...)
				refresh(false)
				env.Draw() <- redraw(r, showing, selected, user.Pos[showing], lineHeight, namesImage, playnextsViewMatches)
			}
		}
	}

	shuffle := func() {
		if showing != "playlist" {
			rand.Seed(time.Now().Unix())
			rand.Shuffle(len(songs), func(i, j int) { songs[i], songs[j] = songs[j], songs[i] })
			namesImage = reload(songs)

			user.Pos[showing] = image.Point{}
			selected = -1
			for i, s := range songs {
				if s.OriginalID == selectedOGID {
					selected = i
				}
				for pi, pnx := range playnexts {
					if s.OriginalID == pnx.OriginalID {
						playnexts[pi].i = i
					}
				}
			}
			yPos := selected * lineHeight
			if selected != -1 {
				user.Pos[showing] = user.Pos[showing].Add(image.Point{0, yPos})
			}
			playnextsViewMatches = playnexts
			playlistImage = reloadPlaynext(playnextsViewMatches)
			env.Draw() <- redraw(r, showing, selected, user.Pos[showing], lineHeight, namesImage, playnextsViewMatches)
			updateBrowserSlider <- user.Pos[showing].Y
			playingPos <- lineHeight * selected
		}
	}

	moveSong := func(m int) {
		if m > 0 && len(playnexts) > 0 {
			s := playnexts[0]
			if s.i >= 0 && s.i < len(songs) {
				for i, sng := range songs {
					if s.OriginalID == sng.OriginalID {
						selected = i

					}
				}

			}
			selectedOGID = s.OriginalID
			playnexts = playnexts[1:]
			selectedOGID = s.OriginalID
			refresh(false)
			song2player <- s.Song
			pausebtn <- true
			playingPos <- lineHeight * selected
			env.Draw() <- redraw(r, showing, selected, user.Pos[showing], lineHeight, namesImage, playnextsViewMatches)
			return
		}
		if selected >= -1 && selected < len(songs)+1 && len(songs) > 0 {
			if selected == 0 && m == -1 {
				return
			}

			selected = selected + m

			if selected < 0 || selected > len(songs)-1 {
				selected = -1
				return
			}
			if selected < 0 || songs[selected].OriginalID == 0 {
				return
			}
			selectedOGID = songs[selected].OriginalID
			song2player <- songs[selected]
			pausebtn <- true
			playingPos <- lineHeight * selected
			env.Draw() <- redraw(r, showing, selected, user.Pos[showing], lineHeight, namesImage, playnextsViewMatches)
		}
	}

	sliderAction := func(y int) {
		user.Pos[showing] = image.Point{0, y}
		env.Draw() <- redraw(r, showing, selected, user.Pos[showing], lineHeight, namesImage, playnextsViewMatches)
	}

	showTracks := func() {
		showing = "tracks"
		refresh(false)
	}

	showLikes := func() {
		showing = "likes"
		refresh(false)
	}

	showPlaylist := func() {
		showing = "playlist"
		refresh(false)
	}

	showLocal := func() {
		showing = "local"
		refresh(false)
	}

	mrbSelectPlaynext := func(e win.MoMove, showing string) {
		if mouseRbPressed && showing != "playlist" {
			if !e.Point.In(r) {
				return
			}
			click := e.Point.Sub(r.Min).Add(user.Pos[showing])
			i := click.Y / lineHeight
			if i < 0 || i >= len(songs) {
				return
			}
			if songs[i].OriginalID == 0 {
				return
			}

			if len(playnexts) > 0 {
				last := playnexts[len(playnexts)-1].i
				if i == last {
					return
				}
			}
			pnx := pnext{songs[i], i}

			playnexts = append(playnexts, pnx)
			playnextsViewMatches = append(playnextsViewMatches, pnx)

			env.Draw() <- redraw(r, showing, selected, user.Pos[showing], lineHeight, namesImage, playnextsViewMatches)
		}
	}

	for {
		select {
		case y := <-listenBrowserSlider:
			sliderAction(y)
		case <-clearPlaylist:
			playnexts = []pnext{}
			refresh(false)
		case newuser := <-reloadUser:
			changeUserProfile(newuser)
		case action := <-action:
			switch action {
			case "refresh":
				refresh(true)
			case "shuffle":
				shuffle()
			case "tracks":
				showTracks()
			case "likes":
				showLikes()
			case "playlist":
				showPlaylist()
			case "local":
				showLocal()
			}
		case m := <-move:
			moveSong(m)
		case e, ok := <-env.Events():
			if !ok {
				close(env.Draw())
				return
			}
			switch e := e.(type) {
			case gui.Resize:
				r = e.Rectangle
				env.Draw() <- redraw(r, showing, selected, user.Pos[showing], lineHeight, namesImage, playnextsViewMatches)
			case win.MoUp:
				clickEvent(e, showing)
				if e.Button == win.ButtonRight {
					mouseRbPressed = false
				}
			case win.MoDown:
				if e.Button == win.ButtonRight {
					mouseRbPressed = true
				}
			case win.MoMove:
				mrbSelectPlaynext(e, showing)
			case win.MoScroll:
				scrollView(e)
			}
		}
	}
}
