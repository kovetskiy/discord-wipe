package main

import (
	"os"
	"strings"
	"time"

	"github.com/reconquest/karma-go"
	"github.com/reconquest/pkg/log"

	"github.com/bwmarrin/discordgo"
)

type App struct {
	ds *discordgo.Session
}

func main() {
	discordToken := os.Getenv("DISCORD_TOKEN")
	if discordToken == "" {
		log.Fatal("DISCORD_TOKEN is not specified")
	}

	ds, err := discordgo.New("Bot " + discordToken)
	if err != nil {
		log.Fatalf(err, "discord session")
	}
	defer ds.Close()

	err = ds.Open()
	if err != nil {
		log.Fatalf(err, "discord open")
	}

	app := &App{
		ds: ds,
	}

	for {
		err = app.Wipe()
		if err != nil {
			log.Error(err)
		}

		time.Sleep(time.Minute)
	}
}

func (app *App) Wipe() error {
	guilds, err := app.ds.UserGuilds(0, "", "")
	if err != nil {
		return karma.Format(err, "list guilds")
	}

	for _, guild := range guilds {
		channels, err := app.ds.GuildChannels(guild.ID)
		if err != nil {
			return karma.Format(err, "list channels: %s", guild.ID)
		}

		for _, channel := range channels {
			if !strings.HasSuffix(channel.Name, "-24h") {
				continue
			}

			limit := 10
			beforeID := ""
			for {
				log.Infof(
					karma.
						Describe("channel_id", channel.ID).
						Describe("after_id", beforeID),
					"list messages: %s / %s",
					guild.Name,
					channel.Name,
				)

				messages, err := app.ds.ChannelMessages(
					channel.ID,
					limit,
					beforeID,
					"",
					"",
				)
				if err != nil {
					return karma.Format(err, "list messages: %s", channel.ID)
				}

				bulkDelete := []string{}
				for _, message := range messages {
					timestamp, err := message.Timestamp.Parse()
					if err != nil {
						return karma.Format(
							err,
							"parse timestamp: %s",
							message.Timestamp,
						)
					}

					if time.Since(timestamp) > time.Minute {
						bulkDelete = append(bulkDelete, message.ID)
					}
				}

				if len(bulkDelete) > 0 {
					log.Infof(nil, "bulk delete: %d messages", len(bulkDelete))
					err = app.ds.ChannelMessagesBulkDelete(
						channel.ID,
						bulkDelete,
					)
					if err != nil {
						return karma.Format(
							err,
							"messages bulk delete: %q",
							bulkDelete,
						)
					}
				}

				if len(messages) == 0 {
					break
				}

				beforeID = messages[0].ID
			}
		}
	}

	return nil
}
