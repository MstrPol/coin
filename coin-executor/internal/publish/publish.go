package publish

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// PushImage pushes a built image ref (docker CLI fallback for non-buildkit engines).
func PushImage(ref string) error {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return fmt.Errorf("empty image ref")
	}
	if _, err := exec.LookPath("docker"); err == nil {
		cmd := exec.Command("docker", "push", ref)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
	if _, err := exec.LookPath("buildctl"); err == nil {
		cmd := exec.Command("buildctl", "build", "--output", "type=image,name="+ref+",push=true")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
	return fmt.Errorf("no docker or buildctl available to push %s", ref)
}
