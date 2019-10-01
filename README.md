[![Master status](https://tc.gortc.io/app/rest/builds/buildType:(id:stun_MasterStatus)/statusIcon.svg)](https://tc.gortc.io/project.html?projectId=stun&tab=projectOverview&guest=1)
[![GoDoc](https://godoc.org/gortc.io/stun?status.svg)](http://godoc.org/gortc.io/stun)
[![codecov](https://codecov.io/gh/gortc/stun/branch/master/graph/badge.svg)](https://codecov.io/gh/gortc/stun)
# STUN
Package stun implements Session Traversal Utilities for NAT (STUN)
[[RFC5389](https://tools.ietf.org/html/rfc5389)] protocol with no
external dependencies and zero allocations in hot paths.Complies to
[gortc principles](https://gortc.io/#principles) as core package.

See [stun server](https://github.com/gortc/stund) for simple usage. Also see
[gortc/turn](https://github.com/gortc/turn) for TURN
[[RFC5766](https://tools.ietf.org/html/rfc5766)] implementation and
[gortcd](https://github.com/gortc/gortcd) for TURN and STUN server. This
repo was merged to [pion/stun](https://github.com/pion/stun) at version
`v1.19.0`.

Please use `v1` version for stun agent and client.

## Supported RFCs
- [x] [RFC 5389](https://tools.ietf.org/html/rfc5389) — Session Traversal Utilities for NAT
- [x] [RFC 5769](https://tools.ietf.org/html/rfc5769) — Test Vectors for STUN
- [x] [RFC 7064](https://tools.ietf.org/html/rfc7064) — STUN URI

# RFC 3489 notes
RFC 5389 obsoletes RFC 3489, so implementation was ignored by purpose, however,
RFC 3489 can be easily implemented as separate package.

# Requirements
Go 1.13 or better is required.

# Benchmarks

v2.0.0-alpha, Intel(R) Core(TM) i7-8700K:

```
goos: linux
goarch: amd64
pkg: gortc.io/stun/v2
PASS
benchmark                                        iter       time/iter      throughput   bytes alloc        allocs
---------                                        ----       ---------      ----------   -----------        ------
BenchmarkMappedAddress_AddTo-6               62213032     19.10 ns/op                        0 B/op   0 allocs/op
BenchmarkAlternateServer_AddTo-6             61888764     19.10 ns/op                        0 B/op   0 allocs/op
BenchmarkMessage_GetNotFound-6              512655426      2.33 ns/op                        0 B/op   0 allocs/op
BenchmarkMessage_Get-6                      428058297      2.81 ns/op                        0 B/op   0 allocs/op
BenchmarkErrorCode_AddTo-6                   31417389     38.20 ns/op                        0 B/op   0 allocs/op
BenchmarkErrorCodeAttribute_AddTo-6          40657971     29.60 ns/op                        0 B/op   0 allocs/op
BenchmarkErrorCodeAttribute_GetFrom-6       205016971      5.85 ns/op                        0 B/op   0 allocs/op
BenchmarkFingerprint_AddTo-6                 29251347     41.10 ns/op    1071.71 MB/s        0 B/op   0 allocs/op
BenchmarkFingerprint_Check-6                 38511943     41.80 ns/op    1244.49 MB/s        0 B/op   0 allocs/op
BenchmarkBuildOverhead/Build-6               10615113    113.00 ns/op                        0 B/op   0 allocs/op
BenchmarkBuildOverhead/BuildNonPointer-6      4450084    251.00 ns/op                      100 B/op   4 allocs/op
BenchmarkBuildOverhead/Raw-6                 12964191     93.90 ns/op                        0 B/op   0 allocs/op
BenchmarkMessage_ForEach-6                   29016141     41.20 ns/op                        0 B/op   0 allocs/op
BenchmarkMessageIntegrity_AddTo-6             1752014    620.00 ns/op      32.26 MB/s        0 B/op   0 allocs/op
BenchmarkMessageIntegrity_Check-6             1730869    670.00 ns/op      47.75 MB/s       32 B/op   1 allocs/op
BenchmarkMessage_Write-6                     79968304     20.70 ns/op    1350.87 MB/s        0 B/op   0 allocs/op
BenchmarkMessageType_Value-6               1000000000      0.45 ns/op                        0 B/op   0 allocs/op
BenchmarkMessage_WriteTo-6                  184335535      7.99 ns/op                        0 B/op   0 allocs/op
BenchmarkMessage_ReadFrom-6                  72998370     31.20 ns/op     640.03 MB/s        0 B/op   0 allocs/op
BenchmarkMessage_ReadBytes-6                100000000     19.10 ns/op    1047.25 MB/s        0 B/op   0 allocs/op
BenchmarkIsMessage-6                        982619967      1.35 ns/op   14854.89 MB/s        0 B/op   0 allocs/op
BenchmarkMessage_NewTransactionID-6           1000000   1210.00 ns/op                        0 B/op   0 allocs/op
BenchmarkMessageFull-6                         803314   2417.00 ns/op                      432 B/op   4 allocs/op
BenchmarkMessageFullHardcore-6               16941045     99.40 ns/op                        0 B/op   0 allocs/op
BenchmarkMessage_WriteHeader-6              138267142      8.84 ns/op                        0 B/op   0 allocs/op
BenchmarkMessage_CloneTo-6                   54132223     50.50 ns/op    1346.79 MB/s        0 B/op   0 allocs/op
BenchmarkMessage_AddTo-6                    633083188      2.75 ns/op                        0 B/op   0 allocs/op
BenchmarkDecode-6                            86508906     31.50 ns/op                        0 B/op   0 allocs/op
BenchmarkUsername_AddTo-6                    31159868     36.30 ns/op                        0 B/op   0 allocs/op
BenchmarkUsername_GetFrom-6                 100000000     11.80 ns/op                        0 B/op   0 allocs/op
BenchmarkNonce_AddTo-6                       70995598     39.60 ns/op                        0 B/op   0 allocs/op
BenchmarkNonce_AddTo_BadLength-6            366213033      6.35 ns/op                        0 B/op   0 allocs/op
BenchmarkNonce_GetFrom-6                    177038905      7.33 ns/op                        0 B/op   0 allocs/op
BenchmarkUnknownAttributes/AddTo-6           40863390     42.50 ns/op                        0 B/op   0 allocs/op
BenchmarkUnknownAttributes/GetFrom-6        100000000     27.60 ns/op                        0 B/op   0 allocs/op
BenchmarkXOR-6                               20526511     53.50 ns/op   19138.02 MB/s                            
BenchmarkXORSafe-6                            3539281    332.00 ns/op    3088.98 MB/s                            
BenchmarkXORFast-6                           26500282     46.30 ns/op   22103.46 MB/s                            
BenchmarkXORMappedAddress_AddTo-6            42675984     29.40 ns/op                        0 B/op   0 allocs/op
BenchmarkXORMappedAddress_GetFrom-6          63697773     18.50 ns/op                        0 B/op   0 allocs/op
ok      gortc.io/stun/v2        72.661s

```

## Build status
[![Build Status](https://travis-ci.com/gortc/stun.svg)](https://travis-ci.com/gortc/stun)
[![Build status](https://ci.appveyor.com/api/projects/status/fw3drn3k52mf5ghw/branch/master?svg=true)](https://ci.appveyor.com/project/ernado/stun-j08g0/branch/master)

## License
BSD 3-Clause License
