package tui

import "strings"
import "github.com/ericklucioh/mobdesk/internal/i18n"

func (m Model) renderOperation() string {
	title := operationTitleLocalized(m.operation, m.localizer)
	var builder strings.Builder
	builder.WriteString(tagStyle.Render(m.text(i18n.TUIBrand, nil)) + "\n" + titleStyle.Render(title) + "\n\n")
	builder.WriteString(operationWaitStyle.Render(m.text(i18n.TUIOperationRunning, nil)))
	if m.operationProgress != "" {
		builder.WriteString("\n" + bodyStyle.Render(m.operationProgress))
	}
	builder.WriteString("\n" + mutedStyle.Render(m.text(i18n.TUIOperationWait, nil)))
	return builder.String()
}

func operationTitle(operation string) string {
	return operationTitleLocalized(operation, i18n.New(i18n.LocaleENUS))
}

func operationTitleLocalized(operation string, localizer i18n.Localizer) string {
	switch operation {
	case "start":
		return localizer.Text(i18n.TUIOperationStart, nil)
	case "stop":
		return localizer.Text(i18n.TUIOperationStop, nil)
	case "setup":
		return localizer.Text(i18n.TUIOperationSetup, nil)
	case "setup-upgrade":
		return localizer.Text(i18n.TUIOperationSetupUpgrade, nil)
	case "update-check":
		return localizer.Text(i18n.TUIOperationUpdateCheck, nil)
	case "update":
		return localizer.Text(i18n.TUIOperationUpdate, nil)
	case "uninstall":
		return localizer.Text(i18n.TUIOperationUninstall, nil)
	case "config-apply":
		return localizer.Text(i18n.TUIOperationConfigApply, nil)
	case "config-remove":
		return localizer.Text(i18n.TUIOperationConfigRemove, nil)
	default:
		return localizer.Text(i18n.TUIOperationInstall, nil)
	}
}
