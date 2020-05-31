variable "subnet" {
  description = "List of subnet to add in the master alias"
  type        = set(string)
  default     = ["1.1.1.1", "1.1.1.2"]
}

resource "opnsense_firewall_alias_util" "test" {
  provider = opnsense.ntnu
  for_each = var.subnet
  name     = "test"
  address  = each.value
}
