package services

import (
	"testing"

	"github.com/namely/broadway/instance"
	"github.com/namely/broadway/store"
	"github.com/stretchr/testify/assert"
)

func TestCreateInstance(t *testing.T) {
	nt := newNotificationTestHelper()
	defer nt.Close()
	store := store.New()
	service := NewInstanceService(store)

	i := &instance.Instance{PlaybookID: "test", ID: "222"}
	createdInstance, err := service.Create(i)
	assert.Nil(t, err)
	assert.Equal(t, "test", createdInstance.PlaybookID)
	assert.Equal(t, instance.StatusNew, createdInstance.Status)
	assert.Contains(t, nt.requestBody, "created")
}

func TestShow(t *testing.T) {
	store := store.New()
	service := NewInstanceService(store)

	i := &instance.Instance{PlaybookID: "test", ID: "222"}
	createdInstance, err := service.Create(i)
	i, err = service.Show(i.PlaybookID, i.ID)
	assert.Nil(t, err)
	assert.Equal(t, createdInstance, i)
}

func TestShowMissingInstance(t *testing.T) {
	store := store.New()
	service := NewInstanceService(store)

	i := &instance.Instance{PlaybookID: "test", ID: "broken"}
	i, err := service.Show(i.PlaybookID, i.ID)
	assert.NotNil(t, err)
	assert.Nil(t, i, "PlaybookID should be empty")
}

func TestAllWithPlaybookID(t *testing.T) {
	nt := newNotificationTestHelper()
	defer nt.Close()
	service := NewInstanceService(store.New())

	i := &instance.Instance{PlaybookID: "none", ID: "none"}
	_, err := service.Create(i)
	if err != nil {
		t.Fail()
	}

	instances, err := service.AllWithPlaybookID(i.PlaybookID)
	assert.Nil(t, err)
	assert.NotEmpty(t, instances)
}

func TestUpdate(t *testing.T) {
	nt := newNotificationTestHelper()
	defer nt.Close()
	instanceService := NewInstanceService(store.New())
	testcases := []struct {
		Scenario           string
		Instance           *instance.Instance
		ExpectedPlaybookID string
		ExpectedID         string
		ExpectedVars       map[string]string
		E                  error
	}{
		{
			"When the instance have all the needed values",
			&instance.Instance{PlaybookID: "foo", ID: "bar"},
			"bar",
			"foo",
			map[string]string{},
			nil,
		},
	}

	for _, testcase := range testcases {
		createdInstance, err := instanceService.Create(testcase.Instance)
		if err != nil {
			t.Fail()
		}
		createdInstance.PlaybookID = testcase.ExpectedPlaybookID
		createdInstance.ID = testcase.ExpectedID
		createdInstance.Vars = testcase.ExpectedVars
		updatedInstance, err := instanceService.Update(createdInstance)

		assert.Equal(t, testcase.ExpectedPlaybookID, updatedInstance.PlaybookID)
		assert.Equal(t, testcase.E, err, testcase.Scenario)
	}
}

func TestDeleteWhenExistentInstance(t *testing.T) {
	nt := newNotificationTestHelper()
	defer nt.Close()
	is := NewInstanceService(store.New())

	i := &instance.Instance{PlaybookID: "foo", ID: "bar"}

	createdInstance, err := is.Create(i)
	if err != nil {
		t.Fail()
	}
	err = is.Delete(createdInstance)
	assert.Nil(t, err, "When existent instance")
}

func TestDeleteWhenNonExistantInstance(t *testing.T) {
	is := NewInstanceService(store.New())
	i := &instance.Instance{PlaybookID: "random", ID: "bar"}

	err := is.Delete(i)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "was not found", "When non-existent instance")
}
