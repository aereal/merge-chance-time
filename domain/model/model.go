package model

import (
	"fmt"
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
	if s == nil {
		return ""
	}
	return s.Repr
}

func (s *CronSchedule) UnmarshalText(text []byte) error {
	repr := string(text)
	parsed, err := cronParser.Parse(repr)
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
	cfg := &RepositoryConfig{StartSchedule: &CronSchedule{}, StopSchedule: &CronSchedule{}}
	if err := cfg.StartSchedule.UnmarshalText(startScheduleRepr); err != nil {
		return nil, err
	}
	if err := cfg.StopSchedule.UnmarshalText(stopScheduleRepr); err != nil {
		return nil, err
	}
	return cfg, nil
}

type RepositoryConfig struct {
	Owner          string
	Name           string
	StartSchedule  *CronSchedule
	StopSchedule   *CronSchedule
	MergeAvailable bool
}

var (
	cronParser   = cron.NewParser(cron.Hour | cron.Dom | cron.Month | cron.Dow)
	baseDuration = time.Hour

	NowFunc = time.Now
)

func (c *RepositoryConfig) ShouldStartOn(expected time.Time) bool {
	expected = expected.Truncate(baseDuration)
	baseTime := expected.Add(baseDuration * -1)
	return timeHasCome(c.StartSchedule, baseTime, expected)
}

func (c *RepositoryConfig) ShouldStopOn(expected time.Time) bool {
	expected = expected.Truncate(baseDuration)
	baseTime := expected.Add(baseDuration * -1)
	return timeHasCome(c.StopSchedule, baseTime, expected)
}

func (c *RepositoryConfig) Valid() error {
	if c.Owner == "" {
		return fmt.Errorf("Owner must not be empty")
	}
	if c.Name == "" {
		return fmt.Errorf("Name must not be empty")
	}
	return nil
}

func timeHasCome(schedule *CronSchedule, baseTime, expectedTime time.Time) bool {
	nextTime := schedule.Next(baseTime)
	return nextTime.Equal(expectedTime)
}
