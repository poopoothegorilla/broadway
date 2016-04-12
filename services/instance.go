package services

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/namely/broadway/instance"
	"github.com/namely/broadway/notification"
	"github.com/namely/broadway/store"
)

// InstanceService definition
type InstanceService struct {
	repo instance.Repository
}

// NewInstanceService creates a new instance service
func NewInstanceService(s store.Store) *InstanceService {
	return &InstanceService{repo: instance.NewRepo(s)}
}

// Create a new instance
func (is *InstanceService) Create(i *instance.Instance) (*instance.Instance, error) {
	glog.Info("Instance Service: Create")
	err := is.repo.Save(i)
	if err != nil {
		return nil, err
	}
	err = sendCreationNotification(i)
	if err != nil {
		return i, err
	}
	return i, nil
}

// Update an instance
func (is *InstanceService) Update(i *instance.Instance) (*instance.Instance, error) {
	glog.Info("Instance Service: Update")
	err := is.repo.Save(i)
	if err != nil {
		return nil, err
	}
	return i, nil
}

// Show takes playbookID and instanceID and returns the matching Instance, if
// any
func (is *InstanceService) Show(playbookID, ID string) (*instance.Instance, error) {
	instance, err := is.repo.FindByID(playbookID, ID)
	if err != nil {
		return instance, err
	}
	return instance, nil
}

// AllWithPlaybookID returns all the instances for an specified playbook id
func (is *InstanceService) AllWithPlaybookID(playbookID string) ([]*instance.Instance, error) {
	return is.repo.FindByPlaybookID(playbookID)
}

func (is *InstanceService) Delete(i *instance.Instance) error {
	ii, err := is.Show(i.PlaybookID, i.ID)
	if err != nil {
		return err
	}

	return is.repo.Delete(ii)
}

func sendCreationNotification(i *instance.Instance) error {
	m := &notification.Message{
		Attachments: []notification.Attachment{
			{
				Text: fmt.Sprintf("New broadway instance was created: %s %s.", i.PlaybookID, i.ID),
			},
		},
	}

	return m.Send()
}
