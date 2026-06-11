# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Automatic background refresh of the OpenCode Zen and Nvidia NIM model lists on
  startup. Non-blocking, runs once per launch, guarded against concurrency, and
  honors `GHOST_DISABLE_PROVIDER_AUTO_UPDATE`. The refreshed catalog is cached
  for the next session; failures leave the existing cache untouched.
- `GHOST_OPENCODE_MODELS_URL` override for the OpenCode Zen models endpoint.
- Remotely-updatable Nvidia NIM model catalog (`nvidia_nim_models.json`) fetched
  from GitHub, with an embedded offline fallback and `GHOST_NVIDIA_NIM_MODELS_URL`
  override.
- Astro-based website and documentation under `site/` (built with Bun).
- Continuous Integration: build, vet, test (with race detector and coverage),
  and `golangci-lint` on every push and pull request.
- GitHub Pages deployment workflow for the website.
- Project governance files: `CONTRIBUTING.md`, `SECURITY.md`,
  `CODE_OF_CONDUCT.md`, issue/PR templates, and `.editorconfig`.
- `LICENSE` (MIT).

### Changed
- Rewrote `README.md` to accurately describe shipped providers and models.
- Standardized the GitHub repository reference to `swadhinbiswas/Ghost`.

### Fixed
- Repository hygiene: added `.gitignore` and removed committed build artifacts,
  tooling archives, and throwaway scripts.

<!--
Release sections below are populated by GoReleaser on tagged releases.
Example:

## [1.0.0] - 2026-06-11
### Added
- Initial public release.
-->
