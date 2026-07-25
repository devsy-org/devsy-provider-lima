package runtime

import "testing"

func TestFromName(t *testing.T) {
	cases := []struct {
		in      string
		want    string
		wantErr bool
	}{
		{in: "", want: Docker},
		{in: Docker, want: Docker},
		{in: Podman, want: Podman},
		{in: Nerdctl, want: Nerdctl},
		{in: "bogus", wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			rt, err := FromName(tc.in)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error for %q", tc.in)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if rt.Name() != tc.want {
				t.Errorf("Name = %q, want %q", rt.Name(), tc.want)
			}
		})
	}
}

// TestProvisioning locks in how each runtime is provisioned: docker uses a
// built-in template locator, podman/nerdctl ship embedded configs. A runtime
// must supply exactly one of the two.
func TestProvisioning(t *testing.T) {
	docker, _ := FromName(Docker)
	if docker.LimaTemplate() == "" || docker.LimaConfig() != nil {
		t.Errorf("docker: want a template locator and no embedded config")
	}
	for _, name := range []string{Podman, Nerdctl} {
		rt, _ := FromName(name)
		if rt.LimaTemplate() != "" || len(rt.LimaConfig()) == 0 {
			t.Errorf("%s: want an embedded config and empty template locator", name)
		}
	}
}
