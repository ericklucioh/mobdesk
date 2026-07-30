package cobra

import (
	"fmt"

	"github.com/ericklucioh/mobdesk/internal/i18n"
	"github.com/ericklucioh/mobdesk/internal/paths"
	"github.com/ericklucioh/mobdesk/internal/status"
)

func requireTermuxRuntime(operation string, localizers ...i18n.Localizer) error {
	if status.IsTermuxRuntime(paths.Current().Prefix) {
		return nil
	}
	return fmt.Errorf("%s", localized(localizers, i18n.ErrorTermuxRequired, map[string]any{"Operation": operation}))
}
