# Сборка и публикация Coin CI образов
# Использование:
#   make build
#   make push REGISTRY=registry.example.com/coin

REGISTRY ?= coin
TAG_PYTHON_UV ?= 3.13
TAG_PYTHON_PIP ?= 3.13
TAG_GO ?= 1.22
TAG_JAVA ?= 17
PLATFORMS ?= linux/amd64

.PHONY: build push python-uv python-pip go jvm-maven jvm-gradle

build: python-uv python-pip go jvm-maven jvm-gradle

python-uv:
	docker build \
		--platform $(PLATFORMS) \
		-t $(REGISTRY)/ci-python-uv:$(TAG_PYTHON_UV) \
		-f python-uv/Dockerfile \
		python-uv

python-pip:
	docker build \
		--platform $(PLATFORMS) \
		-t $(REGISTRY)/ci-python-pip:$(TAG_PYTHON_PIP) \
		-f python-pip/Dockerfile \
		python-pip

go:
	docker build \
		--platform $(PLATFORMS) \
		-t $(REGISTRY)/ci-go:$(TAG_GO) \
		-f go/Dockerfile \
		go

jvm-maven:
	docker build \
		--platform $(PLATFORMS) \
		-t $(REGISTRY)/ci-jvm-maven:$(TAG_JAVA) \
		-f jvm-maven/Dockerfile \
		jvm-maven

jvm-gradle:
	docker build \
		--platform $(PLATFORMS) \
		-t $(REGISTRY)/ci-jvm-gradle:$(TAG_JAVA) \
		-f jvm-gradle/Dockerfile \
		jvm-gradle

push: build
	docker push $(REGISTRY)/ci-python-uv:$(TAG_PYTHON_UV)
	docker push $(REGISTRY)/ci-python-pip:$(TAG_PYTHON_PIP)
	docker push $(REGISTRY)/ci-go:$(TAG_GO)
	docker push $(REGISTRY)/ci-jvm-maven:$(TAG_JAVA)
	docker push $(REGISTRY)/ci-jvm-gradle:$(TAG_JAVA)
