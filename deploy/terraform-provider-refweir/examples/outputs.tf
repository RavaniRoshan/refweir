output "namespace_id" {
  description = "Assigned unique ULID identifier of the created namespace"
  value       = refweir_namespace.platform.id
}

output "repository_clone_url_http" {
  description = "Smart HTTP clone URL for the created repository"
  value       = "${var.refweir_endpoint}/${refweir_namespace.platform.slug}/${refweir_repository.payment_gateway.name}.git"
}

output "repository_clone_url_ssh" {
  description = "Pure-Go SSH clone URL for the created repository"
  value       = "git@${replace(replace(var.refweir_endpoint, "http://", ""), "https://", "")}:${refweir_namespace.platform.slug}/${refweir_repository.payment_gateway.name}.git"
}
