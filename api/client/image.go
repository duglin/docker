package client

import (
	"fmt"
)

// CmdImage is the top-level func to deal with 'docker image' commands
func (cli *DockerCli) CmdImage(args ...string) error {
	if len(args) == 0 {
		return fmt.Errorf("SHOW 'docker image' HELP")
	}
	switch {
	case args[0] == "create":
		return cli.CmdCommit(args[1:]...)

	case args[0] == "history":
		return cli.CmdHistory(args[1:]...)

	case args[0] == "list":
		return cli.CmdImages(args[1:]...)

	case args[0] == "import":
		return cli.CmdImport(args[1:]...)

	case args[0] == "inspect":
		// Need to force it to check just images
		return cli.CmdInspect(args[1:]...)

	case args[0] == "load":
		return cli.CmdLoad(args[1:]...)

	case args[0] == "pull":
		return cli.CmdPull(args[1:]...)

	case args[0] == "push":
		return cli.CmdPush(args[1:]...)

	case args[0] == "rm":
		return cli.CmdRmi(args[1:]...)

	case args[0] == "save":
		return cli.CmdSave(args[1:]...)

	case args[0] == "search":
		return cli.CmdSearch(args[1:]...)

	case args[0] == "tag":
		return cli.CmdTag(args[1:]...)
	}

	return fmt.Errorf("Unknown command %q", args[0])
}
