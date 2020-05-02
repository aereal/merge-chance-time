// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package dto

type MergeChanceSchedule struct {
	StartHour *int `json:"startHour"`
	StopHour  *int `json:"stopHour"`
}

type MergeChanceScheduleToUpdate struct {
	StartHour *int `json:"startHour"`
	StopHour  *int `json:"stopHour"`
}

type MergeChanceSchedules struct {
	Sunday    *MergeChanceSchedule `json:"sunday"`
	Monday    *MergeChanceSchedule `json:"monday"`
	Tuesday   *MergeChanceSchedule `json:"tuesday"`
	Wednesday *MergeChanceSchedule `json:"wednesday"`
	Thursday  *MergeChanceSchedule `json:"thursday"`
	Friday    *MergeChanceSchedule `json:"friday"`
	Saturday  *MergeChanceSchedule `json:"saturday"`
}

type MergeChanceSchedulesToUpdate struct {
	Sunday    *MergeChanceScheduleToUpdate `json:"sunday"`
	Monday    *MergeChanceScheduleToUpdate `json:"monday"`
	Tuesday   *MergeChanceScheduleToUpdate `json:"tuesday"`
	Wednesday *MergeChanceScheduleToUpdate `json:"wednesday"`
	Thursday  *MergeChanceScheduleToUpdate `json:"thursday"`
	Friday    *MergeChanceScheduleToUpdate `json:"friday"`
	Saturday  *MergeChanceScheduleToUpdate `json:"saturday"`
}

type RepositoryConfig struct {
	Schedules      *MergeChanceSchedules `json:"schedules"`
	MergeAvailable bool                  `json:"mergeAvailable"`
}

type RepositoryConfigToUpdate struct {
	Schedules *MergeChanceSchedulesToUpdate `json:"schedules"`
}
