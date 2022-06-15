CONTROLLER_TAG=1.0.0
CONTROLLER_IMAGE=josericardomcastro/nodechecker-controller
IMAGE_CODE_GENERATOR=josericardomcastro/kube-code-generator-tools:0.1.0
PROJECT_PACKAGE=github.com/josericardomcastro/nodechecker-controller

# Generate code api
generate-api:
	sudo docker run -it --rm \
 		-v ${PWD}:/go/src/${PROJECT_PACKAGE} \
		-e PROJECT_PACKAGE=${PROJECT_PACKAGE} \
		-e CLIENT_GENERATOR_OUT=${PROJECT_PACKAGE}/pkg/generated \
		-e APIS_PKG=${PROJECT_PACKAGE}/pkg/apis \
		-e GROUPS_VERSION="nodecontroller:v1" \
		-e GENERATION_TARGETS="all" \
		${IMAGE_CODE_GENERATOR}

# Generate crd
SOURCE_PROJECT=/go/src/${PROJECT_PACKAGE}
generate-crd:
	sudo docker run -it --rm \
	-v ${PWD}:${SOURCE_PROJECT} \
	-e GO_PROJECT_ROOT=${SOURCE_PROJECT} \
	-e CRD_TYPES_PATH=${SOURCE_PROJECT}/pkg/apis \
	-e CRD_OUT_PATH=${SOURCE_PROJECT}/manifests \
	${IMAGE_CODE_GENERATOR} ./generate-crd.sh


# Build binary
build-bin:
	docker run --rm  \
		-e CGO_ENABLED=0 \
		-v ${PWD}:/go/src/${PROJECT_NAME} \
		-w /go/src/${PROJECT_NAME} \
		-v ${GOPATH}/pkg:/go/pkg \
		golang:1.16 go build -o ${PROJECT_NAME} .

# Build image
build-image:
	docker build -t ${CONTROLLER_IMAGE}:${CONTROLLER_TAG} .

# Build image
push-image:
	docker push ${CONTROLLER_IMAGE}:${CONTROLLER_TAG}

fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...
