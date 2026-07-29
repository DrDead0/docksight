package system

import (
	"runtime"

	"github.com/rodriguecyber/docksight/apps/cli/cmd/internal/ui"
)

func checkOs()  {
	// if runtime.GOOS != "linux" {
	// 	fmt.Println("runtime os", runtime.GOOS)
	// 	ui.Info("Docksight currently supports linux only")
	// 	return nil
	// }

	// ui.Info("Linux dected")
	// return nil
    ui.Info("check Os skipped")

}

func CheckArchitecture() error {
	switch runtime.GOARCH {
	case "amd64":
		ui.Info("✓ Architecture: amd64")
	case "arm64":
		ui.Info("✓ Architecture: arm64")
	default:
		 ui.Error("unsupported architecture: %s" +runtime.GOARCH)
		 return  nil

	}
	return nil
}
