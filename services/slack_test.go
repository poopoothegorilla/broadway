package services

import (
	"testing"

	"github.com/namely/broadway/deployment"
	"github.com/namely/broadway/instance"
	"github.com/namely/broadway/store"
	"github.com/stretchr/testify/assert"
)

func TestSetvarExecute(t *testing.T) {
	is := NewInstanceService(store.New())
	testcases := []struct {
		Scenario     string
		Arguments    string
		Instance     *instance.Instance
		Playbooks    map[string]*deployment.Playbook
		ExpectedVars map[string]string
		ExpectedMsg  string
		E            error
	}{
		{
			"When an instance sets a new variable value from the playbook",
			"setvar helloplaybook setvar1 bird=chickadee",
			&instance.Instance{PlaybookID: "helloplaybook", ID: "setvar1", Vars: map[string]string{"bird": "", "word": ""}},
			map[string]*deployment.Playbook{"helloplaybook": &deployment.Playbook{ID: "helloplaybook", Vars: []string{"word", "bird"}}},
			map[string]string{"bird": "chickadee", "word": ""},
			"Instance helloplaybook setvar1 updated its variables",
			nil,
		},
		{
			"When an instance sets new variable values from the playbook",
			"setvar helloplaybook setvar2 bird=gander word=test",
			&instance.Instance{PlaybookID: "helloplaybook", ID: "setvar2", Vars: map[string]string{"bird": "", "word": ""}},
			map[string]*deployment.Playbook{"helloplaybook": &deployment.Playbook{ID: "helloplaybook", Vars: []string{"word", "bird"}}},
			map[string]string{"bird": "gander", "word": "test"},
			"Instance helloplaybook setvar2 updated its variables",
			nil,
		},
		{
			"When the instance's playbook does not define a variable",
			"setvar helloplaybook setvar3 newvar=val1",
			&instance.Instance{PlaybookID: "helloplaybook", ID: "setvar3", Vars: map[string]string{"bird": "", "word": ""}},
			map[string]*deployment.Playbook{"helloplaybook": &deployment.Playbook{ID: "helloplaybook", Vars: []string{"word", "bird"}}},
			map[string]string{"bird": "", "word": ""},
			"Playbook helloplaybook does not define those variables",
			&InvalidSetVar{},
		},
		{
			"When an argument text sets 'key='",
			"setvar helloplaybook setvar4 bird=",
			&instance.Instance{PlaybookID: "helloplaybook", ID: "setvar4", Vars: map[string]string{"bird": "", "word": ""}},
			map[string]*deployment.Playbook{"helloplaybook": &deployment.Playbook{ID: "helloplaybook", Vars: []string{"word", "bird"}}},
			map[string]string{"bird": "", "word": ""},
			"Instance helloplaybook setvar4 updated its variables",
			nil,
		},
		{
			"When the argument sets '=value'",
			"setvar helloplaybook setvar5 =broken",
			&instance.Instance{PlaybookID: "helloplaybook", ID: "setvar5", Vars: map[string]string{"bird": "", "word": ""}},
			map[string]*deployment.Playbook{"helloplaybook": &deployment.Playbook{ID: "helloplaybook", Vars: []string{"word", "bird"}}},
			map[string]string{"bird": "", "word": ""},
			"Playbook helloplaybook does not define those variables",
			&InvalidSetVar{},
		},
		{
			"When just the setvar command is sent",
			"setvar",
			&instance.Instance{PlaybookID: "helloplaybook", ID: "setvar6", Vars: map[string]string{"bird": "", "word": ""}},
			map[string]*deployment.Playbook{"helloplaybook": &deployment.Playbook{ID: "helloplaybook", Vars: []string{"word", "bird"}}},
			map[string]string{"bird": "", "word": ""},
			"",
			&InvalidSetVar{},
		},
	}

	for _, testcase := range testcases {
		_, err := is.Create(testcase.Instance)
		command := BuildSlackCommand(testcase.Arguments, is, testcase.Playbooks)
		if err != nil {
			t.Log(err)
		}

		msg, err := command.Execute()
		assert.Equal(t, testcase.ExpectedMsg, msg, testcase.Scenario)
		assert.Equal(t, testcase.E, err, testcase.Scenario)

		updatedInstance, _ := is.Show(testcase.Instance.PlaybookID, testcase.Instance.ID)
		assert.Equal(t, testcase.ExpectedVars, updatedInstance.Vars, testcase.Scenario)
	}
}

func TestHelpExecute(t *testing.T) {
	testcases := []struct {
		Scenario string
		Args     string
		Expected string
		E        error
	}{
		{
			"When passing help command",
			"help",
			`/broadway help: This message
/broadway deploy myPlaybookID myInstanceID: Deploy a new instance
/broadway setvar myPlaybookID myInstanceID var1=val1 githash=8ad33dad env=prod`,
			nil,
		},
		{
			"When non existent command",
			"none",
			`/broadway help: This message
/broadway deploy myPlaybookID myInstanceID: Deploy a new instance
/broadway setvar myPlaybookID myInstanceID var1=val1 githash=8ad33dad env=prod`,
			nil,
		},
	}
	is := NewInstanceService(store.New())
	for _, testcase := range testcases {
		command := BuildSlackCommand(testcase.Args, is, nil)
		msg, err := command.Execute()
		assert.Nil(t, err)
		assert.Equal(t, testcase.Expected, msg)
	}
}
