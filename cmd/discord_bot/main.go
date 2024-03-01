package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
	"github.com/go-co-op/gocron/v2"
	"github.com/gookit/event"
	"github.com/joho/godotenv"
	"github.com/life4/genesis/slices"
	"github.com/poseisharp/khairul-bot/internal/app/features"
	feature_adzan "github.com/poseisharp/khairul-bot/internal/app/features/adzan"
	feature_jadwal "github.com/poseisharp/khairul-bot/internal/app/features/jadwal"
	feature_reminder_worker "github.com/poseisharp/khairul-bot/internal/app/features/reminder_worker"
	"github.com/poseisharp/khairul-bot/internal/app/services"
	"github.com/poseisharp/khairul-bot/internal/domain/aggregates"
	interface_features "github.com/poseisharp/khairul-bot/internal/interfaces"
	"github.com/poseisharp/khairul-bot/internal/persistent/repositories"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	s         *discordgo.Session
	GuildID   string = ""
	commands  []interface_features.FeatureCommand
	scheduler gocron.Scheduler

	reminderHandler *feature_reminder_worker.ReminderWorkerHandler
)

func init() {
	var err error
	if err = godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	if s, err = discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN")); err != nil {
		log.Fatal("Error creating discord session")
	}

	db, err := initDb()
	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}

	scheduler, err = gocron.NewScheduler()
	if err != nil {
		log.Fatalf("Error initializing scheduler: %v", err)
	}

	serverRepository := repositories.NewServerRepository(db)
	reminderRepository := repositories.NewReminderRepository(db)
	presetRepository := repositories.NewPresetRepository(db)

	serverService := services.NewServerService(serverRepository)
	prayerService := services.NewPrayerService()
	reminderService := services.NewReminderService(reminderRepository)
	presetService := services.NewPresetService(presetRepository)

	discordService := services.NewDiscordService(s)

	reminderHandler = feature_reminder_worker.NewReminderWorkerHandler(&scheduler, reminderService, prayerService, discordService)

	event.On("worker.run_reminder", event.ListenerFunc(reminderHandler.RunReminder))

	commands = []interface_features.FeatureCommand{
		features.NewPingCommand(),
		feature_jadwal.NewJadwalPresetCommand(serverService, presetService),
		feature_jadwal.NewJadwalCommand(prayerService, serverService, presetService),
		feature_jadwal.NewJadwalManualCommand(prayerService),

		feature_adzan.NewAdzanCommand(&scheduler, serverService, reminderService, presetService),
	}

	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		for _, command := range commands {
			if err := command.Handle(s, i); err != nil {
				log.Printf("Error handling '%v' command: %v", i.ApplicationCommandData().Name, err)
			}
		}
	})

	s.AddHandler(func(s *discordgo.Session, g *discordgo.GuildCreate) {
		serverService.CreateServerIfNotExists(aggregates.Server{
			ID: g.ID,
		})
	})

	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})
}

func main() {
	scheduler.NewJob(gocron.DailyJob(1, gocron.NewAtTimes(gocron.NewAtTime(0, 0, 0))), gocron.NewTask(reminderHandler.SetupReminder))

	scheduler.Start()
	defer func() {
		if err := scheduler.Shutdown(); err != nil {
			log.Println("Error shutting down scheduler:", err)
		}
	}()

	defer event.CloseWait()

	if err := s.Open(); err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}

	log.Println("Adding commands...")
	registeredCommands, err := s.ApplicationCommandBulkOverwrite(s.State.User.ID, GuildID, slices.MapAsync(commands, 0, func(v interface_features.FeatureCommand) *discordgo.ApplicationCommand {
		return v.DiscordCommand()
	}))

	if err != nil {
		log.Fatalf("Cannot add commands: %v", err)
	}

	defer s.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("Press Ctrl+C to exit")
	<-stop

	log.Println("Removing commands...")
	for _, v := range registeredCommands {
		if err := s.ApplicationCommandDelete(s.State.User.ID, GuildID, v.ID); err != nil {
			log.Panicf("Cannot delete '%v' command: %v", v.Name, err)
		}
	}

	log.Println("Gracefully shutting down.")
}

func initDb() (db *gorm.DB, err error) {
	db, err = gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		return
	}

	if err := db.AutoMigrate(&aggregates.Server{}); err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&aggregates.Preset{}); err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&aggregates.Reminder{}); err != nil {
		return nil, err
	}

	return
}
