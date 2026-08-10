module github.com/vyuvaraj/pranor/learn/wasm

go 1.22

require (
	github.com/tetratelabs/wazero v1.8.0
	github.com/vyuvaraj/pranor/learn/api v0.0.0
)

replace github.com/vyuvaraj/pranor/learn/api => ../api
