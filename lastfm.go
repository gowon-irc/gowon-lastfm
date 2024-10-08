package main

import (
	"fmt"
	"strings"

	"github.com/imroc/req/v3"
)

const (
	lastfmAPIURL = "https://ws.audioscrobbler.com/2.0"
)

func colourString(in, colour string) string {
	return fmt.Sprintf("{%s}%s{clear}", colour, in)
}

func colourList(in []string) (out []string) {
	out = []string{}

	colours := []string{"green", "red", "blue", "orange", "magenta", "cyan", "yellow"}
	cl := len(colours)

	for n, i := range in {
		c := colours[n%cl]
		o := colourString(i, c)
		out = append(out, o)
	}

	return out
}

type lastfmUserGetRecentTracks struct {
	Recenttracks Recenttracks `json:"recenttracks"`
}

func (j lastfmUserGetRecentTracks) String() string {
	return j.Recenttracks.String()
}

// Recenttracks represents the metadata returned in the lastfm json
type Recenttracks struct {
	User   User    `json:"@attr"`
	Tracks []Track `json:"track"`
}

func (r Recenttracks) String() string {
	if len(r.Tracks) == 0 {
		return "No tracks found for user"
	}

	track := r.Tracks[0]

	return fmt.Sprintf("{green}{clear} %s %s: %s {green}{clear}", r.User, track.action(), track)
}

// User represents the user information returned in the lastfm json
type User struct {
	User string `json:"user"`
}

func (u User) String() string {
	return u.User
}

// Artist represents the artist information returned in the lastfm json
type Artist struct {
	Name string `json:"#text"`
}

func (a Artist) String() string {
	return a.Name
}

// TrackAttr contains the track metadata, with now playing information
type TrackAttr struct {
	Nowplaying string `json:"nowplaying"`
}

// Album represents the album information returned in the lastfm json
type Album struct {
	Name string `json:"#text"`
}

func (a Album) String() string {
	return a.Name
}

// Track represents the track information returned in the lastfm json
type Track struct {
	Artist     Artist     `json:"artist"`
	Nowplaying *TrackAttr `json:"@attr,omitempty"`
	Album      Album      `json:"album"`
	Name       string     `json:"name"`
}

func (t Track) String() string {
	return fmt.Sprintf("%s - %s (%s)", t.Artist, t.Name, t.Album)
}

func (t Track) action() string {
	if t.Nowplaying == nil {
		return "last listened to"
	}
	return "is listening to"
}

func lastfmNewestScrobble(client *req.Client, user string) (msg string, err error) {
	j := &lastfmUserGetRecentTracks{}

	_, err = client.R().
		SetQueryParam("method", "user.getrecenttracks").
		SetQueryParam("user", user).
		SetQueryParam("format", "json").
		SetQueryParam("limit", "1").
		SetSuccessResult(&j).
		Get(lastfmAPIURL)

	if err != nil {
		return "", err
	}

	return j.String(), nil
}

type lastfmUserGetTopArtists struct {
	TopArtists struct {
		Artists []struct {
			Name      string `json:"name"`
			PlayCount string `json:"playcount"`
		} `json:"artist"`
	} `json:"topartists"`
}

func lastfmTopArtists(client *req.Client, user, period string) (msg string, err error) {
	j := &lastfmUserGetTopArtists{}

	_, err = client.R().
		SetQueryParam("method", "user.gettopartists").
		SetQueryParam("user", user).
		SetQueryParam("format", "json").
		SetQueryParam("limit", "10").
		SetQueryParam("period", period).
		SetSuccessResult(&j).
		Get(lastfmAPIURL)

	if err != nil {
		return "", err
	}

	artists := []string{}
	for _, a := range j.TopArtists.Artists {
		artists = append(artists, fmt.Sprintf("%s (%s)", a.Name, a.PlayCount))
	}

	cl := colourList(artists)

	return strings.Join(cl, ", "), err
}

func lastfmTopArtistsAllTime(client *req.Client, user string) (msg string, err error) {
	topartists, err := lastfmTopArtists(client, user, "overall")

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s's top artists: %s", user, topartists), nil
}

func lastfmTopArtistsWeekly(client *req.Client, user string) (msg string, err error) {
	topartists, err := lastfmTopArtists(client, user, "7day")

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s's top artists (last week): %s", user, topartists), nil
}

func lastfmTopArtistsMonthly(client *req.Client, user string) (msg string, err error) {
	topartists, err := lastfmTopArtists(client, user, "1month")

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s's top artists (last month): %s", user, topartists), nil
}

func lastfmTopArtists3Monthly(client *req.Client, user string) (msg string, err error) {
	topartists, err := lastfmTopArtists(client, user, "3month")

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s's top artists (last 3 months): %s", user, topartists), nil
}

func lastfmTopArtists6Monthly(client *req.Client, user string) (msg string, err error) {
	topartists, err := lastfmTopArtists(client, user, "6month")

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s's top artists (last 6 months): %s", user, topartists), nil
}

func lastfmTopArtistsYearly(client *req.Client, user string) (msg string, err error) {
	topartists, err := lastfmTopArtists(client, user, "12month")

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s's top artists (last year): %s", user, topartists), nil
}
