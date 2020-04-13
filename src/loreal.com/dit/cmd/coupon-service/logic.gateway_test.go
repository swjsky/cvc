package main

import (
	"reflect"
	"testing"

	"loreal.com/dit/utils"
)

func Test_jsonBodyfilter(t *testing.T) {
	type args struct {
		bodyObject *map[string]interface{}
		mixupFn    func(interface{}) interface{}
		keys       []string
	}
	tests := []struct {
		name string
		args args
		want map[string]interface{}
	}{
		// TODO: Add test cases.
		{
			name: "case1",
			args: args{
				bodyObject: &map[string]interface{}{
					"kt": "a",
					"k1": map[string]interface{}{
						"kt": "aa",
					},
					"k2": map[string]interface{}{
						"k21": map[string]interface{}{
							"kt":  "aaa",
							"kt1": "aaaa",
						},
					},
				},
				mixupFn: mixupString,
				keys:    []string{"kt", "kt1"},
			},
			want: map[string]interface{}{
				"kt": "*",
				"k1": map[string]interface{}{
					"kt": "**",
				},
				"k2": map[string]interface{}{
					"k21": map[string]interface{}{
						"kt":  "***",
						"kt1": "****",
					},
				},
			},
		},
		{
			name: "case2",
			args: args{
				bodyObject: &map[string]interface{}{
					"kt": "a",
					"k1": map[string]interface{}{
						"kt": "aa",
					},
					"k2": map[string]interface{}{
						"k21": map[string]interface{}{
							"kt":  "aaa",
							"kt1": "aaaa",
						},
					},
				},
				mixupFn: nil,
				keys:    []string{"kt", "kt1"},
			},
			want: map[string]interface{}{
				"k1": map[string]interface{}{},
				"k2": map[string]interface{}{
					"k21": map[string]interface{}{},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBodyfilter(tt.args.bodyObject, tt.args.mixupFn, tt.args.keys...)
			if !reflect.DeepEqual(*tt.args.bodyObject, tt.want) {
				t.Error("\n\nwant:\n", utils.MarshalJSON(tt.want, true), "\ngot:\n", utils.MarshalJSON(*tt.args.bodyObject, true))
			}
		})
	}
}
