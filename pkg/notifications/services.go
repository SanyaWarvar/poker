package notifications

import "github.com/google/uuid"

type INotificationService interface {
	MarkReaded(notificationId, userId uuid.UUID) error
	CreateNotification(item Notification) error
	GetNotReadedNotifiesByUserId(userId uuid.UUID) ([]Notification, error)
}

type NotificationService struct {
	repo INotificationRepository
}

func NewNotificationService(repo INotificationRepository) *NotificationService {
	return &NotificationService{repo: repo}
}

func (s *NotificationService) MarkReaded(notificationId, userId uuid.UUID) error {
	return s.repo.MarkReaded(notificationId, userId)
}

func (s *NotificationService) CreateNotification(item Notification) error {
	return s.repo.CreateNotification(item)
}

func (s *NotificationService) GetNotReadedNotifiesByUserId(userId uuid.UUID) ([]Notification, error) {
	return s.repo.GetNotReadedNotifiesByUserId(userId)
}
