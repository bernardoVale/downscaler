package main

import (
	"reflect"
	"testing"
)

func Test_sleepCandidates(t *testing.T) {
	type args struct {
		active map[string]int
		all    map[string]string
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "Should return only ingresses that are not active",
			args: args{
				active: map[string]int{
					"foo/bar": 1, "bar/bar": 1,
				},
				all: map[string]string{
					"bar/bar": "bar/bar", "hey/go": "hey/go", "foo/bar": "foo/bar",
				},
			},
			want: map[string]string{"hey/go": "hey/go"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sleepCandidates(tt.args.active, tt.args.all); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("sleepCandidates() = %v, want %v", got, tt.want)
			}
		})
	}
}
