terraform {
  required_providers {
    domain-management = {
      source = "myklst/domain-management"
    }
  }
}

provider "domain-management" {
  endpoint = "http://localhost:10800"
}
