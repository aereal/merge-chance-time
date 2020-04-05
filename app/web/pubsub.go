package web

import (
	"encoding/json"
	"time"
)

type PubSubPayload struct {
	Message      *PubSubMessage `json:"message"`
	Subscription string         `json:"subscription"`
}

type PubSubMessage struct {
	Data        json.RawMessage `json:"data"`
	ID          string          `json:"messageId"`
	PublishTime PublishTime     `json:"publishTime"`
}

type PublishTime time.Time

func (t *PublishTime) UnmarshalText(text []byte) error {
	parsed, err := time.ParseInLocation(time.RFC3339Nano, string(text), time.Local)
	if err != nil {
		return err
	}
	*t = PublishTime(parsed)
	return nil
}

func (t PublishTime) MarshalText() ([]byte, error) {
	return []byte(time.Time(t).Format(time.RFC3339Nano)), nil
}

func (t PublishTime) String() string {
	return time.Time(t).Format(time.RFC3339Nano)
}
