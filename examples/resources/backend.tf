# Example: Backend with servers and listener

resource "corex_backend" "web" {
  name      = "web-pool"
  mode      = "http"
  protocol  = "http"
  algorithm = "roundrobin"

  health_check_enabled  = true
  health_check_interval = 5000
  health_check_uri      = "/health"
  health_check_method   = "GET"

  retries    = 3
  redispatch = true
}

resource "corex_server" "web1" {
  backend_id = corex_backend.web.id
  name       = "web-server-1"
  address    = "10.0.0.1"
  port       = 8080
  weight     = 100
  check      = true
}

resource "corex_server" "web2" {
  backend_id = corex_backend.web.id
  name       = "web-server-2"
  address    = "10.0.0.2"
  port       = 8080
  weight     = 100
  check      = true
}

resource "corex_listener" "https" {
  name           = "https-listener"
  bind_address   = "0.0.0.0"
  bind_port      = 443
  mode           = "http"
  protocol       = "https"
  ssl_enabled    = true
  http2          = true
  default_backend_id = corex_backend.web.id
}
