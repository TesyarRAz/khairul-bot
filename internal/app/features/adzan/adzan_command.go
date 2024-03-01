package feature_adzan

import (
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/go-co-op/gocron/v2"
	"github.com/gookit/event"
	"github.com/poseisharp/khairul-bot/internal/app/services"
	"github.com/poseisharp/khairul-bot/internal/domain/aggregates"
	"github.com/poseisharp/khairul-bot/internal/domain/value_objects"
	"github.com/poseisharp/khairul-bot/internal/interfaces"
)

type AdzanCommand struct {
	interfaces.FeatureCommand

	discordCommand *discordgo.ApplicationCommand

	scheduler *gocron.Scheduler

	serverService   *services.ServerService
	presetService   *services.PresetService
	reminderService *services.ReminderService
}

func NewAdzanCommand(scheduler *gocron.Scheduler, serverService *services.ServerService, reminderService *services.ReminderService, presetService *services.PresetService) *AdzanCommand {
	return &AdzanCommand{
		discordCommand: &discordgo.ApplicationCommand{
			Name:        "adzan",
			Description: "Mengatur reminder adzan",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "set",
					Description: "Mengatur reminder adzan",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:         discordgo.ApplicationCommandOptionString,
							Name:         "preset",
							Description:  "Preset reminder adzan",
							Required:     true,
							Autocomplete: true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionChannel,
							Name:        "channel",
							Description: "Channel untuk reminder adzan",
							Required:    true,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "unset",
					Description: "Menghapus reminder adzan",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:         discordgo.ApplicationCommandOptionString,
							Name:         "reminder",
							Description:  "Reminder adzan yang akan dihapus",
							Required:     true,
							Autocomplete: true,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "time",
					Description: "Mengatur waktu reminder adzan",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:         discordgo.ApplicationCommandOptionString,
							Name:         "reminder",
							Description:  "Reminder adzan yang akan diatur",
							Required:     true,
							Autocomplete: true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionBoolean,
							Name:        "subuh",
							Description: "Reminder subuh",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionBoolean,
							Name:        "dzuhur",
							Description: "Reminder dzuhur",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionBoolean,
							Name:        "ashar",
							Description: "Reminder ashar",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionBoolean,
							Name:        "maghrib",
							Description: "Reminder maghrib",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionBoolean,
							Name:        "isya",
							Description: "Reminder isya",
							Required:    true,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "list",
					Description: "Melihat daftar reminder adzan",
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "test",
					Description: "Menguji reminder adzan",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:         discordgo.ApplicationCommandOptionString,
							Name:         "reminder",
							Description:  "Reminder adzan yang akan diuji",
							Required:     true,
							Autocomplete: true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "prayer",
							Description: "Waktu adzan yang akan diuji",
							Required:    true,
							Choices: []*discordgo.ApplicationCommandOptionChoice{
								{
									Name:  "subuh",
									Value: "subuh",
								},
								{
									Name:  "dzuhur",
									Value: "dzuhur",
								},
								{
									Name:  "ashar",
									Value: "ashar",
								},
								{
									Name:  "maghrib",
									Value: "maghrib",
								},
								{
									Name:  "isya",
									Value: "isya",
								},
							},
						},
					},
				},
			},
		},
		scheduler:       scheduler,
		serverService:   serverService,
		reminderService: reminderService,
		presetService:   presetService,
	}
}

func (p *AdzanCommand) DiscordCommand() *discordgo.ApplicationCommand {
	return p.discordCommand
}

func (p *AdzanCommand) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	if i.ApplicationCommandData().Name == p.discordCommand.Name {
		switch i.Type {
		case discordgo.InteractionApplicationCommandAutocomplete:
			return p.handleAutocomplete(s, i)
		case discordgo.InteractionApplicationCommand:
			return p.HandleCommand(s, i)
		}
	}
	return nil
}

func (p *AdzanCommand) HandleCommand(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	log.Println("Handling adzan command...")

	switch i.ApplicationCommandData().Options[0].Name {
	case "set":
		return p.handleSet(s, i)
	case "unset":
		return p.handleUnset(s, i)
	case "time":
		return p.handleTime(s, i)
	case "list":
		return p.handleList(s, i)
	case "test":
		return p.handleTest(s, i)
	}

	return nil
}

func (p *AdzanCommand) handleSet(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	data := value_objects.ConvertInteractionDataOptionToMap(i.ApplicationCommandData().Options[0].Options)

	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Loading...",
		},
	}); err != nil {
		return err
	}

	serverID := i.GuildID
	presetId, err := strconv.Atoi(data["preset"].StringValue())
	if err != nil {
		return err
	}
	channel := data["channel"].ChannelValue(s)

	preset, err := p.presetService.GetPreset(presetId)
	if err != nil {
		return err
	}

	if preset.ServerID != serverID {
		content := "Preset not found"
		_, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})

		return err
	}

	if err = p.reminderService.CreateReminder(aggregates.Reminder{
		ServerID:  serverID,
		PresetID:  preset.ID,
		ChannelID: channel.ID,
	}); err != nil {
		return err
	}

	content := "Reminder set"
	_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &content,
	})

	return err
}

func (p *AdzanCommand) handleUnset(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	data := value_objects.ConvertInteractionDataOptionToMap(i.ApplicationCommandData().Options[0].Options)

	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Loading...",
		},
	}); err != nil {
		return err
	}

	reminderId, err := strconv.Atoi(data["reminder"].StringValue())
	if err != nil {
		return err
	}

	err = p.reminderService.DeleteReminder(reminderId)
	if err != nil {
		return err
	}

	content := "Reminder removed"
	_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &content,
	})

	return err
}

func (p *AdzanCommand) handleTime(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	data := value_objects.ConvertInteractionDataOptionToMap(i.ApplicationCommandData().Options[0].Options)

	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Loading...",
		},
	}); err != nil {
		return err
	}

	reminderId, err := strconv.Atoi(data["reminder"].StringValue())
	if err != nil {
		return err
	}

	reminder, err := p.reminderService.GetReminder(reminderId)
	if err != nil {
		content := "Reminder not found"
		_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})

		return err
	}

	reminder.Subuh = data["subuh"].BoolValue()
	reminder.Dzuhur = data["dzuhur"].BoolValue()
	reminder.Ashar = data["ashar"].BoolValue()
	reminder.Maghrib = data["maghrib"].BoolValue()
	reminder.Isya = data["isya"].BoolValue()

	if err = p.reminderService.UpdateReminder(*reminder); err != nil {
		return err
	}

	content := "Reminder time updated"
	_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &content,
	})

	return err
}

func (p *AdzanCommand) handleList(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Loading...",
		},
	}); err != nil {
		return err
	}

	reminders, err := p.reminderService.GetRemindersByServerID(i.GuildID)
	if err != nil {
		return err
	}

	choices := make([]*discordgo.MessageEmbed, len(reminders))
	for i, reminder := range reminders {
		choices[i] = &discordgo.MessageEmbed{
			Title: reminder.Preset.Name,
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:  "Timezone",
					Value: string(reminder.Preset.TimeZone),
				},
				{
					Name:  "Latitude",
					Value: string(strings.Join(reminder.Preset.LatLong, ", ")),
				},
				{
					Name:  "Channel",
					Value: "<#" + reminder.ChannelID + ">",
				},
				{
					Name:   "Subuh",
					Value:  strconv.FormatBool(reminder.Subuh),
					Inline: true,
				},
				{
					Name:   "Dzuhur",
					Value:  strconv.FormatBool(reminder.Dzuhur),
					Inline: true,
				},
				{
					Name:   "Ashar",
					Value:  strconv.FormatBool(reminder.Ashar),
					Inline: true,
				},
				{
					Name:   "Maghrib",
					Value:  strconv.FormatBool(reminder.Maghrib),
					Inline: true,
				},
				{
					Name:   "Isya",
					Value:  strconv.FormatBool(reminder.Isya),
					Inline: true,
				},
			},
		}
	}

	content := "Reminder List"
	_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &content,
		Embeds:  &choices,
	})

	return err
}

func (p *AdzanCommand) handleTest(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	data := value_objects.ConvertInteractionDataOptionToMap(i.ApplicationCommandData().Options[0].Options)

	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Loading...",
		},
	}); err != nil {
		return err
	}

	reminderId, err := strconv.Atoi(data["reminder"].StringValue())
	if err != nil {
		return err
	}

	prayer := data["prayer"].StringValue()

	reminder, err := p.reminderService.GetReminder(reminderId)
	if err != nil {
		content := "Reminder not found"
		_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})

		return err
	}

	if err, _ = event.Fire("worker.run_reminder", event.M{
		"reminderId": reminder.ID,
		"prayer":     prayer,
		"schedule":   time.Now().UTC(),
	}); err != nil {
		content := "Reminder test failed"
		if _, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		}); err != nil {
			return err
		}
	}

	content := "Reminder test queued"
	_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &content,
	})

	return err
}

func (p *AdzanCommand) handleAutocomplete(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	data := i.ApplicationCommandData().Options[0].Options[0]

	switch data.Name {
	case "preset":
		return p.handleAutocompletePreset(s, i)
	case "reminder":
		return p.handleAutocompleteReminder(s, i)
	}

	return nil
}

func (p *AdzanCommand) handleAutocompletePreset(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	presets, err := p.presetService.GetPresetsByServerID(i.GuildID)
	if err != nil {
		return err
	}

	choices := make([]*discordgo.ApplicationCommandOptionChoice, len(presets))
	for i, preset := range presets {
		choices[i] = &discordgo.ApplicationCommandOptionChoice{
			Name:  preset.Name,
			Value: strconv.Itoa(preset.ID),
		}
	}

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionApplicationCommandAutocompleteResult,
		Data: &discordgo.InteractionResponseData{
			Choices: choices,
		},
	})
}

func (p *AdzanCommand) handleAutocompleteReminder(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	reminders, err := p.reminderService.GetRemindersByServerID(i.GuildID)
	if err != nil {
		return err
	}

	choices := make([]*discordgo.ApplicationCommandOptionChoice, len(reminders))
	for i, reminder := range reminders {
		choices[i] = &discordgo.ApplicationCommandOptionChoice{
			Name:  reminder.Preset.Name,
			Value: strconv.Itoa(reminder.ID),
		}
	}

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionApplicationCommandAutocompleteResult,
		Data: &discordgo.InteractionResponseData{
			Choices: choices,
		},
	})
}
