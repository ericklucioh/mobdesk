package tui

import "strings"

func (m Model) renderOperation() string {
	title := operationTitle(m.operation)
	var builder strings.Builder
	builder.WriteString(tagStyle.Render("MOBDESK") + "\n" + titleStyle.Render(title) + "\n\n")
	builder.WriteString(operationWaitStyle.Render("Operação em andamento"))
	if m.operationProgress != "" {
		builder.WriteString("\n" + bodyStyle.Render(m.operationProgress))
	}
	builder.WriteString("\n" + mutedStyle.Render("Aguarde a conclusão do comando."))
	return builder.String()
}

func operationTitle(operation string) string {
	switch operation {
	case "start":
		return "Iniciando workstation"
	case "stop":
		return "Parando workstation"
	case "setup":
		return "Preparando o ambiente"
	case "setup-upgrade":
		return "Atualizando o Termux"
	case "update-check":
		return "Verificando atualização"
	case "update":
		return "Atualizando Mobdesk"
	default:
		return "Instalando ferramenta"
	}
}
