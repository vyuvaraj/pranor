# PranorConsole Makefile

.PHONY: build run test docker-build clean

build:
	go build -o pranor-console.exe main.go

run: build
	./pranorconsole.exe --port=8083

test:
	go test -v ./...

docker-build:
	docker build -t pranor-console:latest .

clean:
	@if exist pranor-console.exe del /f /q pranor-console.exe
