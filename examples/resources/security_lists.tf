# Example: Security list with entries

resource "corex_network_list" "blocked_ips" {
  name        = "blocked-ips"
  description = "Known malicious IP addresses"
}

resource "corex_network_list_entry" "bad_ip_1" {
  list_id = corex_network_list.blocked_ips.id
  value   = "10.0.0.0/24"
  note    = "Internal test range"
}

resource "corex_network_list_entry" "bad_ip_2" {
  list_id = corex_network_list.blocked_ips.id
  value   = "192.168.1.50"
  note    = "Known attacker"
}

# ASN list
resource "corex_asn_list" "blocked_asns" {
  name        = "blocked-asns"
  description = "Blocked autonomous systems"
}

resource "corex_asn_list_entry" "bad_asn" {
  list_id = corex_asn_list.blocked_asns.id
  value   = "AS12345"
  note    = "Abuse provider"
}

# Geo list
resource "corex_geo_list" "blocked_countries" {
  name        = "blocked-countries"
  description = "Blocked countries"
}

resource "corex_geo_list_entry" "blocked_country" {
  list_id = corex_geo_list.blocked_countries.id
  value   = "CN"
  note    = "Geo-blocking"
}
