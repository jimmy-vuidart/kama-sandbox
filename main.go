package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"syscall"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <claude|agy|command> [arguments...]\n", os.Args[0])
		os.Exit(1)
	}
	cmd := os.Args[1]
	args := os.Args[2:]

	if cmd == "claude" || cmd == "agy" {
		if !slices.Contains(args, "--dangerously-skip-permissions") {
			args = append(args, "--dangerously-skip-permissions")
		}
	}

	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, "sbox:", err)
		os.Exit(1)
	}

	runtimeDir := os.Getenv("XDG_RUNTIME_DIR")
	if runtimeDir == "" {
		runtimeDir = fmt.Sprintf("/run/user/%d", os.Getuid())
	}
	waylandDisplay := os.Getenv("WAYLAND_DISPLAY")
	if waylandDisplay == "" {
		waylandDisplay = "wayland-0"
	}

	pwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "sbox:", err)
		os.Exit(1)
	}

	bwrapPath, err := exec.LookPath("bwrap")
	if err != nil {
		fmt.Fprintln(os.Stderr, "sbox:", err)
		os.Exit(1)
	}

	argv := bwrapArgv(cmd, args, home, runtimeDir, waylandDisplay, pwd)
	if err := syscall.Exec(bwrapPath, argv, os.Environ()); err != nil {
		fmt.Fprintln(os.Stderr, "sbox:", err)
		os.Exit(1)
	}
}

func bwrapArgv(cmd string, args []string, home, runtimeDir, waylandDisplay, pwd string) []string {
	waylandSocket := filepath.Join(runtimeDir, waylandDisplay)

	// argv[0] est le nom de la commande sandboxée, pas "bwrap" : herdr identifie
	// l'agent d'un pane par l'argv[0] de son process foreground, et ce process
	// reste bwrap (il fork la cible au lieu de l'exec en place). Sans ça l'agent
	// n'apparaît jamais dans `herdr agent list`.
	argv := []string{
		filepath.Base(cmd),
		"--ro-bind", "/usr", "/usr",
		"--ro-bind", "/lib", "/lib",
		"--ro-bind", "/lib64", "/lib64",
		"--ro-bind", "/bin", "/bin",
		"--ro-bind", "/etc", "/etc",
		"--ro-bind-try", "/opt", "/opt",
		"--ro-bind-try", "/run/systemd/resolve", "/run/systemd/resolve",
		"--proc", "/proc",
		"--dev", "/dev",
		"--tmpfs", "/tmp",
		"--bind-try", waylandSocket, waylandSocket,
		"--bind-try", filepath.Join(runtimeDir, "quickshell"), filepath.Join(runtimeDir, "quickshell"),
		"--bind-try", filepath.Join(runtimeDir, "hypr"), filepath.Join(runtimeDir, "hypr"),
		"--dir", home,
	}

	for _, path := range configMounts(home) {
		argv = append(argv, "--bind-try", path, path)
	}

	argv = append(argv,
		"--bind", pwd, pwd,
		"--chdir", pwd,
		"--die-with-parent",
		"--new-session",
		"--", cmd,
	)
	return append(argv, args...)
}

func configMounts(home string) []string {
	return []string{
		filepath.Join(home, ".claude"),
		filepath.Join(home, ".claude.json"),
		filepath.Join(home, ".config", "claude"),
		filepath.Join(home, ".gemini"),
		filepath.Join(home, ".config", "agy"),
		filepath.Join(home, ".config", "herdr"),
		filepath.Join(home, ".local", "state", "herdr"),
		filepath.Join(home, ".gitconfig"),
	}
}
