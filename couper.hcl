server "oidc-gate" {

  // actual application
  endpoint "/**" {
    access_control = ["AccessToken"]
    proxy {
      websockets = true
      set_request_headers = {
        cookie = request.headers.cookie
        authorization = request.headers.authorization
      }
      backend {
        origin = env.ORIGIN
        hostname = env.ORIGIN_HOSTNAME != "" ? env.ORIGIN_HOSTNAME : request.host
      }
    }
  }

  // OIDC start login
  endpoint "/_couper/oidc/start" {
    response {
      status = 303
      headers = {
        cache-control = "no-cache,no-store"
        location = "${beta_oauth_authorization_url("oidc")}&state=${url_encode(relative_url(request.query.url[0]))}"
        set-cookie = "${env.VERIFIER_COOKIE_NAME}=${beta_oauth_verifier()};HttpOnly;Secure;Path=/_couper/oidc/callback"
      }
    }
  }

  // OIDC login callback
  endpoint "/_couper/oidc/callback" {
    access_control = ["oidc"]

    response {
      status = 303
      headers = {
        cache-control = "no-cache,no-store"
        set-cookie = [
          "${env.TOKEN_COOKIE_NAME}=${jwt_sign("AccessToken", {})}; HttpOnly; Secure; Path=/", # cannot use Max-Age=${env.TOKEN_TTL} here as long as TOKEN_TTL is a duration, because an integer is expected for Max-Age
          "${env.VERIFIER_COOKIE_NAME}=;HttpOnly;Secure;Path=/_couper/oidc/callback;Max-Age=0"
        ]
        location = relative_url(request.query.state[0])
      }
    }
  }
}

definitions {
  beta_oidc "oidc" {
    configuration_url = env.OIDC_CONFIGURATION_URL
    client_id = env.OIDC_CLIENT_ID
    client_secret = env.OIDC_CLIENT_SECRET
    redirect_uri = "/_couper/oidc/callback"
    verifier_value = request.cookies[env.VERIFIER_COOKIE_NAME]
  }

  jwt "AccessToken" {
    signature_algorithm = "HS256"
    key = env.TOKEN_SECRET
    signing_ttl = env.TOKEN_TTL
    cookie = env.TOKEN_COOKIE_NAME

    error_handler {
      response {
        status = 403
        headers = {
          cache-control = "no-cache,no-store"
          content-type = "text/html"
        }
        body = <<-EOB
<!DOCTYPE html><html><head>
<script>location.href = "/_couper/oidc/start?url=${url_encode(relative_url(request.url))}"</script>
<meta http-equiv="refresh" content="0;url=/_couper/oidc/start?url=${url_encode(relative_url(request.url))}"
</head><body><h1>Authentication required</h1>
<p><a href="/_couper/oidc/start?url=${url_encode(relative_url(request.url))}">Proceed to login</a></p>
<p>Authentication powered by <a href="https://github.com/avenga/couper-oidc-gateway" target="_blank">Couper OIDC Gateway</a></p>
</body></html>
EOB
      }
    }
  }
}

settings {
  accept_forwarded_url = ["proto", "host"]
  request_id_accept_from_header = "ingress-request-id"
}

defaults {
  environment_variables = {
    OIDC_CLIENT_ID = ""
    OIDC_CLIENT_SECRET = ""
    OIDC_CONFIGURATION_URL = ""
    TOKEN_SECRET = "asdf"
    TOKEN_TTL = "1h"
    TOKEN_COOKIE_NAME = "_couper_access_token"
    VERIFIER_COOKIE_NAME = "_couper_authvv"
    ORIGIN = ""
    ORIGIN_HOSTNAME = ""
  }
}
