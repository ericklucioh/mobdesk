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

	OutputStartWarning          MessageID = "output.start.warning"
	OutputStartCompleted        MessageID = "output.start.completed"
	OutputStartLocalSSH         MessageID = "output.start.local_ssh"
	OutputStartRemoteSSH        MessageID = "output.start.remote_ssh"
	OutputStartReady            MessageID = "output.start.ready"
	OutputStopStale             MessageID = "output.stop.stale"
	OutputStopAlready           MessageID = "output.stop.already"
	OutputStopCompleted         MessageID = "output.stop.completed"
	OutputSetupCompleted        MessageID = "output.setup.completed"
	OutputSetupBase             MessageID = "output.setup.base"
	OutputSetupSSH              MessageID = "output.setup.ssh"
	OutputInstallInstalled      MessageID = "output.install.installed"
	OutputInstallAlready        MessageID = "output.install.already"
	OutputInstallStorageWarning MessageID = "output.install.storage_warning"
	OutputUninstallRemoved      MessageID = "output.uninstall.removed"
	OutputUninstallAlready      MessageID = "output.uninstall.already"
	OutputUpdateCurrent         MessageID = "output.update.current"
	OutputUpdateAvailable       MessageID = "output.update.available"
	OutputUpdateUpdated         MessageID = "output.update.updated"
	OutputVersion               MessageID = "output.version"
	OutputLogsEmpty             MessageID = "output.logs.empty"
	OutputLogsNameEmpty         MessageID = "output.logs.name_empty"
	OutputLogsLabel             MessageID = "output.logs.label"
	OutputLogsError             MessageID = "output.logs.error"
	OutputLogsMissing           MessageID = "output.logs.missing"
	OutputLogsContentEmpty      MessageID = "output.logs.content_empty"
	OutputConfigApplied         MessageID = "output.config.applied"
	OutputConfigRemoved         MessageID = "output.config.removed"
	OutputConfigFailed          MessageID = "output.config.failed"

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
	ServiceInstallRepair      MessageID = "service.install.repair"
	ServiceInstallUpdate      MessageID = "service.install.update"
	ServiceInstallTool        MessageID = "service.install.tool"
	ServiceInstallHash        MessageID = "service.install.hash"
	ServiceInstallLock        MessageID = "service.install.lock"
	ServiceInstallWait        MessageID = "service.install.wait"
	ServiceInstallStorage     MessageID = "service.install.storage"
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

	TUIBrand                       MessageID = "tui.brand"
	TUIStateStopped                MessageID = "tui.state.stopped"
	TUIStateRunning                MessageID = "tui.state.running"
	TUIStateOn                     MessageID = "tui.state.on"
	TUIStateOff                    MessageID = "tui.state.off"
	TUIHeaderHome                  MessageID = "tui.header.home"
	TUIHeaderClose                 MessageID = "tui.header.close"
	TUIFooterScroll                MessageID = "tui.footer.scroll"
	TUIFooterFocus                 MessageID = "tui.footer.focus"
	TUIFooterAct                   MessageID = "tui.footer.act"
	TUIFooterRefresh               MessageID = "tui.footer.refresh"
	TUIFooterQuit                  MessageID = "tui.footer.quit"
	TUIHomeTag                     MessageID = "tui.home.tag"
	TUIHomeRemoteTitle             MessageID = "tui.home.remote_title"
	TUIHomeRemoteBody              MessageID = "tui.home.remote_body"
	TUIHomeStatusTitle             MessageID = "tui.home.status_title"
	TUIHomeStatusDetail            MessageID = "tui.home.status_detail"
	TUIHomeShellTitle              MessageID = "tui.home.shell_title"
	TUIHomeShellDetail             MessageID = "tui.home.shell_detail"
	TUIHomeWorkstationTitle        MessageID = "tui.home.workstation_title"
	TUIHomeStatusLabel             MessageID = "tui.home.status_label"
	TUIHomeStart                   MessageID = "tui.home.start"
	TUIHomeStop                    MessageID = "tui.home.stop"
	TUIHomeSSHAccess               MessageID = "tui.home.ssh_access"
	TUIHomeSetupTitle              MessageID = "tui.home.setup_title"
	TUIHomeSetupDetail             MessageID = "tui.home.setup_detail"
	TUIHomeAppsTitle               MessageID = "tui.home.apps_title"
	TUIHomeAppsDetail              MessageID = "tui.home.apps_detail"
	TUIHomeShellCardTitle          MessageID = "tui.home.shell_card_title"
	TUIHomeShellCardDetail         MessageID = "tui.home.shell_card_detail"
	TUIHomeSystemTitle             MessageID = "tui.home.system_title"
	TUIHomeSystemDetail            MessageID = "tui.home.system_detail"
	TUIStatusTag                   MessageID = "tui.status.tag"
	TUIStatusTitle                 MessageID = "tui.status.title"
	TUIStatusLoading               MessageID = "tui.status.loading"
	TUIStatusLoadingDetail         MessageID = "tui.status.loading_detail"
	TUIStatusHost                  MessageID = "tui.status.host"
	TUIStatusRuntime               MessageID = "tui.status.runtime"
	TUIStatusAndroidHost           MessageID = "tui.status.android_host"
	TUIStatusUnknownArchitecture   MessageID = "tui.status.unknown_architecture"
	TUIStatusRemoteHost            MessageID = "tui.status.remote_host"
	TUIStatusUbuntuDetail          MessageID = "tui.status.ubuntu_detail"
	TUIStatusRemoteUbuntuDetail    MessageID = "tui.status.remote_ubuntu_detail"
	TUIStatusSSHHost               MessageID = "tui.status.ssh_host"
	TUIStatusResources             MessageID = "tui.status.resources"
	TUIStatusFreeBattery           MessageID = "tui.status.free_battery"
	TUIStatusInstallations         MessageID = "tui.status.installations"
	TUIStatusAlerts                MessageID = "tui.status.alerts"
	TUIStatusAlertsShort           MessageID = "tui.status.alerts_short"
	TUIStatusRefresh               MessageID = "tui.status.refresh"
	TUIStatusRefreshShort          MessageID = "tui.status.refresh_short"
	TUIStatusBack                  MessageID = "tui.status.back"
	TUIStatusEnvironment           MessageID = "tui.status.environment"
	TUIStatusOverall               MessageID = "tui.status.overall"
	TUIStatusVerified              MessageID = "tui.status.verified"
	TUIStatusDetails               MessageID = "tui.status.details"
	TUIStatusItem                  MessageID = "tui.status.item"
	TUIStatusTableState            MessageID = "tui.status.table_state"
	TUIStatusArchitecture          MessageID = "tui.status.architecture"
	TUIStatusWorkspace             MessageID = "tui.status.workspace"
	TUIStatusSSHPort               MessageID = "tui.status.ssh_port"
	TUIStatusWakeLock              MessageID = "tui.status.wake_lock"
	TUIStatusBattery               MessageID = "tui.status.battery"
	TUIStatusWiFi                  MessageID = "tui.status.wifi"
	TUIStatusAvailable             MessageID = "tui.status.available"
	TUIStatusInactive              MessageID = "tui.status.inactive"
	TUIStatusYes                   MessageID = "tui.status.yes"
	TUIStatusNo                    MessageID = "tui.status.no"
	TUIStatusConnected             MessageID = "tui.status.connected"
	TUIStatusDisconnected          MessageID = "tui.status.disconnected"
	TUIStatusSSHRunning            MessageID = "tui.status.ssh_running"
	TUIStatusSSHStopped            MessageID = "tui.status.ssh_stopped"
	TUIStatusNetworkUnavailable    MessageID = "tui.status.network_unavailable"
	TUIStatusInstallationsCount    MessageID = "tui.status.installations_count"
	TUIStatusAlertsCount           MessageID = "tui.status.alerts_count"
	TUIStatusBatteryNormal         MessageID = "tui.status.battery_normal"
	TUIStatusBatteryLow            MessageID = "tui.status.battery_low"
	TUISetupTag                    MessageID = "tui.setup.tag"
	TUISetupTitle                  MessageID = "tui.setup.title"
	TUISetupBody                   MessageID = "tui.setup.body"
	TUISetupContinue               MessageID = "tui.setup.continue"
	TUISetupUpgrade                MessageID = "tui.setup.upgrade"
	TUISetupDirectories            MessageID = "tui.setup.directories"
	TUISetupDirectoriesDetail      MessageID = "tui.setup.directories_detail"
	TUISetupPackages               MessageID = "tui.setup.packages"
	TUISetupPackagesDetail         MessageID = "tui.setup.packages_detail"
	TUISetupUbuntu                 MessageID = "tui.setup.ubuntu"
	TUISetupUbuntuDetail           MessageID = "tui.setup.ubuntu_detail"
	TUISetupWorkspace              MessageID = "tui.setup.workspace"
	TUISetupWorkspaceDetail        MessageID = "tui.setup.workspace_detail"
	TUISetupAdvanced               MessageID = "tui.setup.advanced"
	TUISetupAdvancedTitle          MessageID = "tui.setup.advanced_title"
	TUISetupAdvancedBody           MessageID = "tui.setup.advanced_body"
	TUISetupAdvancedHint           MessageID = "tui.setup.advanced_hint"
	TUIToolsTag                    MessageID = "tui.tools.tag"
	TUIToolsTitle                  MessageID = "tui.tools.title"
	TUIToolsBody                   MessageID = "tui.tools.body"
	TUIToolsRemoteTitle            MessageID = "tui.tools.remote_title"
	TUIToolsRemoteBody             MessageID = "tui.tools.remote_body"
	TUIToolStateInstalled          MessageID = "tui.tool.state.installed"
	TUIToolStateInstalling         MessageID = "tui.tool.state.installing"
	TUIToolStateInstall            MessageID = "tui.tool.state.install"
	TUIShellTag                    MessageID = "tui.shell.tag"
	TUIShellTitle                  MessageID = "tui.shell.title"
	TUIShellBody                   MessageID = "tui.shell.body"
	TUIShellOpen                   MessageID = "tui.shell.open"
	TUIShellOpenDetail             MessageID = "tui.shell.open_detail"
	TUIShellBack                   MessageID = "tui.shell.back"
	TUIShellBackDetail             MessageID = "tui.shell.back_detail"
	TUISystemTag                   MessageID = "tui.system.tag"
	TUISystemTitle                 MessageID = "tui.system.title"
	TUISystemBody                  MessageID = "tui.system.body"
	TUISystemUpdate                MessageID = "tui.system.update"
	TUISystemCheck                 MessageID = "tui.system.check"
	TUISystemAdvanced              MessageID = "tui.system.advanced"
	TUISystemAdvancedTitle         MessageID = "tui.system.advanced_title"
	TUISystemAdvancedBody          MessageID = "tui.system.advanced_body"
	TUISystemResult                MessageID = "tui.system.result"
	TUISystemUpdateHint            MessageID = "tui.system.update_hint"
	TUISystemVersion               MessageID = "tui.system.version"
	TUISystemChannel               MessageID = "tui.system.channel"
	TUISystemPlatform              MessageID = "tui.system.platform"
	TUISystemCurrent               MessageID = "tui.system.current"
	TUISystemAvailable             MessageID = "tui.system.available"
	TUISystemUpdated               MessageID = "tui.system.updated"
	TUISystemFailed                MessageID = "tui.system.failed"
	TUIOperationStart              MessageID = "tui.operation.start"
	TUIOperationStop               MessageID = "tui.operation.stop"
	TUIOperationSetup              MessageID = "tui.operation.setup"
	TUIOperationSetupUpgrade       MessageID = "tui.operation.setup_upgrade"
	TUIOperationUpdateCheck        MessageID = "tui.operation.update_check"
	TUIOperationUpdate             MessageID = "tui.operation.update"
	TUIOperationUninstall          MessageID = "tui.operation.uninstall"
	TUIOperationConfigApply        MessageID = "tui.operation.config_apply"
	TUIOperationConfigRemove       MessageID = "tui.operation.config_remove"
	TUIOperationInstall            MessageID = "tui.operation.install"
	TUIOperationRunning            MessageID = "tui.operation.running"
	TUIOperationWait               MessageID = "tui.operation.wait"
	TUIOperationPreparingInstall   MessageID = "tui.operation.preparing_install"
	TUIOperationCompleted          MessageID = "tui.operation.completed"
	TUIOperationVerified           MessageID = "tui.operation.verified"
	TUIOperationInstalled          MessageID = "tui.operation.installed"
	TUIConfirmationExit            MessageID = "tui.confirmation.exit"
	TUIConfirmationStop            MessageID = "tui.confirmation.stop"
	TUIConfirmationDestructive     MessageID = "tui.confirmation.destructive"
	TUIConfirmationYes             MessageID = "tui.confirmation.yes"
	TUIConfirmationNo              MessageID = "tui.confirmation.no"
	TUIActionCancelled             MessageID = "tui.action.cancelled"
	TUIHostRestriction             MessageID = "tui.host.restriction"
	TUIHostOnlyTitle               MessageID = "tui.host_only.title"
	TUIHostOnlySetup               MessageID = "tui.host_only.setup"
	TUIHostOnlyTools               MessageID = "tui.host_only.tools"
	TUIHostOnlySystem              MessageID = "tui.host_only.system"
	TUIHostOnlyHomeStatus          MessageID = "tui.host_only.home_status"
	TUIHostOnlyHomeShell           MessageID = "tui.host_only.home_shell"
	TUIPopupTag                    MessageID = "tui.popup.tag"
	TUIPopupState                  MessageID = "tui.popup.state"
	TUIPopupSource                 MessageID = "tui.popup.source"
	TUIPopupVersion                MessageID = "tui.popup.version"
	TUIPopupUsage                  MessageID = "tui.popup.usage"
	TUIPopupDependencies           MessageID = "tui.popup.dependencies"
	TUIPopupConfig                 MessageID = "tui.popup.config"
	TUIPopupPaths                  MessageID = "tui.popup.paths"
	TUIPopupPlugins                MessageID = "tui.popup.plugins"
	TUIPopupConfigUnavailable      MessageID = "tui.popup.config_unavailable"
	TUIPopupStorage                MessageID = "tui.popup.storage"
	TUIPopupStorageTotal           MessageID = "tui.popup.storage_total"
	TUIPopupStorageShort           MessageID = "tui.popup.storage_short"
	TUIPopupActions                MessageID = "tui.popup.actions"
	TUIPopupInstall                MessageID = "tui.popup.install"
	TUIPopupReinstall              MessageID = "tui.popup.reinstall"
	TUIPopupUninstall              MessageID = "tui.popup.uninstall"
	TUIPopupApplyConfig            MessageID = "tui.popup.apply_config"
	TUIPopupRemoveConfig           MessageID = "tui.popup.remove_config"
	TUIPopupClose                  MessageID = "tui.popup.close"
	TUIPopupInstallShort           MessageID = "tui.popup.install_short"
	TUIPopupUninstallShort         MessageID = "tui.popup.uninstall_short"
	TUIPopupApplyConfigShort       MessageID = "tui.popup.apply_config_short"
	TUIPopupRemoveConfigShort      MessageID = "tui.popup.remove_config_short"
	TUIPopupAlreadyInstalled       MessageID = "tui.popup.already_installed"
	TUIPopupDetectedReason         MessageID = "tui.popup.detected_reason"
	TUIPopupInstallFirst           MessageID = "tui.popup.install_first"
	TUIPopupConflict               MessageID = "tui.popup.conflict"
	TUIPopupNotApplied             MessageID = "tui.popup.not_applied"
	TUIPopupConfirm                MessageID = "tui.popup.confirm"
	TUIPopupDetected               MessageID = "tui.popup.detected"
	TUIPopupMobdesk                MessageID = "tui.popup.mobdesk"
	TUIPopupNotInstalled           MessageID = "tui.popup.not_installed"
	TUIPopupNotDetected            MessageID = "tui.popup.not_detected"
	TUIPopupAppAvailable           MessageID = "tui.popup.app_available"
	TUIPopupConfigState            MessageID = "tui.popup.config_state"
	TUIPopupAppStateAvailable      MessageID = "tui.popup.app_state_available"
	TUIPopupAppStateInstalling     MessageID = "tui.popup.app_state_installing"
	TUIPopupAppStateInstalled      MessageID = "tui.popup.app_state_installed"
	TUIPopupAppStateUninstalling   MessageID = "tui.popup.app_state_uninstalling"
	TUIPopupAppStateUninstalled    MessageID = "tui.popup.app_state_uninstalled"
	TUIPopupAppStatePartial        MessageID = "tui.popup.app_state_partial"
	TUIPopupAppStateFailed         MessageID = "tui.popup.app_state_failed"
	TUIPopupConfigUnavailableState MessageID = "tui.popup.config_state_unavailable"
	TUIPopupConfigNotApplied       MessageID = "tui.popup.config_state_not_applied"
	TUIPopupConfigApplying         MessageID = "tui.popup.config_state_applying"
	TUIPopupConfigApplied          MessageID = "tui.popup.config_state_applied"
	TUIPopupConfigRemoving         MessageID = "tui.popup.config_state_removing"
	TUIPopupConfigRemoved          MessageID = "tui.popup.config_state_removed"
	TUIPopupConfigModified         MessageID = "tui.popup.config_state_modified"
	TUIPopupConfigConflict         MessageID = "tui.popup.config_state_conflict"
	TUIPopupConfigFailed           MessageID = "tui.popup.config_state_failed"
	TUIErrorUnexpectedOperation    MessageID = "tui.error.unexpected_operation"
	TUIErrorUnexpectedStatus       MessageID = "tui.error.unexpected_status"
	TUIErrorInvalidResponse        MessageID = "tui.error.invalid_response"
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
	OutputInstallStorageWarning,
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
	ServiceInstallRecord, ServiceInstallVerify, ServiceInstallRepair, ServiceInstallUpdate, ServiceInstallTool,
	ServiceInstallHash, ServiceInstallLock, ServiceInstallWait, ServiceInstallStorage, ServiceConfigError,
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

var tuiMessageIDs = []MessageID{
	TUIBrand, TUIStateStopped, TUIStateRunning, TUIStateOn, TUIStateOff, TUIHeaderHome, TUIHeaderClose,
	TUIFooterScroll, TUIFooterFocus, TUIFooterAct, TUIFooterRefresh, TUIFooterQuit,
	TUIHomeTag, TUIHomeRemoteTitle, TUIHomeRemoteBody, TUIHomeStatusTitle, TUIHomeStatusDetail,
	TUIHomeShellTitle, TUIHomeShellDetail, TUIHomeWorkstationTitle, TUIHomeStatusLabel, TUIHomeStart,
	TUIHomeStop, TUIHomeSSHAccess, TUIHomeSetupTitle, TUIHomeSetupDetail, TUIHomeAppsTitle,
	TUIHomeAppsDetail, TUIHomeShellCardTitle, TUIHomeShellCardDetail, TUIHomeSystemTitle, TUIHomeSystemDetail,
	TUIStatusTag, TUIStatusTitle, TUIStatusLoading, TUIStatusLoadingDetail, TUIStatusHost, TUIStatusRuntime,
	TUIStatusAndroidHost, TUIStatusUnknownArchitecture, TUIStatusRemoteHost, TUIStatusUbuntuDetail,
	TUIStatusRemoteUbuntuDetail, TUIStatusSSHHost, TUIStatusResources, TUIStatusFreeBattery,
	TUIStatusInstallations, TUIStatusAlerts, TUIStatusAlertsShort, TUIStatusRefresh, TUIStatusBack,
	TUIStatusEnvironment, TUIStatusOverall, TUIStatusVerified, TUIStatusDetails, TUIStatusItem, TUIStatusRefreshShort,
	TUIStatusTableState, TUIStatusArchitecture, TUIStatusWorkspace, TUIStatusSSHPort, TUIStatusWakeLock,
	TUIStatusBattery, TUIStatusWiFi, TUIStatusAvailable, TUIStatusInactive, TUIStatusYes, TUIStatusNo,
	TUIStatusConnected, TUIStatusDisconnected, TUIStatusSSHRunning, TUIStatusSSHStopped,
	TUIStatusNetworkUnavailable, TUIStatusInstallationsCount, TUIStatusAlertsCount,
	TUIStatusBatteryNormal, TUIStatusBatteryLow,
	TUISetupTag, TUISetupTitle, TUISetupBody, TUISetupContinue, TUISetupUpgrade, TUISetupDirectories,
	TUISetupDirectoriesDetail, TUISetupPackages, TUISetupPackagesDetail, TUISetupUbuntu, TUISetupUbuntuDetail,
	TUISetupWorkspace, TUISetupWorkspaceDetail, TUISetupAdvanced, TUISetupAdvancedTitle, TUISetupAdvancedBody,
	TUISetupAdvancedHint, TUIToolsTag, TUIToolsTitle, TUIToolsBody, TUIToolsRemoteTitle, TUIToolsRemoteBody,
	TUIToolStateInstalled, TUIToolStateInstalling, TUIToolStateInstall, TUIShellTag, TUIShellTitle, TUIShellBody,
	TUIShellOpen, TUIShellOpenDetail, TUIShellBack, TUIShellBackDetail, TUISystemTag, TUISystemTitle,
	TUISystemBody, TUISystemUpdate, TUISystemCheck, TUISystemAdvanced, TUISystemAdvancedTitle,
	TUISystemAdvancedBody, TUISystemResult, TUISystemUpdateHint, TUISystemVersion, TUISystemChannel,
	TUISystemPlatform, TUISystemCurrent, TUISystemAvailable, TUISystemUpdated, TUISystemFailed,
	TUIOperationStart, TUIOperationStop, TUIOperationSetup, TUIOperationSetupUpgrade, TUIOperationUpdateCheck,
	TUIOperationUpdate, TUIOperationUninstall, TUIOperationConfigApply, TUIOperationConfigRemove,
	TUIOperationInstall, TUIOperationRunning, TUIOperationWait, TUIOperationPreparingInstall,
	TUIOperationCompleted, TUIOperationVerified, TUIOperationInstalled, TUIConfirmationExit,
	TUIConfirmationStop, TUIConfirmationDestructive, TUIConfirmationYes, TUIConfirmationNo, TUIActionCancelled,
	TUIHostRestriction, TUIHostOnlyTitle, TUIHostOnlySetup, TUIHostOnlyTools, TUIHostOnlySystem,
	TUIHostOnlyHomeStatus, TUIHostOnlyHomeShell, TUIPopupTag, TUIPopupState, TUIPopupSource, TUIPopupVersion, TUIPopupUsage,
	TUIPopupDependencies, TUIPopupConfig, TUIPopupPaths, TUIPopupPlugins, TUIPopupConfigUnavailable,
	TUIPopupStorage, TUIPopupStorageTotal, TUIPopupStorageShort, TUIPopupActions, TUIPopupInstall, TUIPopupReinstall, TUIPopupUninstall,
	TUIPopupApplyConfig, TUIPopupRemoveConfig, TUIPopupClose, TUIPopupInstallShort, TUIPopupUninstallShort,
	TUIPopupApplyConfigShort, TUIPopupRemoveConfigShort, TUIPopupAlreadyInstalled, TUIPopupDetectedReason,
	TUIPopupInstallFirst, TUIPopupConflict, TUIPopupNotApplied, TUIPopupConfirm, TUIPopupDetected, TUIPopupMobdesk,
	TUIPopupNotInstalled, TUIPopupNotDetected, TUIPopupAppAvailable, TUIPopupConfigState,
	TUIPopupAppStateAvailable, TUIPopupAppStateInstalling, TUIPopupAppStateInstalled,
	TUIPopupAppStateUninstalling, TUIPopupAppStateUninstalled, TUIPopupAppStatePartial, TUIPopupAppStateFailed,
	TUIPopupConfigUnavailableState, TUIPopupConfigNotApplied, TUIPopupConfigApplying, TUIPopupConfigApplied,
	TUIPopupConfigRemoving, TUIPopupConfigRemoved, TUIPopupConfigModified, TUIPopupConfigConflict, TUIPopupConfigFailed,
	TUIErrorUnexpectedOperation, TUIErrorUnexpectedStatus, TUIErrorInvalidResponse,
}

func init() {
	requiredMessageIDs = append(requiredMessageIDs, tuiMessageIDs...)
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
