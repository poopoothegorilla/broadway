package instance

import (
	"testing"

	"github.com/namely/broadway/store"
	"github.com/stretchr/testify/assert"
)

type DummyStore struct{}

func (ds *DummyStore) Value(path string) string {
	return "malformed_json"
}

func (ds *DummyStore) SetValue(path, value string) error {
	return nil
}

func (ds *DummyStore) Values(path string) map[string]string {
	return map[string]string{"foo": "foo"}
}

func (ds *DummyStore) Delete(path string) error {
	return nil
}

func TestFindByPath(t *testing.T) {
	repo := NewRepo(store.New())
	i := &Instance{PlaybookID: "test", ID: "222"}
	err := repo.Save(i)
	assert.Nil(t, err)

	instance, err := repo.FindByPath(i.Path())
	assert.Nil(t, err)
	assert.NotNil(t, instance)
	assert.Equal(t, "test", instance.PlaybookID)
}

func TestFindByPathWhenTheInstanceDoesNotExist(t *testing.T) {
	repo := NewRepo(store.New())
	i := Instance{PlaybookID: "notcreated", ID: "222"}

	instance, err := repo.FindByPath(i.Path())
	assert.NotNil(t, err)
	assert.Equal(t, "Instance with path: "+i.Path()+" was not found", err.Error())
	assert.Nil(t, instance)
}

func TestFindByPathWhenMalformedData(t *testing.T) {
	repo := NewRepo(&DummyStore{})
	i := Instance{PlaybookID: "notcreated", ID: "222"}

	_, err := repo.FindByPath(i.Path())
	assert.NotNil(t, err)
	assert.Equal(t, "Saved data for this instance is malformed", err.Error())
}

func TestFindByID(t *testing.T) {
	repo := NewRepo(store.New())
	i := &Instance{PlaybookID: "created", ID: "222"}

	err := repo.Save(i)
	if err != nil {
		t.Error(err)
	}
	instance, err := repo.FindByID(i.PlaybookID, i.ID)
	assert.Nil(t, err)
	assert.Equal(t, "created", instance.PlaybookID)
}

func TestFindByPlaybookIDOne(t *testing.T) {
	repo := NewRepo(store.New())
	i := &Instance{PlaybookID: "one", ID: "222"}
	err := repo.Save(i)

	if err != nil {
		t.Fail()
	}
	instances, err := repo.FindByPlaybookID(i.PlaybookID)
	assert.Nil(t, err)
	assert.Len(t, instances, 1)
}

func TestFindByPlaybookIDMany(t *testing.T) {
	repo := NewRepo(store.New())
	i := &Instance{PlaybookID: "many", ID: "1of2"}
	repo.Save(i)
	j := &Instance{PlaybookID: "many", ID: "2of2"}
	repo.Save(j)

	instances, err := repo.FindByPlaybookID(i.PlaybookID)
	assert.Nil(t, err)
	assert.Len(t, instances, 2)
}

func TestFindByPlaybookIDNoExistent(t *testing.T) {
	repo := NewRepo(&DummyStore{})
	instances, err := repo.FindByPlaybookID("notcreated")

	assert.NotNil(t, err)
	assert.Empty(t, instances)

	assert.Equal(t, "Saved data for this instance is malformed", err.Error())
}

func TestDelete(t *testing.T) {
	repo := NewRepo(store.New())
	i := &Instance{PlaybookID: "anewone", ID: "withid"}
	err := repo.Save(i)
	if err != nil {
		t.Fail()
	}
	err = repo.Delete(i)
	assert.Nil(t, err)
}
