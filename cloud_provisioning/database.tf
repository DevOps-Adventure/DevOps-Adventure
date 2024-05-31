resource "digitalocean_database_cluster" "db_minitwit" {
  name       = "db_minitwit"
  engine     = "mysql"
  version    = "8"
  size       = "db-s-1vcpu-1gb"
  region     = "fra1"
  node_count = 1
  tags       = ["production"]
}

resource "digitalocean_database_cluster" "db_minitwit_backup" {
  name       = "db_minitwin_backup"
  engine     = "mysql"
  version    = "8"
  size       = "db-s-1vcpu-1gb"
  region     = "fra1"
  node_count = 1
  tags       = ["production"]

  backup_restore {
    database_name = "db_minitwin_backup"
  }

  depends_on = [
    digitalocean_database_cluster.db_minitwit
  ]
}
