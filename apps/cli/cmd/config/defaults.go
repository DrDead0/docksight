package config

func Default () Config{
	return Config{
		InstallationDir: "/opt/docksight",
		DataDir: "/var/lib/docksight",
		Port: 2002,
		Version: "latest",
	}
}