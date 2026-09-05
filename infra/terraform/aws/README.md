# LogPulse AWS Terraform

This directory describes the AWS lab stage. It creates only the lab
network, ECR, EKS, GitHub OIDC deploy role, EKS access entry, add-ons and a
monthly budget alert. It intentionally does not create RDS, MSK, ElastiCache,
OpenSearch, NAT Gateway, Ingress or a service mesh.

## Before running it

1. Install Terraform and AWS CLI v2.
2. Sign in with an AWS identity that can create the listed resources.
3. Copy `terraform.tfvars.example` to `terraform.tfvars` and set the GitHub
   repository and budget email.
4. Review the cost estimate and the plan. The EKS control plane and EC2 nodes
   can incur charges. The checked-in example uses `t3.large`; the Free Tier
   lab run used a local ignored `terraform.tfvars` with `t3.small` x 2.

After the account and credentials are ready:

```powershell
terraform init
terraform fmt -check
terraform validate
terraform plan
terraform apply
terraform destroy
```

Do not commit `terraform.tfvars` or credentials. Runtime evidence for this lab
is kept locally; it does not claim production SLA or managed stateful services.
