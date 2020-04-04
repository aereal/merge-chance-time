package model

import (
	"time"

	"github.com/robfig/cron/v3"
)

type CronSchedule struct {
	Repr string
	Spec cron.Schedule
}

func (s *CronSchedule) Next(t time.Time) time.Time {
	return s.Spec.Next(t)
}

func (s *CronSchedule) String() string {
	return s.Repr
}

func (s *CronSchedule) UnmarshalText(text []byte) error {
	repr := string(text)
	parsed, err := parser.Parse(repr)
	if err != nil {
		return err
	}
	*s = CronSchedule{
		Spec: parsed,
		Repr: repr,
	}
	return nil
}

func (s CronSchedule) MarshalText() ([]byte, error) {
	return []byte(s.Repr), nil
}

func NewRepositoryConfig(startScheduleRepr, stopScheduleRepr []byte) (*RepositoryConfig, error) {
	cfg := &RepositoryConfig{}
	if err := cfg.StartSchedule.UnmarshalText(startScheduleRepr); err != nil {
		return nil, err
	}
	if err := cfg.StopSchedule.UnmarshalText(stopScheduleRepr); err != nil {
		return nil, err
	}
	return cfg, nil
}

type RepositoryConfig struct {
	StartSchedule *CronSchedule `json:"startSchedule"`
	StopSchedule  *CronSchedule `json:"stopSchedule"`
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

func timeHasCome(schedule *CronSchedule, previousTime time.Time) bool {
	now := NowFunc().Round(baseDuration)
	nextTime := schedule.Next(previousTime)
	return nextTime.Equal(now)
}
