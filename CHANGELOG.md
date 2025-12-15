# Changelog

## [1.7.0](https://github.com/usetero/cli/compare/v1.6.0...v1.7.0) (2025-12-15)


### Features

* **cmd:** add tero reset subcommand ([1438789](https://github.com/usetero/cli/commit/14387892fa448cb1f03fcaa14cdedc78848fec2c))


### Bug Fixes

* **auth:** refresh token without org when --no-org is set ([1f892ba](https://github.com/usetero/cli/commit/1f892ba5b7a6ae92beef51c38f34785c70d7acec))
* **styles:** detect terminal background for theme selection ([3cf0f2b](https://github.com/usetero/cli/commit/3cf0f2b89e226281b0fdb5e1f5e75791dd0394e3))

## [1.6.0](https://github.com/usetero/cli/compare/v1.5.0...v1.6.0) (2025-12-15)


### Features

* **auth:** add org selection and switch command ([a113c52](https://github.com/usetero/cli/commit/a113c520aac0ba0b1b4600006891824fb53100c1))
* **auth:** add styles and tests to auth commands ([8ad0f3e](https://github.com/usetero/cli/commit/8ad0f3e075805b2c49ed69a17b4bc605f4a76fde))


### Bug Fixes

* **ci:** use conventional Docker Hub secret names ([c1dd141](https://github.com/usetero/cli/commit/c1dd1413bddf398502b9cac5d4d7efc2c1ea9d70))
* **ci:** use correct Docker Hub secret names ([f51b02f](https://github.com/usetero/cli/commit/f51b02f3ab14cfd11c89e033ca9f2414fc7c4d0f))

## [1.5.0](https://github.com/usetero/cli/compare/v1.4.0...v1.5.0) (2025-12-15)


### Features

* **auth:** add auth commands for API authentication ([7b7e70f](https://github.com/usetero/cli/commit/7b7e70f869e6d24629621c6ddad2651ee33491d4))
* **release:** add Docker Hub publishing ([6664122](https://github.com/usetero/cli/commit/6664122265c29c33cffbdc5c11150a02b35aefbf))
* **release:** add Scoop bucket publishing for Windows ([1a6895c](https://github.com/usetero/cli/commit/1a6895c1f80aa076d0f706949eab02729cc1fc61))


### Bug Fixes

* **ci:** don't cancel in-progress signoff on master pushes ([0ae0394](https://github.com/usetero/cli/commit/0ae0394e42ecee85a2744fb526161fed08d8110b))
* **ci:** skip signoff for release-please commits ([91ad45e](https://github.com/usetero/cli/commit/91ad45e11144a9b3fef541947955d6e5b1cde83d))
* **deps:** update go dependencies (minor/patch) ([#16](https://github.com/usetero/cli/issues/16)) ([eb4a5b5](https://github.com/usetero/cli/commit/eb4a5b5c3ecb6f1503a01973cd780fdb1fff30ed))
* **deps:** upgrade gopkg.in/yaml.v2 to v3 ([f23df38](https://github.com/usetero/cli/commit/f23df386a3f14935b06d157a53af21f26a2cd22e))

## [1.4.0](https://github.com/usetero/cli/compare/v1.3.0...v1.4.0) (2025-12-09)


### Features

* **auth:** auto-open browser when device auth starts ([a18653f](https://github.com/usetero/cli/commit/a18653fdd209c34d723059041ed37d734f1f1103))
* **auth:** auto-refresh expired access tokens ([71bceca](https://github.com/usetero/cli/commit/71bcecaf9d3235bb44258035c5b29e7b53cf596b))
* **auth:** refresh token with org scope after org creation ([8b28daf](https://github.com/usetero/cli/commit/8b28dafa5727aef8e144105a2e7073a64787bff3))
* **cli:** add --reset flag to clear preferences and auth ([c66af84](https://github.com/usetero/cli/commit/c66af840e9e568486212f090dddf83a57255cb43))
* **config:** namespace credentials and config by environment ([74f05a8](https://github.com/usetero/cli/commit/74f05a82ca45f041cd39095b295cf73531238f65))


### Bug Fixes

* **auth:** improve expired device code error message ([b50d793](https://github.com/usetero/cli/commit/b50d793ad4569d73fc082388b0006b0ae65dc2e4))
* **auth:** use body style for instruction text ([6e91c3f](https://github.com/usetero/cli/commit/6e91c3f19322fa8272b1fb6718b07a0707e45e00))
* **cli:** --reset flag continues into TUI instead of exiting ([c274ad1](https://github.com/usetero/cli/commit/c274ad181317a8f2bf8d476938036b741e86bda7))
* **client:** extract message from genqlient HTTPError ([c470a2e](https://github.com/usetero/cli/commit/c470a2ef3dee3bb40639f2a547ea8a0721e33139))
* **client:** regenerate client with workosOrganizationID field ([9e9aee3](https://github.com/usetero/cli/commit/9e9aee38715976f4540e53b1d02c4b4261216c12))
* **client:** update API client token after org-scoped refresh ([bbeac04](https://github.com/usetero/cli/commit/bbeac0414f73ea73b68996557589462ec647375c))
* **datadog:** prevent browser opening during tests ([11f3b97](https://github.com/usetero/cli/commit/11f3b9721fdacd30bee0e7718eb5e2e2cfa2eba5))
* **install:** remove nested color codes from success message ([a2201db](https://github.com/usetero/cli/commit/a2201db5597b2b4fce6245f81051179058e240a3))
* **onboarding:** add context and step counter to Datadog flow ([f882347](https://github.com/usetero/cli/commit/f88234700e61daff4998669db0b0bc9725aa4862))
* **onboarding:** clarify Datadog region question ([7f967c9](https://github.com/usetero/cli/commit/7f967c9ef560bbe564839c9930af891d3eb0bc99))
* **onboarding:** consistent text styling across screens ([0fd2dc5](https://github.com/usetero/cli/commit/0fd2dc5618fb02151c24d6a05bfc3b3b11e3686b))
* **onboarding:** correct text hierarchy - prompts white, explanations muted ([1f2630b](https://github.com/usetero/cli/commit/1f2630beef875b3e8178ad75dce1ff93743912db))
* **onboarding:** improve Datadog connection flow copy ([8beab8e](https://github.com/usetero/cli/commit/8beab8eecb3c569a6cd1849da3f8d3a9556bd074))
* **org:** refresh token with org scope when selecting existing org ([63053e4](https://github.com/usetero/cli/commit/63053e487274272052747700b31a673d3d09cb72))
* **org:** show loading state during token refresh after auto-selection ([97eaebf](https://github.com/usetero/cli/commit/97eaebfa2e8966313d37666c91a62b8b068d96af))
* **org:** show org name when auto-selecting ([a6cea1c](https://github.com/usetero/cli/commit/a6cea1ce6d6806bd49fc60da8495ed763060d81a))
* **org:** use spinner loader during token refresh ([351b8e8](https://github.com/usetero/cli/commit/351b8e8958515842bf6c79510a20733ebf9509aa))
* resolve linter errors in tests and workos ([9bb90e9](https://github.com/usetero/cli/commit/9bb90e9c38d9680e3becdd29e2e4773900c966a9))

## [1.3.0](https://github.com/usetero/cli/compare/v1.2.3...v1.3.0) (2025-12-03)


### Features

* **onboarding:** add TERO_SKIP_TO_APP env var to skip completion screen ([e948998](https://github.com/usetero/cli/commit/e948998013c54d3c9fec15af7f1f37cd22cea390))


### Bug Fixes

* **install:** remove confirmation prompt for piped installs ([7d1c86f](https://github.com/usetero/cli/commit/7d1c86fed11ff1198ebeb137e32aa1569d721b98))
* specify Formula directory in goreleaser brew config ([b64aa60](https://github.com/usetero/cli/commit/b64aa60139819804faaf24c061d57b1bb46aa45e))

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
