package audrey

import (
	"fmt"

	"github.com/dghubble/go-twitter/twitter"
)

type DataStore interface {
	CreateURL(string) (int, error)
}

func NewTweetHandler(db DataStore) func(*twitter.Tweet) {
	return func(t *twitter.Tweet) {
		fmt.Println("Hi")
	}
}
