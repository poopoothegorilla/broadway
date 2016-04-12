package services

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/namely/broadway/deployment"
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

// PlaybookNotFound indicates a problem due to Broadway not knowing about a
// playbook
type PlaybookNotFound struct {
	playbookID string
}

func (e *PlaybookNotFound) Error() string {
	return fmt.Sprintf("Can't make instance because playbook %s is missing\n", e.playbookID)
}

// InvalidVar indicates a problem setting or updating an instance var that is not declared in that instance's playbook
type InvalidVar struct {
	playbookID string
	key        string
}

func (e *InvalidVar) Error() string {
	return fmt.Sprintf("Playbook %s does not declare a var named %s\n", e.playbookID, e.key)
}

// Create a new instance
func (is *InstanceService) Create(i *instance.Instance) (*instance.Instance, error) {
	glog.Infof("InstanceService attempting to create %s/%s...\n", i.PlaybookID, i.ID)
	pb, ok := deployment.AllPlaybooks[i.PlaybookID]
	if !ok {
		return nil, &PlaybookNotFound{i.PlaybookID}
	}
	// Set all vars declared in playbook to default empty string
	vars := make(map[string]string)
	for _, pv := range pb.Vars {
		vars[pv] = ""
	}
	// Abort if new instance tries to set vars not declared in playbook
	for k, v := range i.Vars {
		_, valid := vars[k]
		if !valid { // k is not listed in playbook
			return nil, &InvalidVar{i.PlaybookID, k}
		}
		vars[k] = v
	}

	i.Vars = vars
	err := is.repo.Save(i)
	if err != nil {
		return nil, err
	}
	err = sendCreationNotification(i)
	if err != nil {
		return nil, err
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

// Delete remove an instance
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
