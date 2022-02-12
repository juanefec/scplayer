package sc

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/juanefec/scplayer/util"
	scp "github.com/juanefec/soundcloud-api"
	"golang.org/x/image/draw"
)

type User struct {
	Username   string
	Av, AvSmol image.Image
	Err        error
}

func GetAll(username string) (User, []Song, []Song, error) {
	if username == "" {
		err := fmt.Errorf("empty username")
		return User{Err: err}, nil, nil, err
	}

	u := User{
		Username: username,
	}

	sc, err := scp.New(scp.APIOptions{})
	if err != nil {
		return User{Err: err}, nil, nil, err
	}

	user, err := sc.GetUser(scp.GetUserOptions{
		ProfileURL: "https://soundcloud.com/" + username,
	})
	if err != nil {
		return User{Err: err}, nil, nil, err
	}

	var (
		ogimg, av, avSmol  image.Image
		likes, tracks      []Song
		errt, errl, errjpg error
	)

	ParalelJobs(func() {
		ogimg, _, errjpg = GetJPEG(user.AvatarURL)
		ParalelJobs(func() {
			av = resizeImg(ogimg, 52, 52)
		}, func() {
			avSmol = resizeImg(ogimg, 26, 26)
		})
	}, func() {
		tracks, errt = getAllTracks(sc, user.ID, "")
	}, func() {
		likes, errl = getAllLikes(sc, user.ID, "")
	})

	switch true {
	case errt != nil:
		u.Err = errt
	case errl != nil:
		u.Err = errl
	case errjpg != nil:
		u.Err = errjpg
	}

	u.Av = av
	u.AvSmol = avSmol

	return u, tracks, likes, nil
}

func resizeImg(img image.Image, w, h int) image.Image {
	r := image.Rect(0, 0, w, h)
	rimg := image.NewRGBA(r)
	if img != nil {
		draw.CatmullRom.Scale(rimg, r, img, img.Bounds(), draw.Over, nil)
	}
	return rimg
}

func GetAvatar(username string, w, h int) (image.Image, error) {
	if username == "" {
		return nil, fmt.Errorf("empty username")
	}

	sc, err := scp.New(scp.APIOptions{})

	if err != nil {
		return nil, err
	}

	user, err := sc.GetUser(scp.GetUserOptions{
		ProfileURL: "https://soundcloud.com/" + username,
	})

	if err != nil {
		return nil, fmt.Errorf("user [%v] not found", username)
	}

	resp, err := http.DefaultClient.Get(user.AvatarURL)
	if err != nil {
		return nil, fmt.Errorf("avatar [%v] not found", user.AvatarURL)
	}

	img, err := jpeg.Decode(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to decode [%v]", user.AvatarURL)
	}

	r := image.Rect(0, 0, w, h)
	rimg := image.NewRGBA(r)

	draw.CatmullRom.Scale(rimg, r, img, img.Bounds(), draw.Over, nil)

	return rimg, nil
}

func GetJPEG(url string) (image.Image, []byte, error) {
	resp, err := http.DefaultClient.Get(url)
	if err != nil {
		return nil, nil, fmt.Errorf("avatar [%v] not found", url)
	}
	raw, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to read raw data [%v]", url)
	}
	img, err := jpeg.Decode(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to decode [%v]", url)
	}

	return img, raw, nil
}

func GetTracksAndLikes(username string) (tracks []Song, likes []Song, err error) {
	if username == "" {
		return nil, nil, fmt.Errorf("empty username")
	}

	sc, err := scp.New(scp.APIOptions{})
	if err != nil {
		return nil, nil, err
	}

	user, err := sc.GetUser(scp.GetUserOptions{
		ProfileURL: "https://soundcloud.com/" + username,
	})

	if err != nil {
		return nil, nil, fmt.Errorf("user [%v] not found", username)
	}

	ParalelJobs(func() {
		tracks, err = getAllTracks(sc, user.ID, "")
	}, func() {
		likes, err = getAllLikes(sc, user.ID, "")
	})

	return tracks, likes, err

}

func GetTracks(username string) ([]Song, error) {
	if username == "" {
		return nil, fmt.Errorf("empty username")
	}

	sc, err := scp.New(scp.APIOptions{})

	if err != nil {
		return nil, err
	}

	user, err := sc.GetUser(scp.GetUserOptions{
		ProfileURL: "https://soundcloud.com/" + username,
	})

	if err != nil {
		return nil, fmt.Errorf("user [%v] not found", username)
	}

	return getAllTracks(sc, user.ID, "")

}

func getAllTracks(sc *scp.API, user int64, offset string) ([]Song, error) {
	ls, err := sc.GetTracklist(scp.GetTracklistOptions{
		ID:     user,
		Type:   "tracklist",
		Limit:  200,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}

	tks, err := ls.GetTracks()
	if err != nil {
		return nil, err
	}

	songs := []Song{}
	for _, trk := range tks {
		if len(trk.Media.Transcodings) > 0 {
			urlsplit := strings.Split(trk.User.PermalinkURL, "/")
			username := urlsplit[len(urlsplit)-1]
			s := Song{
				Title:      trk.Title,
				Artist:     trk.User.Username,
				Username:   username,
				OriginalID: int(trk.ID),
				coverUrl:   trk.ArtworkURL,
				duration:   time.Duration(trk.FullDurationMS * 1000000),
				data:       trk.Media.Transcodings[0],
			}

			songs = append(songs, s)
		}
	}

	// Recursion disabled for developing, it takes some time to bring all tracks sometimes.
	//
	// After mapping SC Tracks to Songs we look for the sc_response.NextHref, this prop
	// contains the url that follows your search, this way you will be able to retrive all
	// the list.
	// Here I just take the offset value form the url.Query() and pass it recusively

	if ls.NextHref != "" {
		url, err := url.Parse(ls.NextHref)
		if err != nil {
			// log.Fatal(err.Error())
			return nil, err
		}

		off := url.Query()["offset"][0]
		at, err := getAllTracks(sc, user, off)
		if err != nil {
			return nil, err
		}
		songs = append(songs, at...)
	}

	return songs, nil
}

func GetLikes(username string) ([]Song, error) {
	if username == "" {
		return nil, fmt.Errorf("empty username")
	}

	sc, err := scp.New(scp.APIOptions{})

	if err != nil {
		return nil, err
	}

	user, err := sc.GetUser(scp.GetUserOptions{
		ProfileURL: "https://soundcloud.com/" + username,
	})

	if err != nil {
		return nil, fmt.Errorf("user [%v] not found", username)
	}

	return getAllLikes(sc, user.ID, "")

}

func getAllLikes(sc *scp.API, user int64, offset string) ([]Song, error) {
	ls, err := sc.GetTracklist(scp.GetTracklistOptions{
		ID:     user,
		Type:   "track",
		Limit:  200,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}

	l, err := ls.GetLikes()
	if err != nil {
		return nil, err
	}

	songs := []Song{}
	for _, li := range l {
		if len(li.Track.Media.Transcodings) > 0 {
			urlsplit := strings.Split(li.Track.User.PermalinkURL, "/")
			username := urlsplit[len(urlsplit)-1]
			s := Song{
				Title:      li.Track.Title,
				Artist:     li.Track.User.Username,
				Username:   username,
				OriginalID: int(li.Track.ID),
				coverUrl:   li.Track.ArtworkURL,
				duration:   time.Duration(li.Track.FullDurationMS * 1000000),
				data:       li.Track.Media.Transcodings[0],
			}

			songs = append(songs, s)
		}
	}

	// Recursion disabled for developing, it takes some time to bring all tracks sometimes.
	//
	// After mapping SC Tracks to Songs we look for the sc_response.NextHref, this prop
	// contains the url that follows your search, this way you will be able to retrive all
	// the list.
	// Here I just take the offset value form the url.Query() and pass it recusively

	// if ls.NextHref != "" {
	// 	url, err := url.Parse(ls.NextHref)
	// 	if err != nil {
	// 		// log.Fatal(err.Error())
	// 		return nil, err
	// 	}

	// 	off := url.Query()["offset"][0]
	// 	al, err := getAllLikes(sc, user, off)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	songs = append(songs, al...)
	// }

	return songs, nil
}

const (
	VolMin, VolMax = -6.5, 1.8
)

type Song struct {
	Title        string
	Artist       string
	OriginalID   int
	Username     string
	Cover        image.Image
	coverUrl     string
	duration     time.Duration
	data         scp.Transcoding
	volume       *effects.Volume
	controller   *beep.Ctrl
	streamer     beep.StreamSeeker
	format       beep.Format
	isDownloaded bool
}

type SongData struct {
	Title      string `json:"title"`
	Artist     string `json:"artist"`
	OriginalID int    `json:"id"`
	Username   string `json:"artist_username"`
}

func (song *SongData) AsSong() Song {
	return Song{
		Title:      song.Title,
		Artist:     song.Artist,
		OriginalID: song.OriginalID,
		Username:   song.Username,
	}
}

func (song *Song) AsSongData() SongData {
	return SongData{
		Title:      song.Title,
		Artist:     song.Artist,
		OriginalID: song.OriginalID,
		Username:   song.Username,
	}
}

func (song *Song) IsDownloaded() bool {
	return song.isDownloaded
}

func (song *Song) Play(vol float64, done chan<- struct{}) error {

	speaker.Init(song.format.SampleRate, song.format.SampleRate.N(time.Second/10))

	ctrl := &beep.Ctrl{
		Streamer: beep.Seq(song.streamer, beep.Callback(func() {
			done <- struct{}{}
		})),
		Paused: false,
	}

	volume := &effects.Volume{
		Streamer: ctrl,
		Base:     2,
		Volume:   0,
		Silent:   false,
	}

	song.controller = ctrl
	song.volume = volume
	song.SetVolume(vol)
	speaker.Play(song.volume)

	return nil
}

func (song *Song) Download(done chan<- int) error {
	// its a prop song to write on the browser probably.
	if song.OriginalID == 0 {
		//done <- 0
		return fmt.Errorf("error")
	}

	fmt.Println("checking localy...")
	if audio, cover, exist := LocalCheck(song.OriginalID); exist {
		err := error(nil)
		ParalelJobs(
			func() {
				song.streamer, song.format, err = mp3.Decode(ioutil.NopCloser(audio))

				buff := beep.NewBuffer(song.format)
				// it takes way too long but its the only way I can Seek the stream later >:(
				buff.Append(song.streamer)
				bstreamer := buff.Streamer(0, buff.Len())
				song.streamer = bstreamer
			},
			func() {
				song.Cover = resizeImg(cover, 40, 40)
			},
		)
		if err == nil {
			song.isDownloaded = true
			done <- song.OriginalID
		}
		fmt.Println("local files found!")
		return err
	}
	fmt.Println("downloading...")

	sc, err := scp.New(scp.APIOptions{})

	if err != nil {
		return err
	}

	var (
		audioStorCp []byte
		cover       image.Image
	)

	ParalelJobs(
		func() {
			buffer := &bytes.Buffer{}

			err = sc.DownloadTrack(song.data, buffer)

			audioStorCp = buffer.Bytes()

			song.streamer, song.format, err = mp3.Decode(ioutil.NopCloser(buffer))

			buff := beep.NewBuffer(song.format)
			// it takes way too long but its the only way I can Seek the stream later >:(
			buff.Append(song.streamer)
			bstreamer := buff.Streamer(0, buff.Len())
			song.streamer = bstreamer
		},
		func() {
			cover, _, err = GetJPEG(song.coverUrl)
			song.Cover = resizeImg(cover, 40, 40)
		},
	)

	if err != nil {
		return err
	}

	fmt.Println("saving locally...")
	err = LocalSave(song.AsSongData(), audioStorCp, cover)
	if err != nil {
		fmt.Println("error saving", err)
	}
	fmt.Println("saved", err)

	song.isDownloaded = true
	done <- song.OriginalID
	return nil
}

func (song *Song) SetVolume(vol float64) float64 {
	if song == nil || song.volume == nil {
		return vol
	}
	if vol < 0.3 {
		song.volume.Silent = true
		return vol
	}
	song.volume.Silent = false

	song.volume.Volume = util.MapFloat(float64(vol), 0, 9, VolMin, VolMax)
	return vol
}

func (song *Song) GetVolume() float64 {
	if song == nil || song.volume == nil {
		return -4.0
	}
	return util.MapFloat(song.volume.Volume, VolMin, VolMax, 0, 9)
}

func (song Song) SetProgress(pos int) {
	if song.streamer == nil {
		fmt.Println("song.streamer nil")
	}
	rpos := util.Map(pos, 0, 100, 0, song.streamer.Len()-1)
	speaker.Lock()
	err := song.streamer.Seek(rpos)
	if err != nil {
		fmt.Println(err)
	}
	speaker.Unlock()
}

func (song Song) Progress() string {
	if song.streamer == nil {
		return ""
	}
	speaker.Lock()
	t := song.format.SampleRate.D(song.streamer.Position())
	speaker.Unlock()

	return DurationToStr(t)
}

func (song Song) Duration() string {
	return DurationToStr(song.duration)
}

func (song Song) ProgressMs() int {
	if song.streamer == nil {
		return 0
	}
	speaker.Lock()
	t := song.format.SampleRate.D(song.streamer.Position())
	speaker.Unlock()

	return int(t / time.Millisecond)
}

func (song Song) DurationMs() int {
	return int(song.duration / time.Millisecond)
}

func (song *Song) Resume() error {
	if song.controller == nil {
		return errors.New("not playing")
	}
	song.controller.Paused = false
	return nil
}

func (song Song) Pause() {
	if song.controller != nil {
		song.controller.Paused = true
	}
}

func ClearSpeaker() {
	speaker.Clear()
}

func DurationToStr(d time.Duration) string {
	h, m, s := 0, 0, 0
	{
		d = d.Round(time.Second)
		hd := d / time.Hour
		d -= hd * time.Hour
		md := d / time.Minute
		d -= md * time.Minute
		sd := d / time.Second
		h, m, s = int(hd), int(md), int(sd)
	}
	if h >= 1 {
		return fmt.Sprintf("%v:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%d:%02d", m, s)
}

func ParalelJobs(jobs ...func()) {
	wg := &sync.WaitGroup{}
	wg.Add(len(jobs))
	for _, job := range jobs {
		go func(wg *sync.WaitGroup, work func()) {
			work()
			wg.Done()
		}(wg, job)
	}
	wg.Wait()
}
