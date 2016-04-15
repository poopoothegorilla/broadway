package services

import (
	"fmt"
	"strings"

	"github.com/golang/glog"
	"github.com/namely/broadway/deployment"
	"github.com/namely/broadway/env"
	"github.com/namely/broadway/store"
)

// SlackCommand represents a user command that came in from Slack
type SlackCommand interface {
	Execute() (string, error)
}

type deployCommand struct {
	pID string
	ID  string
	is  *InstanceService
}

func (c *deployCommand) Execute() (string, error) {
	// todo: Load these from deployment package like playbooks
	ms := NewManifestService(env.ManifestsPath)
	AllManifests, err := ms.LoadManifestFolder()
	if err != nil {
		glog.Error(err)
	}

	ds := NewDeploymentService(store.New(), deployment.AllPlaybooks, AllManifests)

	i, err := c.is.Show(c.pID, c.ID)
	if err != nil {
		msg := fmt.Sprintf("Failed to deploy instance %s/%s: Instance not found", i.PlaybookID, i.ID)
		glog.Error(msg)
		return msg, err
	}

	go func() {
		glog.Infof("Asynchronously deploying %s/%s...", i.PlaybookID, i.ID)
		err := ds.DeployAndNotify(i)
		if err != nil {
			glog.Errorf("Slack command failed to deploy instance %s/%s:\n%s\n", i.PlaybookID, i.ID, err)
			return
		}
		glog.Infof("Slack command successfully deployed instance %s/%s", i.PlaybookID, i.ID)
		return
	}()

	return fmt.Sprintf("Started deployment of %s/%s", i.PlaybookID, i.ID), nil
}

// InvalidSetVar error presentation for invalid setvar syntax
type InvalidSetVar struct{}

func (e *InvalidSetVar) Error() string {
	return "Syntax error, example: setvar playbook1 instance10 var1=value"
}

// InvalidDeploy represents an error error for invalid deploy syntax
type InvalidDeploy struct{}

func (e *InvalidDeploy) Error() string {
	return "Syntax error, example: deploy playbook1 instance10"
}

type setvarCommand struct {
	args      []string
	playbooks map[string]*deployment.Playbook
	is        *InstanceService
}

func (c *setvarCommand) Execute() (string, error) {
	var commandMsg string
	if len(c.args) < 4 {
		return commandHints, &InvalidSetVar{}
	}
	kvs := c.args[3:] // from e.g. "setvar foo bar var1=val1 var2=val2"
	i, err := c.is.Show(c.args[1], c.args[2])
	if err != nil {
		glog.Warningf("Cannot setvars for not found instance %s/%s\n", c.args[1], c.args[2])
		return "", err
	}
	for _, kv := range kvs {
		tmp := strings.SplitN(kv, "=", 2)
		if len(tmp) != 2 {
			glog.Warningf("Setvar tried to parse badly formatted variable: %s", kv)
			return "", &InvalidSetVar{}
		}
		if c.playbookContainsVar(i.PlaybookID, tmp[0]) {
			i.Vars[tmp[0]] = tmp[1]
			commandMsg = fmt.Sprintf("Instance %s/%s updated its variables", i.PlaybookID, i.ID)
		} else {
			return fmt.Sprintf("Playbook %s does not define those variables", i.PlaybookID), &InvalidSetVar{}
		}
	}
	_, err = c.is.Update(i)
	if err != nil {
		glog.Errorf("Failed to save instance %s/%s with new vars\n", c.args[1], c.args[2])
		return "", err
	}
	return commandMsg, nil
}

func (c *setvarCommand) playbookContainsVar(playbookID, name string) bool {
	if p, ok := c.playbooks[playbookID]; ok {
		for _, playbookVar := range p.Vars {
			if playbookVar == name {
				return true
			}
		}
	}
	return false
}

// CommandHints slack commands help hints
const commandHints = `/broadway help: This message
/broadway deploy myPlaybookID myInstanceID: Deploy a new instance
/broadway setvar myPlaybookID myInstanceID var1=val1 githash=8ad33dad env=prod`

// Help slack command
type helpCommand struct{}

func (c *helpCommand) Execute() (string, error) {
	return commandHints, nil
}

// InvalidDelete definition of syntax error
type InvalidDelete struct{}

func (e *InvalidDelete) Error() string {
	return ""
}

// Delete slack command
type deleteCommand struct {
	args []string
	is   *InstanceService
}

func (c *deleteCommand) Execute() (string, error) {
	if len(c.args) < 3 {
		return "", &InvalidDelete{}
	}
	iToDelete, err := c.is.Show(c.args[1], c.args[2])

	if err != nil {
		return "", err
	}
	err = c.is.Delete(iToDelete)
	if err != nil {
		return "", err
	}
	return "Instance successfully deleted", nil
}

// BuildSlackCommand takes a string and some context and creates a SlackCommand
func BuildSlackCommand(payload string, is *InstanceService, playbooks map[string]*deployment.Playbook) SlackCommand {
	terms := strings.Split(payload, " ")
	switch terms[0] {
	case "setvar": // setvar foo bar var1=val1 var2=val2
		return &setvarCommand{args: terms, is: is, playbooks: playbooks}
	case "deploy":
		if len(terms) < 3 {
			return &helpCommand{}
		}
		return &deployCommand{
			pID: terms[1],
			ID:  terms[2],
			is:  is,
		}
	case "delete":
		return &deleteCommand{args: terms, is: is}
	default:
		return &helpCommand{}
	}
}
