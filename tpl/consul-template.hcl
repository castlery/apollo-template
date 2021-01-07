log_level = "warn"

wait {
  min = "5s"
  max = "10s"
}

vault {
  address = "http://vault.castlery.internal:8200"

  grace        = "15s"
  unwrap_token = false
  renew_token  = true
}

template {
  source      = "/Users/caipeijun/go/src/apollo-go/tpl/application-default.yml.tmpl"
  destination = "/Users/caipeijun/go/src/apollo-go/tpl/application-default.yml"

  error_on_missing_key = true
  perms  = 0644
  backup = true
}

template {
  source      = "/Users/caipeijun/go/src/apollo-go/tpl/sentry.properties.tmpl"
  destination = "/Users/caipeijun/go/src/apollo-go/tpl/sentry.properties"

  error_on_missing_key = true
  perms  = 0644
  backup = true
}