package cobra

import (
	"fmt"

	"github.com/ericklucioh/mobdesk/internal/paths"
	"github.com/ericklucioh/mobdesk/internal/status"
)

func requireTermuxRuntime(operation string) error {
	if status.IsTermuxRuntime(paths.Current().Prefix) {
		return nil
	}
	return fmt.Errorf("%s deve ser executado no Termux; saia da sessão Ubuntu e execute o comando no host", operation)
}
