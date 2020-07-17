package cli

import "testing"

func TestCall(t *testing.T) {
	cli := NewCLI()

	cli.TestCommand("call sum int 1 int 2")

}
