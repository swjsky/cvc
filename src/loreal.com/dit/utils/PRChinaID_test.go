package utils

import (
	"reflect"
	"testing"
	"time"
)

func TestPRChinaID_Parse(t *testing.T) {
	tests := []struct {
		name string
		i    *PRChinaID
		want *PRChinaID
	}{
		// test cases:
		{
			name: "15 digit ID",
			i:    &PRChinaID{Code: "340204770812231"},
			want: &PRChinaID{
				Code:     "340204770812231",
				Location: 340204,
				Year:     1977,
				Month:    8,
				Day:      12,
				Birthday: time.Date(1977, 8, 12, 0, 0, 0, 0, time.Local),
				Serial:   231,
				Gender:   Male,
				CheckSum: -1,
			},
		},
		{
			name: "18 digit ID",
			i:    &PRChinaID{Code: "340204197708122318"},
			want: &PRChinaID{
				Code:     "340204197708122318",
				Location: 340204,
				Year:     1977,
				Month:    8,
				Day:      12,
				Birthday: time.Date(1977, 8, 12, 0, 0, 0, 0, time.Local),
				Serial:   231,
				Gender:   Male,
				CheckSum: 8,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.i.Parse()
			t.Logf("ID.Parse() = %v", got)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ID.Parse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPRChinaIDLocation_RegionCode(t *testing.T) {
	tests := []struct {
		name string
		l    PRChinaIDLocation
		want int
	}{
		// TODO: Add test cases.
		{
			name: "Case1",
			l:    PRChinaIDLocation(310104),
			want: 310000,
		},
		{
			name: "case2",
			l:    PRChinaIDLocation(421182),
			want: 420000,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.l.RegionCode(); got != tt.want {
				t.Errorf("PRChinaIDLocation.region() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPRChinaIDLocation_CityCode(t *testing.T) {
	tests := []struct {
		name string
		l    PRChinaIDLocation
		want int
	}{
		// TODO: Add test cases.
		{
			name: "Case1",
			l:    PRChinaIDLocation(310104),
			want: 310100,
		},
		{
			name: "case2",
			l:    PRChinaIDLocation(421182),
			want: 421100,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.l.CityCode(); got != tt.want {
				t.Errorf("PRChinaIDLocation.CityCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPRChinaIDGender_String(t *testing.T) {
	tests := []struct {
		name string
		g    PRChinaIDGender
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.g.String(); got != tt.want {
				t.Errorf("PRChinaIDGender.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPRChinaIDLocation_String(t *testing.T) {
	tests := []struct {
		name       string
		l          PRChinaIDLocation
		wantResult string
	}{
		// TODO: Add test cases.
		{
			name:       "Case1",
			l:          PRChinaIDLocation(310104),
			wantResult: "上海市徐汇区",
		},
		{
			name:       "Case2",
			l:          PRChinaIDLocation(421182),
			wantResult: "上海市徐汇区",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotResult := tt.l.String(); gotResult != tt.wantResult {
				t.Errorf("PRChinaIDLocation.String() = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}
