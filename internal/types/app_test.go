package types

import (
	"testing"
)

func TestApp_GetNamespace(t *testing.T) {
	tests := []struct {
		name string
		i    App
		want string
	}{
		{
			name: "Get App",
			i:    App("foo/bar"),
			want: "foo",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.i.Namespace(); got != tt.want {
				t.Errorf("App.GetNamespace() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestApp_GetName(t *testing.T) {
	tests := []struct {
		name string
		i    App
		want string
	}{
		{
			name: "Test getname",
			i:    App("bar/biz"),
			want: "biz",
		},
		{
			name: "Should return empty instead of index err",
			i:    App("foo"),
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.i.Name(); got != tt.want {
				t.Errorf("App.Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestApp_Key(t *testing.T) {
	tests := []struct {
		name string
		app  App
		want string
	}{
		{
			name: "normal usecase",
			app:  App("foo/bar"),
			want: "downscaler:foo:bar",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.app.Key(); got != tt.want {
				t.Errorf("App.Key() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewApp(t *testing.T) {
	type args struct {
		queue string
	}
	tests := []struct {
		name    string
		args    args
		want    App
		wantErr bool
	}{
		{
			name:    "Test new App",
			args:    args{"sleeping:foo:bar"},
			want:    App("foo/bar"),
			wantErr: false,
		},
		{
			name:    "Test new App error",
			args:    args{"foo:bar"},
			want:    App(""),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewApp(tt.args.queue)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewApp() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("NewApp() = %v, want %v", got, tt.want)
			}
		})
	}
}
