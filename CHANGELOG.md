# Changelog

## [1.2.7](https://github.com/VU-ASE/roverlib-go/compare/v1.2.6...v1.2.7) (2025-03-29)


### Bug Fixes

* updated rovercom dependency ([d37948a](https://github.com/VU-ASE/roverlib-go/commit/d37948ad1241c1bc8572e655a49de20a4190a064))

## [1.2.6](https://github.com/VU-ASE/roverlib-go/compare/v1.2.5...v1.2.6) (2024-12-30)


### Bug Fixes

* use subscribe socket for tuning, instead of request socket ([93c27d1](https://github.com/VU-ASE/roverlib-go/commit/93c27d1d4e4275b0b2aa61a3f28bb4e33f3f70d3))

## [1.2.5](https://github.com/VU-ASE/roverlib-go/compare/v1.2.4...v1.2.5) (2024-12-30)


### Bug Fixes

* show tuning address ([5187547](https://github.com/VU-ASE/roverlib-go/commit/5187547785947780c26a51907d2d34532e9ba960))

## [1.2.4](https://github.com/VU-ASE/roverlib-go/compare/v1.2.3...v1.2.4) (2024-12-30)


### Bug Fixes

* keep retrying tuning subscription ([3dec11d](https://github.com/VU-ASE/roverlib-go/commit/3dec11d74cc9a599754b5c87fc12a03711334edf))

## [1.2.3](https://github.com/VU-ASE/roverlib-go/compare/v1.2.2...v1.2.3) (2024-12-30)


### Bug Fixes

* more explicit tuning logs ([2047e36](https://github.com/VU-ASE/roverlib-go/commit/2047e36543c0157a1b26a37e3f7814ef5fa4e169))

## [1.2.2](https://github.com/VU-ASE/roverlib-go/compare/v1.2.1...v1.2.2) (2024-12-30)


### Bug Fixes

* replace localhost with * for write streams ([af4e1e2](https://github.com/VU-ASE/roverlib-go/commit/af4e1e2f44af8123b698fc93f4eb0567913e27ab))

## [1.2.1](https://github.com/VU-ASE/roverlib-go/compare/v1.2.0...v1.2.1) (2024-12-30)


### Bug Fixes

* set correct subscribe value for tuning receiver socket ([49ef9bd](https://github.com/VU-ASE/roverlib-go/commit/49ef9bd7171ad2efbe4b47236aa3d52f73058c9a))

## [1.2.0](https://github.com/VU-ASE/roverlib-go/compare/v1.1.7...v1.2.0) (2024-12-28)


### Features

* compliant with merged int/float rovervalidate schema ([b4d656e](https://github.com/VU-ASE/roverlib-go/commit/b4d656e841cc4b392ff20458a95e568a3ea18a44))

## [1.1.7](https://github.com/VU-ASE/roverlib-go/compare/v1.1.6...v1.1.7) (2024-12-10)


### Bug Fixes

* subscribe to all topics when reading from stream ([71d2233](https://github.com/VU-ASE/roverlib-go/commit/71d22334d64cad573bc746ef1cf79258159b0108))

## [1.1.6](https://github.com/VU-ASE/roverlib-go/compare/v1.1.5...v1.1.6) (2024-12-10)


### Bug Fixes

* bind instead of connect on write stream, typestate pattern ([633c3b0](https://github.com/VU-ASE/roverlib-go/commit/633c3b011fa2d9f4c81c123eaab53cbba10ef659))

## [1.1.5](https://github.com/VU-ASE/roverlib-go/compare/v1.1.4...v1.1.5) (2024-12-01)


### Bug Fixes

* initialize mutex before making configuration accessible ([5016754](https://github.com/VU-ASE/roverlib-go/commit/501675451ac67cba1d7c5edd5969ece25f7932f1))

## [1.1.4](https://github.com/VU-ASE/roverlib-go/compare/v1.1.3...v1.1.4) (2024-12-01)


### Bug Fixes

* more descriptive env var parsing errors ([c3de3d6](https://github.com/VU-ASE/roverlib-go/commit/c3de3d688807cc1eb0affa1fc2a0517bea55b20d))

## [1.1.3](https://github.com/VU-ASE/roverlib-go/compare/v1.1.2...v1.1.3) (2024-12-01)


### Bug Fixes

* use `panic` instead of `log` before logging is set up ([010e25e](https://github.com/VU-ASE/roverlib-go/commit/010e25e27e38b03b287bacaa4b9e570e6c4bf9af))

## [1.1.2](https://github.com/VU-ASE/roverlib-go/compare/v1.1.1...v1.1.2) (2024-11-30)


### Bug Fixes

* updated roverlib to latest version of rovercom ([52a2e03](https://github.com/VU-ASE/roverlib-go/commit/52a2e036e6103aab24dace261a4ec9d2adf79e63))

## [1.1.1](https://github.com/VU-ASE/roverlib-go/compare/v1.1.0...v1.1.1) (2024-11-30)


### Bug Fixes

* release new version to avoid checksum mismatch ([ed4c421](https://github.com/VU-ASE/roverlib-go/commit/ed4c4212e8a5f3c9e93f1307333114b59934b482))

## [1.1.0](https://github.com/VU-ASE/roverlib-go/compare/v1.0.3...v1.1.0) (2024-11-30)


### Features

* added input/output stream functionality based on rovercom ([73bb808](https://github.com/VU-ASE/roverlib-go/commit/73bb808f3c4651121d14daa838988ccef2dc0f67))
* base implementation of roverlib in the roverd system ([3a38afd](https://github.com/VU-ASE/roverlib-go/commit/3a38afde081633febe2828797fe265205739550b))
* implement tuning updates ([7c4d34b](https://github.com/VU-ASE/roverlib-go/commit/7c4d34baed8de8c1245761bcb5968cd1b66e58d0))
* parsing rover/service yamls and validation + validation setup ([e17b173](https://github.com/VU-ASE/roverlib-go/commit/e17b173c9a8dda2c497a8de71411b284e2acf265))
* release roverlib as roverlib-go ([d6d8c47](https://github.com/VU-ASE/roverlib-go/commit/d6d8c47591d89d78772e5a3383632261c013b306))
* service.yaml parsing and testing ([dee2933](https://github.com/VU-ASE/roverlib-go/commit/dee2933c400a9ba7e2d48417b5d286d42fe5b03c))
* switch to devcontainer-based approach ([c10c245](https://github.com/VU-ASE/roverlib-go/commit/c10c2456be6abda30e915aa752715f02363a52f2))

## [1.0.3](https://github.com/VU-ASE/roverlib/compare/v1.0.2...v1.0.3) (2024-07-29)


### Bug Fixes

* change back to CoreInfo ([de53fc2](https://github.com/VU-ASE/roverlib/commit/de53fc2b53db3ecbca577b062b13db4120e7ce0c))

## [1.0.2](https://github.com/VU-ASE/roverlib/compare/v1.0.1...v1.0.2) (2024-07-29)


### Bug Fixes

* error emssages and systemmanagerInfo struct rename ([d916dce](https://github.com/VU-ASE/roverlib/commit/d916dce4fc6ab72c15131bbc2aca23d2b9820f9c))
* some lingering sysman names changed to core ([c1de31b](https://github.com/VU-ASE/roverlib/commit/c1de31b5fac942ad43259633b4f8ccbab54415ca))

## [1.0.1](https://github.com/VU-ASE/roverlib/compare/v1.0.0...v1.0.1) (2024-07-25)


### Bug Fixes

* added lint back ([afe9f3e](https://github.com/VU-ASE/roverlib/commit/afe9f3eaaec824187b9c3bb42336589405d46359))
* upgrade to rovercom ([8467b00](https://github.com/VU-ASE/roverlib/commit/8467b00300db6a4deac7bb6e52615bae13a2a314))

## 1.0.0 (2024-07-25)


### ⚠ BREAKING CHANGES

* inital commit

### Features

* inital commit ([de015ec](https://github.com/VU-ASE/roverlib/commit/de015ec27b9216b78522467289749a8d8f5995da))


### Bug Fixes

* test CI ([93a132c](https://github.com/VU-ASE/roverlib/commit/93a132cde590da475429370a4f8f4332096a63e6))

## Changelog
