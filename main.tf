module "github_repository" {
  source                    = "./modules/github-repository"
  github_owner              = var.GITHUB_OWNER
  github_token              = var.GITHUB_TOKEN
  repository_name           = var.FLUX_GITHUB_REPO
  public_key_openssh        = module.tls_private_key.public_key_openssh
  public_key_openssh_title  = "flux0"
  secret_github_token       = "TF_GITHUB_TOKEN"
  secret_github_token_value = var.GITHUB_TOKEN
}

module "gke_cluster" {
  source         = "./modules/gke-cluster"
  GOOGLE_REGION  = var.GOOGLE_REGION
  GOOGLE_PROJECT = var.GOOGLE_PROJECT
  GKE_NUM_NODES  = 2
}

module "flux_bootstrap" {
  source            = "github.com/den-vasyliev/tf-fluxcd-flux-bootstrap"
  github_repository = "${var.GITHUB_OWNER}/${var.FLUX_GITHUB_REPO}"
  private_key       = module.tls_private_key.private_key_pem
  config_path       = module.gke_cluster.kubeconfig
  github_token = var.GITHUB_TOKEN
}

module "tls_private_key" {
  source = "github.com/den-vasyliev/tf-hashicorp-tls-keys"
}