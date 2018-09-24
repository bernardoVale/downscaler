package main

import (
	"testing"
)

func TestIngress_GetNamespace(t *testing.T) {
	tests := []struct {
		name string
		i    Ingress
		want string
	}{
		{
			name: "Get Ingress",
			i:    Ingress("foo/bar"),
			want: "foo",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.i.GetNamespace(); got != tt.want {
				t.Errorf("Ingress.GetNamespace() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIngress_GetName(t *testing.T) {
	tests := []struct {
		name string
		i    Ingress
		want string
	}{
		{
			name: "Test getname",
			i:    Ingress("bar/biz"),
			want: "biz",
		},
		{
			name: "Should return empty instead of index err",
			i:    Ingress("foo"),
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.i.GetName(); got != tt.want {
				t.Errorf("Ingress.GetName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIngress_AsQueue(t *testing.T) {
	tests := []struct {
		name string
		i    Ingress
		want string
	}{
		{
			name: "normal usecase",
			i:    Ingress("foo/bar"),
			want: "foo:bar",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.i.AsQueue(); got != tt.want {
				t.Errorf("Ingress.AsQueue() = %v, want %v", got, tt.want)
			}
		})
	}
}
