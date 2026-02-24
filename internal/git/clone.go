package git

import (
	"fmt"
	"os/exec"
)

// Clone clones a git repository to the destination path.
// Uses the system git command for SSH key support.
func Clone(url, destPath string) error {
	cmd := exec.Command("git", "clone", url, destPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %s", err, string(output))
	}
	return nil
}
