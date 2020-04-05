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
	Owner         string
	Name          string
	StartSchedule *CronSchedule
	StopSchedule  *CronSchedule
}

var (
	parser       = cron.NewParser(cron.Hour | cron.Dom | cron.Month | cron.Dow)
	baseDuration = time.Hour

	NowFunc = time.Now
)

func (c *RepositoryConfig) ShouldStartOn(baseTime time.Time) bool {
	baseTime = baseTime.Truncate(baseDuration)
	expectedTime := baseTime.Add(baseDuration)
	return timeHasCome(c.StartSchedule, baseTime, expectedTime)
}

func (c *RepositoryConfig) ShouldStopOn(baseTime time.Time) bool {
	baseTime = baseTime.Truncate(baseDuration)
	expectedTime := baseTime.Add(baseDuration)
	return timeHasCome(c.StopSchedule, baseTime, expectedTime)
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
