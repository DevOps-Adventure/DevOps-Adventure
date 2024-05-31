resource "digitalocean_droplet" "proxy" {
    image = "docker-20-04"
    name = "proxy"
    region = "fra1"
    size = "s-2vcpu-4gb"
    ssh_keys = [digitalocean_ssh_key.minitwit.fingerprint]

    connection {
        user = "root"
        host = self.ipv4_address
        type = "ssh"
        private_key = file(var.pvt_key)
        timeout = "2m"
    }

    provisioner "remote-exec" {
        inline = [
        # HTTP
        "ufw allow 80",
        # HTTPS
        "ufw allow 443",
        # SSH
        "ufw allow 22",

        # initialize docker swarm cluster
        "docker compose up -d"
        ]
    }
}

output "proxy-ip-address" {
  value = digitalocean_droplet.proxy.ipv4_address
}