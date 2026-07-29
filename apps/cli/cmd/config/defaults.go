package config

func Default () Config{
	return Config{
		InstallationDir: "/opt/docksight",
		BundleDir: "/tmp/docksight/bundle",
		DataDir: "/var/lib/docksight",
		Port: 2002,
		Version: "latest",
	}
}