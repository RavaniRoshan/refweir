variable "refweir_endpoint" {
  description = "Base URL of the self-hosted refweir.dev forge instance"
  type        = string
  default     = "http://127.0.0.1:8080"
}

variable "refweir_token" {
  description = "Personal access token or admin API key with full platform management scopes"
  type        = string
  sensitive   = true
}
