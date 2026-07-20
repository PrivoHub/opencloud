SHELL := bash
NAME := web
WEB_ASSETS_VERSION = v6.2.0
WEB_ASSETS_BRANCH = main

ifneq (, $(shell command -v go 2> /dev/null)) # suppress `command not found warnings` for non go targets in CI
include ../../.bingo/Variables.mk
endif

include ../../.make/default.mk
include ../../.make/go.mk
include ../../.make/release.mk
include ../../.make/docs.mk

.PHONY: node-generate-dev
node-generate-dev: pull-assets

.PHONY: node-generate-prod
node-generate-prod: download-assets


.PHONY: pull-assets
pull-assets:
	echo "using unreleased assets from branch $(WEB_ASSETS_BRANCH), this should not be used for official releases"
	git clean -xfd assets
	@if [ ! -d tmp/assets/core/origin ]; then \
		echo "no web assets found, initializing"; \
		git clone -b main --depth 1 https://github.com/opencloud-eu/web.git tmp/assets/core/origin; \
	else \
		echo "web assets found, updating"; \
		cd tmp/assets/core/origin; \
		if [[ `git status --porcelain` ]]; then echo "There are uncommitted changes in tmp/assets/core/origin"; exit 1; fi; \
		git checkout $(WEB_ASSETS_BRANCH); \
        git pull; \
	fi;
	make -C tmp/assets/core/origin release
	tar xfv tmp/assets/core/origin/release/web.tar.gz -C assets/core/
	$(MAKE) privohub-patch-filesize
	mkdir -p ../../third-party-licenses/node/web
	tar xfv tmp/assets/core/origin/release/third-party-licenses.tar.gz -C ../../third-party-licenses/node/web/

.PHONY: download-assets
download-assets:
	rm -rf assets/core assets/apps && mkdir -p assets/core assets/apps
	curl --fail -slL -o- https://github.com/opencloud-eu/web/releases/download/$(WEB_ASSETS_VERSION)/web.tar.gz | tar xzf - -C assets/core/
	$(MAKE) privohub-patch-filesize

# PrivoHub: the upstream SPA formats sizes with filesize({base: 10}) → decimal
# GB. Rewrite the minified bundle to base 2, which the filesize lib resolves to
# the IEC standard (KiB/MiB/GiB). Fails the build if the pattern is missing so
# a web bump that changes the bundle can't silently revert to GB.
.PHONY: privohub-patch-filesize
privohub-patch-filesize:
	@files=$$(grep -rl ',base:10,output:' assets/core/js); \
	if [ -z "$$files" ]; then \
		echo "privohub-patch-filesize: pattern ',base:10,output:' not found in assets/core/js — web bundle changed, update the patch"; \
		exit 1; \
	fi; \
	sed -i 's/,base:10,output:/,base:2,output:/g' $$files; \
	echo "privohub-patch-filesize: patched to IEC units (GiB) in: $$files"

.PHONY: ci-node-save-licenses
ci-node-save-licenses:
	@mkdir -p ../../third-party-licenses/node/web
	@curl --fail -slL -o- https://github.com/opencloud-eu/web/releases/download/$(WEB_ASSETS_VERSION)/third-party-licenses.tar.gz | tar xzf - -C ../../third-party-licenses/node/web
