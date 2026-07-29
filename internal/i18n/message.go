package i18n

import (
	"bytes"
	"errors"
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

	AppGoDescription          MessageID = "app.go.description"
	AppPythonDescription      MessageID = "app.python.description"
	AppNodeDescription        MessageID = "app.node.description"
	AppCDescription           MessageID = "app.c.description"
	AppCPPDescription         MessageID = "app.cpp.description"
	AppLuaDescription         MessageID = "app.lua.description"
	AppGitDescription         MessageID = "app.git.description"
	AppGHDescription          MessageID = "app.gh.description"
	AppTmuxDescription        MessageID = "app.tmux.description"
	AppZellijDescription      MessageID = "app.zellij.description"
	AppMicroDescription       MessageID = "app.micro.description"
	AppLazygitDescription     MessageID = "app.lazygit.description"
	AppTreeDescription        MessageID = "app.tree.description"
	AppTTTDescription         MessageID = "app.ttt.description"
	AppHtopDescription        MessageID = "app.htop.description"
	AppNcduDescription        MessageID = "app.ncdu.description"
	AppInxiDescription        MessageID = "app.inxi.description"
	AppSpeedtestDescription   MessageID = "app.speedtest_cli.description"
	AppPostingDescription     MessageID = "app.posting.description"
	AppYaziDescription        MessageID = "app.yazi.description"
	AppTuifiDescription       MessageID = "app.tuifi.description"
	AppNeovimDescription      MessageID = "app.neovim.description"
	AppOpencodeDescription    MessageID = "app.opencode_cli.description"
	AppCodexDescription       MessageID = "app.codex_cli.description"
	AppClaudeDescription      MessageID = "app.claudecode_cli.description"
	AppLeetgoDescription      MessageID = "app.leetgo.description"
	ProfileLazyVimDescription MessageID = "profile.lazyvim.description"

	ServiceInstallUnsupported MessageID = "service.install.unsupported"
	ServiceInstallDependency  MessageID = "service.install.dependency"
	ServiceInstallState       MessageID = "service.install.state"
	ServiceInstallLogs        MessageID = "service.install.logs"
	ServiceInstallRecord      MessageID = "service.install.record"
	ServiceInstallVerify      MessageID = "service.install.verify"
	ServiceInstallUpdate      MessageID = "service.install.update"
	ServiceInstallTool        MessageID = "service.install.tool"
	ServiceInstallHash        MessageID = "service.install.hash"
	ServiceInstallLock        MessageID = "service.install.lock"
	ServiceInstallWait        MessageID = "service.install.wait"
	ServiceConfigError        MessageID = "service.config.error"
	ServiceConfigProgress     MessageID = "service.config.progress"
	ServiceConfigPlugin       MessageID = "service.config.plugin"
	ServiceUninstallError     MessageID = "service.uninstall.error"
	ServiceUninstallProgress  MessageID = "service.uninstall.progress"
	ServiceUninstallDetected  MessageID = "service.uninstall.detected"
	ServiceUninstallState     MessageID = "service.uninstall.state"
	ServiceUninstallShared    MessageID = "service.uninstall.shared"
	ServiceWorkstationError   MessageID = "service.workstation.error"
	ServiceWorkstationWarning MessageID = "service.workstation.warning"
	ServiceWorkstationPID     MessageID = "service.workstation.pid"
	ServiceUpdateError        MessageID = "service.update.error"
	ServiceLogsError          MessageID = "service.logs.error"
	ServiceExecError          MessageID = "service.exec.error"

	StatusTitle             MessageID = "status.title"
	StatusSummary           MessageID = "status.summary"
	StatusUpdated           MessageID = "status.updated"
	StatusHost              MessageID = "status.host"
	StatusRuntime           MessageID = "status.runtime"
	StatusArchitecture      MessageID = "status.architecture"
	StatusWakeLock          MessageID = "status.wake_lock"
	StatusTermuxAPI         MessageID = "status.termux_api"
	StatusStorage           MessageID = "status.storage"
	StatusDeviceStorage     MessageID = "status.device_storage"
	StatusSetup             MessageID = "status.setup"
	StatusComplete          MessageID = "status.complete"
	StatusUbuntu            MessageID = "status.ubuntu"
	StatusAccessible        MessageID = "status.accessible"
	StatusWorkspace         MessageID = "status.workspace"
	StatusSSH               MessageID = "status.ssh"
	StatusPort              MessageID = "status.port"
	StatusRunning           MessageID = "status.running"
	StatusNetwork           MessageID = "status.network"
	StatusAddresses         MessageID = "status.addresses"
	StatusDevice            MessageID = "status.device"
	StatusBattery           MessageID = "status.battery"
	StatusWiFi              MessageID = "status.wifi"
	StatusInstallations     MessageID = "status.installations"
	StatusConfiguration     MessageID = "status.configurations"
	StatusAlerts            MessageID = "status.alerts"
	StatusState             MessageID = "status.state"
	StatusError             MessageID = "status.error"
	StatusLog               MessageID = "status.log"
	StatusAvailable         MessageID = "status.available"
	StatusMissing           MessageID = "status.missing"
	StatusYes               MessageID = "status.yes"
	StatusNo                MessageID = "status.no"
	StatusNone              MessageID = "status.none"
	StatusBatteryAPIMissing MessageID = "status.battery_api_missing"
	StatusConnected         MessageID = "status.connected"
	StatusDisconnected      MessageID = "status.disconnected"
	StatusOverallHealthy    MessageID = "status.overall.healthy"
	StatusOverallDegraded   MessageID = "status.overall.degraded"
	StatusOverallError      MessageID = "status.overall.error"
	StatusOverallUnknown    MessageID = "status.overall.unknown"
	StatusCheckOK           MessageID = "status.check.ok"
	StatusCheckWarning      MessageID = "status.check.warning"
	StatusCheckError        MessageID = "status.check.error"
	StatusCheckMissing      MessageID = "status.check.missing"
	StatusCheckUnknown      MessageID = "status.check.unknown"
	StatusAppState          MessageID = "status.app_state"
	StatusConfigState       MessageID = "status.config_state"
	StatusAlertCounts       MessageID = "status.alert_counts"
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
	AppGoDescription, AppPythonDescription, AppNodeDescription, AppCDescription,
	AppCPPDescription, AppLuaDescription, AppGitDescription, AppGHDescription,
	AppTmuxDescription, AppZellijDescription, AppMicroDescription, AppLazygitDescription,
	AppTreeDescription, AppTTTDescription, AppHtopDescription, AppNcduDescription,
	AppInxiDescription, AppSpeedtestDescription, AppPostingDescription, AppYaziDescription,
	AppTuifiDescription, AppNeovimDescription, AppOpencodeDescription, AppCodexDescription,
	AppClaudeDescription, AppLeetgoDescription, ProfileLazyVimDescription,
	ServiceInstallUnsupported, ServiceInstallDependency, ServiceInstallState, ServiceInstallLogs,
	ServiceInstallRecord, ServiceInstallVerify, ServiceInstallUpdate, ServiceInstallTool,
	ServiceInstallHash, ServiceInstallLock, ServiceInstallWait, ServiceConfigError,
	ServiceConfigProgress, ServiceConfigPlugin, ServiceUninstallError, ServiceUninstallProgress,
	ServiceWorkstationError, ServiceWorkstationWarning, ServiceWorkstationPID, ServiceUpdateError, ServiceLogsError,
	ServiceUninstallDetected, ServiceUninstallState, ServiceUninstallShared, ServiceExecError,
	StatusTitle, StatusSummary, StatusUpdated, StatusHost, StatusRuntime, StatusArchitecture,
	StatusWakeLock, StatusTermuxAPI, StatusStorage, StatusDeviceStorage, StatusSetup,
	StatusComplete, StatusUbuntu, StatusAccessible, StatusWorkspace, StatusSSH, StatusPort,
	StatusRunning, StatusNetwork, StatusAddresses, StatusDevice, StatusBattery, StatusWiFi,
	StatusInstallations, StatusConfiguration, StatusAlerts, StatusState, StatusError, StatusLog,
	StatusAvailable, StatusMissing, StatusYes, StatusNo, StatusNone, StatusBatteryAPIMissing,
	StatusConnected, StatusDisconnected, StatusOverallHealthy, StatusOverallDegraded,
	StatusOverallError, StatusOverallUnknown, StatusCheckOK, StatusCheckWarning, StatusCheckError,
	StatusCheckMissing, StatusCheckUnknown, StatusAppState, StatusConfigState, StatusAlertCounts,
}

// MessageError carries a stable code and translatable presentation data across
// service boundaries. Its cause remains available for diagnostics and command
// output while callers choose the locale at the presentation boundary.
type MessageError struct {
	ID    MessageID
	Code  string
	Data  map[string]any
	Cause error
}

func NewError(id MessageID, code string, data map[string]any, cause error) error {
	if data == nil {
		data = map[string]any{}
	}
	if _, ok := data["Detail"]; !ok {
		data["Detail"] = ""
	}
	if cause != nil {
		if _, ok := data["Detail"]; !ok {
			data["Detail"] = cause.Error()
		}
	}
	return &MessageError{ID: id, Code: code, Data: data, Cause: cause}
}

func (e *MessageError) Error() string {
	message := New(LocaleENUS).Text(e.ID, e.Data)
	if e.Cause != nil && !bytes.Contains([]byte(message), []byte(e.Cause.Error())) {
		message += ": " + e.Cause.Error()
	}
	return message
}
func (e *MessageError) Unwrap() error { return e.Cause }

func ErrorCode(err error) string {
	var target *MessageError
	if errors.As(err, &target) {
		return target.Code
	}
	return ""
}

func ErrorMessageID(err error) MessageID {
	var target *MessageError
	if errors.As(err, &target) {
		return target.ID
	}
	return ""
}

func (l Localizer) Error(err error) string {
	if err == nil {
		return ""
	}
	var target *MessageError
	if errors.As(err, &target) {
		message := l.Text(target.ID, target.Data)
		if target.Cause != nil && !bytes.Contains([]byte(message), []byte(target.Cause.Error())) {
			message += ": " + target.Cause.Error()
		}
		return message
	}
	return err.Error()
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
