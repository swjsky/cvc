package account

import (
	"testing"
	"time"
)

func Test_removeExpires(t *testing.T) {
	now := time.Now()
	type args struct {
		list *[]time.Time
	}
	tests := []struct {
		name           string
		args           args
		expectedLength int
	}{
		// TODO: Add test cases.
		{
			name: "case1",
			args: args{
				list: &[]time.Time{
					now.Add(time.Second * 60),
					now.Add(time.Second * 60),
					now.Add(time.Second * 60),
				},
			},
			expectedLength: 3,
		},
		{
			name: "case2",
			args: args{
				list: &[]time.Time{
					now.Add(-time.Second * 60),
					now.Add(time.Second * 60),
					now.Add(time.Second * 60),
				},
			},
			expectedLength: 2,
		},
		{
			name: "case3",
			args: args{
				list: &[]time.Time{
					now.Add(-time.Second * 60),
					now.Add(time.Second * 60),
					now.Add(-time.Second * 60),
				},
			},
			expectedLength: 1,
		},
		{
			name: "case4",
			args: args{
				list: &[]time.Time{
					now.Add(-time.Second * 60),
					now.Add(-time.Second * 60),
					now.Add(-time.Second * 60),
				},
			},
			expectedLength: 0,
		},
		{
			name: "case5",
			args: args{
				list: &[]time.Time{
					now.Add(time.Second * 60),
					now.Add(-time.Second * 60),
					now.Add(-time.Second * 60),
				},
			},
			expectedLength: 1,
		},
		{
			name: "case6",
			args: args{
				list: &[]time.Time{
					now.Add(-time.Second * 60),
					now.Add(-time.Second * 60),
					now.Add(time.Second * 60),
					now.Add(-time.Second * 60),
				},
			},
			expectedLength: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			removeExpires(tt.args.list)
			if len(*tt.args.list) != tt.expectedLength {
				t.Errorf("failed, list: %v", *tt.args.list)
			}
			for _, item := range *tt.args.list {
				if item != now.Add(time.Second*60) {
					t.Errorf("failed, list: %v", *tt.args.list)
				}
			}
		})
	}
}
