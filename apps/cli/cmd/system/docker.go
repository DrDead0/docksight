package system



import (
	"os/exec"
	"github.com/rodriguecyber/docksight/apps/cli/cmd/internal/ui"

)

func CheckDocker() error {
  cmd := exec.Command("docker", "--version")

  if err := cmd.Run(); err != nil {
      ui.Error("Docker not installed")
	  return nil
  }

  ui.Info("Docker installed")
  return  nil
}

func CheckDockerCompose() error {
  cmd := exec.Command("docker", "compose", "version")

  if err := cmd.Run(); err != nil {
      ui.Error("Docker compose not found")
	  return nil
  }

  ui.Info("Docker Compose installed")
  return  nil
}