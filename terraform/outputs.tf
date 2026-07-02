output "cluster_name" {
  description = "EKS cluster name."
  value       = module.eks.cluster_name
}

output "cluster_endpoint" {
  description = "Endpoint for the EKS Kubernetes API."
  value       = module.eks.cluster_endpoint
}

output "cluster_region" {
  description = "AWS region of the cluster."
  value       = var.region
}

output "configure_kubectl" {
  description = "Command to update your kubeconfig for this cluster."
  value       = "aws eks update-kubeconfig --region ${var.region} --name ${module.eks.cluster_name}"
}

output "vpc_id" {
  description = "ID of the VPC hosting the cluster."
  value       = module.vpc.vpc_id
}
