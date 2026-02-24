# Changelog

## [1.19.0](https://github.com/usetero/cli/compare/v1.18.0...v1.19.0) (2026-02-24)


### Features

* show estimated savings in waste tab headline ([357ce17](https://github.com/usetero/cli/commit/357ce175b0820c5765414c60b7ab4d0476d82189))


### Bug Fixes

* compliance tab timeout — switch to pre-aggregated cache table ([c3197ab](https://github.com/usetero/cli/commit/c3197ab3e5c981e096f5f9c086aadf6d146ee651))

## [1.18.0](https://github.com/usetero/cli/compare/v1.17.0...v1.18.0) (2026-02-24)


### Features

* add policy card component and show tool ([568723f](https://github.com/usetero/cli/commit/568723ffc9dea51336b60068e00cc2ae7370f5d6))
* add Quality tab and enrich all analysis tabs with category descriptions ([4def4cd](https://github.com/usetero/cli/commit/4def4cdb2651fef9e725db3704eb2cd4d52dab75))


### Bug Fixes

* add spacer line between chat messages and input bar ([e05b723](https://github.com/usetero/cli/commit/e05b72386813231152dc7da271405bf4a574fe21))
* clean up orphaned messages from DB on cancel ([47205b1](https://github.com/usetero/cli/commit/47205b1f965c28513c2998ebe67eed43bd46bb54))
* clean up waste table Impact column and lower discovery thresholds ([2534d9a](https://github.com/usetero/cli/commit/2534d9a6f5239c3d91a83b2874bafe636e7596f9))
* display "$0/yr" instead of "—" for zero-value impact in waste category table ([e6b6557](https://github.com/usetero/cli/commit/e6b655769303f036c17ecbd227cb56d79c7e3e41))
* eliminate JOINs in waste category detail query using denormalized columns ([42dedd9](https://github.com/usetero/cli/commit/42dedd9c2d9a2b3ef1198d61e8781368a87d5874))
* only toggle tool/thinking blocks when clicking the header line ([936cefc](https://github.com/usetero/cli/commit/936cefc596fdd74526bbd5084d48a7dfff502f3d))
* restore category subtitle in policy header and use orange focus border ([550409c](https://github.com/usetero/cli/commit/550409cb83ef1cc61973f552622fa861b314b61d))
* tighten policy card layout — remove subtitle, rationale divider, cap examples ([de8d2fe](https://github.com/usetero/cli/commit/de8d2fe4c7b438fd3913cac3c899a5fcf8bca869))
* use pre-computed cache table for waste category statuses ([c5e453b](https://github.com/usetero/cli/commit/c5e453bd645c31cf3eba22c4487f47e84fcf18df))

## [1.17.0](https://github.com/usetero/cli/compare/v1.16.0...v1.17.0) (2026-02-20)


### Features

* add log event detail drill-down to Services tab ([0b9890c](https://github.com/usetero/cli/commit/0b9890ca825207743aae737027de44108c645774))
* selectable log events in drawer detail views submit context-specific prompts ([145c239](https://github.com/usetero/cli/commit/145c23996b3ac9ec2295529df231e62142c8fe25))


### Bug Fixes

* replace sticky userScrolled flag with AtBottom() snapshot for smoother scrolling ([930703e](https://github.com/usetero/cli/commit/930703e9c16c0a48a0dc5e990edcb29268be5835))
* show helpful guidance when all services are disabled in drawer tabs ([64fc372](https://github.com/usetero/cli/commit/64fc372fda7ed275c18eabccaf657eef9d6cb028))

## [1.16.0](https://github.com/usetero/cli/compare/v1.15.0...v1.16.0) (2026-02-19)


### Features

* add discovery progress indicators and PII observed/at-risk split ([5663eea](https://github.com/usetero/cli/commit/5663eeacc12e397157ebbc2446547ec3db596de8))
* async statusbar polling, impact_type + per-severity volumes, Impact column ([dc64f87](https://github.com/usetero/cli/commit/dc64f8755d5733db4be061849deb2e4ba94ad0e8))
* replace PII tab with comprehensive Compliance tab covering 4 categories ([8d2c568](https://github.com/usetero/cli/commit/8d2c56851d5ea9aba1773f5934ea7b7b94992471))
* show cost percentage per waste category in Est. Savings column ([8f92b53](https://github.com/usetero/cli/commit/8f92b5319f69ba35864db4733697b74305a9389d))


### Bug Fixes

* color 'at risk' compliance status with warning color to match dot ([ba88810](https://github.com/usetero/cli/commit/ba888107480e232536ba17b20170639fbf366f2c))
* **deps:** update module github.com/mattn/go-sqlite3 to v1.14.34 ([#42](https://github.com/usetero/cli/issues/42)) ([fe9ca4b](https://github.com/usetero/cli/commit/fe9ca4b57c170df94a1aa56fa354f4f8cc1171b7))
* distinguish null from zero for all measurement fields ([804721d](https://github.com/usetero/cli/commit/804721dff04c4bb960cd67bb7832fce58749b18d))
* force-refresh sync token on server auth rejection ([7ea6fe5](https://github.com/usetero/cli/commit/7ea6fe5f948b713f9e4bfeb4d68ee000acee6478))
* remove redundant error toasts from query tool ([977a4d0](https://github.com/usetero/cli/commit/977a4d00c4a82f15724d253aec9ce4eda074ab81))
* show actual error message on stream failures and fix statusbar overflow ([17ece74](https://github.com/usetero/cli/commit/17ece74a566b4b60b10a7d580ee214cdeb0c57ba))

## [1.15.0](https://github.com/usetero/cli/compare/v1.14.0...v1.15.0) (2026-02-13)


### Features

* add dedicated PII tab to statusbar ([8c12d91](https://github.com/usetero/cli/commit/8c12d919d1fea75e04a8f0e1a7e94b4cf751d52c))
* add keyboard navigation and drill-down to waste drawer ([c0d1667](https://github.com/usetero/cli/commit/c0d1667c3b5a3d1bd58486b3247869e05b1454ea))
* add set_service_enabled chat tool ([fe88058](https://github.com/usetero/cli/commit/fe88058b0ce07f537b98311df4dda91c29af25dc))
* copy selected text to clipboard with toast feedback ([bb01fe4](https://github.com/usetero/cli/commit/bb01fe4fe0f88dea53fb50cfd6f7633401293ad6))
* redesign statusbar with services and waste tabs ([ac1f245](https://github.com/usetero/cli/commit/ac1f24514369ea3953bb362867859fdef6c71420))
* show environment label in status bar for non-production envs ([4668b85](https://github.com/usetero/cli/commit/4668b85caf379d3b9afd91adad0d0df44dccc53f))
* show PII type categories in statusbar instead of field paths ([9f98c13](https://github.com/usetero/cli/commit/9f98c13c75d04590d940f44e5b2a5237e86f83a5))


### Bug Fixes

* add spacer between sync status headline and description ([469d0e3](https://github.com/usetero/cli/commit/469d0e3f320b0e5d9ee2f0706280aa4b87eb2baa))
* check rows.Err after EXPLAIN QUERY PLAN iteration ([6d4d530](https://github.com/usetero/cli/commit/6d4d530e4d0fda9c6fbaaa48b8154e492a63c119))
* clarify pending counts as policies in waste view ([8df2869](https://github.com/usetero/cli/commit/8df2869918ed99dceea4a8fdcc6902855d46531b))
* clicking outside drawer closes it and passes click through ([ac858e3](https://github.com/usetero/cli/commit/ac858e3205ddc28adeee6a1ad6e6e3bbe9b13bd9))
* code block background matches panel surface ([9bf7caa](https://github.com/usetero/cli/commit/9bf7caaf5b54eddada815c8dd676ad8b8e56062e))
* complete Chroma syntax highlighting for code blocks ([c948523](https://github.com/usetero/cli/commit/c948523f65cd93d6d2ee2e7fcca555be2aba4abb))
* **deps:** update module github.com/workos/workos-go/v4 to v6 ([#36](https://github.com/usetero/cli/issues/36)) ([5ef549f](https://github.com/usetero/cli/commit/5ef549f66c7ee4ca60c3ac4597a623d91fb530f4))
* display friendly category names in waste table ([0fb0d06](https://github.com/usetero/cli/commit/0fb0d06cf078901a7c87b2d717178b51712c9c89))
* eliminate black rectangles in markdown code blocks ([d9ef4f1](https://github.com/usetero/cli/commit/d9ef4f1ae444e920baf8e52f21774b5bc89bbbe5))
* empty state shows setup guidance when no services are enabled ([bb6520b](https://github.com/usetero/cli/commit/bb6520bd78de782301843e28052ca77529f4d4af))
* hide catalog drawer when all services are disabled ([6c17498](https://github.com/usetero/cli/commit/6c174987a768f4ed55e78c676fa2bffa0802be84))
* move clipboard.WriteAll to tea.Cmd to avoid blocking Update ([eefa102](https://github.com/usetero/cli/commit/eefa10275491923af40f71eee2d9453fa1a205a0))
* reject queries with full table scan JOINs ([12238a6](https://github.com/usetero/cli/commit/12238a64c3b4daf102052e814a0cfdf4ce915ee9))
* show org name in status bar during onboarding ([eb7e76c](https://github.com/usetero/cli/commit/eb7e76cefdfc6f4b9d0485c6013e7b9b5f48a0de))
* show org name in status bar during onboarding ([f492541](https://github.com/usetero/cli/commit/f4925416b832c51fd849b6624045e12bf7c70db1))
* show total volume in services summary, reorder statusbar segments ([ffd4e09](https://github.com/usetero/cli/commit/ffd4e09c361c4260138202d4866b52fe1c79e035))
* syntax highlighting uses distinct colors for keys vs strings ([807822c](https://github.com/usetero/cli/commit/807822c7b1646f4f00b2fc6fb6e5a57aa6da2a30))
* text highlight no longer bleeds into left border ([79c4636](https://github.com/usetero/cli/commit/79c463689bfef12ae83cf3fc31634a473de87aa6))
* tool error state collapsed by default with chevron toggle ([d1406ba](https://github.com/usetero/cli/commit/d1406ba084cc1474d59ddd56acaeb0b9ef59c442))
* use config defaults for PowerSync endpoint in generate script ([fd04fc5](https://github.com/usetero/cli/commit/fd04fc54c235dffe7129696ab38a534a8508bdce))
* validate service existence before enabling/disabling ([3256d6e](https://github.com/usetero/cli/commit/3256d6ee8d4c2cee2ab4daeac6e3300720177ebf))

## [1.14.0](https://github.com/usetero/cli/compare/v1.13.0...v1.14.0) (2026-02-11)


### Features

* stream error handling with StateFailed and DB cleanup ([a3bf473](https://github.com/usetero/cli/commit/a3bf473c4ce6a59a0e5108ca574786969ec344a0))


### Bug Fixes

* cap query tool results to prevent token limit errors ([fb6d116](https://github.com/usetero/cli/commit/fb6d1165344084143dc7de7dd5962dcbfc1f121c))
* preserve chat messages when stream errors on subsequent turns ([5e90a49](https://github.com/usetero/cli/commit/5e90a49562b474accf4c313724fc402c5486fd5b))
* single write connection prevents flaky notification tests ([92806e5](https://github.com/usetero/cli/commit/92806e58c6cc6e63655aafbdf9bb92e813d46295))

## [1.13.0](https://github.com/usetero/cli/compare/v1.12.0...v1.13.0) (2026-02-11)


### Features

* WAL mode with separate read/write connection pools ([0ffe144](https://github.com/usetero/cli/commit/0ffe1445d999f6a23e02da71ca0c0bb573be7b88))

## [1.12.0](https://github.com/usetero/cli/compare/v1.11.1...v1.12.0) (2026-02-11)


### Features

* dynamic drawer height with row clipping and +N more ([d4b7627](https://github.com/usetero/cli/commit/d4b7627fa6457dd2046b3692730d5d6ec433c53e))


### Bug Fixes

* drawer shrinks to fit content instead of filling page ([5c8f347](https://github.com/usetero/cli/commit/5c8f347cb688f5b75cd769691b9564191c8c152d))
* fill statusbar with diagonals when drawer hint is hidden ([0f76244](https://github.com/usetero/cli/commit/0f7624427ca91ee87c9aab8d34dd32b65d676557))
* hide drawer hint until data is loaded ([c869ce3](https://github.com/usetero/cli/commit/c869ce3a9f81160cfdaf2f400a15e814132bbb8a))
* prevent duplicate API streams when tools complete during turn transition ([017ae6e](https://github.com/usetero/cli/commit/017ae6e9ef55f96682674aa2822a1031f97f5629))
* show empty state when all catalog services are disabled ([99c73c1](https://github.com/usetero/cli/commit/99c73c1e45310d8135e44e83f7270c0331139619))
* suppress drawer until data is loaded ([5db3e5c](https://github.com/usetero/cli/commit/5db3e5c7acd8d0ba0f5252c3d899d0a4e02802b3))

## [1.11.1](https://github.com/usetero/cli/compare/v1.11.0...v1.11.1) (2026-02-10)


### Bug Fixes

* duplicate stream completion and cross-turn tool leaking ([3b946b0](https://github.com/usetero/cli/commit/3b946b04ff818101e6ff5aed9ccbea6ce7098847))

## [1.11.0](https://github.com/usetero/cli/compare/v1.10.0...v1.11.0) (2026-02-10)


### Features

* add detailed logging to chat client ([e10fd9d](https://github.com/usetero/cli/commit/e10fd9d416d50dfb79920df79a1f2a08c94674f0))
* add nested command palette with theme selection ([a822da8](https://github.com/usetero/cli/commit/a822da88448b60a52edf082057cc11008f9b7e76))
* display costs as yearly instead of monthly ([af8b5d4](https://github.com/usetero/cli/commit/af8b5d4471409902c3d55ca3e773aa2aea7b6d1d))
* graceful stream error handling ([12e34ea](https://github.com/usetero/cli/commit/12e34ea6aea79115cbc5a122d2c837e568c3d9c6))
* improve catalog drawer ([aaa8105](https://github.com/usetero/cli/commit/aaa8105e17b731144f8b0c9da8a954d805b6f588))
* improve message list focus UX ([199b67b](https://github.com/usetero/cli/commit/199b67b23f845c3037b8af237d016c4faf160d10))
* show query execution time in tool result title ([fa44c3b](https://github.com/usetero/cli/commit/fa44c3b6a0ca24adecaaf822072a8fd45ff29981))


### Bug Fixes

* disable strikethrough in markdown rendering ([2b8db94](https://github.com/usetero/cli/commit/2b8db94508eaeb5124279b5857bde83d058e7fe4))
* include chat endpoint in token audience ([c9dbdba](https://github.com/usetero/cli/commit/c9dbdba99c6b6f89272b4975293e36336702a488))
* increase watch test timeouts to prevent flakiness ([8d565ca](https://github.com/usetero/cli/commit/8d565ca515740ddf53b2bd3e5efab505c618277d))
* race between persist and fire tool results ([a3f3231](https://github.com/usetero/cli/commit/a3f3231d4d5e8701d23c7d3150f3d95d4ee1123a))
* reduce max sync retry backoff from 30s to 10s ([1785294](https://github.com/usetero/cli/commit/1785294c17573051dc9a007ae8ab7d4fc8b95065))
* revert hardcoded version to dev ([e34e468](https://github.com/usetero/cli/commit/e34e468a3986840571f13045b0ca10161e372ddd))
* set org and workspace on statusbar after onboarding ([b2c6c41](https://github.com/usetero/cli/commit/b2c6c419b9bb5306d954e568360c494d243daa33))
* upload assets to existing release-please release ([90aca09](https://github.com/usetero/cli/commit/90aca09f5df1a0f72c96314c968fcd67c5a90435))
* write log file to ~/.tero instead of /tmp ([d1f283e](https://github.com/usetero/cli/commit/d1f283e90d8b779ae32521e088ecbca1e848a76e))

## [1.10.0](https://github.com/usetero/cli/compare/v1.9.0...v1.10.0) (2026-02-09)


### Features

* replace goreleaser with native CGO builds on platform runners ([8a59146](https://github.com/usetero/cli/commit/8a591462ae94f559d74ef89873fab9154fd92d8e))


### Bug Fixes

* mark prerelease tags, skip homebrew update for rc tags ([4f80be8](https://github.com/usetero/cli/commit/4f80be80d60f5003ccbc345fd4867c2bb3caa2d0))
* use macos-latest with cross-compile for darwin/amd64 ([66e9dce](https://github.com/usetero/cli/commit/66e9dce3c8a471989b8e8a02ef7514606cb17fe2))

## [1.9.0](https://github.com/usetero/cli/compare/tero-v1.8.0...tero-v1.9.0) (2026-02-09)


### Features

* add chat focus toggle, table options, and AccentAlt color ([dcc1664](https://github.com/usetero/cli/commit/dcc1664cd14f8cdb8dd8a7753525df9bfa3e5546))
* add command bar, theme refactor, and testing improvements ([48e3e63](https://github.com/usetero/cli/commit/48e3e630e0b4b6603687297c07d18506629edc7b))
* add context window usage to statusbar and refactor API service inputs ([68c2401](https://github.com/usetero/cli/commit/68c240107de68dc4d82887df1279822c976f9661))
* add conversation update/delete and improve logging ([a94d762](https://github.com/usetero/cli/commit/a94d762d972748765c05e1de598b66aec7389e7f))
* add cost column to policy status category table ([95ab73d](https://github.com/usetero/cli/commit/95ab73de9c640b0478bbec0fe1156e90b7d9011e))
* add debug subcommand for diagnostics ([1d55c14](https://github.com/usetero/cli/commit/1d55c1413162a2e8ca48a2c92e527675c5ca6883))
* add environment-based API endpoint configuration ([30c440d](https://github.com/usetero/cli/commit/30c440d0e21c873251b0406a17f97a08608aa2ff))
* add install script for single-line installation ([42566fb](https://github.com/usetero/cli/commit/42566fb3d3e38215600d4f82820bc65b9a768ff1))
* add messagelist test infrastructure and tests ([c7267bd](https://github.com/usetero/cli/commit/c7267bdbe43912bc245ba11f4cf7d3140756c1b5))
* add mouse scrolling and text selection with highlighting ([0988aa7](https://github.com/usetero/cli/commit/0988aa7da937a35501bc01ce7aa1c7011ec38828))
* add PowerSync integration for local-first data sync ([4708259](https://github.com/usetero/cli/commit/4708259bee9733503079fedf04bd73c117ee0e1e))
* add SQLite change subscriptions and chat service ([d20b755](https://github.com/usetero/cli/commit/d20b755979dc40a8319e7d6f16770d162240c305))
* add statusbar drawer with tabs and right-aligned ctrl+d hint ([9b40c01](https://github.com/usetero/cli/commit/9b40c01b3995c11636cafa5251766a330a163292))
* add toast notification system for user-facing errors ([04d1d02](https://github.com/usetero/cli/commit/04d1d02f74330286894a363e7a8acff5f8c206a0))
* add tools architecture with query tool and schema generator ([a9fd4b9](https://github.com/usetero/cli/commit/a9fd4b9e5b5b7b62c0d84b5c10045bf3999c5d44))
* add upload event channel for TUI integration ([8ab23f3](https://github.com/usetero/cli/commit/8ab23f35fc4c4d321f10cd800bb859943f5733c2))
* add X-Account-ID header to API requests ([946266b](https://github.com/usetero/cli/commit/946266b38356a93f243740ab25e934f6c347e309))
* admin tool for org membership, typed token claims ([f14835b](https://github.com/usetero/cli/commit/f14835b604ce74fbedcca327822d2a8db5c2dc81))
* **auth:** add auth commands for API authentication ([4a72f7a](https://github.com/usetero/cli/commit/4a72f7af4123f6237feccee00c793e55dfc27344))
* **auth:** add org selection and switch command ([ff90d80](https://github.com/usetero/cli/commit/ff90d8099746f2dad1fffc84b9f6f7b76cbbd25c))
* **auth:** add styles and tests to auth commands ([8566a04](https://github.com/usetero/cli/commit/8566a0435158f073dc133680a6b9628c44ad4bfb))
* **auth:** auto-open browser when device auth starts ([69640c0](https://github.com/usetero/cli/commit/69640c0f9616ac7856280a8970e41bb3280c4fa3))
* **auth:** auto-refresh expired access tokens ([e0564e2](https://github.com/usetero/cli/commit/e0564e29aeeb0db2fddac9e8891eb1ccd03d4d1f))
* **auth:** refresh token with org scope after org creation ([2c8834c](https://github.com/usetero/cli/commit/2c8834c467391ef57d2526bc934425da8ad9182d))
* block padding, elevated surfaces, explicit bg on all styled text ([583d1f3](https://github.com/usetero/cli/commit/583d1f3673c23f7c1f1d9830747cc4eff53e7a14))
* cancel active round on new message or esc, fix tool spinner persistence ([4d4c91f](https://github.com/usetero/cli/commit/4d4c91fe9373d25835c813b7686340d90c40c1a1))
* **cli:** add --reset flag to clear preferences and auth ([f4036e6](https://github.com/usetero/cli/commit/f4036e60ec1e3d3a12a24c69e2d344bef6556479))
* client-generated UUIDs and streaming model metadata ([21dc12a](https://github.com/usetero/cli/commit/21dc12a0cc3f32471ff2c5edfa27d606ab1565c5))
* **cmd:** add tero reset subcommand ([91859e8](https://github.com/usetero/cli/commit/91859e8e7d06e39d53521fbea176cdf3e215f60c))
* command palette triggered by / on empty input ([e8a5345](https://github.com/usetero/cli/commit/e8a5345f7b01f71587f001002ba20635f23ac426))
* **config:** namespace credentials and config by environment ([3e07a7b](https://github.com/usetero/cli/commit/3e07a7bffbadcdef906c10198c691bc80c86d52c))
* esc opens quit dialog when no active round to cancel ([57a1891](https://github.com/usetero/cli/commit/57a189179d86ccb23f5a834c0cc2b02205d6724a))
* focus command bar on mouse click ([b2d7650](https://github.com/usetero/cli/commit/b2d765081d5afc48b6a10c1f35c4ad1d0113bb63))
* handle idempotent upload failures (404 on delete, conflict on create) ([9513d72](https://github.com/usetero/cli/commit/9513d72ad6c9e4a9f3283a9205e490e4890521ce))
* implement cursor positioning with marker extraction ([942ce28](https://github.com/usetero/cli/commit/942ce28bc469de0be1e34036a545090044757c6b))
* make WorkOS client ID configurable via environment variable ([a8c8f70](https://github.com/usetero/cli/commit/a8c8f70b2d25fe95419fe1f307e9baeda5c19499))
* move thinking indicator into message list ([654db05](https://github.com/usetero/cli/commit/654db05c22d194d2bc48bcbe71cb0f9f81f64b85))
* **onboarding:** add TERO_SKIP_TO_APP env var to skip completion screen ([cb7cdce](https://github.com/usetero/cli/commit/cb7cdcedd88aac0dd926d86e0655a1eb3378875a))
* org/account switching, preference cascading clears, table column reordering ([aec63d5](https://github.com/usetero/cli/commit/aec63d5ecfe8106ef3abdbefcdbfa98194502399))
* policy status drawer, context-aware chat empty state, personalized input bar ([7f85d84](https://github.com/usetero/cli/commit/7f85d84f288961b5b4e4064ff51f843f6f7a1f29))
* powersync upload protocol and chat streaming ([9775926](https://github.com/usetero/cli/commit/977592690e837bf0e2ce912189d8e2d9c1a61443))
* **release:** add Docker Hub publishing ([ec4917c](https://github.com/usetero/cli/commit/ec4917c30db371ce1c35a3fa431c77157245a926))
* **release:** add Scoop bucket publishing for Windows ([d0ee803](https://github.com/usetero/cli/commit/d0ee8032950b13ae48a0ec22bde48fad6dceead4))
* show upgrade message when version changes ([c9de3f1](https://github.com/usetero/cli/commit/c9de3f1351fdfa69e9d4d7f9e4e9a4fb830a5ca0))
* simplify PowerSync/SQLite schema generation ([b463a4a](https://github.com/usetero/cli/commit/b463a4a52ebdec80d19c9df7d3c6f3ad6bd5ee79))
* tool block collapse by default with chevron, restyle thinking block, quit dialog, window title, statusbar fixes ([b0c7bfe](https://github.com/usetero/cli/commit/b0c7bfe27b2f73953c2ff866e0629bfdc1abda9d))
* **tui:** redesign theme system and unify discovery step ([94c129b](https://github.com/usetero/cli/commit/94c129b8b8577b0077ac06add0c2d5cbc7a3532b))
* use ready_for_use to gate onboarding instead of status ([5bdf62c](https://github.com/usetero/cli/commit/5bdf62cdeb40282c971f554c794d0c363227e400))
* waste badge component, risk colors, show all categories, sort by pending count ([3dc1d84](https://github.com/usetero/cli/commit/3dc1d843f1bbce76dd5a4b1437dc0f490c1a49a9))
* wire up tool execution with Turn abstraction ([a7e756d](https://github.com/usetero/cli/commit/a7e756d06ad3c068a228e5817e8966c54c2ab736))


### Bug Fixes

* add context propagation to sqlite interface ([f2e16aa](https://github.com/usetero/cli/commit/f2e16aab382394aacc1ad46fd8fb866e279d1b95))
* address lint errors in client tests ([033322e](https://github.com/usetero/cli/commit/033322e3e4dbed948494fdc1b84fd030c703ecb3))
* allow onboarding to proceed when some services have errors ([c3290d8](https://github.com/usetero/cli/commit/c3290d8a745136560fea7bab2a836a963e24a513))
* assistant message flickering and thinking animation ([cc37f2b](https://github.com/usetero/cli/commit/cc37f2bd8f4087edb0e83b767ad0929b0d31c612))
* **auth:** improve expired device code error message ([585cddd](https://github.com/usetero/cli/commit/585cddd33b6e0d77bf807b83033ac0bcb6b03c36))
* **auth:** refresh token without org when --no-org is set ([e72a651](https://github.com/usetero/cli/commit/e72a6516805a3269c08140019d38103b69c1cd81))
* **auth:** use body style for instruction text ([b60cf98](https://github.com/usetero/cli/commit/b60cf98331be07bdf12613f21b088ed4968a4c0b))
* cancel session context on shutdown, TERO_ENV taskfile, env var renames ([478c78c](https://github.com/usetero/cli/commit/478c78c74098cabd09565491239fd89d6d757455))
* check ps_sync_state instead of ps_buckets for sync completion ([b9b0fde](https://github.com/usetero/cli/commit/b9b0fde3b76021ac050ed2a1f7e77eceafa21be9))
* **ci:** don't cancel in-progress signoff on master pushes ([f4a4d6d](https://github.com/usetero/cli/commit/f4a4d6dcf2e21d30132841880256ec43cf631a0a))
* **ci:** prevent master pushes from cancelling each other ([4effc4a](https://github.com/usetero/cli/commit/4effc4ad880390f3baf5581a89e0601102261127))
* **ci:** skip signoff for release-please commits ([94bcf5e](https://github.com/usetero/cli/commit/94bcf5e6bebe9e33a13d57fdbf7f12e2a432e797))
* **ci:** use conventional Docker Hub secret names ([5b19541](https://github.com/usetero/cli/commit/5b1954197274bffffca8bb1cbfd0a92d3e438282))
* **ci:** use correct Docker Hub secret names ([8e65aa8](https://github.com/usetero/cli/commit/8e65aa8d19b83a81d3a7d196e6fcd3383ee10ff7))
* **cli:** --reset flag continues into TUI instead of exiting ([aa1c16e](https://github.com/usetero/cli/commit/aa1c16ed461ddb2454172a7fdc9196454d845ac2))
* **client:** extract message from genqlient HTTPError ([b8c8c5a](https://github.com/usetero/cli/commit/b8c8c5a865db57f1b7fcdfb389b831726255610b))
* **client:** regenerate client with workosOrganizationID field ([0f8cd92](https://github.com/usetero/cli/commit/0f8cd92c606dae673df4a834ec262e608d61b2df))
* **client:** update API client token after org-scoped refresh ([fbdbfea](https://github.com/usetero/cli/commit/fbdbfea9d9ba74691045990477ece41750e5595f))
* **datadog:** prevent browser opening during tests ([07e170d](https://github.com/usetero/cli/commit/07e170dac21184cb872fa796a5cb4c36db45af68))
* **deps:** update go dependencies (minor/patch) ([#16](https://github.com/usetero/cli/issues/16)) ([c4a98e5](https://github.com/usetero/cli/commit/c4a98e5eed23bc42116861a808f30f5d85da7dd0))
* **deps:** update module github.com/charmbracelet/x/ansi to v0.11.3 ([#22](https://github.com/usetero/cli/issues/22)) ([9b68d36](https://github.com/usetero/cli/commit/9b68d365363aab0cfc6aaaaa7e965d5d39480016))
* **deps:** upgrade gopkg.in/yaml.v2 to v3 ([3a308ae](https://github.com/usetero/cli/commit/3a308aee7e9d80d286d562e02fc677a6c4bbadff))
* drawer sits flush under statusbar, border matches slash color ([09dad5e](https://github.com/usetero/cli/commit/09dad5e803d7890c979b3830183639db61be61b7))
* enable staging WorkOS client ID by default for development ([8cb3b9c](https://github.com/usetero/cli/commit/8cb3b9ce651250373d9be28837ee509a94fda507))
* **install:** remove confirmation prompt for piped installs ([63a8c82](https://github.com/usetero/cli/commit/63a8c8290f82281540077b042d537a4d392f41a1))
* **install:** remove nested color codes from success message ([c225f97](https://github.com/usetero/cli/commit/c225f97e5897e60badb7b0e9f9a811eb61bda7f3))
* lint issues and update to golangci-lint v2 config ([072c7bd](https://github.com/usetero/cli/commit/072c7bdaefab7b56294ef596492187176486217f))
* Loading checks sync state directly instead of relying on messages ([585c8f8](https://github.com/usetero/cli/commit/585c8f8c18cc3639c47efd15aeb05ecd5621050b))
* **onboarding:** add context and step counter to Datadog flow ([aba531e](https://github.com/usetero/cli/commit/aba531e9b756aa9dec45b48d1da6d15babfd635e))
* **onboarding:** clarify Datadog region question ([8acdc0e](https://github.com/usetero/cli/commit/8acdc0ebe4b6fc256fdde9941540cee960a98c2f))
* **onboarding:** consistent text styling across screens ([cdf5be6](https://github.com/usetero/cli/commit/cdf5be6fe452eeef438488138c146ec3c14bccec))
* **onboarding:** correct text hierarchy - prompts white, explanations muted ([543bdf9](https://github.com/usetero/cli/commit/543bdf90b3e716343799be4b6f00803773f40190))
* **onboarding:** improve Datadog connection flow copy ([a699766](https://github.com/usetero/cli/commit/a6997663b1fd637f74955f3b7c27a27e8892cdc4))
* **org:** refresh token with org scope when selecting existing org ([2213227](https://github.com/usetero/cli/commit/22132274f15a9e5d4b5feece2833b7a2c7d59854))
* **org:** show loading state during token refresh after auto-selection ([9dc3851](https://github.com/usetero/cli/commit/9dc38512fc08650020d2a0199167b4a9d1c88f53))
* **org:** show org name when auto-selecting ([cbedf8d](https://github.com/usetero/cli/commit/cbedf8dc708e6d498d464b98711a97a014799dcc))
* **org:** use spinner loader during token refresh ([60befb1](https://github.com/usetero/cli/commit/60befb16a928b8f9277844aa5f569ae736bf03c8))
* reduce statusbar diagonal slashes from 3 to 2 ([833e964](https://github.com/usetero/cli/commit/833e96403c0bad1ae704b1ec0b5d9682ef60bf77))
* remove install:test from signoff (runs before release exists) ([b5b9913](https://github.com/usetero/cli/commit/b5b9913fa5fb29b9792b538d5cff32c240f7497b))
* remove invalid rlcp field from goreleaser config ([8a6a7aa](https://github.com/usetero/cli/commit/8a6a7aa773eed3e3d4b3edb5948112c3b2c25fea))
* remove log output from get_latest_version function ([89df1ad](https://github.com/usetero/cli/commit/89df1ad095f231e8534ae4d02dca6d572b48b1ed))
* remove unsupported folder field from brews config ([6b4eec9](https://github.com/usetero/cli/commit/6b4eec9be92b65990a5c44a94c1bdffb06237aa9))
* resolve linter errors in tests and workos ([c63df9d](https://github.com/usetero/cli/commit/c63df9d7952328420690b304a306906f72bea162))
* schema generation uses local control plane and writes to correct paths ([a732106](https://github.com/usetero/cli/commit/a7321067b9d43cf7bee5e688fa50eb3319f82e20))
* scroll gap mismatch, error body styling, and scroll amounts ([5ae16a9](https://github.com/usetero/cli/commit/5ae16a903108a897e0353fedffa134c9ad7b6d83))
* show context window percentage in statusbar ([b3deb16](https://github.com/usetero/cli/commit/b3deb1605f32a06a0f8159ac686d8a7ea1700791))
* specify Formula directory in goreleaser brew config ([b79d28a](https://github.com/usetero/cli/commit/b79d28a664e4b04df21bdfcfe3d58b90fb8eafc0))
* strict JSON unmarshaling for PowerSync instructions ([fd9f713](https://github.com/usetero/cli/commit/fd9f71382a4a11c2e85844329b6afa031ecfe6ba))
* **styles:** detect terminal background for theme selection ([c80b387](https://github.com/usetero/cli/commit/c80b38779c84d03ffc94d92ffb1f84ea86e36037))
* unignore powersync extension binaries for embed ([6bdb398](https://github.com/usetero/cli/commit/6bdb3980557b9829b7b4fd1509b4829f3f9f1d22))
* update GraphQL schema for datadog account status changes ([9801d66](https://github.com/usetero/cli/commit/9801d6675b57eb771cf7eb2f33d068f22a4fb75e))
* use 'folder' instead of deprecated 'directory' in goreleaser ([881cbeb](https://github.com/usetero/cli/commit/881cbeb985573eb95ed4da571922699c8c7f659f))
* use config and manifest files for release-please ([b7d49dc](https://github.com/usetero/cli/commit/b7d49dc79bc745485d667656e62310a7c6e21c05))


### Reverts

* install test should fail on errors, not swallow them ([89442b1](https://github.com/usetero/cli/commit/89442b16361e7ad4a69bfb163b621ed90b6c1185))
* remove separate test-install workflow ([874a9ea](https://github.com/usetero/cli/commit/874a9eafcd5ee6a221b05a7505bcdf971eed1c90))
* switch back from homebrew_casks to brews for CLI tool ([4ec664c](https://github.com/usetero/cli/commit/4ec664c24dfa5637b11e61d4df426bfcb56bc5c6))

## [1.8.0](https://github.com/usetero/cli/compare/v1.7.0...v1.8.0) (2026-01-13)


### Features

* **tui:** redesign theme system and unify discovery step ([10c99a1](https://github.com/usetero/cli/commit/10c99a1b46b220f3c1365d65c9215e17860fd014))


### Bug Fixes

* address lint errors in client tests ([179b2cc](https://github.com/usetero/cli/commit/179b2ccc2a059416f75aeeaf8b19f71eb5e96d31))
* **ci:** prevent master pushes from cancelling each other ([075c36d](https://github.com/usetero/cli/commit/075c36d8fd6738e50a9d9defed5a923dcbd945d3))

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
