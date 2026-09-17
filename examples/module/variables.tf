variable "name" {
  description = "Name of the deployment."
  type        = string
}

variable "instance_count" {
  description = "Number of instances."
  type        = number
  default     = 1
}

variable "enabled" {
  description = "Whether the resource is enabled."
  type        = bool
  default     = true
}

variable "tags" {
  description = "Tags applied to resources."
  type        = map(string)
  default     = {}
}

variable "subnets" {
  description = "Subnet CIDRs."
  type        = list(string)
  default     = []
}

variable "settings" {
  description = "Structured settings."
  type = object({
    tier    = string
    replicas = number
  })
  default = {
    tier     = "standard"
    replicas = 2
  }
}

variable "anything" {
  description = "Untyped value."
  default     = null
}
