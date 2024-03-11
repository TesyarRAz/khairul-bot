package services

import (
	"time"

	"github.com/hablullah/go-prayer"
	"github.com/poseisharp/khairul-bot/internal/domain/value_objects"
)

type PrayerService struct {
}

func NewPrayerService() *PrayerService {
	return &PrayerService{}
}

func (s *PrayerService) Calculate(timezone value_objects.TimeZone, latLong value_objects.LatLong) []prayer.Schedule {
	schedule, _ := prayer.Calculate(prayer.Config{
		Latitude:           latLong.Latitude(),
		Longitude:          latLong.Longitude(),
		Timezone:           timezone.LoadLocation(),
		TwilightConvention: prayer.Kemenag(),
		AsrConvention:      prayer.Shafii,
		PreciseToSeconds:   true,
		Corrections:        prayer.ScheduleCorrections{},
	}, time.Now().UTC().Year())

	return schedule
}
