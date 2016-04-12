package deployment

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/golang/glog"
	"github.com/namely/broadway/env"
)

const MockPlaybookContents string = `---
id: project-playbook
name: The Project 
meta:
  team: Project Devs
  email: devs@project.com
  slack: devs
vars:
  - version
  - assets_version
  - owner
tasks:
  - name: Deploy Postgres
    manifests:
      - hello
      - hello
  - name: Deploy Redis
    manifests:
      - hello
      - hello
  - name: Database Setup
    pod_manifest: hello
    wait_for:
      - success
    when: new_deployment
  - name: Database Migration
    pod_manifest: hello
    wait_for:
      - success
  - name: Deploy Project
    manifests:
      - hello
      - hello
      - hello
`
const MockPlaybookContentsIncomplete = `---
name: The Project 
`
const MockPlaybookFilename = "test-playbook.yml"

var MockPlaybookBytes = []byte(MockPlaybookContents)
var MockIncompletePlaybookBytes = []byte(MockPlaybookContentsIncomplete)
var rootDir string
var playbookDir string
var mockPlaybookPath string

func SetupTestFixtures() {
	// Find project root:
	cwd, err := os.Getwd()
	if err != nil {
		glog.Fatal(err)
	}
	rootDir = filepath.Join(cwd, "..")

	// Set path variables:
	playbookDir = filepath.Join(rootDir, "playbooks")
	mockPlaybookPath = filepath.Join(env.PlaybooksPath, MockPlaybookFilename)

	// Create folders:
	err = os.MkdirAll(playbookDir, os.ModePerm)
	if err != nil {
		glog.Fatalf("Failed to create broadway/playbooks folder: %s", err)
	}

	// Write mock playbook and manifest::
	f, err := os.Create(mockPlaybookPath)
	if err != nil {
		glog.Fatalf("Failed to write mock test playbook: %s", err)
	}
	_, err = f.Write(MockPlaybookBytes)
	if err != nil {
		panic(err)
	}
	err = f.Close()
	if err != nil {
		panic(err)
	}
}

// TeardownTestFixtures will remove the mock files but leave the folders
func TeardownTestFixtures() {
	err := os.Remove(mockPlaybookPath)
	if err != nil {
		glog.Error(err)
	}
}

func TestMain(m *testing.M) {
	SetupTestFixtures()
	testresult := m.Run()
	TeardownTestFixtures()
	os.Exit(testresult)
}

func TestReadPlaybookFromDisk(t *testing.T) {
	playbook, err := ReadPlaybookFromDisk(mockPlaybookPath)
	if err != nil {
		t.Error(err)
		return
	}
	if !bytes.Equal(playbook, MockPlaybookBytes) {
		t.Error(errors.New("Playbook read from disk differs from Playbook written to disk"))
		return
	}
}

func TestParsePlaybook(t *testing.T) {
	var err error
	ParsedPlaybook, err := ParsePlaybook(MockPlaybookBytes)
	if err != nil {
		t.Error(err)
		return
	}
	if ParsedPlaybook.Name != "The Project" {
		t.Error(errors.New("Parsed Playbook has incorrect Name field"))
		return
	}
}

func TestParsePlaybookMalformed(t *testing.T) {
	_, err := ParsePlaybook([]byte("asdf"))
	if err == nil {
		t.Error(errors.New("Parsing asdf succeeded, expected failure"))
		return
	}
}

func TestParsePlaybookIncomplete(t *testing.T) {
	_, err := ParsePlaybook(MockIncompletePlaybookBytes)
	if err != nil {
		t.Error(errors.New("Parsing well-formed, incomplete playbook failed, expected success"))
		return
	}
}

func TestValidatePlaybookPasses(t *testing.T) {
	ValidTask1 := Task{
		Name:      "task",
		Manifests: []string{"hello"},
	}
	ValidTask2 := Task{
		Name:        "task",
		PodManifest: "hello",
	}
	ParsedPlaybook, _ := ParsePlaybook(MockPlaybookBytes) // already checked err in previous test

	testcases := []struct {
		scenario string
		playbook *Playbook
	}{
		{
			"Validate Valid Playbook",
			&Playbook{
				ID:    "playbook id 1",
				Name:  "playbook 1",
				Tasks: []Task{ValidTask1, ValidTask2},
			},
		},
		{
			"Parse And Validate Test Playbook",
			ParsedPlaybook,
		},
	}
	for _, testcase := range testcases {
		playbook := testcase.playbook
		err := playbook.Validate()
		if err != nil {
			t.Errorf("Scenario %s\nExpected: No error\nActual:\n%s", testcase.scenario, err.Error())
		}
	}
}

func TestValidatePlaybookFailures(t *testing.T) {
	InvalidTask1 := Task{
		Manifests: []string{"hello"},
	}
	InvalidTask2 := Task{
		Name: "task",
	}

	testcases := []struct {
		scenario    string
		playbook    *Playbook
		expectedErr string
	}{
		{
			"Validate Playbook Without ID",
			&Playbook{},
			"Playbook missing required ID",
		},
		{
			"Validate Playbook Without Name",
			&Playbook{
				ID: "playbook id 1",
			},
			"Playbook missing required Name",
		},
		{
			"Validate Playbook With Zero Tasks",
			&Playbook{
				ID:    "playbook id 1",
				Name:  "playbook 1",
				Tasks: []Task{},
			},
			"Playbook requires at least 1 task",
		},
		{
			"Validate Playbook With Tasks Missing Names",
			&Playbook{
				ID:    "playbook id 1",
				Name:  "playbook 1",
				Tasks: []Task{InvalidTask1},
			},
			"Task missing required Name",
		},
		{
			"Validate Playbook With Tasks Missing Manifests",
			&Playbook{
				ID:    "playbook id 1",
				Name:  "playbook 1",
				Tasks: []Task{InvalidTask2},
			},
			"Task requires at least one manifest or a pod manifest",
		},
	}

	for _, testcase := range testcases {
		playbook := testcase.playbook
		err := playbook.Validate()
		if testcase.expectedErr != err.Error() {
			t.Errorf("Scenario %s\nExpected:\n%s\nActual:\n%s", testcase.scenario, testcase.expectedErr, err.Error())
		}
	}
}

func TestLoadPlaybookFolder(t *testing.T) {
	pbs, err := LoadPlaybookFolder(env.PlaybooksPath)
	if err != nil {
		t.Errorf("LoadPlaybookFolder failed to load playbooks: %s\n", err)
	}
	if len(pbs) == 0 {
		t.Error("LoadPlaybookFolder failed to load")
	}
}
func TestTaskManifestsPresentPasses(t *testing.T) {
	testcases := []struct {
		scenario string
		task     Task
	}{
		{
			"Task With Existing Manifests",
			Task{
				Name:      "task 2",
				Manifests: []string{"hello"},
			},
		},
		{
			"Task With Only Pod Manifest",
			Task{
				Name:        "task 2",
				PodManifest: "hello",
			},
		},
		{
			"Task With Both Manifests And Pod Manifest",
			Task{
				Name:        "task 2",
				PodManifest: "hello",
				Manifests:   []string{"hello"},
			},
		},
	}
	for _, testcase := range testcases {
		task := testcase.task
		err := task.ManifestsPresent()
		if err != nil {
			t.Errorf("Scenario %s\nExpected: Success\nActual:\n%s", testcase.scenario, err.Error())
		}
	}
}
func TestTaskManifestsPresentFailures(t *testing.T) {
	testcases := []struct {
		scenario string
		task     Task
	}{
		{
			"Task With Missing Manifests",
			Task{
				Name:      "task 1",
				Manifests: []string{"pod0"},
			},
		},
	}

	for _, testcase := range testcases {
		task := testcase.task
		err := task.ManifestsPresent()
		if err == nil {
			t.Errorf("Scenario %s\nExpected: File does not exist\nActual: Success", testcase.scenario)
		} else if !os.IsNotExist(err) { // it was the wrong error!
			t.Errorf("Scenario %s\nExpected: File does not exist\nActual:\n%s", testcase.scenario, err.Error())
		}
	}
}
