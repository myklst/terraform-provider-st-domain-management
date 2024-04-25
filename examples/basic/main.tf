terraform {
  required_version = ">= 1.0, < 2.0"
  required_providers {
    domain-management = {
      source  = "example.local/myklst/domain-management"
      version = "0.1.0"
    }
  }
}

provider "domain-management" {
  endpoint = "http://localhost:10800"
}

import {
  id = "{\"domain\":\"pgb-st.com\",\"annotations\":{\"OmegaLUL\":\"\"}}"
  to = domain-management_domain_annotations.example["pgb-st.com"]
}

import {
  id = "{\"domain\":\"pgf-st.com\",\"annotations\":{\"OmegaLUL\":\"\"}}"
  to = domain-management_domain_annotations.example["pgf-st.com"]
}

data "domain-management_domain_filter" "example" {
  domain_labels = {
    "common/status" = "new"
  }
  domain_tags = {
    brand   = "pg"
    country = "default"
  }
}

resource "domain-management_domain_annotations" "example" {
  for_each = {
    for index, value in data.domain-management_domain_filter.example.domains :
    value => value
  }

  domain = each.value
  annotations = {
    "top-level/module-specific/annotationnnnnn" = {
      numberExample = 69
      floatExample  = 69.69
      stringExample = "hello"
      boolExample   = false
      objectExample = {
        this = {
          is = {
            an = "object"
          }
        }
      }
      listStringExample = tolist(["a", "b", "c"])
      listNumberExample = tolist([1, 2, 3])
      listBoolExample   = tolist([true, false, true])
      listObjectExample = tolist([
        {
          type   = "weekday"
          status = "active"
        },
        {
          type   = "weekend"
          status = "dormant"
        },
      ])
    }
  }
}
