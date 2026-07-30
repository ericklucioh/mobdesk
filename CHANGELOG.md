# Changelog

All notable changes to Mobdesk are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and releases use
[Semantic Versioning](https://semver.org/).

## [Unreleased]

### Added

- Public project documentation in the root README.
- Issue templates, pull request guidance and a Code of Conduct.
- Private vulnerability reporting through `contato@ericklucioh.com`.

### Fixed

- TUI home focus now cycles through only rendered controls.
- TUI setup progress now reflects persisted setup phases instead of fixed
  completion markers.

## [0.6.0] - 2026-07-29

### Added

- English-first localization with `en-US` and `pt-BR` catalogs.
- Localized CLI, JSON presentation, services, app profiles and Bubble Tea TUI.
- Bilingual tests and documentation migration to English as the canonical
  technical language.

### Changed

- Technical documentation and repository filenames use English names.
- Locale selection is available through `--locale` and `MOBDESK_LOCALE`.

[Unreleased]: https://github.com/ericklucioh/mobdesk/compare/v0.6.0...HEAD
[0.6.0]: https://github.com/ericklucioh/mobdesk/releases/tag/v0.6.0
