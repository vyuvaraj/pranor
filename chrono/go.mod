module github.com/vyuvaraj/pranor/chrono

go 1.23.0

require (
	github.com/redis/go-redis/v9 v9.7.0
	github.com/vyuvaraj/pranor/core v0.0.0
)

require (
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/golang-jwt/jwt/v5 v5.3.1 // indirect
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
