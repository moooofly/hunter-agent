<a name="0.6.0"></a>
# [0.6.0](https://github.com/moooofly/hunter-agent/compare/v0.5.0...v0.6.0) (2018-09-17)


### Features

* support Prometheus metrics ([b2365b8](https://github.com/moooofly/hunter-agent/commit/b2365b8))



<a name="0.5.0"></a>
# [0.5.0](https://github.com/moooofly/hunter-agent/compare/v0.4.0...v0.5.0) (2018-09-12)


### Bug Fixes

* fix configuration reload issue, close [#7](https://github.com/moooofly/hunter-agent/issues/7) ([e6e6a73](https://github.com/moooofly/hunter-agent/commit/e6e6a73))


### Features

* support metrics output when hunter agent existing ([1a3d9e0](https://github.com/moooofly/hunter-agent/commit/1a3d9e0))
* support queue-size dynamic reloading ([7b4a52b](https://github.com/moooofly/hunter-agent/commit/7b4a52b))



<a name="0.4.0"></a>
# [0.4.0](https://github.com/moooofly/hunter-agent/compare/v0.3.1...v0.4.0) (2018-09-11)


### Features

* support flow control ability ([dff0a3c](https://github.com/moooofly/hunter-agent/commit/dff0a3c))
* support multiple kafka partitions by traceid of spans ([28a822b](https://github.com/moooofly/hunter-agent/commit/28a822b))



<a name="0.3.1"></a>
# [0.3.1](https://github.com/moooofly/hunter-agent/compare/v0.3.0...v0.3.1) (2018-09-07)


### Bug Fixes

* fix goroutines dump issue, close [#6](https://github.com/moooofly/hunter-agent/issues/6) ([947736b](https://github.com/moooofly/hunter-agent/commit/947736b))


### Features

* add build flags and version log ([d771d26](https://github.com/moooofly/hunter-agent/commit/d771d26))



<a name="0.3.0"></a>
# [0.3.0](https://github.com/moooofly/hunter-agent/compare/v0.2.0...v0.3.0) (2018-08-22)


### Features

* add Dockerfile ([f33dddf](https://github.com/moooofly/hunter-agent/commit/f33dddf))
* do version control with glide ([542b181](https://github.com/moooofly/hunter-agent/commit/542b181))
* update Makefile for dockerization ([326888b](https://github.com/moooofly/hunter-agent/commit/326888b))



<a name="0.2.0"></a>
# [0.2.0](https://github.com/moooofly/hunter-agent/compare/v0.1.0...v0.2.0) (2018-08-20)


### Features

* add dumpproto related files ([af52f99](https://github.com/moooofly/hunter-agent/commit/af52f99))
* add more flags, add grpc server, add metrics server ([fa1c04f](https://github.com/moooofly/hunter-agent/commit/fa1c04f))
* change value of kafka message from traceproto to dumpproto definition ([c71eeb1](https://github.com/moooofly/hunter-agent/commit/c71eeb1))
* construct daemon infrastructure inspired by docker daemon ([66b62bb](https://github.com/moooofly/hunter-agent/commit/66b62bb))



