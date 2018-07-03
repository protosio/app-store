package registry

import (
	"log"
	"time"
)

type Target struct {
	MediaType  string `json:"mediaType"`
	Size       int    `json:"size"`
	Digest     string `json:"digest"`
	Length     int    `json:"length"`
	Repository string `json:"repository"`
	URL        string `json:"url"`
	Tag        string `json:"tag"`
}

type Event struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Action    string    `json:"action"`
	Target    Target    `json:"target"`
}

type Events struct {
	Events []Event `json:"events"`
}

// ProcessEvents takes an events array and process all the events of type "push"
func ProcessEvents(events []Event) {
	for _, event := range events {
		if event.Action != "push" {
			log.Println("Ignoring event of type " + event.Action)
		}
		log.Println(event.Target)
	}
}
