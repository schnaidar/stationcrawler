package stationcrawler

import (
	"time"
	"net/http"
	"github.com/gitschneider/radiowatch"
	"github.com/Jeffail/gabs"
	"errors"
	"fmt"
	"html"
)

type dasDingSongInfo struct {
	Artist        string `json:"artist"`
	ArtistDetails interface{} `json:"artistDetails"`
	CoverURL      string `json:"coverUrl"`
	CurrentLikes  string `json:"currentLikes"`
	DetailPageURL string `json:"detailPageUrl"`
	Duration      string `json:"duration"`
	HookFile      string `json:"hookFile"`
	ID            string `json:"id"`
	LikeURL       string `json:"likeUrl"`
	PlayedAtTs    string `json:"playedAtTs"`
	ScheduledAtTs string `json:"scheduledAtTs"`
	SocialMessage string `json:"socialMessage"`
	Title         string `json:"title"`
}

type DasDingCrawler struct {
	crawler
}

func (d *DasDingCrawler) Crawl() (*radiowatch.TrackInfo, error) {
	resp, err := http.Get("http://www.dasding.de/ext/playlist/gen2_pl.json")
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return nil, err
	}

	jsonParsed, err := gabs.ParseJSONBuffer(resp.Body)
	if err != nil {
		return nil, err
	}
	songs, err := jsonParsed.Children()
	if err != nil {
		return nil, err
	}

	var song *gabs.Container
	for _, song = range songs {
		// is there even a song being played?
		songTitle, ok := song.Search("title").Data().(string)
		if !ok {
			return nil, fmt.Errorf("No song is played at the moment!")
		}
		if songTitle == "" {
			return nil, fmt.Errorf("No song is played at the moment!")
		}

		// Check if the current entry is already scheduled
		playedAt := song.Search("playedAtTs").Data()
		if playedAt == nil{
			continue
		}

		scheduledAt, ok := song.Search("scheduledAtTs").Data().(string)
		if !ok {
			return nil, fmt.Errorf("Error while converting scheduledAt timestamp. Expected string, got %s",
				song.Search("scheduledAtTs").Data())
		}
		start, err := parseUnixString(scheduledAt)
		if err != nil {
			return nil, err
		}

		duration, ok := song.Search("duration").Data().(string)
		if !ok {
			return nil, fmt.Errorf("Error while converting duration. Expected tring, got %s",
				song.Search("duration").Data())
		}

		length, err := time.ParseDuration(duration + "s")
		if err != nil {
			return nil, err
		}

		end := start.Add(length)
		if isNow(start, end) {
			d.setNextCrawlTime(end)

			artist := html.UnescapeString(song.Search("artist").Data().(string))
			//if strings.Contains(artist, ",") {
			//	parts := strings.Split(artist, ",")
			//	for i, j := 0, len(parts) - 1; i < j; i, j = i +1, j -1 {
			//		parts[i], parts[j] = parts[j], parts[i]
			//	}
			//	artist = strings.Join(parts, ",")
			//}

			return &radiowatch.TrackInfo{
				Artist: artist,
				CrawlTime: time.Now(),
				Station: d.name,
				Title: html.UnescapeString(songTitle),
			}, nil
		}
	}

	return nil, errors.New("No song is played at the moment.")
}

func (d *DasDingCrawler)Name() string {
	return d.name
}

func NewDasDing() *DasDingCrawler {
	return &DasDingCrawler{newCrawler("Das Ding")}
}