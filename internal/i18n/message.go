package i18n

import (
	"bytes"
	"text/template"
)

// MessageID identifies a translatable user-facing message.
type MessageID string

const (
	ErrorInvalidLocale     MessageID = "error.invalid_locale"
	ErrorInvalidArgs       MessageID = "error.invalid_args"
	ErrorUnknownCommand    MessageID = "error.unknown_command"
	ErrorInvalidFlag       MessageID = "error.invalid_flag"
	ErrorOperationFailed   MessageID = "error.operation_failed"
	ErrorTermuxRequired    MessageID = "error.termux_required"
	ErrorSetupIncomplete   MessageID = "error.setup_incomplete"
	ErrorUbuntuUnavailable MessageID = "error.ubuntu_unavailable"
	ErrorPTYStart          MessageID = "error.pty_start"
	ErrorCommandFailed     MessageID = "error.command_failed"
	ErrorReadOutput        MessageID = "error.read_output"

	RootShort       MessageID = "root.short"
	RootLong        MessageID = "root.long"
	RootExample     MessageID = "root.example"
	HelpUsage       MessageID = "help.usage"
	HelpExamples    MessageID = "help.examples"
	HelpCommands    MessageID = "help.commands"
	HelpFlags       MessageID = "help.flags"
	HelpGlobalFlags MessageID = "help.global_flags"

	CommandStartShort          MessageID = "command.start.short"
	CommandStartExample        MessageID = "command.start.example"
	CommandStopShort           MessageID = "command.stop.short"
	CommandStopExample         MessageID = "command.stop.example"
	CommandSetupShort          MessageID = "command.setup.short"
	CommandSetupExample        MessageID = "command.setup.example"
	CommandInstallShort        MessageID = "command.install.short"
	CommandInstallUse          MessageID = "command.install.use"
	CommandInstallExample      MessageID = "command.install.example"
	CommandUninstallShort      MessageID = "command.uninstall.short"
	CommandUninstallUse        MessageID = "command.uninstall.use"
	CommandUninstallExample    MessageID = "command.uninstall.example"
	CommandStatusShort         MessageID = "command.status.short"
	CommandStatusExample       MessageID = "command.status.example"
	CommandConfigShort         MessageID = "command.config.short"
	CommandConfigApplyShort    MessageID = "command.config_apply.short"
	CommandConfigApplyUse      MessageID = "command.config_apply.use"
	CommandConfigApplyExample  MessageID = "command.config_apply.example"
	CommandConfigRemoveShort   MessageID = "command.config_remove.short"
	CommandConfigRemoveUse     MessageID = "command.config_remove.use"
	CommandConfigRemoveExample MessageID = "command.config_remove.example"
	CommandShellShort          MessageID = "command.shell.short"
	CommandShellExample        MessageID = "command.shell.example"
	CommandVersionShort        MessageID = "command.version.short"
	CommandVersionExample      MessageID = "command.version.example"
	CommandUpdateShort         MessageID = "command.update.short"
	CommandUpdateExample       MessageID = "command.update.example"
	CommandTUIShort            MessageID = "command.tui.short"
	CommandTUIExample          MessageID = "command.tui.example"
	CommandLogsShort           MessageID = "command.logs.short"
	CommandLogsExample         MessageID = "command.logs.example"

	FlagLocaleDescription        MessageID = "flag.locale.description"
	FlagHelpDescription          MessageID = "flag.help.description"
	FlagJSONDescription          MessageID = "flag.json.description"
	FlagProgressDescription      MessageID = "flag.progress.description"
	FlagUpgradeSystemDescription MessageID = "flag.upgrade_system.description"
	FlagStrictDescription        MessageID = "flag.strict.description"
	FlagCheckDescription         MessageID = "flag.check.description"
	FlagMockDescription          MessageID = "flag.mock.description"
	FlagMockScenarioDescription  MessageID = "flag.mock_scenario.description"
	FlagLogsNameDescription      MessageID = "flag.logs_name.description"
	FlagLogsLinesDescription     MessageID = "flag.logs_lines.description"
	ValidationExactArgs          MessageID = "validation.exact_args"
	ValidationNoArgs             MessageID = "validation.no_args"

	OutputStartWarning     MessageID = "output.start.warning"
	OutputStartCompleted   MessageID = "output.start.completed"
	OutputStartLocalSSH    MessageID = "output.start.local_ssh"
	OutputStartRemoteSSH   MessageID = "output.start.remote_ssh"
	OutputStartReady       MessageID = "output.start.ready"
	OutputStopStale        MessageID = "output.stop.stale"
	OutputStopAlready      MessageID = "output.stop.already"
	OutputStopCompleted    MessageID = "output.stop.completed"
	OutputSetupCompleted   MessageID = "output.setup.completed"
	OutputSetupBase        MessageID = "output.setup.base"
	OutputSetupSSH         MessageID = "output.setup.ssh"
	OutputInstallInstalled MessageID = "output.install.installed"
	OutputInstallAlready   MessageID = "output.install.already"
	OutputUninstallRemoved MessageID = "output.uninstall.removed"
	OutputUninstallAlready MessageID = "output.uninstall.already"
	OutputUpdateCurrent    MessageID = "output.update.current"
	OutputUpdateAvailable  MessageID = "output.update.available"
	OutputUpdateUpdated    MessageID = "output.update.updated"
	OutputVersion          MessageID = "output.version"
	OutputLogsEmpty        MessageID = "output.logs.empty"
	OutputLogsNameEmpty    MessageID = "output.logs.name_empty"
	OutputLogsLabel        MessageID = "output.logs.label"
	OutputLogsError        MessageID = "output.logs.error"
	OutputLogsMissing      MessageID = "output.logs.missing"
	OutputLogsContentEmpty MessageID = "output.logs.content_empty"
	OutputConfigApplied    MessageID = "output.config.applied"
	OutputConfigRemoved    MessageID = "output.config.removed"
	OutputConfigFailed     MessageID = "output.config.failed"

	LocaleEnglishName      MessageID = "locale.english_name"
	LocalePortugueseBRName MessageID = "locale.portuguese_br_name"
	ErrorMissingMessage    MessageID = "error.missing_message"
)

var requiredMessageIDs = []MessageID{
	ErrorInvalidLocale,
	ErrorInvalidArgs,
	ErrorUnknownCommand,
	ErrorInvalidFlag,
	ErrorOperationFailed,
	ErrorTermuxRequired,
	ErrorSetupIncomplete,
	ErrorUbuntuUnavailable,
	ErrorPTYStart,
	ErrorCommandFailed,
	ErrorReadOutput,
	RootShort,
	RootLong,
	RootExample,
	HelpUsage,
	HelpExamples,
	HelpCommands,
	HelpFlags,
	HelpGlobalFlags,
	CommandStartShort,
	CommandStartExample,
	CommandStopShort,
	CommandStopExample,
	CommandSetupShort,
	CommandSetupExample,
	CommandInstallShort,
	CommandInstallUse,
	CommandInstallExample,
	CommandUninstallShort,
	CommandUninstallUse,
	CommandUninstallExample,
	CommandStatusShort,
	CommandStatusExample,
	CommandConfigShort,
	CommandConfigApplyShort,
	CommandConfigApplyUse,
	CommandConfigApplyExample,
	CommandConfigRemoveShort,
	CommandConfigRemoveUse,
	CommandConfigRemoveExample,
	CommandShellShort,
	CommandShellExample,
	CommandVersionShort,
	CommandVersionExample,
	CommandUpdateShort,
	CommandUpdateExample,
	CommandTUIShort,
	CommandTUIExample,
	CommandLogsShort,
	CommandLogsExample,
	FlagLocaleDescription,
	FlagHelpDescription,
	FlagJSONDescription,
	FlagProgressDescription,
	FlagUpgradeSystemDescription,
	FlagStrictDescription,
	FlagCheckDescription,
	FlagMockDescription,
	FlagMockScenarioDescription,
	FlagLogsNameDescription,
	FlagLogsLinesDescription,
	ValidationExactArgs,
	ValidationNoArgs,
	OutputStartWarning,
	OutputStartCompleted,
	OutputStartLocalSSH,
	OutputStartRemoteSSH,
	OutputStartReady,
	OutputStopStale,
	OutputStopAlready,
	OutputStopCompleted,
	OutputSetupCompleted,
	OutputSetupBase,
	OutputSetupSSH,
	OutputInstallInstalled,
	OutputInstallAlready,
	OutputUninstallRemoved,
	OutputUninstallAlready,
	OutputUpdateCurrent,
	OutputUpdateAvailable,
	OutputUpdateUpdated,
	OutputVersion,
	OutputLogsEmpty,
	OutputLogsNameEmpty,
	OutputLogsLabel,
	OutputLogsError,
	OutputLogsMissing,
	OutputLogsContentEmpty,
	OutputConfigApplied,
	OutputConfigRemoved,
	OutputConfigFailed,
	LocaleEnglishName,
	LocalePortugueseBRName,
	ErrorMissingMessage,
}

// Localizer renders messages from one immutable locale catalog.
type Localizer struct {
	Locale   Locale
	messages map[MessageID]string
}

// New returns a localizer for locale. Catalogs are embedded in the binary and
// validated during package initialization.
func New(locale Locale) Localizer {
	if locale != LocalePTBR {
		locale = LocaleENUS
	}
	return Localizer{Locale: locale, messages: catalogs[locale]}
}

// Text renders a message with optional template data. Message templates use
// Go's text/template syntax so translations can reorder named values safely.
func (l Localizer) Text(id MessageID, data any) string {
	message, ok := l.messages[id]
	if !ok {
		message = catalogs[LocaleENUS][ErrorMissingMessage]
		data = map[string]any{"ID": id}
	}
	if data == nil {
		return message
	}
	templateValue, err := template.New(string(id)).Option("missingkey=error").Parse(message)
	if err != nil {
		return message
	}
	var output bytes.Buffer
	if err := templateValue.Execute(&output, data); err != nil {
		return message
	}
	return output.String()
}

// RequiredMessageIDs returns a copy of the IDs that every catalog must contain.
func RequiredMessageIDs() []MessageID {
	return append([]MessageID(nil), requiredMessageIDs...)
}
