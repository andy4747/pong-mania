terraform {
    required_providers {
        aws = {
            source = "hashicorp/aws"
            version = "~>5.66.0"
        }
        tls = {
            source = "hashicorp/tls"
            version = "~>4.0.0"
        }
        local = {
            source = "hashicorp/local"
            version = "~>2.5.1"
        }
        null = {
            source  = "hashicorp/null"
            version = "~> 3.1"
        }
    }
    required_version = ">=1.8.0"
}
