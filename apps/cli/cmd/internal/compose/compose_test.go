package compose

import (
	"testing"
)

func TestParseServicesJSONLines(t *testing.T) {

	// docker compose v2 emits one object per line.
	output := `{"Name":"docksight-postgres","Service":"postgres","State":"running","Health":"healthy","ExitCode":0}
{"Name":"docksight-redis","Service":"redis","State":"running","Health":"healthy","ExitCode":0}
{"Name":"docksight-web","Service":"web","State":"running","Health":"","ExitCode":0}`

	services, err := parseServices([]byte(output))

	if err != nil {
		t.Fatal(err)
	}

	if len(services) != 3 {
		t.Fatalf("got %d services, want 3", len(services))
	}

	if services[0].Service != "postgres" || services[0].Health != "healthy" {
		t.Fatalf("unexpected first service: %+v", services[0])
	}
}

func TestParseServicesJSONArray(t *testing.T) {

	// Older releases emit a single array.
	output := `[{"Name":"docksight-redis","Service":"redis","State":"running","Health":"healthy"}]`

	services, err := parseServices([]byte(output))

	if err != nil {
		t.Fatal(err)
	}

	if len(services) != 1 || services[0].Service != "redis" {
		t.Fatalf("unexpected services: %+v", services)
	}
}

func TestParseServicesEmpty(t *testing.T) {

	services, err := parseServices([]byte("  \n"))

	if err != nil {
		t.Fatal(err)
	}

	if len(services) != 0 {
		t.Fatalf("expected no services, got %+v", services)
	}
}

func TestParseServicesMalformed(t *testing.T) {

	if _, err := parseServices([]byte("not json")); err == nil {
		t.Fatal("expected an error for unparseable output")
	}
}

func TestReady(t *testing.T) {

	cases := []struct {
		name    string
		service Service
		want    bool
	}{
		{
			name:    "running with a passing healthcheck",
			service: Service{State: "running", Health: "healthy"},
			want:    true,
		},
		{
			// web and nginx declare no healthcheck.
			name:    "running without a healthcheck",
			service: Service{State: "running"},
			want:    true,
		},
		{
			// The window where postgres accepts no connections yet.
			name:    "running but still starting",
			service: Service{State: "running", Health: "starting"},
			want:    false,
		},
		{
			name:    "running but unhealthy",
			service: Service{State: "running", Health: "unhealthy"},
			want:    false,
		},
		{
			name:    "restarting",
			service: Service{State: "restarting"},
			want:    false,
		},
		{
			name:    "exited",
			service: Service{State: "exited", ExitCode: 1},
			want:    false,
		},
	}

	for _, testCase := range cases {

		t.Run(testCase.name, func(t *testing.T) {

			if got := testCase.service.Ready(); got != testCase.want {
				t.Fatalf("Ready() = %v, want %v", got, testCase.want)
			}
		})
	}
}

func TestFailed(t *testing.T) {

	cases := []struct {
		name    string
		service Service
		want    bool
	}{
		{
			name:    "crashed",
			service: Service{State: "exited", ExitCode: 1},
			want:    true,
		},
		{
			name:    "dead",
			service: Service{State: "dead"},
			want:    true,
		},
		{
			// A one-shot migration container that finished its job.
			name:    "exited cleanly",
			service: Service{State: "exited", ExitCode: 0},
			want:    false,
		},
		{
			// Crash-looping: still has a chance to settle.
			name:    "restarting",
			service: Service{State: "restarting"},
			want:    false,
		},
		{
			name:    "running",
			service: Service{State: "running", Health: "healthy"},
			want:    false,
		},
	}

	for _, testCase := range cases {

		t.Run(testCase.name, func(t *testing.T) {

			if got := testCase.service.Failed(); got != testCase.want {
				t.Fatalf("Failed() = %v, want %v", got, testCase.want)
			}
		})
	}
}

func TestDescribe(t *testing.T) {

	withHealth := Service{Service: "postgres", State: "running", Health: "unhealthy"}

	if got := withHealth.Describe(); got != "postgres: running (unhealthy)" {
		t.Fatalf("got %q", got)
	}

	withoutHealth := Service{Service: "web", State: "running"}

	if got := withoutHealth.Describe(); got != "web: running" {
		t.Fatalf("got %q", got)
	}

	// The exit code is the first thing you need when a service crashes.
	crashed := Service{Service: "server", State: "exited", ExitCode: 3}

	if got := crashed.Describe(); got != "server: exited (code 3)" {
		t.Fatalf("got %q", got)
	}
}
