SHELL := bash
NAME := web

# PrivoHub Drive builds its web client from our own fork rather than from
# upstream's prebuilt release tarball. The only divergence from upstream is
# source-level (storage units in packages/web-pkg/src/helpers/filesize.ts),
# so a plain build of the fork yields a correctly-formatted, license-compliant
# bundle with NO post-build patching of the minified assets.
#
# WEB_ASSETS_REF must stay equal to the branch name in PrivoHub/opencloud-web,
# which itself tracks an upstream web tag (see that repo's PRIVOHUB.md). Bumping
# the web client = repoint this ref and the fork branch together. The tag it
# tracks is upstream's own WEB_ASSETS_VERSION for the server tag this fork sits
# on: server v7.3.0 pairs with web v7.2.0.
WEB_ASSETS_REPO ?= https://github.com/PrivoHub/opencloud-web.git
WEB_ASSETS_REF  ?= privohub/v7.2.0

ifneq (, $(shell command -v go 2> /dev/null)) # suppress `command not found warnings` for non go targets in CI
include ../../.bingo/Variables.mk
endif

include ../../.make/default.mk
include ../../.make/go.mk
include ../../.make/release.mk
include ../../.make/docs.mk

.PHONY: node-generate-dev
node-generate-dev: build-web-assets

.PHONY: node-generate-prod
node-generate-prod: build-web-assets

# Clone the PrivoHub web fork at the pinned ref, build its release tarball from
# source, and unpack it into assets/core (+ the third-party licenses). This
# replaces upstream's download-of-a-prebuilt-tarball + our old sed patch: the
# storage-units change now lives in the fork's source, so what ships is genuinely
# the corresponding source (AGPL-clean) and an upstream rebundle can't silently
# revert it.
.PHONY: build-web-assets
build-web-assets:
	rm -rf assets/core assets/apps tmp/privohub-web
	mkdir -p assets/core assets/apps
	git clone --depth 1 --branch $(WEB_ASSETS_REF) $(WEB_ASSETS_REPO) tmp/privohub-web
	$(MAKE) -C tmp/privohub-web release
	tar xf tmp/privohub-web/release/web.tar.gz -C assets/core/
	mkdir -p ../../third-party-licenses/node/web
	tar xf tmp/privohub-web/release/third-party-licenses.tar.gz -C ../../third-party-licenses/node/web/

.PHONY: ci-node-save-licenses
ci-node-save-licenses: build-web-assets
