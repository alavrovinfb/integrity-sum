package data

import (
	"io"
	"reflect"
	"testing"
)

func TestFileStorage_parseRecord(t *testing.T) {
	type fields struct {
		r io.Reader
	}
	type args struct {
		rec string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *HashDataOutput
		wantErr bool
	}{
		{
			name: "parse record test",
			fields: fields{
				r: nil,
			},
			args: args{rec: "f37852d0113de30fa6bfc3d9b180ef99383c06739530dd482a8538503afd5a58  etc/nginx/fastcgi_params"},
			want: &HashDataOutput{
				Hash:         "f37852d0113de30fa6bfc3d9b180ef99383c06739530dd482a8538503afd5a58",
				FullFileName: "etc/nginx/fastcgi_params",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := &ChecksumsReader{
				r: tt.fields.r,
			}
			got, err := fs.parseRecord(tt.args.rec)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseRecord() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseRecord() got = %v, want %v", got, tt.want)
			}
		})
	}
}
