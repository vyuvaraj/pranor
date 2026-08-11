module github.com/vyuvaraj/pranor/tools/agentctl

go 1.25.0

replace github.com/vyuvaraj/pranor/eval => ../../eval

replace github.com/vyuvaraj/pranor/decision => ../../decision

replace github.com/vyuvaraj/pranor/core => ../../core

replace github.com/vyuvaraj/pranor/auth => ../../auth

replace github.com/vyuvaraj/pranor/lang => ../../lang

replace github.com/vyuvaraj/pranor/cache => ../../cache

replace github.com/vyuvaraj/pranor/vault => ../../vault

replace github.com/vyuvaraj/pranor/pool => ../../pool

replace github.com/vyuvaraj/pranor/trace => ../../trace

replace github.com/vyuvaraj/pranor/learn/api => ../../learn/api

require (
	github.com/vyuvaraj/pranor/decision v0.0.0-00010101000000-000000000000
	github.com/vyuvaraj/pranor/eval v0.0.0-00010101000000-000000000000
)
