package shell

import (
	"fmt"
	"os"
	"os/exec"
)

func RunSetupCommands(dir string, commands []string) error {
	for _, s := range commands {
		fmt.Fprintf(os.Stderr, "Running: %s\n", s)
		shell := os.Getenv("SHELL")
		if shell == "" {
			shell = "sh"
		}
		cmd := exec.Command(shell, "-ic", s)
		cmd.Dir = dir
		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("setup command %q failed: %w", s, err)
		}
	}
	return nil
}

func PrintCdMarker(path string) {
	fmt.Printf("__WT_CD__:%s\n", path)
}

func PrintEnvExports(env map[string]string) {
	for k, v := range env {
		fmt.Printf("__WT_ENV__:%s=%s\n", k, v)
	}
}
