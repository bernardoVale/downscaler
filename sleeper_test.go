package main

import (
	"reflect"
	"testing"
)

func Test_sleepCandidates(t *testing.T) {
	type args struct {
		active map[string]int
		all    map[string]bool
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "Should return only ingresses that are not active",
			args: args{
				active: map[string]int{
					"foo/bar": 1, "bar/bar": 1,
				},
				all: map[string]bool{
					"bar/bar": true, "hey/go": true, "foo/bar": true,
				},
			},
			want: []string{"hey/go"},
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
