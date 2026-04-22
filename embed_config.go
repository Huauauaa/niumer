package main

import "embed"

//go:embed configs/config.yaml configs/config.dev.yaml configs/config.prod.yaml
var embeddedAppConfig embed.FS
