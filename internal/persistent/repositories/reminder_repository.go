package repositories

import (
	"github.com/poseisharp/khairul-bot/internal/domain/aggregates"
	"gorm.io/gorm"
)

type ReminderRepository struct {
	db *gorm.DB
}

func NewReminderRepository(db *gorm.DB) *ReminderRepository {
	return &ReminderRepository{
		db: db,
	}
}

func (r *ReminderRepository) Delete(id int) error {
	return r.db.Delete(&aggregates.Reminder{}, id).Error
}

func (r *ReminderRepository) FindAll() ([]aggregates.Reminder, error) {
	var reminders []aggregates.Reminder

	if err := r.db.Preload("Preset").Find(&reminders).Error; err != nil {
		return nil, err
	}

	return reminders, nil
}

func (r *ReminderRepository) FindOne(id int) (*aggregates.Reminder, error) {
	var reminder aggregates.Reminder

	if err := r.db.Preload("Preset").First(&reminder, id).Error; err != nil {
		return nil, err
	}

	return &reminder, nil
}

func (r *ReminderRepository) Store(reminder aggregates.Reminder) error {
	return r.db.Create(&reminder).Error
}

func (r *ReminderRepository) Update(reminder aggregates.Reminder) error {
	return r.db.Save(&reminder).Error
}

func (r *ReminderRepository) FindByServerID(serverID string) ([]aggregates.Reminder, error) {
	var reminders []aggregates.Reminder

	if err := r.db.Preload("Preset").Where("server_id = ?", serverID).Find(&reminders).Error; err != nil {
		return nil, err
	}

	return reminders, nil
}
