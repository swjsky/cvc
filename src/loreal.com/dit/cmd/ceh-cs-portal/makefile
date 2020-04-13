NAME = ceh-cs-portal
BUILDPATH = `go env GOPATH`
OUTPUT = ${BUILDPATH}/bin/loreal.com/${NAME}
PACKAGES = ${BUILDPATH}/src/loreal.com/dit/cmd/${NAME}
GIT_COMMIT = `git rev-parse HEAD | cut -c1-7`
DT = `date +'%Y%m%d-%H%M%S'`
VERSION = V0.5
BUILD_OPTIONS = -ldflags "-X main.Version=$(VERSION) -X main.CommitID=$(DT)"

default:
	mkdir -p ${OUTPUT}
	go build ${BUILD_OPTIONS} -o ${OUTPUT}/${NAME} ${PACKAGES}

windows:
	mkdir -p ${OUTPUT}
	go build ${BUILD_OPTIONS} -o ${OUTPUT}/${NAME}.exe ${PACKAGES}

race:
	mkdir -p ${OUTPUT}
	go build ${BUILD_OPTIONS} -race -o ${OUTPUT}/${NAME}-race ${PACKAGES}

health:
	curl http://localhost:1521/health
