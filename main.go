package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/urfave/cli/v2"
)

// Config holds the configuration
var Config struct {
	dg       *discordgo.Session
	token    string
	filename string
	guild    string
	bot      string
	isBot    bool
	message  string
}

// IDs are the user IDs
var banCount uint

func main() {
	log.SetPrefix("[DG Ban] ")
	log.SetFlags(0)

	app := cli.App{
		Name:    "dgban",
		Authors: []*cli.Author{{Name: "Xyatt#6392", Email: "Xyatt#6392gmail.com"}},
		Usage:   "Discord moderation bot",
		Action: func(_ *cli.Context) error {
			setupBot()
			if Config.guild != "" {
				banUsers()
			}
			sc := make(chan os.Signal, 1)
			signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
			<-sc
			Config.dg.Close()
			return nil
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "guild",
				Aliases:     []string{"g"},
				Destination: &Config.guild,
			},
			&cli.StringFlag{
				Name:        "file",
				Aliases:     []string{"f"},
				Destination: &Config.filename,
				Value:       "userID.json",
			},
			&cli.StringFlag{
				Name:        "message",
				Aliases:     []string{"m"},
				Destination: &Config.message,
				Value:       "!banhere",
			},
			&cli.StringFlag{
				Name:        "token",
				Aliases:     []string{"t"},
				Destination: &Config.token,
				Required:    true,
				Usage:       "token required for the bot to work",
			},
			&cli.BoolFlag{
				Name:        "bot",
				Aliases:     []string{"b"},
				Destination: &Config.isBot,
				Value:       false,
				Usage:       "set bot to true to use a bot token instead of a user token",
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatalln(err)
	}
}

func setupBot() {
	Config.token = Config.bot + Config.token
	var err error
	Config.dg, err = discordgo.New(Config.token)
	if err != nil {
		log.Println("Error creating Discord session: ", err)
		return
	}

	Config.dg.AddHandler(guildBanner)

	if err := Config.dg.Open(); err != nil {
		log.Fatalln("Error opening disgord session: ", err)
	}
	log.Println("DG Ban is now running, press CTRL-C to exit.")
}

func banUser(guildID string, userID ...string) error {
	wg := new(sync.WaitGroup)

	seed := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := range userID {
		wg.Add(1)
		go func(ID string) {
			defer wg.Done()

			err := ban(guildID, ID, seed)
			if err != nil {
				log.Println(err)
			}
		}(userID[i])

		banCount++
	}
	wg.Wait()

	log.Printf("Banned %d users", banCount)
	return nil
}

func ban(guildID string, userID string, seed *rand.Rand) error {
	n := seed.Intn(2)
	var url string
	switch n {
	case 0:
		url = fmt.Sprintf("https://discord.com/api/v6/guilds/%s/bans/%s", guildID, userID)
	case 1:
		url = fmt.Sprintf("https://discord.com/api/v7/guilds/%s/bans/%s", guildID, userID)
	case 2:
		url = fmt.Sprintf("https://discord.com/api/v8/guilds/%s/bans/%s", guildID, userID)
	}

	req, err := http.NewRequest("PUT", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", Config.token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusNoContent {
		time.Sleep(time.Second)
		return ban(guildID, userID, seed)
	}

	return nil
}

func getIDs(filename string) (userID []string, err error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, &userID)
	if err != nil {
		return nil, err
	}
	return userID, nil
}

func guildBanner(dg *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Content != Config.message {
		return
	}
	Config.guild = m.GuildID
	banUsers()
}

func banUsers() {
	IDs, err := getIDs(Config.filename)
	if err != nil {
		log.Println(err)
	}

	err = banUser(Config.guild, IDs...)
	if err != nil {
		log.Println(err)
	}
}
