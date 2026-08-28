variable "project_name" {
  type        = string
  description = "Prefix used for AWS resource names."
  default     = "logpulse"
}

variable "aws_region" {
  type        = string
  description = "AWS region for the lab."
  default     = "ap-northeast-1"
}

variable "vpc_cidr" {
  type        = string
  description = "CIDR for the VPC."
  default     = "10.20.0.0/16"
}

variable "availability_zones" {
  type        = list(string)
  description = "Exactly two availability zones for the public subnets."
  default     = ["ap-northeast-1a", "ap-northeast-1c"]

  validation {
    condition     = length(var.availability_zones) == 2
    error_message = "The lab uses exactly two availability zones."
  }
}

variable "public_subnet_cidrs" {
  type        = list(string)
  description = "CIDRs for the two public subnets."
  default     = ["10.20.1.0/24", "10.20.2.0/24"]

  validation {
    condition     = length(var.public_subnet_cidrs) == 2
    error_message = "The lab uses exactly two public subnets."
  }
}

variable "eks_version" {
  type        = string
  description = "EKS Kubernetes version."
  default     = "1.33"
}

variable "node_instance_type" {
  type        = string
  description = "Instance type for the single managed node."
  default     = "t3.large"
}

variable "node_count" {
  type        = number
  description = "Number of managed nodes for the lab runtime."
  default     = 1

  validation {
    condition     = var.node_count >= 1 && var.node_count <= 3
    error_message = "The lab uses between one and three managed nodes."
  }
}

variable "github_repository" {
  type        = string
  description = "GitHub repository in OWNER/REPOSITORY form for OIDC."
  default     = "OWNER/REPOSITORY"
}

variable "github_branch" {
  type        = string
  description = "Branch allowed to assume the deploy role."
  default     = "master"
}

variable "budget_email" {
  type        = string
  description = "Optional address for the monthly budget alert."
  default     = ""
}

variable "monthly_budget_usd" {
  type        = string
  description = "Monthly budget limit in USD."
  default     = "25"
}

variable "tags" {
  type        = map(string)
  description = "Tags added to every resource."
  default = {
    Project   = "logpulse"
    ManagedBy = "terraform"
    Stage     = "lab"
  }
}
