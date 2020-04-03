package repo

import "testing"

func Test_keyOf(t *testing.T) {
	type args struct {
		owner string
		name  string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "OK",
			args: args{
				owner: "aereal",
				name:  "merge-chance-time",
			},
			want: "29efad4fe9346ead48cc42971c174f262f412e42c3fc0e3707249a3417b1dc10",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := keyOf(tt.args.owner, tt.args.name); got != tt.want {
				t.Errorf("keyOf() = %q, want %q", got, tt.want)
			}
		})
	}
}
