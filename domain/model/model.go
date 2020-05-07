package model

import (
	"fmt"
	"time"
)

type MergeChanceSchedule struct {
	StartHour int
	StopHour  int
}

var WholeDay = &MergeChanceSchedule{StartHour: 0, StopHour: 23}

type MergeChanceSchedules struct {
	Sunday    *MergeChanceSchedule
	Monday    *MergeChanceSchedule
	Tuesday   *MergeChanceSchedule
	Wednesday *MergeChanceSchedule
	Thursday  *MergeChanceSchedule
	Friday    *MergeChanceSchedule
	Saturday  *MergeChanceSchedule
}

func (s *MergeChanceSchedules) ForWeekday(wd time.Weekday) *MergeChanceSchedule {
	switch wd {
	case time.Sunday:
		return s.Sunday
	case time.Monday:
		return s.Monday
	case time.Tuesday:
		return s.Tuesday
	case time.Wednesday:
		return s.Wednesday
	case time.Thursday:
		return s.Thursday
	case time.Friday:
		return s.Friday
	case time.Saturday:
		return s.Saturday
	default:
		return nil
	}
}

type RepositoryConfig struct {
	Owner          string
	Name           string
	Schedules      *MergeChanceSchedules
	MergeAvailable bool
}

func (c *RepositoryConfig) ShouldStartOn(expected time.Time) bool {
	if c.MergeAvailable {
		return false
	}
	schedule := c.Schedules.ForWeekday(expected.Weekday())
	if schedule == nil {
		return false
	}
	return expected.Hour() == schedule.StartHour
}

func (c *RepositoryConfig) ShouldStopOn(expected time.Time) bool {
	if !c.MergeAvailable {
		return false
	}
	schedule := c.Schedules.ForWeekday(expected.Weekday())
	if schedule == nil {
		return false
	}
	return expected.Hour() == schedule.StopHour
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
