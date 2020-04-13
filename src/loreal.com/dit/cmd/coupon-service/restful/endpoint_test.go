package restful

import (
	"reflect"
	"testing"
)

func Test_trimURIPrefix(t *testing.T) {
	type args struct {
		uri     string
		stopTag string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		// TODO: Add test cases.
		{
			name: "case1",
			args: args{
				uri:     "/crm/api/store/",
				stopTag: "store",
			},
			want: []string{},
		},
		{
			name: "case2",
			args: args{
				uri:     "/crm/api/store/332/",
				stopTag: "store",
			},
			want: []string{"332"},
		},
		{
			name: "case3",
			args: args{
				uri:     "/crm/api/store/332/1222/11",
				stopTag: "store",
			},
			want: []string{"332", "1222", "11"},
		},
		{
			name: "case4",
			args: args{
				uri:     "/crm/api/store",
				stopTag: "store",
			},
			want: []string{},
		},
		{
			name: "case5",
			args: args{
				uri:     "/crm/api/store",
				stopTag: "store1",
			},
			want: []string{"crm", "api", "store"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := trimURIPrefix(tt.args.uri, tt.args.stopTag); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("trimURIPrefix() = %v, want %v", got, tt.want)
			}
		})
	}
}
