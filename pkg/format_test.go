package pkg

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFormatNum(t *testing.T) {
	type args struct {
		num     int
		bitSize int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test 1",
			args: args{
				num:     123,
				bitSize: 3,
			},
			want: "123",
		},
		{
			name: "test 2",
			args: args{
				num:     123,
				bitSize: 5,
			},
			want: "00123",
		},
		{
			name: "test 3",
			args: args{
				num:     123,
				bitSize: 2,
			},
			want: "123",
		},
		{
			name: "test 4",
			args: args{
				num:     123,
				bitSize: 1,
			},
			want: "123",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, FormatNum(tt.args.num, tt.args.bitSize), "FormatNum(%v, %v)", tt.args.num, tt.args.bitSize)
		})
	}
}
