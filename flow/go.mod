module github.com/vyuvaraj/pranor/flow

go 1.25.0

require (
	github.com/go-sql-driver/mysql v1.10.0
	github.com/go-stomp/stomp/v3 v3.1.5
	github.com/lib/pq v1.12.3
	github.com/tetratelabs/wazero v1.12.0
	github.com/vyuvaraj/pranor/core v0.0.0
	modernc.org/sqlite v1.54.0
)

require (
	filippo.io/edwards25519 v1.2.0 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/golang-jwt/jwt/v5 v5.3.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/mattn/go-isatty v0.0.22 // indirect
	github.com/ncruces/go-strftime v1.0.0 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	golang.org/x/sys v0.46.0 // indirect
	modernc.org/libc v1.74.1 // indirect
	modernc.org/mathutil v1.7.1 // indirect
	modernc.org/memory v1.11.0 // indirect
)

replace github.com/vyuvaraj/pranor/lang => ../lang

replace github.com/vyuvaraj/pranor/auth => ../auth

replace github.com/vyuvaraj/pranor/cache => ../cache

replace github.com/vyuvaraj/pranor/deploy => ../deploy

replace github.com/vyuvaraj/pranor/console => ../console

replace github.com/vyuvaraj/pranor/chrono => ../chrono

replace github.com/vyuvaraj/pranor/flow => ../flow

replace github.com/vyuvaraj/pranor/gate => ../gate

replace github.com/vyuvaraj/pranor/lock => ../lock

replace github.com/vyuvaraj/pranor/notify => ../notify

replace github.com/vyuvaraj/pranor/mesh => ../mesh

replace github.com/vyuvaraj/pranor/pool => ../pool

replace github.com/vyuvaraj/pranor/pulse => ../pulse

replace github.com/vyuvaraj/pranor/hub => ../hub

replace github.com/vyuvaraj/pranor/secret => ../secret

replace github.com/vyuvaraj/pranor/core => ../core

replace github.com/vyuvaraj/pranor/vault => ../vault

replace github.com/vyuvaraj/pranor/trace => ../trace

replace github.com/vyuvaraj/pranor/tunnel => ../tunnel
