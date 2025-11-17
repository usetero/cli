# Changelog

## [1.2.3](https://github.com/usetero/cli/compare/v1.2.2...v1.2.3) (2025-11-17)


### Bug Fixes

* remove unsupported folder field from brews config ([bb5291e](https://github.com/usetero/cli/commit/bb5291e92506d02ffe27dca115fa4de080ed1389))


### Reverts

* install test should fail on errors, not swallow them ([ba4233e](https://github.com/usetero/cli/commit/ba4233e8df769609e3279fad2137e695ae9e7315))

## [1.2.2](https://github.com/usetero/cli/compare/v1.2.1...v1.2.2) (2025-11-17)


### Reverts

* switch back from homebrew_casks to brews for CLI tool ([a3311ed](https://github.com/usetero/cli/commit/a3311ed78a780e1e5a647d10c56ebdb4a504e3ff))

## [1.2.1](https://github.com/usetero/cli/compare/v1.2.0...v1.2.1) (2025-11-17)


### Bug Fixes

* use 'folder' instead of deprecated 'directory' in goreleaser ([0918ba4](https://github.com/usetero/cli/commit/0918ba40780bc206fd72ff9e1639fe980d063d09))

## [1.2.0](https://github.com/usetero/cli/compare/v1.1.1...v1.2.0) (2025-11-17)


### Features

* add install script for single-line installation ([d7ce2b4](https://github.com/usetero/cli/commit/d7ce2b4933b2b0b9593400dded9edad5a1910456))
* make WorkOS client ID configurable via environment variable ([a719553](https://github.com/usetero/cli/commit/a719553606a81e2ed11a0c7e774deda90775ca5c))
* show upgrade message when version changes ([ff72bc7](https://github.com/usetero/cli/commit/ff72bc77ef9ee3c057bcbc1333c4473090e2d68c))


### Bug Fixes

* enable staging WorkOS client ID by default for development ([981bce2](https://github.com/usetero/cli/commit/981bce2db95eabb79d946a2c5c59b230071fd9b5))
* remove log output from get_latest_version function ([b253eec](https://github.com/usetero/cli/commit/b253eec6e3335bc9c8cf5ed3fd7c4525f7959ec1))


### Reverts

* remove separate test-install workflow ([69fabb6](https://github.com/usetero/cli/commit/69fabb6d4a29087f1be7b4fb3537601b7cf46293))

## [1.1.1](https://github.com/usetero/cli/compare/v1.1.0...v1.1.1) (2025-11-15)


### Bug Fixes

* remove invalid rlcp field from goreleaser config ([ebc1cc9](https://github.com/usetero/cli/commit/ebc1cc99641bff6bc82c8bdd58417ac371c65455))

## [1.1.0](https://github.com/usetero/cli/compare/v1.0.0...v1.1.0) (2025-11-15)


### Features

* add environment-based API endpoint configuration ([8d3dd46](https://github.com/usetero/cli/commit/8d3dd4608526d7da2ab572303521d3283707bf7b))

## 1.0.0 (2025-11-13)


### Features

* initial release ([f60c2b2](https://github.com/usetero/cli/commit/f60c2b237f0bd91f9c23d1c4f2acd4403c814b46))
