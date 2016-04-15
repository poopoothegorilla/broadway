package services

import (
	"errors"
	"fmt"

	"github.com/golang/glog"
	"github.com/namely/broadway/deployment"
	"github.com/namely/broadway/instance"
	"github.com/namely/broadway/notification"
	"github.com/namely/broadway/store"
)

// DeploymentService implements the Broadway logic for deployments
type DeploymentService struct {
	repo      instance.Repository
	playbooks map[string]*deployment.Playbook
	manifests map[string]*deployment.Manifest
}

// NewDeploymentService creates a new DeploymentService
func NewDeploymentService(s store.Store, ps map[string]*deployment.Playbook, ms map[string]*deployment.Manifest) *DeploymentService {
	return &DeploymentService{
		repo:      instance.NewRepo(s),
		playbooks: ps,
		manifests: ms,
	}
}

func vars(i *instance.Instance) map[string]string {
	vs := map[string]string{}
	for k, v := range i.Vars {
		vs[k] = v
	}

	vs["playbook_id"] = i.PlaybookID
	vs["instance_id"] = i.ID
	vs["id"] = i.ID
	vs["instance_status"] = string(i.Status)

	return vs
}

// DeployAndNotify attempts to deploy an instance. It reports success or failure
// through the notification service as well as returning an error.
func (d *DeploymentService) DeployAndNotify(i *instance.Instance) error {
	playbook, ok := d.playbooks[i.PlaybookID]
	if !ok {
		msg := fmt.Sprintf("Can't deploy %s/%s: Playbook missing", i.PlaybookID, i.ID)
		notify(i, msg)
		return errors.New(msg)
	}

	config, err := deployment.Config()
	if err != nil {
		msg := fmt.Sprintf("Can't deploy %s/%s: Internal error", i.PlaybookID, i.ID)
		notify(i, msg)
		return err
	}

	deployer, err := deployment.NewKubernetesDeployment(config, playbook, vars(i), d.manifests)
	if err != nil {
		msg := fmt.Sprintf("Can't deploy %s/%s: Internal error", i.PlaybookID, i.ID)
		notify(i, msg)
		return err
	}

	if i.Status == instance.StatusDeploying {
		msg := fmt.Sprintf("Can't deploy %s/%s: Instance is being deployed already.", i.PlaybookID, i.ID)
		notify(i, msg)
		return errors.New(msg)
	}

	if i.Status == instance.StatusDeleting {
		msg := fmt.Sprintf("Can't deploy %s/%s: Instance is being deleted already.", i.PlaybookID, i.ID)
		notify(i, msg)
		return errors.New(msg)
	}

	i.Status = instance.StatusDeploying
	err = d.repo.Save(i)
	if err != nil {
		glog.Errorf("Failed to save instance status Deploying for %s/%s, continuing deployment. Error: %s\n", i.PlaybookID, i.ID, err.Error())
	}

	errD := deployer.Deploy()
	if errD != nil {
		// Mark the instance as problematic:
		i.Status = instance.StatusError
		err := d.repo.Save(i)
		if err != nil {
			glog.Errorf("Failed to save instance.StatusError for %s/%s; not sending notification:\n%s\n", i.PlaybookID, i.ID, err.Error())
			return err
		}

		// Report the problem:
		msg := fmt.Sprintf("Deploying %s/%s failed: %s\n", i.PlaybookID, i.ID, errD.Error())
		glog.Error(msg)
		m := notification.NewMessage(false, msg)
		err = m.Send()
		if err != nil {
			return err
		}

		return errD
	}

	// It worked, notify success:
	msg := fmt.Sprintf("Instance %s/%s deployed successfully", i.PlaybookID, i.ID)
	notify(i, msg)

	i.Status = instance.StatusDeployed
	err = d.repo.Save(i)
	if err != nil {
		glog.Errorf("DeploymentService failed to save instance status Deployed for %s/%s:\n%s\n", i.PlaybookID, i.ID, err.Error())
		return err
	}

	return nil
}

func notify(i *instance.Instance, msg string) {
	m := notification.NewMessage(false, msg)
	err := m.Send()
	if err != nil {
		glog.Warningf("Failed to send notification from DeploymentService:\n%+v\n", m)
	}
}
