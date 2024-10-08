package main

import (
	"fmt"
	"log"
	"net/http"

	"strings"

	"github.com/boltdb/bolt"
	"github.com/gin-gonic/gin"
	"github.com/gowon-irc/go-gowon"
	"github.com/imroc/req/v3"
	"github.com/jessevdk/go-flags"
)

type Options struct {
	APIKey string `short:"k" long:"api-key" env:"GOWON_LASTFM_API_KEY" required:"true" description:"last.fm api key"`
	KVPath string `short:"K" long:"kv-path" env:"GOWON_LASTFM_KV_PATH" default:"kv.db" description:"path to kv db"`
}

const (
	moduleName = "lastfm"
	moduleHelp = "show last listened tracks on last.fm"
)

func setUser(kv *bolt.DB, nick, user []byte) error {
	err := kv.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(moduleName))
		return b.Put([]byte(nick), []byte(user))
	})
	return err
}

func getUser(kv *bolt.DB, nick []byte) (user []byte, err error) {
	err = kv.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(moduleName))
		v := b.Get([]byte(nick))
		user = v
		return nil
	})
	return user, err
}

func parseArgs(msg string) (command, user string) {
	fields := strings.Fields(msg)

	if len(fields) >= 1 {
		command = fields[0]
	}

	if len(fields) >= 2 {
		user = fields[1]
	}

	return command, user
}

func setUserHandler(kv *bolt.DB, nick, user string) (string, error) {
	if user == "" {
		return "Error: username needed", nil
	}

	err := setUser(kv, []byte(nick), []byte(user))
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("set %s's user to %s", nick, user), nil
}

type commandFunc func(*req.Client, string) (string, error)

func CommandHandler(client *req.Client, kv *bolt.DB, nick, user string, f commandFunc) (string, error) {
	if user != "" {
		return f(client, user)
	}

	savedUser, err := getUser(kv, []byte(nick))
	if err != nil {
		return "", err
	}

	if len(savedUser) == 0 {
		return "Error: username needed", nil
	}

	return f(client, string(savedUser))
}

func lastfmHandler(client *req.Client, kv *bolt.DB, m *gowon.Message) (string, error) {
	command, user := parseArgs(m.Args)

	switch command {
	case "s", "set":
		return setUserHandler(kv, m.Nick, user)
	case "l", "scrobbles":
		return CommandHandler(client, kv, m.Nick, user, lastfmNewestScrobble)
	case "ta", "topartists":
		return CommandHandler(client, kv, m.Nick, user, lastfmTopArtistsAllTime)
	case "taw", "topartistsweekly":
		return CommandHandler(client, kv, m.Nick, user, lastfmTopArtistsWeekly)
	case "tam", "topartistsmonthly":
		return CommandHandler(client, kv, m.Nick, user, lastfmTopArtistsMonthly)
	case "ta3m", "topartists3monhtly":
		return CommandHandler(client, kv, m.Nick, user, lastfmTopArtists3Monthly)
	case "ta6m", "topartists6monhtly":
		return CommandHandler(client, kv, m.Nick, user, lastfmTopArtists6Monthly)
	case "tay", "topartistsyearly":
		return CommandHandler(client, kv, m.Nick, user, lastfmTopArtistsYearly)
	}

	commands := []string{
		"set (s)",
		"scrobbles (l)",
		"topartists (ta)",
		"topartistsweekly (taw)",
		"topartistsmonthly (tam)",
		"topartists3monthly (ta3m)",
		"topartists6monthly (ta6m)",
		"topartistsyearly (tay)",
	}

	cl := colourList(commands)

	return fmt.Sprintf("Available commands: %s", strings.Join(cl, ", ")), nil
}

func main() {
	log.Printf("%s starting\n", moduleName)

	opts := Options{}
	if _, err := flags.Parse(&opts); err != nil {
		log.Fatal(err)
	}

	kv, err := bolt.Open(opts.KVPath, 0666, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer kv.Close()

	err = kv.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(moduleName))
		return err
	})
	if err != nil {
		log.Fatal(err)
	}

	httpClient := req.C().
		SetCommonQueryParam("api_key", opts.APIKey)

	r := gin.Default()
	r.POST("/message", func(c *gin.Context) {
		var m gowon.Message

		if err := c.BindJSON(&m); err != nil {
			log.Println("Error: unable to bind message to json", err)
			return
		}

		out, err := lastfmHandler(httpClient, kv, &m)
		if err != nil {
			log.Println(err)
			m.Msg = "{red}Error when looking up last.fm tracks{clear}"
			c.IndentedJSON(http.StatusInternalServerError, &m)
		}

		m.Msg = out
		c.IndentedJSON(http.StatusOK, &m)
	})

	r.GET("/help", func(c *gin.Context) {
		c.IndentedJSON(http.StatusOK, &gowon.Message{
			Module: moduleName,
			Msg:    moduleHelp,
		})
	})

	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
