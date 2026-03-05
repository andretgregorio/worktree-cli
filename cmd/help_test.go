package cmd

import (
	"os/exec"
	"strings"
	"testing"
)

func buildBinary(t *testing.T) string {
	t.Helper()
	binPath := t.TempDir() + "/wt-bin"
	cmd := exec.Command("go", "build", "-o", binPath, ".")
	cmd.Dir = ".."
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to build binary: %v\n%s", err, out)
	}
	return binPath
}

func TestHelp_Command(t *testing.T) {
	bin := buildBinary(t)

	out, err := exec.Command(bin, "help").CombinedOutput()
	if err != nil {
		t.Fatalf("wt help failed: %v\n%s", err, out)
	}

	output := string(out)
	assertContains(t, output, "wt - Git Worktree Manager")
	assertContains(t, output, "create <repo> <branch>")
	assertContains(t, output, "remove <repo> [worktree]")
	assertContains(t, output, "init")
	assertContains(t, output, "help")
	assertNotContains(t, output, "completion")
}

func TestHelp_Flag(t *testing.T) {
	bin := buildBinary(t)

	out, err := exec.Command(bin, "--help").CombinedOutput()
	if err != nil {
		t.Fatalf("wt --help failed: %v\n%s", err, out)
	}

	output := string(out)
	assertContains(t, output, "wt - Git Worktree Manager")
	assertContains(t, output, "create <repo> <branch>")
	assertContains(t, output, "init")
}

func TestHelp_NoArgs(t *testing.T) {
	bin := buildBinary(t)

	out, err := exec.Command(bin).CombinedOutput()
	if err != nil {
		t.Fatalf("wt (no args) failed: %v\n%s", err, out)
	}

	output := string(out)
	assertContains(t, output, "wt - Git Worktree Manager")
	assertContains(t, output, "create <repo> <branch>")
}

func TestHelp_ShowsCreateFlags(t *testing.T) {
	bin := buildBinary(t)

	out, _ := exec.Command(bin, "help").CombinedOutput()
	output := string(out)

	assertContains(t, output, "--base")
	assertContains(t, output, "--dir")
	assertContains(t, output, "--no-setup")
	assertContains(t, output, "--dry-run")
}

func TestHelp_ShowsExamples(t *testing.T) {
	bin := buildBinary(t)

	out, _ := exec.Command(bin, "help").CombinedOutput()
	output := string(out)

	assertContains(t, output, "wt create groups feat/new-feature")
	assertContains(t, output, "wt remove groups feat-new-feature")
	assertContains(t, output, "wt remove groups")
	assertContains(t, output, "wt init")
}

func assertContains(t *testing.T, output, expected string) {
	t.Helper()
	if !strings.Contains(output, expected) {
		t.Errorf("output missing %q, got:\n%s", expected, output)
	}
}

func assertNotContains(t *testing.T, output, unexpected string) {
	t.Helper()
	if strings.Contains(output, unexpected) {
		t.Errorf("output should not contain %q, got:\n%s", unexpected, output)
	}
}
