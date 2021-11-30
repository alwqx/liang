#! /bin/bash
# refer to https://geektutu.com/post/hpg-benchmark.html

# go test -benchmem -bench .

go test -bench .

<<EOF
$ ./bench.sh
goos: linux
goarch: amd64
pkg: liang/internal/service
cpu: Intel(R) Core(TM) i7-7600U CPU @ 2.80GHz
BenchmarkBalanceNetloadPriority_BNPScore10-4              148446              7569 ns/op
BenchmarkBalanceNetloadPriority_BNPScore100-4              37587             31591 ns/op
BenchmarkBalanceNetloadPriority_BNPScore1000-4               850           1515090 ns/op
BenchmarkBalanceNetloadPriority_BNPScore3000-4               100          11901127 ns/op
BenchmarkBalanceNetloadPriority_BNPScore5000-4                37          35008001 ns/op
BenchmarkBalanceNetloadPriority_BNPScore7000-4                18          63258157 ns/op
BenchmarkBalanceNetloadPriority_BNPScore9000-4                10         102124195 ns/op
BenchmarkBalanceNetloadPriority_BNPScore10000-4                8         127820304 ns/op
BenchmarkCMDNPriority_Score10-4                            82564             15187 ns/op
BenchmarkCMDNPriority_Score100-4                            9423            128768 ns/op
BenchmarkCMDNPriority_Score1000-4                           1015           1228016 ns/op
BenchmarkCMDNPriority_Score3000-4                            264           4256656 ns/op
BenchmarkCMDNPriority_Score5000-4                            186           6389366 ns/op
BenchmarkCMDNPriority_Score7000-4                            100          11807567 ns/op
BenchmarkCMDNPriority_Score9000-4                             99          12157084 ns/op
BenchmarkCMDNPriority_Score10000-4                            74          13622703 ns/op
PASS
ok      liang/internal/service  24.979s
EOF