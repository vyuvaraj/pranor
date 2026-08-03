module github.com/vyuvaraj/pranor/chrono

go 1.25.0

require (
	github.com/redis/go-redis/v9 v9.20.0
	github.com/vyuvaraj/pranor/core v0.0.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/golang-jwt/jwt/v5 v5.3.1 // indirect
	github.com/klauspost/cpuid/v2 v2.3.0 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/stretchr/testify v1.11.1 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	golang.org/x/sys v0.46.0 // indirect
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
