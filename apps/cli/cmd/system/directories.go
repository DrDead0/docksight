package system

import (
  "os"
  "github.com/rodriguecyber/docksight/apps/cli/cmd/internal/ui"
)

const InstallationDir = "/opt/docksight"

func CreateInstallationDirectory() error{
if err := os.MkdirAll(InstallationDir, 0755); err != nil {
return  err
}
ui.Success("Finished create Installation directory" + InstallationDir)
return  nil
}