package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/boltdb/bolt"
	"github.com/gin-gonic/gin"
	"github.com/gowon-irc/go-gowon"
	"github.com/jessevdk/go-flags"
)

type Options struct {
	Prefix string `short:"P" long:"prefix" env:"GOWON_PREFIX" default:"." description:"prefix for commands"`
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

func lastfmHandler(apiKey string, kv *bolt.DB, m *gowon.Message) (string, error) {
	fields := strings.Fields(m.Args)

	if len(fields) >= 2 && fields[0] == "set" {
		err := setUser(kv, []byte(m.Nick), []byte(fields[1]))
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("set %s's user to %s", m.Nick, fields[1]), nil
	}

	if len(fields) >= 1 {
		user := strings.Fields(m.Args)[0]
		return lastfm(user, apiKey)
	}

	user, err := getUser(kv, []byte(m.Nick))
	if err != nil {
		return "", err
	}

	if len(user) > 0 {
		return lastfm(string(user), apiKey)
	}

	return "Error: username needed", nil
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

	r := gin.Default()
	r.POST("/message", func(c *gin.Context) {
		var m gowon.Message

		if err := c.BindJSON(&m); err != nil {
			log.Println("Error: unable to bind message to json", err)
			return
		}

		out, err := lastfmHandler(opts.APIKey, kv, &m)
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
