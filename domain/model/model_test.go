package model

import (
	"testing"
	"time"
)

func mustParseTime(repr string) time.Time {
	t, err := time.Parse(time.RFC3339, repr)
	if err != nil {
		panic(err)
	}
	return t
}

func TestRepositoryConfig_ShouldStartOn(t *testing.T) {
	type args struct {
		expected time.Time
	}
	tests := []struct {
		name string
		cfg  *RepositoryConfig
		args args
		want bool
	}{
		{
			name: "OK",
			cfg: &RepositoryConfig{
				MergeAvailable: false,
				Schedules: &MergeChanceSchedules{
					Sunday: &MergeChanceSchedule{
						StartHour: 0,
						StopHour:  23,
					},
					Monday: &MergeChanceSchedule{
						StartHour: 0,
						StopHour:  23,
					},
					Tuesday: &MergeChanceSchedule{
						StartHour: 0,
						StopHour:  23,
					},
					Wednesday: &MergeChanceSchedule{
						StartHour: 0,
						StopHour:  23,
					},
					Thursday: &MergeChanceSchedule{
						StartHour: 0,
						StopHour:  23,
					},
					Friday: &MergeChanceSchedule{
						StartHour: 0,
						StopHour:  23,
					},
					Saturday: &MergeChanceSchedule{
						StartHour: 0,
						StopHour:  23,
					},
				},
			},
			args: args{
				expected: mustParseTime("2020-02-01T00:00:00Z"),
			},
			want: true,
		},
		{
			name: "already started",
			cfg: &RepositoryConfig{
				MergeAvailable: true,
				Schedules: &MergeChanceSchedules{
					Sunday: &MergeChanceSchedule{
						StartHour: 0,
						StopHour:  23,
					},
					Monday: &MergeChanceSchedule{
						StartHour: 0,
						StopHour:  23,
					},
					Tuesday: &MergeChanceSchedule{
						StartHour: 0,
						StopHour:  23,
					},
					Wednesday: &MergeChanceSchedule{
						StartHour: 0,
						StopHour:  23,
					},
					Thursday: &MergeChanceSchedule{
						StartHour: 0,
						StopHour:  23,
					},
					Friday: &MergeChanceSchedule{
						StartHour: 0,
						StopHour:  23,
					},
					Saturday: &MergeChanceSchedule{
						StartHour: 0,
						StopHour:  23,
					},
				},
			},
			args: args{
				expected: mustParseTime("2020-02-01T00:00:00Z"),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cfg.ShouldStartOn(tt.args.expected); got != tt.want {
				t.Errorf("RepositoryConfig.ShouldStartOn() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRepositoryConfig_ShouldStopOn(t *testing.T) {
	type args struct {
		expected time.Time
	}
	tests := []struct {
		name string
		cfg  *RepositoryConfig
		args args
		want bool
	}{
		{
			name: "OK",
			cfg: &RepositoryConfig{
				MergeAvailable: true,
				Schedules: &MergeChanceSchedules{
					Sunday: &MergeChanceSchedule{
						StartHour: 0,
						StopHour:  23,
					},
					Monday: &MergeChanceSchedule{
						StartHour: 0,
						StopHour:  23,
					},
					Tuesday: &MergeChanceSchedule{
						StartHour: 0,
						StopHour:  23,
					},
					Wednesday: &MergeChanceSchedule{
						StartHour: 0,
						StopHour:  23,
					},
					Thursday: &MergeChanceSchedule{
						StartHour: 0,
						StopHour:  23,
					},
					Friday: &MergeChanceSchedule{
						StartHour: 0,
						StopHour:  23,
					},
					Saturday: &MergeChanceSchedule{
						StartHour: 0,
						StopHour:  23,
					},
				},
			},
			args: args{
				expected: mustParseTime("2020-02-01T23:00:00Z"),
			},
			want: true,
		},
		{
			name: "already stopped",
			cfg: &RepositoryConfig{
				MergeAvailable: false,
				Schedules: &MergeChanceSchedules{
					Sunday: &MergeChanceSchedule{
						StartHour: 0,
						StopHour:  23,
					},
					Monday: &MergeChanceSchedule{
						StartHour: 0,
						StopHour:  23,
					},
					Tuesday: &MergeChanceSchedule{
						StartHour: 0,
						StopHour:  23,
					},
					Wednesday: &MergeChanceSchedule{
						StartHour: 0,
						StopHour:  23,
					},
					Thursday: &MergeChanceSchedule{
						StartHour: 0,
						StopHour:  23,
					},
					Friday: &MergeChanceSchedule{
						StartHour: 0,
						StopHour:  23,
					},
					Saturday: &MergeChanceSchedule{
						StartHour: 0,
						StopHour:  23,
					},
				},
			},
			args: args{
				expected: mustParseTime("2020-02-01T23:00:00Z"),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cfg.ShouldStopOn(tt.args.expected); got != tt.want {
				t.Errorf("RepositoryConfig.ShouldStopOn() = %v, want %v", got, tt.want)
			}
		})
	}
}
