package config

import "testing"

func TestGetEnv(t *testing.T) {
	t.Setenv("OMDB_KEY", "TestKey")
	type args struct {
		key      string
		fallback string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "getenv test",
			args: args{
				key:      "OMDB_KEY",
				fallback: "123123",
			},
			want: "TestKey",
		},
		{
			name: "fallback test",
			args: args{
				key:      "unvalid_key",
				fallback: "123123",
			},
			want: "123123",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := GetEnv(tt.args.key, tt.args.fallback); got != tt.want {
				t.Errorf("GetEnv() = %v, want %v", got, tt.want)
			}
		})
	}
}
