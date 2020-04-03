package model

import (
	"testing"
	"time"
)

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
				StartSchedule: "* * * *",
			},
			args: args{
				previousTime: NowFunc().Add(baseDuration * -1),
			},
			want: true,
		},
		{
			name: "with minutes",
			config: &RepositoryConfig{
				StartSchedule: "* * * *",
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
