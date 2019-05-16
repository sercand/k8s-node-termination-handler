# Copyright 2015 Google Inc. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and

GO := go
DOCKER := docker
TAG?=$(shell git rev-parse --short HEAD)
REGISTRY?=gcr.io/otsimocloud
IMAGE=gke-node-termination-handler

all: presubmit build

build:
	@echo ">> building using docker => ${REGISTRY}/${IMAGE}:${TAG}"
	@$(DOCKER) build -t ${REGISTRY}/${IMAGE}:${TAG} -f Dockerfile.build .

format:
	@echo ">> formatting code"
	@$(GO) fmt ./...

vet:
	@echo ">> vetting code"
	@$(GO) vet $(go list)

presubmit: vet
	@echo ">> checking go formatting"
	@./build/check_gofmt.sh .
	@echo ">> checking file boilerplate"
	@./build/check_boilerplate.sh

push:
	docker push ${REGISTRY}/${IMAGE}:${TAG}

buildandpush:
	@echo ">> building using docker => ${REGISTRY}/${IMAGE}:${TAG}"
	@$(DOCKER) build -t ${REGISTRY}/${IMAGE}:${TAG} -f Dockerfile.build .
	@$(DOCKER) push ${REGISTRY}/${IMAGE}:${TAG}

.PHONY: all format vet presubmit build container push buildandpush
