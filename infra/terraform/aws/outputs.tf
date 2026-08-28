output "vpc_id" {
  value       = module.vpc.vpc_id
  description = "VPC ID."
}

output "ecr_repository_url" {
  value       = aws_ecr_repository.app.repository_url
  description = "Immutable ECR repository URL."
}

output "eks_cluster_name" {
  value       = module.eks.cluster_name
  description = "EKS cluster name."
}

output "eks_cluster_endpoint" {
  value       = module.eks.cluster_endpoint
  description = "EKS API endpoint."
}

output "github_deploy_role_arn" {
  value       = aws_iam_role.github_deploy.arn
  description = "GitHub Actions OIDC role ARN."
}
