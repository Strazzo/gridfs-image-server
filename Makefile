all: 
	go test -coverprofile /dev/null ...VoycerAG/config
	go test -coverprofile /dev/null ...VoycerAG/server

coverage-server:
	go test -coverprofile bin/coverage.out  .../VoycerAG/server
	go tool cover -html=bin/coverage.out

coverage-config:
	go test -coverprofile bin/coverage.out  .../VoycerAG/config
	go tool cover -html=bin/coverage.out