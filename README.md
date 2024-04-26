Terraform Custom Provider for Domain Management
===============================================

This Terraform custom provider is designed for managing domains.

Supported Versions
------------------

| Terraform version | minimum provider version |maxmimum provider version
| ---- | ---- | ----|
| >= 1.3.x	| 0.1.1	| latest |

Requirements
------------

-	[Terraform](https://www.terraform.io/downloads.html) 1.3.x
-	[Go](https://golang.org/doc/install) 1.21 (to build the provider plugin)

Local Installation
------------------

1. Run make file `make install-local-domain-management` to install the provider under ~/.terraform.d/plugins.

2. The provider source should be change to the path that configured in the *Makefile*:

    ```
    terraform {
      required_providers {
        domain-management = {
          source = "example.local/myklst/domain-management"
        }
      }
    }

    provider "domain-management" {
      url = "http://localhost:10800
    }
    ```

## Resources
- **domain-management_domain_annotations**

	This resource is desgined to accept a domain name, and a Dynamic Terraform Object with a `JSON` like syntax, which will be assigned as annotations to the domain.

- #### Data Type Rules
1. Dot `.` and dollar sign `$` are not allowed for key name.
2. `Null` value or empty value on right hand side is not allowed.
3. `Object` cannot be empty.
4. `List / Set / Tuple` cannot have length of zero.
5. `List / Set / Tuple` must have contents that are of a single type.
6. `List / Set / Tuple` of objects, the object must have the exact same keys and nested keys.

- #### Data structure recommendations
1. Flat data structure is accepted.
```terraform
resource "..." "example" {
  domain = test.com
  annotations = {
		"common/devops/status" = true
		"common/devops/last-run" = "yesterday"
  }
}
```

2. Object data structure is also accepted.
```terraform
resource "..." "example" {
  domain = test.com
  annotations = {
		"common/devops" = {
			status = true
			"last-run" = "yesterday"
		}
  }
}
```

3. Note that in the two examples above `common/devops` and `common/devops/status` are stored as two separate root keys in the database. They will not be merged.
4. When in doubt, follow Kubernetes labels and annotations style.
5. Use `tolist([1,2,3])` to avoid using tuple. Tuple is not ordered and will cause terraform to see it as configuration drift.

- #### Terraform Resource Lifecyle
1. Each terraform module is responsible for their own annotations.
2. Module A would not, and should not interfere with annotations of Module B.
3. Each module takes ownership of single or multiple root keys, by having the equivalent key in its statefile.

- `DB Data`
```
annotations = {
	common/A = {...}
	common/B = {...}
	common/C = {...}
}
```

- `ModuleA`
```
resource "..." "example" {
  domain = test.com
  annotations = {
		"common/A" = {
			status = true
		}
  }
}
```

- `ModuleB`
```
resource "..." "example" {
  domain = test.com
  annotations = {
		"common/B" = {
			status = false
		}
  }
}
```
4. If root key exists, further `Create` reqest of the same key will fail.
5. `Update` is used to update the entire right hand side of a key.
6. `Update` cannot be used on a non-existent root key.
7. In Terraform's update lifecycle, root keys may be created, updated or deleted.
		Each will be handled by separate API calls.


## Data Sources
- **domain-management_domain_filter**

	This resource is desgined to filter domains based on key value, namely filter by labels and / or filter by tags. This data source returns a
	list of string of domain names for further use in `domain-management_domain_annotations`.
