package instance_repository

import (
	"context"
	"fmt"
	"time"

	"github.com/evolution-foundation/evolution-go/pkg/config"
	instance_model "github.com/evolution-foundation/evolution-go/pkg/instance/model"
	"github.com/gomessguii/logger"
	"github.com/google/uuid"
	"gorm.io/gorm"

	label_model "github.com/evolution-foundation/evolution-go/pkg/label/model"
	label_repository "github.com/evolution-foundation/evolution-go/pkg/label/repository"

	message_model "github.com/evolution-foundation/evolution-go/pkg/message/model"
	message_repository "github.com/evolution-foundation/evolution-go/pkg/message/repository"
)

type InstanceRepository interface {
	Create(instance instance_model.Instance) (*instance_model.Instance, error)
	GetInstanceByID(instanceId string) (*instance_model.Instance, error)
	GetConnectedInstanceByID(instanceId string) (*instance_model.Instance, error)
	GetInstanceByToken(token string) (*instance_model.Instance, error)
	GetInstanceByTokenContext(ctx context.Context, token string) (*instance_model.Instance, error)
	GetInstanceByName(name string) (*instance_model.Instance, error)
	Update(*instance_model.Instance) error
	UpdateConnected(userId string, status bool, disconnectReason string) error
	UpdateQrcode(userId string, qr string) error
	UpdateProxy(userId string, proxy string) error
	UpdateJid(userId string, jid string) error
	GetAllConnectedInstances() ([]*instance_model.Instance, error)
	GetAllConnectedInstancesByClientName(clientName string) ([]*instance_model.Instance, error)
	GetAll(clientName string) ([]*instance_model.Instance, error)
	Delete(instanceId string) error
	GetAdvancedSettings(instanceId string) (*instance_model.AdvancedSettings, error)
	UpdateAdvancedSettings(instanceId string, settings *instance_model.AdvancedSettings) error
}

type instanceRepository struct {
	db           *gorm.DB
	queryTimeout time.Duration
	labelRepo    label_repository.LabelRepository
	messageRepo  message_repository.MessageRepository
}

func (i *instanceRepository) dbWithTimeout(ctx context.Context) (*gorm.DB, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}
	if i.queryTimeout <= 0 {
		return i.db.WithContext(ctx), func() {}
	}
	ctx, cancel := context.WithTimeout(ctx, i.queryTimeout)
	return i.db.WithContext(ctx), cancel
}

func (i *instanceRepository) Create(instance instance_model.Instance) (*instance_model.Instance, error) {
	db, cancel := i.dbWithTimeout(context.Background())
	defer cancel()

	if err := db.Create(&instance).Error; err != nil {
		return nil, err
	}
	return &instance, nil
}

func (i *instanceRepository) GetInstanceByToken(token string) (*instance_model.Instance, error) {
	return i.GetInstanceByTokenContext(context.Background(), token)
}

func (i *instanceRepository) GetInstanceByTokenContext(ctx context.Context, token string) (*instance_model.Instance, error) {
	db, cancel := i.dbWithTimeout(ctx)
	defer cancel()

	var instance instance_model.Instance
	err := db.Where("token = ?", token).First(&instance).Error
	if err != nil {
		return nil, err
	}

	return &instance, nil
}

func (i *instanceRepository) GetInstanceByName(name string) (*instance_model.Instance, error) {
	db, cancel := i.dbWithTimeout(context.Background())
	defer cancel()

	var instance instance_model.Instance
	err := db.Where("name = ?", name).First(&instance).Error
	if err != nil {
		return nil, err
	}

	return &instance, nil
}

func (i *instanceRepository) GetInstanceByID(instanceId string) (*instance_model.Instance, error) {
	// Valida o formato do UUID
	if _, err := uuid.Parse(instanceId); err != nil {
		return nil, fmt.Errorf("invalid UUID format: %v", err)
	}

	db, cancel := i.dbWithTimeout(context.Background())
	defer cancel()

	var instance instance_model.Instance
	err := db.Where("id = ?", instanceId).First(&instance).Error
	if err != nil {
		return nil, err
	}

	return &instance, nil
}

func (i *instanceRepository) GetConnectedInstanceByID(instanceId string) (*instance_model.Instance, error) {
	db, cancel := i.dbWithTimeout(context.Background())
	defer cancel()

	var instance instance_model.Instance
	err := db.Where("id = ? AND connected = ?", instanceId, true).First(&instance).Error
	if err != nil {
		return nil, err
	}

	return &instance, nil
}

func (i *instanceRepository) Update(instance *instance_model.Instance) error {
	db, cancel := i.dbWithTimeout(context.Background())
	defer cancel()

	err := db.Save(&instance).Error
	if err != nil {
		logger.LogError("Error updating instance in DB: %v", err)
	}
	return err
}

func (i *instanceRepository) UpdateConnected(userId string, status bool, disconnectReason string) error {
	db, cancel := i.dbWithTimeout(context.Background())
	defer cancel()
	return db.Model(&instance_model.Instance{}).Where("id = ?", userId).Update("connected", status).Update("disconnect_reason", disconnectReason).Error
}

func (i *instanceRepository) UpdateQrcode(userId string, qr string) error {
	db, cancel := i.dbWithTimeout(context.Background())
	defer cancel()
	return db.Model(&instance_model.Instance{}).Where("id = ?", userId).Update("qrcode", qr).Error
}

func (i *instanceRepository) UpdateProxy(userId string, proxy string) error {
	db, cancel := i.dbWithTimeout(context.Background())
	defer cancel()
	return db.Model(&instance_model.Instance{}).Where("id = ?", userId).Update("proxy", proxy).Error
}

func (i *instanceRepository) UpdateJid(userId string, jid string) error {
	db, cancel := i.dbWithTimeout(context.Background())
	defer cancel()
	return db.Model(&instance_model.Instance{}).Where("id = ?", userId).Update("jid", jid).Error
}

func (i *instanceRepository) GetAllConnectedInstances() ([]*instance_model.Instance, error) {
	db, cancel := i.dbWithTimeout(context.Background())
	defer cancel()

	var instances []*instance_model.Instance
	err := db.Where("connected = ?", true).Find(&instances).Error
	if err != nil {
		return nil, err
	}

	return instances, nil
}

func (i *instanceRepository) GetAllConnectedInstancesByClientName(clientName string) ([]*instance_model.Instance, error) {
	db, cancel := i.dbWithTimeout(context.Background())
	defer cancel()

	var instances []*instance_model.Instance
	err := db.Where("connected = ? AND client_name = ?", true).Where("client_name = ?", clientName).Find(&instances).Error
	if err != nil {
		return nil, err
	}

	return instances, nil
}

func (i *instanceRepository) GetAll(clientName string) ([]*instance_model.Instance, error) {
	db, cancel := i.dbWithTimeout(context.Background())
	defer cancel()

	var instances []*instance_model.Instance
	err := db.Where("client_name = ?", clientName).Find(&instances).Error
	if err != nil {
		return nil, err
	}

	return instances, nil
}

func (i *instanceRepository) Delete(instanceId string) error {
	db, cancel := i.dbWithTimeout(context.Background())
	defer cancel()

	return db.Transaction(func(tx *gorm.DB) error {
		// Deleta todas as labels associadas à instância
		if err := tx.Where("instance_id = ?", instanceId).Delete(&label_model.Label{}).Error; err != nil {
			return fmt.Errorf("erro ao deletar labels: %v", err)
		}

		// Deleta todas as mensagens associadas à instância
		if err := tx.Where("source = ?", instanceId).Delete(&message_model.Message{}).Error; err != nil {
			return fmt.Errorf("erro ao deletar mensagens: %v", err)
		}

		// Deleta a instância
		if err := tx.Where("id = ?", instanceId).Delete(&instance_model.Instance{}).Error; err != nil {
			return fmt.Errorf("erro ao deletar instância: %v", err)
		}

		return nil
	})
}

func (i *instanceRepository) GetAdvancedSettings(instanceId string) (*instance_model.AdvancedSettings, error) {
	// Valida o formato do UUID
	if _, err := uuid.Parse(instanceId); err != nil {
		return nil, fmt.Errorf("invalid UUID format: %v", err)
	}

	db, cancel := i.dbWithTimeout(context.Background())
	defer cancel()

	var instance instance_model.Instance
	err := db.Select("always_online, reject_call, msg_reject_call, read_messages, ignore_groups, ignore_status").
		Where("id = ?", instanceId).First(&instance).Error
	if err != nil {
		return nil, err
	}

	settings := &instance_model.AdvancedSettings{
		AlwaysOnline:  instance.AlwaysOnline,
		RejectCall:    instance.RejectCall,
		MsgRejectCall: instance.MsgRejectCall,
		ReadMessages:  instance.ReadMessages,
		IgnoreGroups:  instance.IgnoreGroups,
		IgnoreStatus:  instance.IgnoreStatus,
	}

	return settings, nil
}

func (i *instanceRepository) UpdateAdvancedSettings(instanceId string, settings *instance_model.AdvancedSettings) error {
	// Valida o formato do UUID
	if _, err := uuid.Parse(instanceId); err != nil {
		return fmt.Errorf("invalid UUID format: %v", err)
	}

	db, cancel := i.dbWithTimeout(context.Background())
	defer cancel()

	updates := map[string]interface{}{
		"always_online":   settings.AlwaysOnline,
		"reject_call":     settings.RejectCall,
		"msg_reject_call": settings.MsgRejectCall,
		"read_messages":   settings.ReadMessages,
		"ignore_groups":   settings.IgnoreGroups,
		"ignore_status":   settings.IgnoreStatus,
	}

	err := db.Model(&instance_model.Instance{}).Where("id = ?", instanceId).Updates(updates).Error
	if err != nil {
		logger.LogError("Error updating advanced settings in DB: %v", err)
		return err
	}

	return nil
}

func NewInstanceRepository(db *gorm.DB) InstanceRepository {
	return NewInstanceRepositoryWithTimeout(db, config.DefaultDBQueryTimeout)
}

func NewInstanceRepositoryWithTimeout(db *gorm.DB, queryTimeout time.Duration) InstanceRepository {
	return &instanceRepository{
		db:           db,
		queryTimeout: queryTimeout,
	}
}
