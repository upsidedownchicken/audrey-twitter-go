package audrey

import "testing"

type TestStore struct {
	URL string
}

func (db *TestStore) CreateURL(url string) (id int, err error) {
	db.URL = url
	return -42, nil
}

func TestNewTweetHandler(t *testing.T) {
	db := &TestStore{}
	handler := NewTweetHandler(db)

	if handler == nil {
		t.Error("I don't know")
	}
}
