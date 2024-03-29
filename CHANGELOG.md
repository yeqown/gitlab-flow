# [](https://github.com/yeqown/gitlab-flow/compare/v1.7.2...v) (2022-09-22)



## [1.7.2](https://github.com/yeqown/gitlab-flow/compare/v1.7.1...v1.7.2) (2022-09-22)


### Features

* **gitlab-op:** beautify callback page based html/template ([8cb4930](https://github.com/yeqown/gitlab-flow/commit/8cb49305aebc3f5cfe6942a9f423441deba38cfd))



## [1.7.1](https://github.com/yeqown/gitlab-flow/compare/v1.7.0...v1.7.1) (2022-03-18)


### Bug Fixes

* blocking some prefixes of milestone/feature/issue title ([a144135](https://github.com/yeqown/gitlab-flow/commit/a144135ac41720cacc6bdf36877f6613e75ccd7b))
* **conf:** ignore git rev-parse error ([d20dcc6](https://github.com/yeqown/gitlab-flow/commit/d20dcc64e16e2cb3d40d710732a5911731a75dd0))


### Features

* **cli/support:** optimise config support function ([efe729b](https://github.com/yeqown/gitlab-flow/commit/efe729bec04b9bf11955a6a4df26b63187c9bf5f))
* **conf:** log format adjust; use git toplevel as cwd firstly ([ca8ff3b](https://github.com/yeqown/gitlab-flow/commit/ca8ff3bb094b3a274fb5fe90814122897cafc7ab))
* **conf/branchs:** add branch enums in configuration which will replace builtin branch enums when it is configured. ([97f8c0f](https://github.com/yeqown/gitlab-flow/commit/97f8c0f4c4c487c3b642f8dfc8dbce8557fabe75))
* **dash:** feature dashboard could parse default feature branch name ([f4e94e0](https://github.com/yeqown/gitlab-flow/commit/f4e94e00c1de35a5d2d0ffc681a837f89cdb9b54))
* **issue:** redesign issue name ([7921272](https://github.com/yeqown/gitlab-flow/commit/79212729a96280351b4d5c6e7d31f82d103837a5))
* **oauth:** move appId and appSecret into build parameters, instead of configuring file. ([9d88182](https://github.com/yeqown/gitlab-flow/commit/9d88182ebfcfc5821209285bb390f704d799931b))
* **oauth:** use oauth2 verification instead of personal token ([d93544a](https://github.com/yeqown/gitlab-flow/commit/d93544ac4f4b4a36b1f991337750d4feb16e98f6))
* **spin:** create spiner ([8512152](https://github.com/yeqown/gitlab-flow/commit/85121522e4d0c30339f0a6ddd2b92aee19448925))
* **types/context:** hide all attributes of context, avoiding wrong modifying from other package. ([d0bc0a2](https://github.com/yeqown/gitlab-flow/commit/d0bc0a2ae7f042833aed88e10265395ca704f3aa))



## [1.6.7](https://github.com/yeqown/gitlab-flow/compare/v1.6.6...v1.6.7) (2021-06-07)


### Features

* **dash:** enhance dash milestone display; and finish one todo. ([82a1496](https://github.com/yeqown/gitlab-flow/commit/82a1496c4593916330e96bd29d651da8f0c3e198))
* **dash:** enhance feature detail information ([f928a20](https://github.com/yeqown/gitlab-flow/commit/f928a20f0859d9980687ca39618fb191a08d7a2d))
* **dash:** project detail has more URL to open ([a540c4d](https://github.com/yeqown/gitlab-flow/commit/a540c4d56ffaf3b482682c47a2e44eb23b61052f))



## [1.6.5](https://github.com/yeqown/gitlab-flow/compare/v1.6.4...v1.6.5) (2021-04-16)


### Bug Fixes

* remove flags in feature sub commands; fix alias of feature branch name. ([92281cc](https://github.com/yeqown/gitlab-flow/commit/92281ccfa626f01a06312d88334c47ce79aae9d7))


### Features

* handle git command result into stdout rather than error message. ([779c393](https://github.com/yeqown/gitlab-flow/commit/779c393152955f06d38989334152a3b7fcaec6c3))
* new sync command ([#9](https://github.com/yeqown/gitlab-flow/issues/9)) ([154b12b](https://github.com/yeqown/gitlab-flow/commit/154b12bc6618b4c18a0e9fdfc03e1c283d003ba2))
* OpFeatureContext support; force create merge request; feature branch name into `opc`; ([e804dbe](https://github.com/yeqown/gitlab-flow/commit/e804dbe050fbd2b75012f1071168639d9fd151bd))
* query resourced ordered by created_at desc ([0553e84](https://github.com/yeqown/gitlab-flow/commit/0553e84ed8c9737b795a5f639211939e27ebda39))
* trim space of title and desc. ([9248eab](https://github.com/yeqown/gitlab-flow/commit/9248eab380d64f14ad9fec8aafbe3146483ed0ac))



## [1.6.3](https://github.com/yeqown/gitlab-flow/compare/v1.6.2...v1.6.3) (2021-02-02)


### Features

* create remote resource enhance... ([93dc87f](https://github.com/yeqown/gitlab-flow/commit/93dc87f3d8be98b5305318d0c04de766df650fbd))
* cwd could be specified from CLI command ([28d6fcb](https://github.com/yeqown/gitlab-flow/commit/28d6fcbb6cd238be600d51217cf5bda2c9925210))
* load projects when CLI command is loading, let user to decide which project should be used ([18e5adf](https://github.com/yeqown/gitlab-flow/commit/18e5adf0a17932fe75e73613fcc0d660db65f923))
* support feature conflicts resolve. ([9818dc5](https://github.com/yeqown/gitlab-flow/commit/9818dc5dd572044e0bf86847bcddb64f4595a3b1))



# [1.6.0](https://github.com/yeqown/gitlab-flow/compare/v1.5.1...v1.6.0) (2021-01-16)


### Features

* global flags supported projectName, openBrowser; optimism FlowContext usage; ([aa7d4e1](https://github.com/yeqown/gitlab-flow/commit/aa7d4e16852f73b988c3108e55efec7472acaa7c))
* global flags supported projectName, openBrowser; optimism FlowContext usage; ([ac3b8e0](https://github.com/yeqown/gitlab-flow/commit/ac3b8e054452c7bc41eb94ee9b9b36a866eea081))
* open issue do not create merge request, close would do this. ([c2277dd](https://github.com/yeqown/gitlab-flow/commit/c2277ddf955bed5dfca3742befcf7bd1792858dc))
* OpenBrowser cross platform support. ([f6c0b98](https://github.com/yeqown/gitlab-flow/commit/f6c0b98edc825b07874d47c02c83fffeabcf9be0))



## [1.5.1](https://github.com/yeqown/gitlab-flow/compare/743c3a12f45077ccad9166aa095dd77fc8218635...v1.5.1) (2021-01-16)


### Bug Fixes

* conf path and db path fixed; repository fixed; add log points; ([5c6d099](https://github.com/yeqown/gitlab-flow/commit/5c6d0996d926834c7f9c8e0ece4231a13b16e60b))
* feature commands fix; ([31ae91d](https://github.com/yeqown/gitlab-flow/commit/31ae91dc121c50eaf8a8536641a8ddc13812a952))
* git operator with space value ([e4d315a](https://github.com/yeqown/gitlab-flow/commit/e4d315a41e91d2c81ba487a5310cb67736464038))
* gitop checkout function ([0852833](https://github.com/yeqown/gitlab-flow/commit/08528331c9a7506407122a8748c26857ac8fe7fc))
* hotfix close with query first. ([d4f0c07](https://github.com/yeqown/gitlab-flow/commit/d4f0c0771ecf19dea4b4f5261b1c3ebfc2439c3c))


### Features

* close-issue; printAndOpenBrowser function; ([b2353da](https://github.com/yeqown/gitlab-flow/commit/b2353da5088247ab52f895c6f2d68a45bf3afca0))
* dash implements and repository debug settings. ([36ce460](https://github.com/yeqown/gitlab-flow/commit/36ce460018c6d9dc938d4254b7cf579a2289997a))
* feature command implements; fix: repository query logic; ([8b5c7a0](https://github.com/yeqown/gitlab-flow/commit/8b5c7a000ddf1aa57dc27ae964a6bc67f54703d7))
* fill logic of flow and its components ([743c3a1](https://github.com/yeqown/gitlab-flow/commit/743c3a12f45077ccad9166aa095dd77fc8218635))
* flow and commands defines ([bd85af3](https://github.com/yeqown/gitlab-flow/commit/bd85af3f81211733d08f9f85249dc8ca468f329f))
* flow and commands defines stage2; ([5db377a](https://github.com/yeqown/gitlab-flow/commit/5db377a66118bc1d9bcb283425d4687ce2e7c874))
* repository adjust and little test; flow implemented few functions; ([2078090](https://github.com/yeqown/gitlab-flow/commit/207809048a2a147f8601d0a9f80e271f70faea4f))



