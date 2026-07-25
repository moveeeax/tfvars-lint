# A module with a required, explicitly non-nullable variable, plus a
# nullable = false variable that does have a default (Terraform substitutes the
# default there, so an explicit null is NOT an error).
variable "account_id" {
  description = "Required and may never be null."
  type        = string
  nullable    = false
}

variable "region" {
  description = "Non-nullable but defaulted; null falls back to the default."
  type        = string
  nullable    = false
  default     = "eu-west-1"
}
