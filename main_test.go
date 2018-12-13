package main

import (
	"github.com/dghubble/go-twitter/twitter"
	"testing"
)

type TestStore struct {
	URL string
}

func (db *TestStore) CreateURL(url string) (id int, err error) {
	db.URL = url
	return -42, nil
}

func TestTweetHandler(t *testing.T) {
	db := &TestStore{}
	handler := TweetHandler(db)
	url := "http://www.example.com/foo"
	tweet := &twitter.Tweet{
		Entities: &twitter.Entities{
			Urls: []twitter.URLEntity{
				twitter.URLEntity{ExpandedURL: url},
			},
		},
		User: &twitter.User{},
	}

	handler(tweet)

	if db.URL != url {
		t.Errorf("Expected %s but got %s", url, db.URL)
	}
}
