package feature_reminder_worker

import (
	"log"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/gookit/event"
	"github.com/poseisharp/khairul-bot/internal/app/services"
)

type ReminderWorkerHandler struct {
	scheduler *gocron.Scheduler

	reminderService *services.ReminderService
	prayerService   *services.PrayerService
	discordService  *services.DiscordService
}

func NewReminderWorkerHandler(scheduler *gocron.Scheduler, reminderService *services.ReminderService, prayerService *services.PrayerService, discordService *services.DiscordService) *ReminderWorkerHandler {
	return &ReminderWorkerHandler{
		scheduler:       scheduler,
		reminderService: reminderService,
		prayerService:   prayerService,
		discordService:  discordService,
	}
}

func (h *ReminderWorkerHandler) SetupReminder() error {
	reminders, err := h.reminderService.GetReminders()
	if err != nil {
		return err
	}

	for _, reminder := range reminders {
		now := time.Now().UTC()

		schedules := h.prayerService.Calculate(reminder.Preset.TimeZone, reminder.Preset.LatLong)
		dayOfYear := now.YearDay() - 1

		var (
			reminderId int
			prayer     string
			schedule   time.Time
		)

		if reminder.Subuh {
			reminderId = reminder.ID
			prayer = "subuh"
			schedule = schedules[dayOfYear].Fajr
		}
		if reminder.Dzuhur {
			reminderId = reminder.ID
			prayer = "dzuhur"
			schedule = schedules[dayOfYear].Zuhr
		}
		if reminder.Ashar {
			reminderId = reminder.ID
			prayer = "ashar"
			schedule = schedules[dayOfYear].Asr
		}
		if reminder.Maghrib {
			reminderId = reminder.ID
			prayer = "maghrib"
			schedule = schedules[dayOfYear].Maghrib
		}
		if reminder.Isya {
			reminderId = reminder.ID
			prayer = "isya"
			schedule = schedules[dayOfYear].Isha
		}

		(*h.scheduler).NewJob(gocron.DurationJob(time.Until(schedule)), gocron.NewTask(h.internalRunReminder, reminderId, prayer, schedule))
	}

	return nil
}

func (h *ReminderWorkerHandler) internalRunReminder(reminderId int, prayer string, schedule time.Time) error {
	log.Println("Handle run reminder")

	reminder, err := h.reminderService.GetReminder(int(reminderId))
	if err != nil {
		log.Println(err)
		return err
	}

	err = h.discordService.SendTextMessage(reminder.ChannelID, "Waktunya sholat "+prayer+" ("+schedule.Format("15:04 MST")+")")

	if err != nil {
		log.Println(err)
	}

	return err
}

func (h *ReminderWorkerHandler) RunReminder(e event.Event) error {
	reminderId := e.Get("reminderId").(int)
	prayer := e.Get("prayer").(string)
	schedule := e.Get("schedule").(time.Time)

	return h.internalRunReminder(reminderId, prayer, schedule)

}
