package netrc

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jdx/go-netrc"
)

// Update updates the .netrc file with the given credentials.
func Update(server, login, password string, force bool) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	netrcPath := filepath.Join(home, ".netrc")

	// Create the file if it doesn't exist
	if _, err := os.Stat(netrcPath); os.IsNotExist(err) {
		file, err := os.Create(netrcPath)
		if err != nil {
			return err
		}
		file.Close()
	}

	n, err := netrc.Parse(netrcPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to parse .netrc file: %w", err)
	}

	if n == nil {
		n = &netrc.Netrc{}
	}

	machine := n.Machine(server)
	if machine != nil && !force {
		return fmt.Errorf("entry for machine '%s' already exists in .netrc file", server)
	}

	if machine == nil {
		n.AddMachine(server, login, password)
	} else {
		machine.Set("login", login)
		machine.Set("password", password)
	}

	return n.Save()
}
