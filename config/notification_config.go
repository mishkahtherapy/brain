package config

type NotificationConfig struct {
	FirebaseServiceAccountPath string
	TherapistAppBaseURL        string
}

func GetNotificationConfig() NotificationConfig {
	return NotificationConfig{
		FirebaseServiceAccountPath: MustGetEnv("BRAIN_FIREBASE_SERVICE_ACCOUNT_PATH"),
		TherapistAppBaseURL:        MustGetEnv("BRAIN_THERAPIST_APP_BASE_URL"),
	}
}
