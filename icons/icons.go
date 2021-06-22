package icons

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/png"
)

// base64 icons
const (
	playBtn64  = "iVBORw0KGgoAAAANSUhEUgAAACAAAAAgCAYAAABzenr0AAAAAXNSR0IArs4c6QAAAH1JREFUWIXt1cEJgDAQRNFVbMb+i0k55hJPouzOTFBhPogn44OsJsI594OOcU1pSQKqz0wFSCEMQAJRAChIFdDGfVdBUMAZDWEBNGRNADK1uMc9zpAKALeJ1oG3gAXQQ4gCXvsM2bUuKWaA+hUzAMlhhACkx3EFIH2xc+4zdY+YEFbYD6UDAAAAAElFTkSuQmCC"
	pauseBtn64 = "iVBORw0KGgoAAAANSUhEUgAAACAAAAAgCAYAAABzenr0AAAAAXNSR0IArs4c6QAAAFNJREFUWIXtljEKwCAQwGL//0x9h10USqc7UHRIQHTIQbhJEJHDlIDTE3MZF4AnEDBpO9xoQPvdq9zUBrZggAEGGGCAARH6OPXzXuECF3xIROQ4L/CnEz9Rryr6AAAAAElFTkSuQmCC"

	forwardBtn64 = "iVBORw0KGgoAAAANSUhEUgAAACAAAAAgCAYAAABzenr0AAAAAXNSR0IArs4c6QAAAIJJREFUWIXtlFEOgCAMQ4fxILv/qbwJfpE0iJSJ+GPfF0mXtWMEMyGEEGvJrGAbbNJrxPTpAGg0o08FOMCkZcT0W/ZAbTFxMElEp0RWgEY4cU9fEuBVIisoOJyTXW/BLUAkQG08otO30GpUg01a9T09M4/IG2BhR4b5nMc/pBBC/IcTIU0bzxaPN5QAAAAASUVORK5CYII="
	backBtn64    = "iVBORw0KGgoAAAANSUhEUgAAACAAAAAgCAYAAABzenr0AAAAAXNSR0IArs4c6QAAAIBJREFUWIXtlFEKwCAMQ+vYQXb/S2036X50SN2MtfMvDwQhmqZSFCGEEILRleZb4K5KPxzSQwGQ8fCreQPUXV0TesPuKFz4KtzTwwF6xq6OLZEh/AVPgCMvSwJ6GDX7sk5ppx3pDQkdyCb2XG385oH0h9kZQMFHGhtm6VdMCCHkBi3MIn+MzxIoAAAAAElFTkSuQmCC"

	shuffleBtn64 = "iVBORw0KGgoAAAANSUhEUgAAACAAAAAgCAYAAABzenr0AAAAAXNSR0IArs4c6QAAAQFJREFUWIXtlrEOgyAQhs9uDD5AH4B06eDUhJj48EZj3By6ND5AH8DB0U7XUORQjkqH8k8GzP0fx3EAkJT0DxJKLtTcKRJDS0Fk5kCRV+8fh6lezfvIZjp340dM0iAUBM3nbsyEkg0AlKa5E4ALoq8aDYWSi83cS0VeLToMZe4quGC5IA431yEoAE481jE0IaKs3ARACKFkEwLAygCeBoQIqXB2J3xcny0AwOV+Lrkx2ACYcjMTUWQ7bnt6xFcBbONcAK8t2Kr2Q7Owp9NxtoJ1uWxB4LdepNRFRgb0NaZAhqnOirzqAeBmg1gNhBrbIHSFPnK4IH3UXmEB+J15UtKWXgFNf+GSDf+zAAAAAElFTkSuQmCC"

	refreshBtn64 = "iVBORw0KGgoAAAANSUhEUgAAACAAAAAgCAYAAABzenr0AAAAAXNSR0IArs4c6QAAAJBJREFUWIVjYBgFo2CkA0ZyNPH3av7HJv6x+DpZ5pFkMS7LiZGn2HJaqKWZgRQ7AmYAJQYRo5eJXMOJAR+LrzMScgReB/D3av6ndcrGaTi6yylxCD08QtABA2b50AX0CjYWWhlMlURMjVAgVD/QtCAiJvvhdQAxJRldwIBURpQYSNM2AbUbJEOrSTYKRsGwBAD6N05wJwJqVAAAAABJRU5ErkJggg=="

	search64 = "iVBORw0KGgoAAAANSUhEUgAAABYAAAAWCAYAAADEtGw7AAAAAXNSR0IArs4c6QAAAIlJREFUOI3V0zEWgCAIBmBo6ETNHa3BozV3ohab7PEIBJUG/63C72dIBDmZPaMyp4Yf4KA17xp80XXbEx26r/NoxctQlkAeUmDii6e9JwjObUu8W/+28XzwnH9FifeCSGfNj9aV7oa1AjTet7VWQgtEoxc28RG4io/CKh4Bi3gU/MEjYY6HJwMAPPsAJhkwO17EAAAAAElFTkSuQmCC"

	gotop64    = "iVBORw0KGgoAAAANSUhEUgAAABYAAAAWCAYAAADEtGw7AAAAAXNSR0IArs4c6QAAAHhJREFUOI3t1cEJwCAMBdBaHMCNnCIZUpeoGzlCT0KgmvxASy/+o3yeiBEDJb6OD3KixdpLrr3kV2EJorgJD4gSN0rcUFyFJTrWUHwJz1APPoU1FMUfMIIieJBzPCusNrC6UUOsC9JOBT8Qbza84Q3/AUer4PnnZG7b/kU1DZxvJwAAAABJRU5ErkJggg=="
	gotosong64 = "iVBORw0KGgoAAAANSUhEUgAAABYAAAAWCAYAAADEtGw7AAAAAXNSR0IArs4c6QAAAJpJREFUOI21k9sRwBAQRROTAnSkCoqkiaQjJeSLMXLtriX30+PMdXB6G+7jh5hVQMrRpRxdP35pYWjM2/BMgRGIyxDMwdp2YhVoYQuShFVBAalTkWAEbWFlXv0qEIwLCS5AjY5pFdIs/7xRiQ9Y8xlQtjXuM3TM+eVOpmos0VUb73JbYnZAkTaVCsn7hpcn2citMVLQbGrj3fAXIE9KyUxzFkIAAAAASUVORK5CYII="

	volumeLVL9 = "iVBORw0KGgoAAAANSUhEUgAAACgAAAAoCAYAAACM/rhtAAAAAXNSR0IArs4c6QAAAJpJREFUWIXtkzENw0AQBHcN61kEQVqTcBUSboMgLEzLKd5fu3hLNyfdIBjtaH1Kp8BYkqiSlrxES9xh6noDfOJoh/z0xC9o4p9sqtwAv2C0Q3564i808Vs2VW6AXzDaIT+WpO1gJv402VS5AX7BaIf89AmPFZlYbbexchf4BaMV8oNPXCeZok7yAPjEdZIp6iQPgE9cJ5mi7f4DR3lB/Gpi44oAAAAASUVORK5CYII="
	volumeLVL8 = "iVBORw0KGgoAAAANSUhEUgAAACgAAAAoCAYAAACM/rhtAAAAAXNSR0IArs4c6QAAAJdJREFUWIXtk7ENhEAQxDxtXRdfwaffBBFNkFIBXdAWBMdlnx3SzkrrCqyxBoo5BHDBFS3yD4HkKjewXzDaIT898cc08YHkKjewXzDaIT898W6a+IvkKjewXzDaIT8CWE7PxGtDcpUb2C8Y7ZCfPuH5s0xM2yRbuQf7BaMV8mOfuE4yRZ3kBewT10mmqJO8gH3iOskUbdMNoCM/+62FTwAAAAAASUVORK5CYII="
	volumeLVL7 = "iVBORw0KGgoAAAANSUhEUgAAACgAAAAoCAYAAACM/rhtAAAAAXNSR0IArs4c6QAAAIpJREFUWIXt0zERgDAUA9DEVl2ggBUTTJjoigJc1FYZoBtby/30Lk9BLrkAZma/IgDUBTU6yBdeIFXDNfINRmeY3zPxKTrxClI1XCPfYHSG+REA9qI58ZFAqoZr5BuMzjC/p8KySU6MlEnZcC/5BqMjzE9+Yp+ki08ygPzEPkkXn2QA+Yl9ki4p8wYEFDn6/t2+bQAAAABJRU5ErkJggg=="
	volumeLVL6 = "iVBORw0KGgoAAAANSUhEUgAAACgAAAAoCAYAAACM/rhtAAAAAXNSR0IArs4c6QAAAHlJREFUWIXt07ENgDAQA0B7rSxDyxJULJGWZbJWKCBduo/0juSbwLJlwMzMzDIRAPqDnh1khgdI1XCDfIPZGfZHALia5sR3AakabpBvMDvD/r4K2yk5MUolZcP95BvMjrA/+Yl9khCfZAH5iX2SEJ9kAfmJfZKQUvkChdgxqPRj6HUAAAAASUVORK5CYII="
	volumeLVL5 = "iVBORw0KGgoAAAANSUhEUgAAACgAAAAoCAYAAACM/rhtAAAAAXNSR0IArs4c6QAAAGdJREFUWIXt07ENgDAQQ1HfWpmIJahYIhNlrVAAHd1FOkf6bwLLliUAAADAWEjSOTSrg/y5miJcw33sG6zOsL+nwnFYTqzWI2zDvewbrI6wP/uJOUkKJ1nAfmJOksJJFrCfmJOktB43heEpBgYCRCcAAAAASUVORK5CYII="
	volumeLVL4 = "iVBORw0KGgoAAAANSUhEUgAAACgAAAAoCAYAAACM/rhtAAAAAXNSR0IArs4c6QAAAFJJREFUWIXt07ENACAMA0FnLSZipEzEWtBQ0gUJI/1NEMV6CQAAAADwr5AkjT4f33HWMsL2uM3+g69P+J/9xERSQiQX2E9MJCVEcoH9xERS0jIWIjkgwYGSdRgAAAAASUVORK5CYII="
	volumeLVL3 = "iVBORw0KGgoAAAANSUhEUgAAACgAAAAoCAYAAACM/rhtAAAAAXNSR0IArs4c6QAAAExJREFUWIXt07ENgDAQA0D/WkzESJmItaChpAtSHnQ3gWXLCQAAAMBvVZLk2M/FOZ5to6ptuFv7BldH+L72EzvJFCd5QfuJnWTKNuoCv0AYkYfAUKgAAAAASUVORK5CYII="
	volumeLVL2 = "iVBORw0KGgoAAAANSUhEUgAAACgAAAAoCAYAAACM/rhtAAAAAXNSR0IArs4c6QAAAEVJREFUWIXt07ERgDAMBMFXW66IklwRbeGEkIzAYtit4EY/SgAAAACAv6okyXlcmzuejVnVNu7W/oK7E76v/cSe5JUxawEczBBh6oDAeQAAAABJRU5ErkJggg=="
	volumeLVL1 = "iVBORw0KGgoAAAANSUhEUgAAACgAAAAoCAYAAACM/rhtAAAAAXNSR0IArs4c6QAAADZJREFUWIXt0DERACAQBLF7WyhCEoqwBQ0lPRSJgp1NAAAAAACAq0qSzL4ed9y1UfVt3PH9wQ06vwgxvE+TowAAAABJRU5ErkJggg=="
	volumeLVL0 = "iVBORw0KGgoAAAANSUhEUgAAACgAAAAoCAYAAACM/rhtAAAAAXNSR0IArs4c6QAAAMtJREFUWIXtWNsWgCAIs/7/n+2942XAsJ10r+q2SAEt5WBTVBXdC5jUmsNGV7MnvtLkUOt2krAw5R1FprWYGUmIfyaYZRLmRcTYJk18qBDLpJnHIhI16VpvjYLXpPvjVvymUOS9mx0VDe/dzNNIOVjRnNYzQUtNWUmXpsMqXSOTIY3ltdWKLSKYugfRfrCHSD6E8Ns8KF1JpGuxdDcj3Q9Kd9TSdxLpW530vTjbHKzTKnWrzEG8SC3Oft1681d4sDGejYpofvV4ebAnHtKLPCEFMFT/AAAAAElFTkSuQmCC"
)

var icons = map[string]string{
	"play":         playBtn64,
	"pause":        pauseBtn64,
	"forward":      forwardBtn64,
	"back":         backBtn64,
	"shuffle":      shuffleBtn64,
	"refresh":      refreshBtn64,
	"search":       search64,
	"gotop":        gotop64,
	"gotosong":     gotosong64,
	"volume_lvl-9": volumeLVL9,
	"volume_lvl-8": volumeLVL8,
	"volume_lvl-7": volumeLVL7,
	"volume_lvl-6": volumeLVL6,
	"volume_lvl-5": volumeLVL5,
	"volume_lvl-4": volumeLVL4,
	"volume_lvl-3": volumeLVL3,
	"volume_lvl-2": volumeLVL2,
	"volume_lvl-1": volumeLVL1,
	"volume_lvl-0": volumeLVL0,
}

func GetIcon(iconName string) image.Image {
	icon64 := icons[iconName]
	return getIcon(icon64)
}

func getIcon(icon string) image.Image {
	bicons, err := base64.StdEncoding.DecodeString(icon)
	if err != nil {
		panic(err)
	}
	i, err := png.Decode(bytes.NewReader(bicons))
	if err != nil {
		panic(err)
	}
	return i
}
