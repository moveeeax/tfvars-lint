# Exercises Terraform 1.3+ type-constraint syntax that plain
# typeexpr.TypeConstraint cannot parse. Both variables are defaulted so the
# fixture counts for valid.tfvars / bad.tfvars stay unchanged.
variable "scaling" {
  description = "Object type using both optional() forms."
  type = object({
    min      = number
    max      = optional(number, 10)
    strategy = optional(string)
  })
  default = {
    min = 1
  }
}
