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

func TestRepositoryConfig_ShouldStartOn(t *testing.T) {
	NowFunc = func() time.Time {
		return mustParseTime("2020-02-01T00:00:00Z")
	}
	defer func() {
		NowFunc = time.Now
	}()

	type args struct {
		nowFunc      func() time.Time
		previousTime time.Time
	}
	tests := []struct {
		name   string
		config *RepositoryConfig
		args   args
		want   bool
	}{
		{
			name: "matched",
			config: &RepositoryConfig{
				StartSchedule: mustParseSchedule("* * * *"),
			},
			args: args{
				previousTime: NowFunc().Add(baseDuration * -1),
			},
			want: true,
		},
		{
			name: "with minutes",
			config: &RepositoryConfig{
				StartSchedule: mustParseSchedule("* * * *"),
			},
			args: args{
				previousTime: NowFunc().Add(baseDuration * -1).Add(time.Minute),
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.config.ShouldStartOn(tt.args.previousTime); got != tt.want {
				t.Errorf("RepositoryConfig.ShouldStartOn() = %v, want %v", got, tt.want)
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
	parsed, err := parser.Parse(repr)
	if err != nil {
		panic(err)
	}
	return &CronSchedule{
		Repr: repr,
		Spec: parsed,
	}
}
