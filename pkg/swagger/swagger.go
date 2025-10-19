package swagger

import "emplacc-api/pkg/swagger/docs"

type Config struct {
	Title       string
	Description string
	Version     string
	Host        string
	BasePath    string
	Schemes     []string
}

func Configure(cfg Config) {
	if cfg.Title != "" {
		docs.SwaggerInfo.Title = cfg.Title
	}
	if cfg.Description != "" {
		docs.SwaggerInfo.Description = cfg.Description
	}
	if cfg.Version != "" {
		docs.SwaggerInfo.Version = cfg.Version
	}
	if cfg.Host != "" {
		docs.SwaggerInfo.Host = cfg.Host
	}
	if cfg.BasePath != "" {
		docs.SwaggerInfo.BasePath = cfg.BasePath
	}
	if len(cfg.Schemes) > 0 {
		docs.SwaggerInfo.Schemes = cfg.Schemes
	}
}
