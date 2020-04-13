PWD= $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
NAME= $(notdir $(PWD))
GOPATH = $(abspath ../../../../../)
# BUILDPATH = `go env GOPATH`
BUILDPATH = ${GOPATH}
# GOPATH = '%CD%/../../'
OUTPUT = ${BUILDPATH}/bin/
PACKAGES = loreal.com/dit/cmd/${NAME}
GIT_COMMIT = `git rev-parse HEAD | cut -c1-7`
DT = `date +'%Y%m%d-%H%M%S'`
VERSION = V0.1
BUILD_OPTIONS = -ldflags "-X main.Version=$(VERSION) -X main.CommitID=$(DT)"

default:
	mkdir -p ${OUTPUT}
	go build ${BUILD_OPTIONS} -o ${OUTPUT}${NAME} ${PACKAGES}

linux:
	mkdir -p ${OUTPUT}
	GOOS=linux GOARCH=amd64 go build ${BUILD_OPTIONS} -o ${OUTPUT}${NAME}-x86-linux ${PACKAGES}

windows:
	mkdir -p ${OUTPUT}
	go build ${BUILD_OPTIONS} -o ${OUTPUT}${NAME}.exe ${PACKAGES}

race:
	mkdir -p ${OUTPUT}
	go build ${BUILD_OPTIONS} -race -o ${OUTPUT}${NAME}-race ${PACKAGES}

health:
	curl http://localhost:1503/health

test:
	go test ${PACKAGES}
