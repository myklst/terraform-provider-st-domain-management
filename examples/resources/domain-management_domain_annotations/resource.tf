resource "domain-management_domain_annotations" "example" {
  domain = "example.xyz"
  annotations = jsonencode({
    "top-level/module-specific/annotations" = {
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
  })
}
