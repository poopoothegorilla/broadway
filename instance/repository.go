package instance

import (
	"encoding/json"
	"fmt"

	"github.com/namely/broadway/env"
	"github.com/namely/broadway/store"
)

// Repository interface
type Repository interface {
	Save(instance *Instance) error
	FindByPath(path string) (*Instance, error)
	FindByID(playbookID, ID string) (*Instance, error)
	FindByPlaybookID(playbookID string) ([]*Instance, error)
	Delete(*Instance) error
}

// Repo handles persistence logic
type Repo struct {
	store store.Store
}

// NotFound instance not found error
type NotFound struct {
	path string
}

func (e NotFound) Error() string {
	return fmt.Sprintf("Instance with path: %s was not found", e.path)
}

// MalformedSavedData malformed saved data
type MalformedSavedData struct{}

func (e MalformedSavedData) Error() string {
	return "Saved data for this instance is malformed"
}

// NewRepo constructor
func NewRepo(s store.Store) *Repo {
	return &Repo{store: s}
}

// Save a new instance as json
func (ir *Repo) Save(instance *Instance) error {
	encoded, err := instance.JSON()
	if err != nil {
		return err
	}
	err = ir.store.SetValue(instance.Path(), encoded)
	if err != nil {
		return err
	}
	return nil
}

// FindByPath find an instance based on it's path
func (ir *Repo) FindByPath(path string) (*Instance, error) {
	var instance Instance

	i := ir.store.Value(path)
	if i == "" {
		return nil, NotFound{path}
	}
	err := json.Unmarshal([]byte(i), &instance)
	if err != nil {
		return nil, MalformedSavedData{}
	}
	return &instance, nil
}

// FindByID finds an instance by playbook and instance ID
func (ir *Repo) FindByID(playbookID, ID string) (*Instance, error) {
	path := env.EtcdPath + "/instances/" + playbookID + "/" + ID
	return ir.FindByPath(path)
}

// FindByPlaybookID finds instances by playbook id
func (ir *Repo) FindByPlaybookID(playbookID string) ([]*Instance, error) {

	data := ir.store.Values(fmt.Sprintf(env.EtcdPath+"/instances/%s", playbookID))
	instances := []*Instance{}
	for _, value := range data {
		var instance Instance
		err := json.Unmarshal([]byte(value), &instance)
		if err != nil {
			return instances, MalformedSavedData{}
		}
		instances = append(instances, &instance)
	}

	return instances, nil
}

// Delete removes an instance
func (ir *Repo) Delete(i *Instance) error {
	return ir.store.Delete(i.Path())
}
