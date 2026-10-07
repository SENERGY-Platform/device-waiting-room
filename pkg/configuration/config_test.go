package configuration

import (
	"io"
	"os"
	"strings"
	"testing"
)

func captureStdout(t *testing.T, f func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	orig := os.Stdout
	os.Stdout = w
	f()
	os.Stdout = orig
	w.Close()
	out, _ := io.ReadAll(r)
	return string(out)
}

func TestHandleEnvironmentVarsHidesSecrets(t *testing.T) {
	t.Setenv("POSTGRES_CONN_STR", "hunter2")
	t.Setenv("API_PORT", "8080")

	config := &Config{}
	out := captureStdout(t, func() { handleEnvironmentVars(config) })

	if config.PostgresConnStr != "hunter2" {
		t.Fatalf("secret not applied: %q", config.PostgresConnStr)
	}
	if strings.Contains(out, "hunter2") {
		t.Fatalf("secret logged: %s", out)
	}
	if !strings.Contains(out, "API_PORT") {
		t.Fatalf("non-secret not logged: %s", out)
	}
}
