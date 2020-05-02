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
	if s == nil {
		return time.Time{}
	}
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

type MergeChanceSchedule struct {
	StartHour *int
	StopHour  *int
}

type MergeChanceSchedules struct {
	Sunday    *MergeChanceSchedule
	Monday    *MergeChanceSchedule
	Tuesday   *MergeChanceSchedule
	Wednesday *MergeChanceSchedule
	Thursday  *MergeChanceSchedule
	Friday    *MergeChanceSchedule
	Saturday  *MergeChanceSchedule
}

type RepositoryConfig struct {
	Owner          string
	Name           string
	Schedules      *MergeChanceSchedules
	MergeAvailable bool
}

var (
	cronParser   = cron.NewParser(cron.Hour | cron.Dom | cron.Month | cron.Dow)
	baseDuration = time.Hour

	NowFunc = time.Now
)

func (c *RepositoryConfig) ShouldStartOn(expected time.Time) bool {
	return false // TODO
}

func (c *RepositoryConfig) ShouldStopOn(expected time.Time) bool {
	return false // TODO
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
