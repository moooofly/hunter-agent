# 2018-09-12 测试报告

## 前置条件

- 测试基于 vagrant 虚拟机，Ubuntu 16.04.3 LTS, Linux 4.4.0-130-generic x86_64, 2c, 1g Mem + 1g Swap
- 使用 [local_example](https://github.com/moooofly/opencensus-go-exporter-agent/tree/master/example/local_example) 作为 client
    - 测试代码中通过 for 无限循环模拟极限压力测试；
	- 模拟调用链路如下：
    	- `simulate_neo_api` ->
        	- `simulate_grpc_client` ->
            	- `simulate_grpc_server` ->
                	- `simulate_grpc_server_call_mysql` ->
                    	- (mysql)
        	- `simulate_neo_api_call_mysql`
            	- (mysql)
        	- `simulate_neo_api_call_redis`
            	- (redis)
    - 测试代码中没有使用人为设置的 `time.Sleep()` ；
    - 当前 agent exporter 的实现为每次只发送一个 Span ，非 batch 方式；
- 使用 [hunter-agent](https://github.com/moooofly/hunter-agent/tree/v0.4.0) 作为 server
    - 当前 flow control 基于 channel 缓冲区长度实现，overflow 后简单丢弃最新到来的 kafka 消息；
    - 当前测试代码在一条 kafka 消息中仅封装一个 span 数据；
- 当前联调使用的 kafka 配置（dev 测试环境）未知，应该是默认配置；

## 当前结论

- 当前 kafka message 处理的速度上限已到达，且受 kafka 配置影响；
- 性能：基于 Unix Domain Socket 的方式略优于基于 TCP Socket 的方式；
    - `tcp + 100`：稳定后，in/out 在 93000~95000 左右，drop 在 6000 左右
    - `unix + 100`：稳定后，in/out 在 85000~90000 左右，drop 在 5500 左右
    - `tcp + 10`：稳定后，in/out 在 45000 左右，drop 在 45000~60000 左右
    - `unix + 10`：稳定后，in/out 在 45000 左右，drop 在 45000~60000 左右
    - `tcp + 1000`：稳定后，in/out 在 88000~95000 左右，drop 一直为 0 ；
    - `unix +1000`：稳定后，in/out 的数值范围在 80858~105163 左右，偶尔出现 drop 情况；

## 测试数据

### queue-size 为 100

#### tcp

```
DEBU[2018-09-11T19:45:32.410438704+08:00] No non-zero metrics in the last 5s
DEBU[2018-09-11T19:45:37.412072384+08:00] No non-zero metrics in the last 5s
DEBU[2018-09-11T19:45:42.408032781+08:00] Non-zero metrics in the last 5s: message-out=62074 overflow-drop=5477 message-in=62074
DEBU[2018-09-11T19:45:47.407402291+08:00] Non-zero metrics in the last 5s: message-in=93956 message-out=93956 overflow-drop=8329
DEBU[2018-09-11T19:45:52.407339330+08:00] Non-zero metrics in the last 5s: message-in=92449 message-out=92449 overflow-drop=5962
DEBU[2018-09-11T19:45:57.416688404+08:00] Non-zero metrics in the last 5s: message-in=95715 message-out=95685 overflow-drop=7753
DEBU[2018-09-11T19:46:02.407619876+08:00] Non-zero metrics in the last 5s: message-in=92063 message-out=92093 overflow-drop=6878
DEBU[2018-09-11T19:46:07.407308105+08:00] Non-zero metrics in the last 5s: message-in=88615 message-out=88615 overflow-drop=5173
DEBU[2018-09-11T19:46:12.407371729+08:00] Non-zero metrics in the last 5s: message-out=93229 overflow-drop=6909 message-in=93229
DEBU[2018-09-11T19:46:17.407517344+08:00] Non-zero metrics in the last 5s: message-in=91843 message-out=91843 overflow-drop=8453
DEBU[2018-09-11T19:46:22.407438980+08:00] Non-zero metrics in the last 5s: overflow-drop=5923 message-in=89362 message-out=89362
DEBU[2018-09-11T19:46:27.409015589+08:00] Non-zero metrics in the last 5s: message-in=92210 message-out=92210 overflow-drop=6709
DEBU[2018-09-11T19:46:32.413416740+08:00] Non-zero metrics in the last 5s: message-in=94785 message-out=94684 overflow-drop=6624
DEBU[2018-09-11T19:46:37.407332666+08:00] Non-zero metrics in the last 5s: message-in=90633 message-out=90734 overflow-drop=5086
DEBU[2018-09-11T19:46:42.407463446+08:00] Non-zero metrics in the last 5s: message-in=93164 message-out=93149 overflow-drop=6480
DEBU[2018-09-11T19:46:47.408133023+08:00] Non-zero metrics in the last 5s: message-in=93083 message-out=93098 overflow-drop=4667
DEBU[2018-09-11T19:46:52.407525233+08:00] Non-zero metrics in the last 5s: message-in=93973 message-out=93973 overflow-drop=5974
DEBU[2018-09-11T19:46:57.407515890+08:00] Non-zero metrics in the last 5s: message-out=93770 overflow-drop=6286 message-in=93771
DEBU[2018-09-11T19:47:02.407308075+08:00] Non-zero metrics in the last 5s: message-out=96437 overflow-drop=7626 message-in=96437
DEBU[2018-09-11T19:47:07.407494506+08:00] Non-zero metrics in the last 5s: overflow-drop=9225 message-in=89739 message-out=89740
DEBU[2018-09-11T19:47:12.415052917+08:00] Non-zero metrics in the last 5s: message-in=94004 message-out=93903 overflow-drop=7057
DEBU[2018-09-11T19:47:17.408423664+08:00] Non-zero metrics in the last 5s: message-in=90425 message-out=90526 overflow-drop=6200
DEBU[2018-09-11T19:47:22.412168168+08:00] Non-zero metrics in the last 5s: message-in=92453 message-out=92453 overflow-drop=6591
DEBU[2018-09-11T19:47:27.407531899+08:00] Non-zero metrics in the last 5s: message-in=94868 message-out=94867 overflow-drop=6585
DEBU[2018-09-11T19:47:32.407360997+08:00] Non-zero metrics in the last 5s: message-in=95121 message-out=95121 overflow-drop=6492
DEBU[2018-09-11T19:47:37.413692752+08:00] Non-zero metrics in the last 5s: message-in=92259 message-out=92260 overflow-drop=8104
DEBU[2018-09-11T19:47:42.407522534+08:00] Non-zero metrics in the last 5s: message-in=91656 message-out=91653 overflow-drop=5058
DEBU[2018-09-11T19:47:47.407584863+08:00] Non-zero metrics in the last 5s: message-in=89570 message-out=89573 overflow-drop=6979
DEBU[2018-09-11T19:47:52.411678674+08:00] Non-zero metrics in the last 5s: message-in=91315 message-out=91315 overflow-drop=8381
DEBU[2018-09-11T19:47:57.407617775+08:00] Non-zero metrics in the last 5s: message-in=84143 message-out=84086 overflow-drop=5889
DEBU[2018-09-11T19:48:02.407530884+08:00] Non-zero metrics in the last 5s: overflow-drop=6579 message-in=84117 message-out=84174
DEBU[2018-09-11T19:48:07.407627617+08:00] Non-zero metrics in the last 5s: overflow-drop=4114 message-in=84894 message-out=84894
DEBU[2018-09-11T19:48:12.407559934+08:00] No non-zero metrics in the last 5s
DEBU[2018-09-11T19:48:17.407617956+08:00] No non-zero metrics in the last 5s
```

#### unix

```
DEBU[2018-09-11T19:49:22.409202300+08:00] No non-zero metrics in the last 5s
DEBU[2018-09-11T19:49:27.408261735+08:00] No non-zero metrics in the last 5s
DEBU[2018-09-11T19:49:32.408816789+08:00] Non-zero metrics in the last 5s: message-in=42730 message-out=42635 overflow-drop=3275
DEBU[2018-09-11T19:49:37.407676708+08:00] Non-zero metrics in the last 5s: message-in=91819 message-out=91914 overflow-drop=6213
DEBU[2018-09-11T19:49:42.407537684+08:00] Non-zero metrics in the last 5s: message-in=90948 message-out=90948 overflow-drop=6105
DEBU[2018-09-11T19:49:47.407869563+08:00] Non-zero metrics in the last 5s: message-in=88028 message-out=88028 overflow-drop=7412
DEBU[2018-09-11T19:49:52.408909254+08:00] Non-zero metrics in the last 5s: message-out=92167 overflow-drop=8014 message-in=92177
DEBU[2018-09-11T19:49:57.407330490+08:00] Non-zero metrics in the last 5s: message-in=86564 message-out=86574 overflow-drop=6141
DEBU[2018-09-11T19:50:02.407828010+08:00] Non-zero metrics in the last 5s: message-out=93000 overflow-drop=5783 message-in=93000
DEBU[2018-09-11T19:50:07.407856789+08:00] Non-zero metrics in the last 5s: message-in=88110 message-out=88110 overflow-drop=6678
DEBU[2018-09-11T19:50:12.408641084+08:00] Non-zero metrics in the last 5s: message-in=92439 message-out=92439 overflow-drop=4339
DEBU[2018-09-11T19:50:17.407386766+08:00] Non-zero metrics in the last 5s: overflow-drop=6158 message-in=93572 message-out=93573
DEBU[2018-09-11T19:50:22.408158716+08:00] Non-zero metrics in the last 5s: message-in=85167 message-out=85166 overflow-drop=4739
DEBU[2018-09-11T19:50:27.407354710+08:00] Non-zero metrics in the last 5s: message-in=86099 message-out=86097 overflow-drop=6401
DEBU[2018-09-11T19:50:32.407406935+08:00] Non-zero metrics in the last 5s: message-in=90022 message-out=90024 overflow-drop=5534
DEBU[2018-09-11T19:50:37.407495484+08:00] Non-zero metrics in the last 5s: message-out=82805 overflow-drop=6648 message-in=82805
DEBU[2018-09-11T19:50:42.407489317+08:00] Non-zero metrics in the last 5s: message-in=90317 message-out=90317 overflow-drop=5781
DEBU[2018-09-11T19:50:47.412103481+08:00] Non-zero metrics in the last 5s: message-out=91649 overflow-drop=7060 message-in=91716
DEBU[2018-09-11T19:50:52.409389709+08:00] Non-zero metrics in the last 5s: overflow-drop=6397 message-in=88750 message-out=88817
DEBU[2018-09-11T19:50:57.407564929+08:00] Non-zero metrics in the last 5s: message-in=91416 message-out=91416 overflow-drop=5634
DEBU[2018-09-11T19:51:02.408184186+08:00] Non-zero metrics in the last 5s: message-in=92181 message-out=92181 overflow-drop=6135
DEBU[2018-09-11T19:51:07.407716137+08:00] Non-zero metrics in the last 5s: message-in=88185 message-out=88185 overflow-drop=5904
DEBU[2018-09-11T19:51:12.407608976+08:00] Non-zero metrics in the last 5s: message-in=94313 message-out=94313 overflow-drop=7766
DEBU[2018-09-11T19:51:17.407581060+08:00] Non-zero metrics in the last 5s: message-in=84609 message-out=84609 overflow-drop=5030
DEBU[2018-09-11T19:51:22.407685261+08:00] Non-zero metrics in the last 5s: message-in=92871 message-out=92871 overflow-drop=5649
DEBU[2018-09-11T19:51:27.407578140+08:00] Non-zero metrics in the last 5s: message-out=90970 overflow-drop=4946 message-in=90970
DEBU[2018-09-11T19:51:32.407515477+08:00] Non-zero metrics in the last 5s: message-in=86658 message-out=86658 overflow-drop=3828
DEBU[2018-09-11T19:51:37.408281267+08:00] Non-zero metrics in the last 5s: message-in=85277 message-out=85277 overflow-drop=5317
DEBU[2018-09-11T19:51:42.407394893+08:00] Non-zero metrics in the last 5s: message-in=85938 message-out=85938 overflow-drop=4774
DEBU[2018-09-11T19:51:47.416961835+08:00] Non-zero metrics in the last 5s: overflow-drop=5797 message-in=82516 message-out=82514
DEBU[2018-09-11T19:51:52.407398228+08:00] Non-zero metrics in the last 5s: overflow-drop=4777 message-in=92756 message-out=92758
DEBU[2018-09-11T19:51:57.407608131+08:00] Non-zero metrics in the last 5s: message-in=88250 message-out=88250 overflow-drop=8007
DEBU[2018-09-11T19:52:02.408621760+08:00] Non-zero metrics in the last 5s: message-in=44087 message-out=44087 overflow-drop=1905
DEBU[2018-09-11T19:52:07.407684133+08:00] No non-zero metrics in the last 5s
DEBU[2018-09-11T19:52:12.408549892+08:00] No non-zero metrics in the last 5s
```

### queue-size 为 10

#### tcp

```
DEBU[2018-09-11T19:55:52.048760131+08:00] No non-zero metrics in the last 5s
DEBU[2018-09-11T19:55:57.048372642+08:00] No non-zero metrics in the last 5s
DEBU[2018-09-11T19:56:02.048311606+08:00] No non-zero metrics in the last 5s
DEBU[2018-09-11T19:56:07.048323788+08:00] Non-zero metrics in the last 5s: message-in=25526 message-out=25526 overflow-drop=30482
DEBU[2018-09-11T19:56:12.051767686+08:00] Non-zero metrics in the last 5s: message-in=49322 message-out=49312 overflow-drop=48160
DEBU[2018-09-11T19:56:17.049317779+08:00] Non-zero metrics in the last 5s: message-in=47934 message-out=47944 overflow-drop=58444
DEBU[2018-09-11T19:56:22.047898607+08:00] Non-zero metrics in the last 5s: message-in=45083 message-out=45083 overflow-drop=61284
DEBU[2018-09-11T19:56:27.047781278+08:00] Non-zero metrics in the last 5s: message-in=47767 message-out=47767 overflow-drop=62239
DEBU[2018-09-11T19:56:32.047809966+08:00] Non-zero metrics in the last 5s: message-out=45494 overflow-drop=59491 message-in=45494
DEBU[2018-09-11T19:56:37.047765605+08:00] Non-zero metrics in the last 5s: message-in=46757 message-out=46757 overflow-drop=57440
DEBU[2018-09-11T19:56:42.051002772+08:00] Non-zero metrics in the last 5s: overflow-drop=49461 message-in=46633 message-out=46633
DEBU[2018-09-11T19:56:47.048554656+08:00] Non-zero metrics in the last 5s: message-in=45278 message-out=45278 overflow-drop=47435
DEBU[2018-09-11T19:56:52.048671428+08:00] Non-zero metrics in the last 5s: message-in=44235 message-out=44235 overflow-drop=48438
DEBU[2018-09-11T19:56:57.090357126+08:00] Non-zero metrics in the last 5s: message-out=36435 overflow-drop=39282 message-in=36435
DEBU[2018-09-11T19:57:02.047874072+08:00] Non-zero metrics in the last 5s: message-in=43243 message-out=43243 overflow-drop=49157
DEBU[2018-09-11T19:57:07.047932810+08:00] Non-zero metrics in the last 5s: message-out=43532 overflow-drop=53877 message-in=43543
DEBU[2018-09-11T19:57:12.048069271+08:00] Non-zero metrics in the last 5s: message-in=44804 message-out=44815 overflow-drop=45388
DEBU[2018-09-11T19:57:17.047893187+08:00] Non-zero metrics in the last 5s: message-in=42158 message-out=42158 overflow-drop=45456
DEBU[2018-09-11T19:57:22.048444923+08:00] Non-zero metrics in the last 5s: message-in=697 message-out=697 overflow-drop=307
DEBU[2018-09-11T19:57:27.051403261+08:00] No non-zero metrics in the last 5s
DEBU[2018-09-11T19:57:32.047923883+08:00] No non-zero metrics in the last 5s
```

#### unix

```
DEBU[2018-09-11T19:53:27.048890594+08:00] No non-zero metrics in the last 5s
DEBU[2018-09-11T19:53:32.048700334+08:00] No non-zero metrics in the last 5s
DEBU[2018-09-11T19:53:37.047900014+08:00] Non-zero metrics in the last 5s: message-in=17762 message-out=17762 overflow-drop=21028
DEBU[2018-09-11T19:53:42.050004527+08:00] Non-zero metrics in the last 5s: message-in=50987 message-out=50987 overflow-drop=53462
DEBU[2018-09-11T19:53:47.049883371+08:00] Non-zero metrics in the last 5s: message-in=49141 message-out=49132 overflow-drop=50400
DEBU[2018-09-11T19:53:52.048117457+08:00] Non-zero metrics in the last 5s: message-out=49433 overflow-drop=54754 message-in=49424
DEBU[2018-09-11T19:53:57.047857679+08:00] Non-zero metrics in the last 5s: message-out=47462 overflow-drop=57334 message-in=47463
DEBU[2018-09-11T19:54:02.059581833+08:00] Non-zero metrics in the last 5s: message-in=52797 message-out=52798 overflow-drop=49776
DEBU[2018-09-11T19:54:07.047824179+08:00] Non-zero metrics in the last 5s: message-in=48810 message-out=48810 overflow-drop=60372
DEBU[2018-09-11T19:54:12.055188512+08:00] Non-zero metrics in the last 5s: overflow-drop=55120 message-in=46388 message-out=46388
DEBU[2018-09-11T19:54:17.051358884+08:00] Non-zero metrics in the last 5s: message-in=41828 message-out=41828 overflow-drop=54530
DEBU[2018-09-11T19:54:22.048453257+08:00] Non-zero metrics in the last 5s: overflow-drop=54502 message-in=47219 message-out=47219
DEBU[2018-09-11T19:54:27.050204622+08:00] Non-zero metrics in the last 5s: message-in=48067 message-out=48067 overflow-drop=42229
DEBU[2018-09-11T19:54:32.047803081+08:00] Non-zero metrics in the last 5s: message-in=46197 message-out=46197 overflow-drop=44170
DEBU[2018-09-11T19:54:37.053013121+08:00] Non-zero metrics in the last 5s: message-in=45676 message-out=45675 overflow-drop=56296
DEBU[2018-09-11T19:54:42.047839123+08:00] Non-zero metrics in the last 5s: message-in=46033 message-out=46034 overflow-drop=54244
DEBU[2018-09-11T19:54:47.051123625+08:00] Non-zero metrics in the last 5s: message-in=50363 message-out=50363 overflow-drop=52906
DEBU[2018-09-11T19:54:52.054049494+08:00] Non-zero metrics in the last 5s: message-in=45038 message-out=45038 overflow-drop=60846
DEBU[2018-09-11T19:54:57.047859567+08:00] Non-zero metrics in the last 5s: message-in=46714 message-out=46714 overflow-drop=47018
DEBU[2018-09-11T19:55:02.048767631+08:00] Non-zero metrics in the last 5s: message-in=47360 message-out=47360 overflow-drop=52969
DEBU[2018-09-11T19:55:07.047757697+08:00] Non-zero metrics in the last 5s: overflow-drop=55299 message-in=42924 message-out=42924
DEBU[2018-09-11T19:55:12.047974532+08:00] Non-zero metrics in the last 5s: message-in=45799 message-out=45799 overflow-drop=49981
DEBU[2018-09-11T19:55:17.048465430+08:00] Non-zero metrics in the last 5s: message-in=46151 message-out=46151 overflow-drop=55951
DEBU[2018-09-11T19:55:22.047981998+08:00] Non-zero metrics in the last 5s: message-in=46110 message-out=46110 overflow-drop=44260
DEBU[2018-09-11T19:55:27.048466712+08:00] Non-zero metrics in the last 5s: message-in=32163 message-out=32163 overflow-drop=37469
DEBU[2018-09-11T19:55:32.049558687+08:00] No non-zero metrics in the last 5s
DEBU[2018-09-11T19:55:37.048162848+08:00] No non-zero metrics in the last 5s
```

### queue-size 为 1000

#### tcp

```
DEBU[2018-09-11T19:58:24.689532633+08:00] No non-zero metrics in the last 5s
DEBU[2018-09-11T19:58:29.689433162+08:00] No non-zero metrics in the last 5s
DEBU[2018-09-11T19:58:34.690265103+08:00] Non-zero metrics in the last 5s: message-in=22701 message-out=22701
DEBU[2018-09-11T19:58:39.690127633+08:00] Non-zero metrics in the last 5s: message-in=95649 message-out=95649
DEBU[2018-09-11T19:58:44.689309646+08:00] Non-zero metrics in the last 5s: message-in=95404 message-out=95404
DEBU[2018-09-11T19:58:49.693279586+08:00] Non-zero metrics in the last 5s: message-in=92240 message-out=92240
DEBU[2018-09-11T19:58:54.690586051+08:00] Non-zero metrics in the last 5s: message-out=92171 message-in=92171
DEBU[2018-09-11T19:58:59.698808158+08:00] Non-zero metrics in the last 5s: message-in=86893 message-out=86893
DEBU[2018-09-11T19:59:04.689360484+08:00] Non-zero metrics in the last 5s: message-in=90307 message-out=90304
DEBU[2018-09-11T19:59:09.689112275+08:00] Non-zero metrics in the last 5s: message-in=88201 message-out=88202
DEBU[2018-09-11T19:59:14.689179489+08:00] Non-zero metrics in the last 5s: message-in=92494 message-out=92496
DEBU[2018-09-11T19:59:19.689595329+08:00] Non-zero metrics in the last 5s: message-in=89602 message-out=89602
DEBU[2018-09-11T19:59:24.693779881+08:00] Non-zero metrics in the last 5s: message-in=89144 message-out=89144
DEBU[2018-09-11T19:59:29.689300967+08:00] Non-zero metrics in the last 5s: message-in=78874 message-out=78874
DEBU[2018-09-11T19:59:34.690954468+08:00] Non-zero metrics in the last 5s: message-in=85818 message-out=85818
DEBU[2018-09-11T19:59:39.689711230+08:00] Non-zero metrics in the last 5s: message-in=97397 message-out=97394
DEBU[2018-09-11T19:59:44.689061528+08:00] Non-zero metrics in the last 5s: message-in=92720 message-out=92723
DEBU[2018-09-11T19:59:49.689081561+08:00] Non-zero metrics in the last 5s: message-in=94120 message-out=94120
DEBU[2018-09-11T19:59:54.689073361+08:00] Non-zero metrics in the last 5s: message-in=95956 message-out=95955
DEBU[2018-09-11T19:59:59.689351584+08:00] Non-zero metrics in the last 5s: message-in=93578 message-out=93579
DEBU[2018-09-11T20:00:04.689149504+08:00] Non-zero metrics in the last 5s: message-in=89562 message-out=89562
DEBU[2018-09-11T20:00:09.689435541+08:00] Non-zero metrics in the last 5s: message-in=14106 message-out=14106
DEBU[2018-09-11T20:00:14.689432398+08:00] No non-zero metrics in the last 5s
DEBU[2018-09-11T20:00:19.689321595+08:00] No non-zero metrics in the last 5s
DEBU[2018-09-11T20:00:24.689458674+08:00] No non-zero metrics in the last 5s
```

#### unix

```
DEBU[2018-09-11T20:00:59.689314080+08:00] No non-zero metrics in the last 5s
DEBU[2018-09-11T20:01:04.689138675+08:00] No non-zero metrics in the last 5s
DEBU[2018-09-11T20:01:09.689478837+08:00] Non-zero metrics in the last 5s: message-in=41226 message-out=41226
DEBU[2018-09-11T20:01:14.689377153+08:00] Non-zero metrics in the last 5s: message-in=80858 message-out=80858 overflow-drop=27126
DEBU[2018-09-11T20:01:19.691663334+08:00] Non-zero metrics in the last 5s: message-in=100532 message-out=100532
DEBU[2018-09-11T20:01:24.694005027+08:00] Non-zero metrics in the last 5s: message-in=103764 message-out=103764
DEBU[2018-09-11T20:01:29.689707369+08:00] Non-zero metrics in the last 5s: message-in=93896 message-out=93896
DEBU[2018-09-11T20:01:34.689089422+08:00] Non-zero metrics in the last 5s: message-out=103155 message-in=103155
DEBU[2018-09-11T20:01:39.689059946+08:00] Non-zero metrics in the last 5s: message-in=104603 message-out=104602
DEBU[2018-09-11T20:01:44.690646577+08:00] Non-zero metrics in the last 5s: message-out=97586 overflow-drop=1974 message-in=97585
DEBU[2018-09-11T20:01:49.690078771+08:00] Non-zero metrics in the last 5s: message-in=103710 message-out=103708
DEBU[2018-09-11T20:01:54.689079722+08:00] Non-zero metrics in the last 5s: message-in=96859 message-out=96861
DEBU[2018-09-11T20:01:59.689219490+08:00] Non-zero metrics in the last 5s: message-in=99801 message-out=99801
DEBU[2018-09-11T20:02:04.690923057+08:00] Non-zero metrics in the last 5s: message-in=100565 message-out=100565
DEBU[2018-09-11T20:02:09.689108205+08:00] Non-zero metrics in the last 5s: message-out=101853 message-in=101853
DEBU[2018-09-11T20:02:14.689456762+08:00] Non-zero metrics in the last 5s: message-in=105163 message-out=105163
DEBU[2018-09-11T20:02:19.689853938+08:00] Non-zero metrics in the last 5s: message-in=103024 message-out=103019
DEBU[2018-09-11T20:02:24.690968932+08:00] Non-zero metrics in the last 5s: message-in=101259 message-out=101251 overflow-drop=18
DEBU[2018-09-11T20:02:29.690502804+08:00] Non-zero metrics in the last 5s: message-in=98197 message-out=98210
DEBU[2018-09-11T20:02:34.689070582+08:00] Non-zero metrics in the last 5s: message-in=99508 message-out=99508
DEBU[2018-09-11T20:02:39.690418878+08:00] Non-zero metrics in the last 5s: message-in=99624 message-out=99624
DEBU[2018-09-11T20:02:44.689235604+08:00] Non-zero metrics in the last 5s: message-in=102212 message-out=102212
DEBU[2018-09-11T20:02:49.689835347+08:00] Non-zero metrics in the last 5s: message-in=102936 message-out=102932
DEBU[2018-09-11T20:02:54.689753276+08:00] Non-zero metrics in the last 5s: message-in=101629 message-out=101633
DEBU[2018-09-11T20:02:59.689060133+08:00] Non-zero metrics in the last 5s: message-in=105158 message-out=105158
DEBU[2018-09-11T20:03:04.689411582+08:00] Non-zero metrics in the last 5s: message-in=98794 message-out=98794
DEBU[2018-09-11T20:03:09.690440816+08:00] Non-zero metrics in the last 5s: message-in=101245 message-out=101245
DEBU[2018-09-11T20:03:14.689200104+08:00] Non-zero metrics in the last 5s: message-in=99772 message-out=99772
DEBU[2018-09-11T20:03:19.689206175+08:00] Non-zero metrics in the last 5s: message-in=103536 message-out=103535
DEBU[2018-09-11T20:03:24.689073136+08:00] Non-zero metrics in the last 5s: message-in=100522 message-out=100523
DEBU[2018-09-11T20:03:29.689723974+08:00] Non-zero metrics in the last 5s: message-out=90816 message-in=91144
DEBU[2018-09-11T20:03:34.690065168+08:00] Non-zero metrics in the last 5s: message-in=96413 message-out=96741
DEBU[2018-09-11T20:03:39.689475520+08:00] Non-zero metrics in the last 5s: message-in=91853 message-out=91853
DEBU[2018-09-11T20:03:44.690298464+08:00] Non-zero metrics in the last 5s: message-in=42159 message-out=42159
DEBU[2018-09-11T20:03:49.689500496+08:00] No non-zero metrics in the last 5s
DEBU[2018-09-11T20:03:54.689550495+08:00] No non-zero metrics in the last 5s
```

## 其他

- 若将日志在 stdout 或 stderr 上输出，则会严重影响压测效果；因此演策指令最后需要使用 `>/dev/null 2>&1` ；

