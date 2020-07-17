package cli

import "testing"

func TestView(t *testing.T) {
	cli := NewCLI()

	cli.TestCommand("view")
}
