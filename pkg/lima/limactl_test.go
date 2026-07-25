package lima

import (
	"reflect"
	"testing"
)

func TestCreateArgs(t *testing.T) {
	cases := []struct {
		name string
		p    CreateParams
		want []string
	}{
		{
			name: "with vm type",
			p: CreateParams{
				Name: "ws", Template: "template://docker-rootful",
				CPUs: "4", Memory: "8", Disk: "60", VMType: "vz",
			},
			want: []string{
				"create", flagNoTTY, "--name=ws", "--cpus=4", "--memory=8",
				"--disk=60", "--vm-type=vz", "template://docker-rootful",
			},
		},
		{
			name: "without vm type omits flag",
			p: CreateParams{
				Name: "ws", Template: "/tmp/lima.yaml",
				CPUs: "2", Memory: "4", Disk: "20",
			},
			want: []string{
				"create", flagNoTTY, "--name=ws", "--cpus=2",
				"--memory=4", "--disk=20", "/tmp/lima.yaml",
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := createArgs(tc.p)
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("createArgs() = %v, want %v", got, tc.want)
			}
			if got[len(got)-1] != tc.p.Template {
				t.Errorf("template must be the last arg, got %q", got[len(got)-1])
			}
		})
	}
}

func TestMapStatus(t *testing.T) {
	cases := []struct {
		in      string
		want    string
		wantErr bool
	}{
		{in: StatusRunning, want: StatusRunning},
		{in: StatusStopped, want: StatusStopped},
		{in: "  Running\n", want: StatusRunning},
		{in: "Starting", want: StatusBusy},
		{in: "", want: StatusBusy},
		{in: "Broken", wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got, err := mapStatus(tc.in)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error for %q", tc.in)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("mapStatus(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestParseInstanceStatus(t *testing.T) {
	out := "ws-a\tRunning\nws-b\tStopped\n"
	cases := []struct {
		name       string
		wantStatus string
		wantFound  bool
	}{
		{name: "ws-a", wantStatus: "Running", wantFound: true},
		{name: "ws-b", wantStatus: "Stopped", wantFound: true},
		{name: "ws-c", wantFound: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			status, found := parseInstanceStatus(out, tc.name)
			if found != tc.wantFound || status != tc.wantStatus {
				t.Fatalf("parseInstanceStatus(%q) = (%q, %v), want (%q, %v)",
					tc.name, status, found, tc.wantStatus, tc.wantFound)
			}
		})
	}

	t.Run("empty output is not found", func(t *testing.T) {
		if _, found := parseInstanceStatus("", "ws"); found {
			t.Fatal("expected not found for empty output")
		}
	})

	t.Run("prefix name is not a false match", func(t *testing.T) {
		if _, found := parseInstanceStatus("ws-abc\tRunning\n", "ws-ab"); found {
			t.Fatal("ws-ab must not match ws-abc")
		}
	})
}
