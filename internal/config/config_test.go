package config

import "testing"

func Test_cleanEnvVar(t *testing.T) {
	type args struct {
		tempVar string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "with double quotes",
			args: args{tempVar: "\"gordon-pn.com\""},
			want: "gordon-pn.com",
		},
		{
			name: "with single quotes",
			args: args{tempVar: "'gordon-pn.com'"},
			want: "gordon-pn.com",
		},
		{
			name: "no quotes",
			args: args{tempVar: "gordon-pn.com"},
			want: "gordon-pn.com",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cleanEnvVar(tt.args.tempVar); got != tt.want {
				t.Errorf("cleanEnvVar() = %v, want %v", got, tt.want)
			}
		})
	}
}
