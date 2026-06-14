package main

import (
	"fmt"
	"strings"

	"github.com/anggasct/rippl/internal/config"
	"github.com/anggasct/rippl/internal/render"
)

var validFormats = map[string]struct{}{
	string(render.FormatText):    {},
	string(render.FormatJSON):    {},
	string(render.FormatAgent):   {},
	string(render.FormatMermaid): {},
	string(render.FormatTUI):     {},
}

func rejectAgentFormat(cfg *config.Config, command string) error {
	if strings.EqualFold(cfg.Output.Format, string(render.FormatAgent)) {
		return &config.ExitError{
			Code: 1,
			Err:  fmt.Errorf("--format agent is not supported for %s", command),
		}
	}
	return nil
}

func validateFormat(format string) error {
	if _, ok := validFormats[strings.ToLower(format)]; !ok {
		return &config.ExitError{
			Code: 1,
			Err:  fmt.Errorf("invalid output format %q", format),
		}
	}
	return nil
}
