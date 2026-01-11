package main

import (
	"flag"
	"log"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file if it exists
	_ = godotenv.Load()

	guildID := flag.String("guild", "", "Guild ID to register commands to (leave empty for global)")
	remove := flag.Bool("remove", false, "Remove all commands instead of registering")
	flag.Parse()

	token := os.Getenv("DISCORD_BOT_TOKEN")
	if token == "" {
		log.Fatal("DISCORD_BOT_TOKEN must be set")
	}

	appID := os.Getenv("DISCORD_APP_ID")
	if appID == "" {
		log.Fatal("DISCORD_APP_ID must be set")
	}

	s, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}

	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "summarize",
			Description: "指定された日のリンクをまとめます",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "date",
					Description: "日付 (MMDD または YYYYMMDD)",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Name:        "with_title",
					Description: "urlにタイトルをつけるかどうか",
					Required:    false,
				},
			},
		},
		{
			Name:        "list",
			Description: "チャンネルごとのTODOリストを管理します",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "create",
					Description: "TODOリストを作成します",
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "add",
					Description: "TODOを追加します",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "content",
							Description: "追加するタスクの内容",
							Required:    true,
						},
					},
				},
			},
		},
	}

	if *remove {
		log.Println("Removing commands...")
		registeredCommands, err := s.ApplicationCommands(appID, *guildID)
		if err != nil {
			log.Fatalf("Could not fetch registered commands: %v", err)
		}
		for _, v := range registeredCommands {
			err := s.ApplicationCommandDelete(appID, *guildID, v.ID)
			if err != nil {
				log.Panicf("Cannot delete '%v' command: %v", v.Name, err)
			}
		}
		log.Println("Successfully removed all commands.")
		return
	}

	log.Println("Registering commands...")
	for _, v := range commands {
		_, err := s.ApplicationCommandCreate(appID, *guildID, v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
		log.Printf("Command '%v' registered!", v.Name)
	}
	log.Println("Done!")
}
