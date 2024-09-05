terraform {
  required_providers {
    st-domain-management = {
      source = "myklst/st-domain-management"
    }
  }
}

provider "st-domain-management" {
  endpoint = "http://localhost:10800"
}
