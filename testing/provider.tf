provider "opnsense" {
  url                  = "https://ingress.ntnu.fap.no"
  key                  = "PlAKFkZKYOwERZvAL+RcC5Y7HgBW7mcIN41bzTPfyhGz/8UlRHRVGaJopxEwP0s0BzmbWYokSq2oz+lM"
  secret               = "YQrM5kRE+u6oVY4DpTFr7rGhwz7fsLw/omqZe/nZi1puZa4ReSRvjkINL7PxLw305h4quKK8A1GBuIAU"
  allow_unverified_tls = true
  alias                = "ntnu"
}
