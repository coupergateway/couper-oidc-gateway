# Couper OIDC Gateway

A ready to use [Couper](https://github.com/avenga/couper) image to spin up an OpenID Connect gateway.

_coming soon_

The gateway uses the Authorization Code Grant Flow to connect to an OpenID Provider.

The redirect endpoint for the flow is `/_couper/oidc/callback`.

You need to register a **confidential** client at the OpenID Provider with the redirect URI `https://<your-gateway-host>/_couper/oidc/callback`. During registration you get a client ID and a client secret.

## Environment Variables

| Variable | Type | Default | Description | Example |
| :------- | :--- | :------ | :---------- | :------ |
| `OIDC_CONFIGURATION_URL` | string | - | The URL of the OpenID configuration at the OpenID Provider | `https://.../.well-known/openid-configuration` |
| `OIDC_CLIENT_ID` | string | - | The client ID of the client registered at the OpenID Provider | - |
| `OIDC_CLIENT_SECRET` | string | - | The client secret | - |
| `TOKEN_SECRET` | string | `asdf` | The secret used for signing the access token (the signature algorithm is `HS256`) | `$e(rE4` |
| `TOKEN_TTL` | duration | `1h` | The time-to-live of the access token | `1h` |
| `TOKEN_COOKIE_NAME` | string | `_couper_access_token` | The name of the cookie storing the access token | `_couper_access_token` |
| `ORIGIN` | string | - | The origin of the service to be protected | `https://www.example.com` |
| `ORIGIN_HOSTNAME` | string | - | The value of the HTTP host header field for the request to the protected service | - |

| Duration units | Description  |
| :------------- | :----------- |
| `ns`           | nanoseconds  |
| `us` (or `µs`) | microseconds |
| `ms`           | milliseconds |
| `s`            | seconds      |
| `m`            | minutes      |
| `h`            | hours        |

## Cookies

The following cookies are involved:

| Name | Description |
| :--- | :---------- |
| access token | The token providing access to the protected service, its name is configurable via `TOKEN_COOKIE_NAME` |
| `authvv` | A verifier used for CSRF protection during the login process |
