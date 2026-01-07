package metrics

import (
	"app/pkg/models"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Repository struct {
	db     *gorm.DB
	logger *logrus.Logger
}

func NewRepository(db *gorm.DB, logger *logrus.Logger) *Repository {
	return &Repository{db: db, logger: logger}
}

func (r *Repository) ensureRow() error {
	var count int64
	if err := r.db.Model(&models.AdClickStats{}).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		row := &models.AdClickStats{ID: 1, Clicks1: 0, Clicks2: 0}
		if err := r.db.Create(row).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) Increment(slot int) error {
	if err := r.ensureRow(); err != nil {
		r.logger.Errorf("Metrics.Increment: ensure row: %v", err)
		return err
	}
	var expr string
	if slot == 2 {
		expr = "clicks2 = clicks2 + 1"
	} else {
		expr = "clicks1 = clicks1 + 1"
	}
	if err := r.db.Exec("UPDATE ad_click_stats SET " + expr + ", updated_at = NOW() WHERE id = 1").Error; err != nil {
		r.logger.Errorf("Metrics.Increment: update failed: %v", err)
		return err
	}
	return nil
}

func (r *Repository) GetTotals() (int64, int64, error) {
	if err := r.ensureRow(); err != nil {
		return 0, 0, err
	}
	var row models.AdClickStats
	if err := r.db.First(&row, 1).Error; err != nil {
		return 0, 0, err
	}
	return row.Clicks1, row.Clicks2, nil
}
