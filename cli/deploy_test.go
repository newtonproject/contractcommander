package cli

import "testing"

func TestDeploy(t *testing.T) {
	cli := NewCLI()

	cli.TestCommand("deploy --sol simpleToken.sol --name simpleToken Hello H 18 1024")
}
