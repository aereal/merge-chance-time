package model

import (
	"testing"
	"time"
)

func TestRepositoryConfig_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "OK",
			input:   `* * * *`,
			wantErr: false,
		},
		{
			name:    "invalid format",
			input:   "...",
			wantErr: true,
		},
		{
			name:    "minutes not supported",
			input:   `5/* * * * *`,
			wantErr: true,
		},
		{
			name:    "empty",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got CronSchedule
			err := got.UnmarshalText([]byte(tt.input))
			gotErr := err != nil
			if tt.wantErr != gotErr {
				t.Errorf("UnmarshalJSON CronSchedule expected error=%+v but got=%+v", tt.wantErr, gotErr)
			}
		})
	}
}

func TestRepositoryConfig_timeHasCome(t *testing.T) {
	type args struct {
		cronSchedule *CronSchedule
		baseTime     time.Time
		expectedTime time.Time
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "matched",
			args: args{
				cronSchedule: mustParseSchedule("* * * *"),
				baseTime:     mustParseTime("2020-01-31T23:00:00Z"),
				expectedTime: mustParseTime("2020-02-01T00:00:00Z"),
			},
			want: true,
		},
		{
			name: "matched (2)",
			args: args{
				cronSchedule: mustParseSchedule("*/2 * * *"),
				baseTime:     mustParseTime("2020-01-31T23:00:00Z"),
				expectedTime: mustParseTime("2020-02-01T00:00:00Z"),
			},
			want: true,
		},
		{
			name: "matched (3)",
			args: args{
				cronSchedule: mustParseSchedule("*/2 * * *"),
				baseTime:     mustParseTime("2020-01-31T23:50:00Z"),
				expectedTime: mustParseTime("2020-02-01T00:00:00Z"),
			},
			want: true,
		},
		{
			name: "not matched",
			args: args{
				cronSchedule: mustParseSchedule("*/2 * * *"),
				baseTime:     mustParseTime("2020-02-01T00:00:00Z"),
				expectedTime: mustParseTime("2020-02-01T01:00:00Z"),
			},
			want: false,
		},
		{
			name: "with minutes",
			args: args{
				cronSchedule: mustParseSchedule("* * * *"),
				baseTime:     mustParseTime("2020-01-31T23:01:00Z"),
				expectedTime: mustParseTime("2020-02-01T00:00:00Z"),
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := timeHasCome(tt.args.cronSchedule, tt.args.baseTime, tt.args.expectedTime)
			if got != tt.want {
				t.Errorf("RepositoryConfig.ShouldStartOn() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCronSchedule_String(t *testing.T) {
	tests := []struct {
		name     string
		schedule *CronSchedule
		want     string
	}{
		{
			name:     "nil",
			schedule: nil,
			want:     "",
		},
		{
			name:     "ok",
			schedule: mustParseSchedule("* * * *"),
			want:     "* * * *",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.schedule.String(); got != tt.want {
				t.Errorf("CronSchedule.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func mustParseTime(repr string) time.Time {
	t, err := time.Parse(time.RFC3339, repr)
	if err != nil {
		panic(err)
	}
	return t
}

func mustParseSchedule(repr string) *CronSchedule {
	parsed, err := cronParser.Parse(repr)
	if err != nil {
		panic(err)
	}
	return &CronSchedule{
		Repr: repr,
		Spec: parsed,
	}
}
