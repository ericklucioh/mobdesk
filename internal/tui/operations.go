package tui

import "strings"

func (m Model) renderOperation() string {
	title, steps := operationSteps(m.operation)
	var builder strings.Builder
	builder.WriteString(tagStyle.Render("MOBDESK") + "\n" + titleStyle.Render(title) + "\n" + m.spinner.View() + " " + mutedStyle.Render("em andamento") + "\n")
	builder.WriteString(m.progress.ViewAs(operationProgress(m.operation)) + "\n\n")
	stepWidth := contentWidth(m.width)
	for _, step := range steps {
		style := stepPendingStyle
		if strings.HasPrefix(step, "✓") {
			style = stepDoneStyle
		} else if strings.HasPrefix(step, "●") {
			style = stepActiveStyle
		}
		builder.WriteString(style.Copy().Width(stepWidth).Render(step) + "\n\n")
	}
	builder.WriteString(mutedStyle.Render("Operação em andamento. Aguarde..."))
	return builder.String()
}

func operationSteps(operation string) (string, []string) {
	switch operation {
	case "start":
		return "Iniciando workstation", []string{"✓  Verificando setup\n   configuração", "✓  Ativando wake-lock\n   execução protegida", "●  Iniciando servidor SSH\n   porta 8022", "○  Ubuntu\n   pronto para abrir shell"}
	case "stop":
		return "Parando workstation", []string{"●  Encerrando servidor SSH", "○  Liberando wake-lock", "○  Preservando workspace"}
	case "setup":
		return "Preparando o ambiente", []string{"✓  Diretórios Mobdesk", "●  Pacotes Termux", "○  Ubuntu persistente", "○  Workspace e SSH"}
	case "setup-upgrade":
		return "Atualizando o Termux", []string{"●  Atualizando índices", "○  Atualizando pacotes", "○  Retomando setup"}
	case "update-check":
		return "Verificando atualização", []string{"●  Consultando canal de release", "○  Validando arquitetura", "○  Comparando versões"}
	case "update":
		return "Atualizando Mobdesk", []string{"✓  Atualização encontrada", "●  Baixando binário", "○  Validando checksum", "○  Substituindo executável"}
	default:
		return "Instalando ferramenta", []string{"●  Verificando versão", "○  Atualizando índices", "○  Instalando no Ubuntu", "○  Validando instalação"}
	}
}

func operationProgress(operation string) float64 {
	switch operation {
	case "start":
		return .55
	case "stop":
		return .7
	case "setup":
		return .35
	case "setup-upgrade":
		return .45
	case "update-check":
		return .6
	case "update":
		return .7
	default:
		return .4
	}
}
