name           = "prod-app"
instance_count = 3
enabled        = false
tags = {
  env  = "prod"
  team = "platform"
}
subnets = ["10.0.1.0/24", "10.0.2.0/24"]
settings = {
  tier     = "premium"
  replicas = 5
}
anything = "whatever"
