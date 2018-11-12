IMAGE ?= mumoshu/kube-n-pods-per-node-scheduler:canary

.PHONY: build
build:
	@IMAGE=$(IMAGE) scripts/build

.PHONY: push
push:
	@IMAGE=$(IMAGE) scripts/push

.PHONY: deploy
deploy:
	@IMAGE=$(IMAGE) scripts/deploy

.PHONY: undeploy
	@IMAGE=$(IMAGE) scripts/undeploy

.PHONY: tryit
tryit:
	@scripts/tryit

.PHONY: update-vendor
update-vendor:
	glide up --strip-vendor
