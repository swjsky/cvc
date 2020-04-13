package utils

import "testing"

func TestSanitize(t *testing.T) {
	type args struct {
		s        string
		keepList []rune
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
		{
			name: "case1",
			args: args{
				s:        "a+`~!@#$%^&-*/|\\//<>,.()[]{}vc123+_11你好啊————，。《》（）【】｛｝·~@#￥%……&*=",
				keepList: []rune{'_'},
			},
			want: "avc123_11你好啊",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Sanitize(tt.args.s, tt.args.keepList...); got != tt.want {
				t.Errorf("Sanitize() = %v, want %v", got, tt.want)
			}
		})
	}
}
