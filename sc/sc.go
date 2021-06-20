package sc

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	scp "github.com/zackradisic/soundcloud-api"
)

type Song struct {
	Title      string
	Artist     string
	OriginalID int
	paused     bool
	duration   time.Duration
	data       scp.Transcoding
	volume     *effects.Volume
	controller *beep.Ctrl
	streamer   beep.StreamSeekCloser
	format     beep.Format
}

func GetLikes(username string) ([]Song, error) {
	sc, err := scp.New(scp.APIOptions{})

	if err != nil {
		log.Fatal(err.Error())
	}

	user, err := sc.GetUser(scp.GetUserOptions{
		ProfileURL: "https://soundcloud.com/" + username,
	})

	if err != nil {
		return nil, fmt.Errorf("user [%v] not found", username)
		//log.Fatal(err.Error())
	}

	return getAllLikes(sc, user.ID, 0), nil

}

func getAllLikes(sc *scp.API, user int64, offset int) []Song {
	ls, err := sc.GetLikes(scp.GetLikesOptions{
		ID:     user,
		Type:   "track",
		Limit:  200,
		Offset: offset,
	})
	if err != nil {
		log.Fatal(err.Error())
	}

	l, err := ls.GetLikes()
	if err != nil {
		log.Fatal(err.Error())
	}

	songs := []Song{}
	for _, li := range l {
		if len(li.Track.Media.Transcodings) > 0 {
			s := Song{
				Title:      li.Track.Title,
				Artist:     li.Track.User.Username,
				OriginalID: int(li.Track.ID),
				duration:   time.Duration(li.Track.FullDurationMS * 1000000),
				data:       li.Track.Media.Transcodings[0],
			}

			songs = append(songs, s)
		}
	}

	// if ls.NextHref != "" {
	// 	url, err := url.Parse(ls.NextHref)
	// 	if err != nil {
	// 		log.Fatal(err.Error())
	// 	}

	// 	off, err := strconv.Atoi(url.Query()["offset"][0])
	// 	if err != nil {
	// 		log.Fatal(err.Error())
	// 	}

	// 	songs = append(songs, getAllLikes(sc, user, off)...)
	// }

	return songs
}

func (song *Song) Play(done chan<- struct{}) error {
	sc, err := scp.New(scp.APIOptions{})

	if err != nil {
		log.Fatal(err.Error())
	}

	buffer := &bytes.Buffer{}

	err = sc.DownloadTrack(song.data, buffer)
	if err != nil {
		log.Fatal(err.Error())
		return err
	}

	streamer, format, err := mp3.Decode(ioutil.NopCloser(buffer))
	if err != nil {
		log.Fatal(err)
		return err
	}

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	ctrl := &beep.Ctrl{
		Streamer: beep.Seq(streamer, beep.Callback(func() {
			buffer.Reset()
			done <- struct{}{}
		})),
		Paused: false,
	}

	volume := &effects.Volume{
		Streamer: ctrl,
		Base:     2,
		Volume:   -4,
		Silent:   false,
	}

	song.format = format
	song.streamer = streamer
	song.controller = ctrl
	song.volume = volume
	speaker.Play(song.volume)

	return nil
}

func (song Song) Progress() string {
	if song.streamer == nil {
		return ""
	}
	speaker.Lock()
	t := song.format.SampleRate.D(song.streamer.Position())
	speaker.Unlock()

	return durationToStr(t)
}

func (song Song) Duration() string {
	return durationToStr(song.duration)
}

func (song Song) ProgressSegs() int {
	if song.streamer == nil {
		return 0
	}
	speaker.Lock()
	t := song.format.SampleRate.D(song.streamer.Position())
	speaker.Unlock()

	return int(t / time.Second)
}

func (song Song) DurationSegs() int {
	return int(song.duration / time.Second)
}

func (song Song) Resume() error {
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

func (song Song) Stop() {
	speaker.Clear()
}

func durationToStr(d time.Duration) string {
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
