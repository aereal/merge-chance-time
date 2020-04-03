package model

import (
	"time"

	"github.com/robfig/cron/v3"
)

type RepositoryConfig struct {
	Schedule      string `json:"schedule"`
	StartSchedule string `json:"startSchedule"`
	StopSchedule  string `json:"stopSchedule"`
}

var (
	parser       = cron.NewParser(cron.Hour | cron.Dom | cron.Month | cron.Dow)
	baseDuration = time.Hour

	NowFunc = time.Now
)

func (c *RepositoryConfig) ShouldStartOn(previousTime time.Time) bool {
	return timeHasCome(c.StartSchedule, previousTime)
}

func (c *RepositoryConfig) ShouldStopOn(previousTime time.Time) bool {
	return timeHasCome(c.StopSchedule, previousTime)
}

func timeHasCome(schedule string, previousTime time.Time) bool {
	s, err := parser.Parse(schedule)
	if err != nil {
		return false
	}
	now := NowFunc().Round(baseDuration)
	nextTime := s.Next(previousTime)
	return nextTime.Equal(now)
}
