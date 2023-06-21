provider "github" {
  owner = var.github_owner
  token = var.github_token
}

resource "github_repository" "this" {
  name       = var.repository_name
  visibility = var.repository_visibility
  auto_init  = true
}

resource "github_repository_deploy_key" "this" {
  title      = var.public_key_openssh_title
  repository = github_repository.this.name
  key        = var.public_key_openssh
  read_only  = false
}

resource "github_actions_secret" "secret_token" {
  repository       = var.repository_name
  secret_name      = var.secret_github_token
  plaintext_value  = var.secret_github_token_value
}