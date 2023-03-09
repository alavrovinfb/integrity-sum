package configs

import (
	"reflect"
	"testing"
)

func Test_parseOpts(t *testing.T) {
	type args struct {
		opts string
	}
	tests := []struct {
		name    string
		args    args
		want    map[string][]string
		wantErr bool
	}{
		{
			name: "Parse monitoring options positive",
			args: args{opts: "nginx=/proc,/dir1,/dir2 redis=/proc,/dir3,/dir4"},
			want: map[string][]string{
				"nginx": {"/proc", "/dir1", "/dir2"},
				"redis": {"/proc", "/dir3", "/dir4"},
			},
			wantErr: false,
		},
		{
			name: "Parse monitoring options extra commas",
			args: args{opts: "nginx=,/proc,/dir1,/dir2, redis=,/proc,/dir3,/dir4,"},
			want: map[string][]string{
				"nginx": {"/proc", "/dir1", "/dir2"},
				"redis": {"/proc", "/dir3", "/dir4"},
			},
			wantErr: false,
		},
		{
			name:    "Empty options string error",
			args:    args{opts: ""},
			wantErr: true,
		},
		{
			name:    "Incorrect key value",
			args:    args{opts: "nginx=/proc,/dir1,/dir2 redis/proc,/dir3,/dir4"},
			wantErr: true,
		},
		{
			name:    "Empty path",
			args:    args{opts: "nginx=/proc,/dir1,/dir2, redis="},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseMonitoringOpts(tt.args.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseMonitoringOpts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseMonitoringOpts() got = %v, want %v", got, tt.want)
			}
		})
	}
}
