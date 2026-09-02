package main

import (
	"slices"
	"testing"
)

// argv[0] doit rester le nom de l'agent : c'est ce que herdr lit pour identifier
// le pane. Le remettre à "bwrap" casse la détection sans rien casser d'autre.
func TestArgv0IsAgentName(t *testing.T) {
	argv := bwrapArgv("/usr/bin/claude", []string{"--foo"}, "/home/u", "/run/user/1000", "wayland-0", "/tmp/proj")

	if argv[0] != "claude" {
		t.Fatalf("argv[0] = %q, veut %q (herdr ne détectera pas l'agent)", argv[0], "claude")
	}

	i := slices.Index(argv, "--")
	if i < 0 {
		t.Fatal("pas de séparateur -- dans les args bwrap")
	}
	if got := argv[i+1:]; !slices.Equal(got, []string{"/usr/bin/claude", "--foo"}) {
		t.Fatalf("commande sandboxée = %q, veut [/usr/bin/claude --foo]", got)
	}
}
