package label_repository

import (
	"context"
	"time"

	"github.com/evolution-foundation/evolution-go/pkg/config"
	label_model "github.com/evolution-foundation/evolution-go/pkg/label/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type LabelRepository interface {
	InsertLabel(label label_model.Label) error
	UpdateLabel(label label_model.Label) error
	GetLabelByID(id string) (*label_model.Label, error)
	DeleteLabel(id string) error
	GetAllLabelsByInstanceID(instanceID string) ([]label_model.Label, error)
	GetAllLabelsByInstanceIDContext(ctx context.Context, instanceID string) ([]label_model.Label, error)
	UpsertLabel(label label_model.Label) error
}

type labelRepository struct {
	db           *gorm.DB
	queryTimeout time.Duration
}

func (l *labelRepository) dbWithTimeout(ctx context.Context) (*gorm.DB, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}
	if l.queryTimeout <= 0 {
		return l.db.WithContext(ctx), func() {}
	}
	ctx, cancel := context.WithTimeout(ctx, l.queryTimeout)
	return l.db.WithContext(ctx), cancel
}

func (l *labelRepository) InsertLabel(label label_model.Label) error {
	db, cancel := l.dbWithTimeout(context.Background())
	defer cancel()
	return db.Create(&label).Error
}

func (l *labelRepository) UpdateLabel(label label_model.Label) error {
	db, cancel := l.dbWithTimeout(context.Background())
	defer cancel()
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"label_id", "label_name", "label_color"}),
	}).Create(&label).Error
}

func (l *labelRepository) GetLabelByID(id string) (*label_model.Label, error) {
	db, cancel := l.dbWithTimeout(context.Background())
	defer cancel()

	var label label_model.Label
	err := db.Where("id = ?", id).First(&label).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &label, nil
}

func (l *labelRepository) DeleteLabel(id string) error {
	db, cancel := l.dbWithTimeout(context.Background())
	defer cancel()
	return db.Where("id = ?", id).Delete(&label_model.Label{}).Error
}

func (l *labelRepository) GetAllLabelsByInstanceID(instanceID string) ([]label_model.Label, error) {
	return l.GetAllLabelsByInstanceIDContext(context.Background(), instanceID)
}

func (l *labelRepository) GetAllLabelsByInstanceIDContext(ctx context.Context, instanceID string) ([]label_model.Label, error) {
	db, cancel := l.dbWithTimeout(ctx)
	defer cancel()

	var labels []label_model.Label
	err := db.Where("instance_id = ?", instanceID).Find(&labels).Error
	if err != nil {
		return nil, err
	}
	return labels, nil
}

func (l *labelRepository) UpsertLabel(label label_model.Label) error {
	db, cancel := l.dbWithTimeout(context.Background())
	defer cancel()

	return db.Where("instance_id = ? AND label_id = ?",
		label.InstanceID,
		label.LabelID,
	).Assign(label_model.Label{
		LabelName:    label.LabelName,
		LabelColor:   label.LabelColor,
		PredefinedId: label.PredefinedId,
	}).FirstOrCreate(&label).Error
}

func NewLabelRepository(db *gorm.DB) LabelRepository {
	return NewLabelRepositoryWithTimeout(db, config.DefaultDBQueryTimeout)
}

func NewLabelRepositoryWithTimeout(db *gorm.DB, queryTimeout time.Duration) LabelRepository {
	return &labelRepository{db: db, queryTimeout: queryTimeout}
}
