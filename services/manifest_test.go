package services

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/namely/broadway/deployment"
	"github.com/namely/broadway/env"
	"github.com/stretchr/testify/assert"
)

var tmpDir string

func Setup() {
	var err error
	tmpDir, err = ioutil.TempDir("", "manifest_test_")
	if err != nil {
		log.Fatal(err)
	}
	content := []byte(testManifest)
	tmpFN := filepath.Join(tmpDir, "test.yml")
	if err := ioutil.WriteFile(tmpFN, content, 0666); err != nil {
		log.Fatal(err)
	}
}
func Teardown() {
	defer os.RemoveAll(tmpDir)
}

func TestMain(m *testing.M) {
	Setup()
	testresult := m.Run()
	Teardown()
	os.Exit(testresult)
}

func TestNewManifestService(t *testing.T) {
	ms := NewManifestService(env.ManifestsPath)
	assert.Equal(t, env.ManifestsPath, ms.rootFolder)
}

func TestRead(t *testing.T) {
	ms := NewManifestService(env.ManifestsPath)
	ms.rootFolder = tmpDir
	assert.Equal(t, tmpDir, ms.rootFolder)
	contents, err := ms.Read("test")
	assert.Contains(t, contents, "ReplicationController")
	assert.Nil(t, err)

	contents, err = ms.Read("missing")
	assert.NotNil(t, err)
}

func TestLoad(t *testing.T) {
	ms := NewManifestService(env.ManifestsPath)
	ms.rootFolder = tmpDir
	m, err := ms.Load("test")
	assert.Nil(t, err)
	assert.Equal(t, m.ID, "test")

	m, err = ms.Load("missing")
	assert.NotNil(t, err)
}

func TestLoadTask(t *testing.T) {
	ms := NewManifestService(env.ManifestsPath)
	ms.rootFolder = tmpDir
	tk := deployment.Task{
		Name: "First step",
		Manifests: []string{
			"test",
		},
	}

	// Load from task with manifests
	pod, mm, err := ms.LoadTask(tk)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(mm))
	assert.Equal(t, mm[0].ID, "test")

	// Load from task with pod manifest
	tk.Manifests = []string{}
	tk.PodManifest = "test"
	pod, mm, err = ms.LoadTask(tk)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(mm))
	assert.Equal(t, "test", pod.ID)

}

func TestNew(t *testing.T) {
	ms := NewManifestService(env.ManifestsPath)
	ms.rootFolder = tmpDir

	m, err := ms.New("testId", "testContent")
	assert.Equal(t, m.ID, "testId")
	assert.Nil(t, err)
}

const testManifest = `apiVersion: v1
kind: ReplicationController
metadata:
  name: test
spec:
  replicas: 1
  selector:
    name: redis
  template:
    metadata:
      labels:
        name: redis
    spec:
      containers:
      - name: redis
        image: kubernetes/redis:v1
`
