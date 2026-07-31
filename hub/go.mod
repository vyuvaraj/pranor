module github.com/vyuvaraj/pranor/hub

go 1.24

require (
	github.com/aws/aws-sdk-go-v2 v1.42.0
	github.com/aws/aws-sdk-go-v2/config v1.32.25
	github.com/aws/aws-sdk-go-v2/credentials v1.19.24
	github.com/aws/aws-sdk-go-v2/service/s3 v1.104.0
	github.com/vyuvaraj/pranor/core v0.0.0
)

require (
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.7.13 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.18.29 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.4.29 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.7.29 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.4.30 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.12 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.9.22 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.13.29 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.19.29 // indirect
	github.com/aws/aws-sdk-go-v2/service/signin v1.2.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.31.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.36.6 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.43.3 // indirect
	github.com/aws/smithy-go v1.27.1 // indirect
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
