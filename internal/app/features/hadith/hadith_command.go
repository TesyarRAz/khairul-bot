package hadith

import (
	"github.com/bwmarrin/discordgo"
	"github.com/poseisharp/khairul-bot/internal/interfaces"
)

type HadithCommand struct {
	interfaces.FeatureCommand

	discordCommand *discordgo.ApplicationCommand
}

func NewHadithCommand() *HadithCommand {
	return &HadithCommand{
		discordCommand: &discordgo.ApplicationCommand{
			Name:        "hadith",
			Description: "Hadith",
			Options:     []*discordgo.ApplicationCommandOption{},
		},
	}
}
