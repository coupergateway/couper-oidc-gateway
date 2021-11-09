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

  // OIDC login callback
  endpoint "/oidc/callback" {
    access_control = ["oidc"]

    response {
      status = 303
      headers = {
        cache-control = "no-cache,no-store"
        set-cookie = [
          "access_token=${jwt_sign("AccessToken", {})}; HttpOnly; Secure; Path=/", # cannot use Max-Age=${env.TOKEN_TTL} here as long as TOKEN_TTL is a duration, because an integer is expected for Max-Age
          "authvv=;HttpOnly;Secure;Path=/oidc/callback;Max-Age=0"
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
    redirect_uri = "/oidc/callback"
    verifier_value = request.cookies.authvv
  }

  jwt "AccessToken" {
    signature_algorithm = "HS256"
    key = env.TOKEN_SECRET
    signing_ttl = env.TOKEN_TTL
    cookie = "access_token"

    error_handler {
      response {
        status = 403
        headers = {
          cache-control = "no-cache,no-store"
          content-type = "text/html"
          set-cookie = "authvv=${beta_oauth_verifier()};HttpOnly;Secure;Path=/oidc/callback"
        }
        body = <<-EOB
<!DOCTYPE html><html><head>
<meta http-equiv="refresh" content="0;url=${beta_oauth_authorization_url("oidc")}&amp;state=${url_encode(relative_url(request.url))}"
</head><body>
<form action="${beta_oauth_authorization_url("oidc")}">
<input type="hidden" name="state" value="${relative_url(request.url)}">
<button type="submit">please log-in!</button>
</form>
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
    TOKEN_TTL = "2m"
    ORIGIN = ""
    ORIGIN_HOSTNAME = ""
  }
}
