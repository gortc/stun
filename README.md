![CI](https://github.com/gortc/stun/workflows/CI/badge.svg)
[![GoDev](https://img.shields.io/badge/go.dev-reference-007d9c)](https://pkg.go.dev/gortc.io/stun)
[![codecov](https://img.shields.io/codecov/c/github/gortc/stun?label=coverage)](https://codecov.io/gh/gortc/stun)
# STUN
Package stun implements Session Traversal Utilities for NAT (STUN) [[RFC5389](https://tools.ietf.org/html/rfc5389)]
protocol and [client](https://godoc.org/gortc.io/stun#Client) with no external dependencies and zero allocations in hot paths.
Client [supports](https://godoc.org/gortc.io/stun#WithRTO) automatic request retransmissions.
Complies to [gortc principles](https://gortc.io/#principles) as core package.

See [example](https://godoc.org/gortc.io/stun#example-Message) and [stun server](https://github.com/gortc/stund) for simple usage.
Also see [gortc/turn](https://github.com/gortc/turn) for TURN [[RFC5766](https://tools.ietf.org/html/rfc5766)] implementation and
[gortcd](https://github.com/gortc/gortcd) for TURN and STUN server. This repo was merged to [pion/stun](https://github.com/pion/stun)
at version `v1.19.0`.

# Example
You can get your current IP address from any STUN server by sending
binding request. See more idiomatic example at `cmd/stun-client`.
```go
package main

import (
	"fmt"

	"gortc.io/stun"
)

func main() {
	// Creating a "connection" to STUN server.
	c, err := stun.Dial("udp", "stun.l.google.com:19302")
	if err != nil {
		panic(err)
	}
	// Building binding request with random transaction id.
	message := stun.MustBuild(stun.TransactionID, stun.BindingRequest)
	// Sending request to STUN server, waiting for response message.
	if err := c.Do(message, func(res stun.Event) {
		if res.Error != nil {
			panic(res.Error)
		}
		// Decoding XOR-MAPPED-ADDRESS attribute from message.
		var xorAddr stun.XORMappedAddress
		if err := xorAddr.GetFrom(res.Message); err != nil {
			panic(err)
		}
		fmt.Println("your IP is", xorAddr.IP)
	}); err != nil {
		panic(err)
	}
}
```

## Supported RFCs
- [x] [RFC 5389](https://tools.ietf.org/html/rfc5389) — Session Traversal Utilities for NAT
- [x] [RFC 5769](https://tools.ietf.org/html/rfc5769) — Test Vectors for STUN
- [x] [RFC 6062](https://tools.ietf.org/html/rfc6062) — TURN extensions for TCP allocations
- [x] [RFC 7064](https://tools.ietf.org/html/rfc7064) — STUN URI
- [x] (TLS-over-)TCP client support
- [ ] [ALTERNATE-SERVER](https://tools.ietf.org/html/rfc5389#section-11) support [#48](https://github.com/gortc/stun/issues/48)
- [ ] [RFC 5780](https://tools.ietf.org/html/rfc5780) — NAT Behavior Discovery Using STUN [#49](https://github.com/gortc/stun/issues/49)

# Stability [![stability-mature](https://img.shields.io/badge/stability-mature-008000.svg)](https://github.com/mkenney/software-guides/blob/master/STABILITY-BADGES.md#mature) ![GitHub tag](https://img.shields.io/github/tag/gortc/stun.svg)

Package is currently stable, no backward incompatible changes are expected
with exception of critical bugs or security fixes.

Additional attributes are unlikely to be implemented in scope of stun package,
the only exception is constants for attribute or message types.

# RFC 3489 notes
RFC 5389 obsoletes RFC 3489, so implementation was ignored by purpose, however,
RFC 3489 can be easily implemented as separate package.

# Requirements
Go 1.14 is currently supported and tested in CI. Should work on 1.13.

# Testing
Client behavior is tested and verified in many ways:
  * End-To-End with long-term credentials
    * **coturn**: The coturn [server](https://github.com/coturn/coturn/wiki/turnserver) (linux)
  * Bunch of code static checkers (linters)
  * Standard unit-tests with coverage reporting (linux {amd64, **arm**64}, windows and darwin)
  * Explicit API backward compatibility [check](https://github.com/gortc/api), see `api` directory

See [TeamCity project](https://tc.gortc.io/project.html?projectId=stun&guest=1) and `e2e` directory
for more information. Also the Wireshark `.pcap` files are available for e2e test in
artifacts for build.

# Benchmarks

Intel(R) Core(TM) i7-8700K:

```
version: 1.22.2
goos: linux
goarch: amd64
pkg: github.com/gortc/stun
PASS
benchmark                                         iter       time/iter      throughput   bytes alloc        allocs
---------                                         ----       ---------      ----------   -----------        ------
BenchmarkMappedAddress_AddTo-12               32489450     38.30 ns/op                        0 B/op   0 allocs/op
BenchmarkAlternateServer_AddTo-12             31230991     39.00 ns/op                        0 B/op   0 allocs/op
BenchmarkAgent_GC-12                            431390   2918.00 ns/op                        0 B/op   0 allocs/op
BenchmarkAgent_Process-12                     35901940     36.20 ns/op                        0 B/op   0 allocs/op
BenchmarkMessage_GetNotFound-12              242004358      5.19 ns/op                        0 B/op   0 allocs/op
BenchmarkMessage_Get-12                      230520343      5.21 ns/op                        0 B/op   0 allocs/op
BenchmarkClient_Do-12                          1282231    943.00 ns/op                        0 B/op   0 allocs/op
BenchmarkErrorCode_AddTo-12                   16318916     75.50 ns/op                        0 B/op   0 allocs/op
BenchmarkErrorCodeAttribute_AddTo-12          21584140     54.80 ns/op                        0 B/op   0 allocs/op
BenchmarkErrorCodeAttribute_GetFrom-12       100000000     11.10 ns/op                        0 B/op   0 allocs/op
BenchmarkFingerprint_AddTo-12                 19368768     64.00 ns/op     687.81 MB/s        0 B/op   0 allocs/op
BenchmarkFingerprint_Check-12                 24167007     49.10 ns/op    1057.99 MB/s        0 B/op   0 allocs/op
BenchmarkBuildOverhead/Build-12                5486252    224.00 ns/op                        0 B/op   0 allocs/op
BenchmarkBuildOverhead/BuildNonPointer-12      2496544    517.00 ns/op                      100 B/op   4 allocs/op
BenchmarkBuildOverhead/Raw-12                  6652118    181.00 ns/op                        0 B/op   0 allocs/op
BenchmarkMessage_ForEach-12                   28254212     35.90 ns/op                        0 B/op   0 allocs/op
BenchmarkMessageIntegrity_AddTo-12             1000000   1179.00 ns/op      16.96 MB/s        0 B/op   0 allocs/op
BenchmarkMessageIntegrity_Check-12              975954   1219.00 ns/op      26.24 MB/s        0 B/op   0 allocs/op
BenchmarkMessage_Write-12                     41040598     30.40 ns/op     922.13 MB/s        0 B/op   0 allocs/op
BenchmarkMessageType_Value-12               1000000000      0.53 ns/op                        0 B/op   0 allocs/op
BenchmarkMessage_WriteTo-12                   94942935     11.30 ns/op                        0 B/op   0 allocs/op
BenchmarkMessage_ReadFrom-12                  43437718     29.30 ns/op     682.87 MB/s        0 B/op   0 allocs/op
BenchmarkMessage_ReadBytes-12                 74693397     15.90 ns/op    1257.42 MB/s        0 B/op   0 allocs/op
BenchmarkIsMessage-12                       1000000000      1.20 ns/op   16653.64 MB/s        0 B/op   0 allocs/op
BenchmarkMessage_NewTransactionID-12            521121   2450.00 ns/op                        0 B/op   0 allocs/op
BenchmarkMessageFull-12                        5389495    221.00 ns/op                        0 B/op   0 allocs/op
BenchmarkMessageFullHardcore-12               12715876     94.40 ns/op                        0 B/op   0 allocs/op
BenchmarkMessage_WriteHeader-12              100000000     11.60 ns/op                        0 B/op   0 allocs/op
BenchmarkMessage_CloneTo-12                   30199020     41.80 ns/op    1626.66 MB/s        0 B/op   0 allocs/op
BenchmarkMessage_AddTo-12                    415257625      2.97 ns/op                        0 B/op   0 allocs/op
BenchmarkDecode-12                            49573747     23.60 ns/op                        0 B/op   0 allocs/op
BenchmarkUsername_AddTo-12                    56282674     22.50 ns/op                        0 B/op   0 allocs/op
BenchmarkUsername_GetFrom-12                 100000000     10.10 ns/op                        0 B/op   0 allocs/op
BenchmarkNonce_AddTo-12                       39419097     35.80 ns/op                        0 B/op   0 allocs/op
BenchmarkNonce_AddTo_BadLength-12            196291666      6.04 ns/op                        0 B/op   0 allocs/op
BenchmarkNonce_GetFrom-12                    120857732      9.93 ns/op                        0 B/op   0 allocs/op
BenchmarkUnknownAttributes/AddTo-12           28881430     37.20 ns/op                        0 B/op   0 allocs/op
BenchmarkUnknownAttributes/GetFrom-12         64907534     19.80 ns/op                        0 B/op   0 allocs/op
BenchmarkXOR-12                               32868506     32.20 ns/op   31836.66 MB/s
BenchmarkXORSafe-12                            5185776    234.00 ns/op    4378.74 MB/s
BenchmarkXORFast-12                           30975679     32.50 ns/op   31525.28 MB/s
BenchmarkXORMappedAddress_AddTo-12            21518028     54.50 ns/op                        0 B/op   0 allocs/op
BenchmarkXORMappedAddress_GetFrom-12          35597667     34.40 ns/op                        0 B/op   0 allocs/op
ok      gortc.io/stun   60.973s
```

## License
BSD 3-Clause License
