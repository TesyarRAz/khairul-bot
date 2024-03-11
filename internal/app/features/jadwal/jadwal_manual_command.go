package feature_jadwal

import (
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/poseisharp/khairul-bot/internal/app/services"
	"github.com/poseisharp/khairul-bot/internal/domain/value_objects"
	"github.com/poseisharp/khairul-bot/internal/interfaces"
)

type JadwalManualCommand struct {
	interfaces.FeatureCommand

	discordCommand *discordgo.ApplicationCommand

	prayerService *services.PrayerService
}

func NewJadwalManualCommand(prayerService *services.PrayerService) *JadwalManualCommand {
	return &JadwalManualCommand{
		discordCommand: &discordgo.ApplicationCommand{
			Name:        "jadwal-manual",
			Description: "Informasi tentang jadwal",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "timezone",
					Description: "Your timezone",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "lat_long",
					Description: "Your lat long",
					Required:    true,
				},
			},
		},
		prayerService: prayerService,
	}
}

func (p *JadwalManualCommand) DiscordCommand() *discordgo.ApplicationCommand {
	return p.discordCommand
}

func (p *JadwalManualCommand) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	if i.ApplicationCommandData().Name == p.discordCommand.Name {
		if i.Type == discordgo.InteractionApplicationCommand {
			return p.HandleCommand(s, i)
		}
	}

	return nil
}

func (p *JadwalManualCommand) HandleCommand(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	log.Println("Handling jadwal command...")

	optionMap := value_objects.ConvertInteractionDataOptionToMap(i.ApplicationCommandData().Options)

	latLong := value_objects.LatLong(strings.Split(optionMap["lat_long"].StringValue(), ","))
	timezone := optionMap["timezone"].StringValue()

	schedule := p.prayerService.Calculate(value_objects.TimeZone(timezone), latLong)
	index := time.Now().UTC().Day() - 1

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Jadwal Sholat",
			Embeds: []*discordgo.MessageEmbed{
				{
					Title: "Latitude & Longitude",
					Fields: []*discordgo.MessageEmbedField{
						{
							Name:   "Latitude",
							Value:  strconv.FormatFloat(latLong.Latitude(), 'f', 6, 32),
							Inline: true,
						},
						{
							Name:   "Longitude",
							Value:  strconv.FormatFloat(latLong.Longitude(), 'f', 6, 32),
							Inline: true,
						},
					},
				},
				{
					Title: "Jadwal Sholat",
					Fields: []*discordgo.MessageEmbedField{
						{
							Name:  "Subuh",
							Value: schedule[index].Fajr.Format("15:04 MST"),
						},
						{
							Name:  "Dzuhur",
							Value: schedule[index].Zuhr.Format("15:04 MST"),
						},
						{
							Name:  "Ashar",
							Value: schedule[index].Asr.Format("15:04 MST"),
						},
						{
							Name:  "Maghrib",
							Value: schedule[index-1].Maghrib.Format("15:04 MST"),
						},
						{
							Name:  "Isya",
							Value: schedule[index-1].Isha.Format("15:04 MST"),
						},
					},
				},
			},
		},
	})
}
