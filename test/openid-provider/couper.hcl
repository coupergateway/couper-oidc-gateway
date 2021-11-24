server "oauth-as" {
  files {
    document_root = "./htdocs"
  }
  api {
    endpoint "/auth" {
      response {
        status = 303
        headers = {
          cache-control = "no-cache, no-store"
          location = "${request.query.redirect_uri[0]}?code=asdf${request.query.state != null ? "&state=${url_encode(request.query.state[0])}" : ""}"
        }
      }
    }
    endpoint "/token" {
      response {
        headers = {
          cache-control = "no-cache, no-store"
        }
        json_body = {
          access_token = jwt_sign("token", {
            aud = "http://host.docker.internal:8081"
            cid = request.headers.authorization != null ? split(":", base64_decode(substr(request.headers.authorization, 6, -1)))[0] : request.form_body.client_id[0]
            sub = "john"
          })
          id_token = jwt_sign("token", {
            aud = request.headers.authorization != null ? split(":", base64_decode(substr(request.headers.authorization, 6, -1)))[0] : request.form_body.client_id[0]
            sub = "john"
          })
          token_type = "Bearer"
          expires_in = 3600
        }
      }
    }
    endpoint "/userinfo" {
      access_control = ["token"]
      response {
        headers = {
          cache-control = "no-cache, no-store"
        }
        json_body = {
          sub = request.context.token.sub
        }
      }
    }
  }
}

definitions {
  jwt_signing_profile "token" {
    signature_algorithm = "HS256"
    ttl = "1h"
    key = "asdf"
    claims = {
      iss = "http://host.docker.internal:8081"
      iat = unixtime()
    }
  }

  jwt "token" {
    signature_algorithm = "HS256"
    key = "asdf"
    header = "authorization"
    required_claims = ["iat", "exp"]
    claims = {
      iss = "http://host.docker.internal:8081"
      aud = "http://host.docker.internal:8081"
    }
  }
}
