# Couper OIDC Gateway

A ready to use [Couper](https://github.com/avenga/couper) image to spin up an OpenID Connect gateway.

_coming soon_

The gateway uses the Authorization Code Grant Flow to connect to an OpenID Provider.

The redirect endpoint for the flow is `/oidc/callback`.

## Environment Variables

| Variable | Type | Description | Example |
| :------- | :--- | :---------- | :------ |
| `OIDC_CONFIGURATION_URL` | string | The URL of the openid configuration | `https://.../.well-known/openid-configuration` |
| `OIDC_CLIENT_ID` | string | The client ID of the client registered at the OpenID Provider | - |
| `OIDC_CLIENT_SECRET` | string | The client secret | - |
| `TOKEN_SECRET` | string | The secret used for signing the token | `$e(rE4` |
| `TOKEN_TTL` | duration | The time-to-live of the token | `2h` |
| `ORIGIN` | string | The origin of the service to be protected | `https://www.example.com` |
| `ORIGIN_HOSTNAME` | string | The value of the HTTP host header field for the request to the protected service | - |

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
| `access_token` | The token providing access to the protected service |
| `authvv` | A verifier used for CSRF protection during the login process |
