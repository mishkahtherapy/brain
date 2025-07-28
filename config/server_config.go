package config

func IsDevelopment() bool {
	return GetEnvOrDefault("BRAIN_ENV", "production") == "development"
}
