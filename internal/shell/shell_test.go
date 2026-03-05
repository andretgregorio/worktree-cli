package shell

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestRunSetupCommands(t *testing.T) {
	dir := t.TempDir()

	t.Run("runs commands in order", func(t *testing.T) {
		commands := []string{
			"echo hello > file1.txt",
			"echo world > file2.txt",
		}

		err := RunSetupCommands(dir, commands)
		if err != nil {
			t.Fatalf("RunSetupCommands() error: %v", err)
		}

		data1, _ := os.ReadFile(filepath.Join(dir, "file1.txt"))
		if string(bytes.TrimSpace(data1)) != "hello" {
			t.Errorf("file1.txt = %q, want %q", string(data1), "hello\n")
		}

		data2, _ := os.ReadFile(filepath.Join(dir, "file2.txt"))
		if string(bytes.TrimSpace(data2)) != "world" {
			t.Errorf("file2.txt = %q, want %q", string(data2), "world\n")
		}
	})

	t.Run("fails on bad command", func(t *testing.T) {
		commands := []string{"false"}
		err := RunSetupCommands(dir, commands)
		if err == nil {
			t.Fatal("expected error for failing command, got nil")
		}
	})

	t.Run("empty commands succeeds", func(t *testing.T) {
		err := RunSetupCommands(dir, nil)
		if err != nil {
			t.Fatalf("RunSetupCommands(nil) error: %v", err)
		}
	})
}

func TestPrintCdMarker(t *testing.T) {
	output := captureStdout(func() {
		PrintCdMarker("/some/path")
	})
	want := "__WT_CD__:/some/path\n"
	if output != want {
		t.Errorf("PrintCdMarker() output = %q, want %q", output, want)
	}
}

func TestPrintEnvExports(t *testing.T) {
	output := captureStdout(func() {
		PrintEnvExports(map[string]string{"FOO": "bar"})
	})
	want := "__WT_ENV__:FOO=bar\n"
	if output != want {
		t.Errorf("PrintEnvExports() output = %q, want %q", output, want)
	}
}

func captureStdout(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	out, _ := io.ReadAll(r)
	return string(out)
}
