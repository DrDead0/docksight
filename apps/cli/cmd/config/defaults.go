package config

// Default returns the standard installation layout.
func Default() Config {

	return Config{
		InstallationDir: "/opt/docksight",
		DataDir:         "/var/lib/docksight",
		BinaryPath:      "/usr/local/bin/docksight",
		Port:            2002,
	}
}
