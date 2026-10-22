![OpenCloud logo](https://raw.githubusercontent.com/opencloud-eu/opencloud/refs/heads/main/opencloud_logo.png)

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

# PrivoHub Drive: OpenCloud fork

This is the **PrivoHub Drive** fork of [OpenCloud](https://github.com/opencloud-eu/opencloud).
The image is built from `opencloud/docker/Dockerfile.multiarch`; each release's source is the
`privohub-v{upstream}` branch of this repository. The fork rebases onto upstream tags and adds a
small set of additive commits: when none of the `PRIVOHUB_*` envs are set, the binary still
behaves like vanilla OpenCloud.

## PrivoHub additions on top of upstream

| Env / setting | What it does |
| --- | --- |
| `PRIVOHUB_PERSONAL_QUOTA_BYTES` | Per-tenant personal-drive quota, applied on every login (cached for 5 minutes per user). Zero or unset disables. |
| `PRIVOHUB_IDM_ADMIN_PASSWORD_FILE` | Path to a file containing the IDM admin password. Created at first boot by the entrypoint at `/var/lib/opencloud/.privohub/idm-admin`; read on every subsequent boot. The password is never logged or set via env. |
| `WEB_OIDC_CLIENT_ID=opencloud-{tenant}` + `OCIS_OIDC_ISSUER=<keycloak-issuer>` | OIDC public-client PKCE against the per-tenant Keycloak realm. No client secret is required; the proxy validates via JWKS. This is upstream-default behavior; the operator only needs to set the two envs. |
| `WEB_UI_THEME_PATH` | Defaults to `/themes/privohub/theme.json` in this fork (was `opencloud/`). Operators can still override. |

## Privacy posture

Upstream OpenCloud has no telemetry: there is no `*_TELEMETRY` env, no analytics SDK, no
phone-home. The single outbound hint in upstream config is a documented `feedbackLink` URL;
this fork removes that default so the only outbound HTTP traffic from a PrivoHub Drive pod is
to its OIDC issuer (Keycloak). Verify with `kubectl exec -- timeout 300 tcpdump -nn -i any
'not host <issuer-ip> and not net <cluster-cidr>'` and expect zero packets.

# Upstream OpenCloud server backend

> [!TIP]
> For general information about OpenCloud and how to install please visit [OpenCloud on Github](https://github.com/opencloud-eu/) and [OpenCloud GmbH](https://opencloud.eu).

This is the main repository of the OpenCloud server.
It contains the golang codebase for the backend services.

## Getting Involved

The OpenCloud server is released under [Apache 2.0](https://github.com/opencloud-eu/opencloud/blob/main/LICENSE).
The project is thrilled to receive contributions in all forms.
Start hacking now, there are many ways to get involved such as:

- Reporting [issues or bugs](https://github.com/opencloud-eu/opencloud/issues)
- Requesting [features](https://github.com/opencloud-eu/opencloud/issues)
- [Writing documentation](https://github.com/opencloud-eu/docs)
- [Writing code or extend our tests](https://github.com/opencloud-eu/opencloud/pulls)
- [Reviewing code](https://github.com/opencloud-eu/opencloud/pulls)
- Helping others in the [community](https://app.element.io/#/room/#opencloud:matrix.org)

Every contribution is meaningful and appreciated!
Please refer to our [Contribution Guidelines](https://github.com/opencloud-eu/opencloud/blob/main/CONTRIBUTING.md) if you want to get started.

## Build OpenCloud

To build the backend, follow these instructions:

Generate the assets needed by e.g., the web UI and the builtin IDP

``` console
make generate
```

Then compile the `opencloud` binary

``` console
make -C opencloud build
```
That will produce the binary `opencloud/bin/opencloud`. It can be started as a local test instance right away with a two step command:

```bash
opencloud/bin/opencloud init && opencloud/bin/opencloud server
```
This creates a server configuration (by default in `$HOME/.opencloud`) and starts the server.

For more setup- and installation options consult the [Development Documentation](https://docs.opencloud.eu/).

## Technology

Important information for contributors about the technology in use.

### Authentication

The OpenCloud backend authenticates users via [OpenID Connect](https://openid.net/connect/) using either an external IdP like [Keycloak](https://www.keycloak.org/) or the embedded [LibreGraph Connect](https://github.com/libregraph/lico) identity provider.

### Database

The OpenCloud backend does not use a database. It stores all data in the filesystem. By default, the root directory of the backend is `$HOME/.opencloud/`.

## Security

If you find a security-related issue, please contact [security@opencloud.eu](mailto:security@opencloud.eu) immediately.
