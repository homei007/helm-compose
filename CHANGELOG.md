# Changelog

## [1.5.0](https://github.com/homei007/helm-compose/compare/1.4.0...1.5.0) (2026-08-04)


### Features

* add Helm 4 compatibility ([cdb788f](https://github.com/homei007/helm-compose/commit/cdb788fa0b172d4670668b7ed345afb5b550379a))
* add release automation ([08f647f](https://github.com/homei007/helm-compose/commit/08f647f9d9275475d76dbc81ed01f8e7124e0cc4))
* add release automation ([#7](https://github.com/homei007/helm-compose/issues/7)) ([f694b09](https://github.com/homei007/helm-compose/commit/f694b0902836f5bbc2ea8b8d4904470748b66878))
* add templating command ([#65](https://github.com/homei007/helm-compose/issues/65)) ([edb7493](https://github.com/homei007/helm-compose/commit/edb7493516593775006a1d9beffea9d9395ebd5f))
* ci/cd ([#3](https://github.com/homei007/helm-compose/issues/3)) ([141a1a8](https://github.com/homei007/helm-compose/commit/141a1a88fc18c82134c2a99297a110811ea97f18))
* implement helm compose up and down commands with concurrent processing ([4a474a0](https://github.com/homei007/helm-compose/commit/4a474a0daebbb481e09896c140a912b103c0fb20))
* increase apiVersion and add better version handling ([777fd82](https://github.com/homei007/helm-compose/commit/777fd822dbe50de3b02d18ab2fff62bef8abad73))
* s3 provider ([#5](https://github.com/homei007/helm-compose/issues/5)) ([42bf48f](https://github.com/homei007/helm-compose/commit/42bf48fa1cd39450291f6f5df8392b72558fbcd4))
* update to go version 1.21 ([91dfd91](https://github.com/homei007/helm-compose/commit/91dfd910540f85ea491603ae76b00625314958fd))
* upgrade to golang 1.22 ([#101](https://github.com/homei007/helm-compose/issues/101)) ([265fcdc](https://github.com/homei007/helm-compose/commit/265fcdc5ceae627bf04f86e81e819d220d0a5292))
* wait option ([#20](https://github.com/homei007/helm-compose/issues/20)) ([4f41b1b](https://github.com/homei007/helm-compose/commit/4f41b1b78be666f537d508c6d68d6009e0a69454))


### Bug Fixes

* deploy documentation for fork pages ([f240b39](https://github.com/homei007/helm-compose/commit/f240b39f51fde10bbc132edb61ffb2aabfb95d6f))
* pin docs build Python version ([2e97252](https://github.com/homei007/helm-compose/commit/2e9725285c7be38bc4d09a0c7e1d9cdf349976da))
* prelease to latest release switch ([73f6b17](https://github.com/homei007/helm-compose/commit/73f6b17af931139e3732ce6ea6a1805843eb69d8))
* remove pr lint job and replace with google bot ([04eee91](https://github.com/homei007/helm-compose/commit/04eee912dc112b999206833ab9dac437f28f8753))
* stop creation of revisions if the helm compose file didn't change ([6e3299d](https://github.com/homei007/helm-compose/commit/6e3299d3d5b7b09a8f3c3eb46b60146770e9cd02))
* support Helm 4 version output ([002a0cd](https://github.com/homei007/helm-compose/commit/002a0cda9e99232d3730beb5f16f6162af51c40b))

## Changes since 1.4.0
## 1.4.0
## 1.3.0
- feat: add templating command
- chore: upgrade to golang 1.22

## 1.2.0
- feat: add wait option
- feat: apiVersion 1.1 for wait option and add better version handling
- fix: stop creation of revisions if the helm compose file didn't change

## 1.1.2
- chore: dependency upgrades

## 1.1.1
- bugfix: when the target bucket is empty no state files are written

## 1.1.0
- feat: add support for s3 as a revision storage provider / backend
- chore: upgrade dependencies to the latest versions
