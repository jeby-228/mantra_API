package config

const (
	DefaultCORSOriginDev     = "http://localhost:5173"
	DefaultCORSOriginPreview = "http://localhost:4173"
)

var DefaultCORSAllowOrigins = []string{DefaultCORSOriginDev, DefaultCORSOriginPreview}
