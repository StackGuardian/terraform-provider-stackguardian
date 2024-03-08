//go:build tools

package tools

import (
	// Documentation generation
	_ "github.com/charmbracelet/glow"
	_ "github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs"
)
