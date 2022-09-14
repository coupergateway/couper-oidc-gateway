# Couper OIDC Gateway

A ready-to-use [Docker image](https://hub.docker.com/r/avenga/couper-oidc-gateway) to spin up an OpenID Connect gateway using the [Couper API Gateway](https://github.com/avenga/couper):

`docker pull avenga/couper-oidc-gateway`

## How it works

The gateway uses the Authorization Code Grant Flow to connect to an OpenID Provider.

The redirect endpoint for the flow is `/_couper/oidc/callback`.

You need to register a **confidential** client at the OpenID Provider with the redirect URI `https://<your-gateway-host>/_couper/oidc/callback`. During registration you get a client ID and a client secret.

Set at least the following variables:

* `OIDC_CONFIGURATION_URL`
* `OIDC_CLIENT_ID`
* `OIDC_CLIENT_SECRET`
* `TOKEN_SECRET`
* `BACKEND_ORIGIN`

## Environment Variables

> **Note**: Along with this container specific variables you can also configure the underlying Couper with its own
> [environment variable options](https://github.com/avenga/couper/blob/master/DOCKER.md#environment-options).
> For example you may have to use a custom ca-certificate with `COUPER_CA_FILE`.

| Variable                                 | Type     | Default                | Description                                                                                   | Example                                        |
|:-----------------------------------------|:---------|:-----------------------|:----------------------------------------------------------------------------------------------|:-----------------------------------------------|
| `OIDC_CONFIGURATION_URL`                 | string   | -                      | The URL of the OpenID configuration at the OpenID Provider                                    | `https://.../.well-known/openid-configuration` |
| `OIDC_CLIENT_ID`                         | string   | -                      | The client ID of the client registered at the OpenID Provider                                 | -                                              |
| `OIDC_CLIENT_SECRET`                     | string   | -                      | The client secret                                                                             | -                                              |
| `TOKEN_SECRET`                           | string   | `asdf`                 | The secret used for signing the access token (the signature algorithm is `HS256`)             | `$e(rE4`                                       |
| `TOKEN_TTL`                              | duration | `1h`                   | The time-to-live of the access token                                                          | `1h`                                           |
| `TOKEN_COOKIE_NAME`                      | string   | `_couper_access_token` | The name of the cookie storing the access token                                               | `_couper_access_token`                         |
| `VERIFIER_COOKIE_NAME`                   | string   | `_couper_authvv`       | The name of the cookie storing the verifier used for CSRF protection during the login process | `_couper_authvv`                               |
| `BACKEND_ORIGIN`                         | string   | -                      | The origin of the service to be protected                                                     | `https://www.example.com`                      |
| `BACKEND_HOSTNAME`                       | string   | -                      | The value of the HTTP host header field for the request to the protected service              | -                                              |
| `BACKEND_CONNECT_TIMEOUT`                | duration | `10s`                  | The total timeout for dialing and connect to the origin                                       | -                                              |
| `BACKEND_TTFB_TIMEOUT`                   | duration | `60s`                  | The duration from writing the full request to the origin and receiving the answer             | -                                              |
| `BACKEND_TIMEOUT`                        | duration | `300s`                 | The total deadline duration a backend request has for write and read/pipe                     | -                                              |
| `BACKEND_DISABLE_CERTIFICATE_VALIDATION` | bool     | `false`                | Disables the peer certificate validation for the protected service                            | -                                              |
| `REQUIRED_EMAIL_DOMAINS`                 | string   | -                      | If non-empty: a list of comma-separated domains, one of which the user's email address (as per the ID token's `email` claim) must have, in order for the user to pass | `example.com,example.org` |

| Duration units | Description  |
|:---------------|:-------------|
| `ns`           | nanoseconds  |
| `us` (or `µs`) | microseconds |
| `ms`           | milliseconds |
| `s`            | seconds      |
| `m`            | minutes      |
| `h`            | hours        |

## Cookies

The following cookies are involved:

| Name          | Description                                                                                                       |
|:--------------|:------------------------------------------------------------------------------------------------------------------|
| access token  | The token providing access to the protected service, its name is configurable via `TOKEN_COOKIE_NAME`             |
| auth verifier | A verifier used for CSRF protection during the login process, its name is configurable via `VERIFIER_COOKIE_NAME` |
