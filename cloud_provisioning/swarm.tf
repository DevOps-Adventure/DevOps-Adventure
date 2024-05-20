# TODO: Figure out how to create a VM to place the Proxy!

# Leader
# create cloud vm
resource "digitalocean_droplet" "minitwit-swarm-leader" {
  image = "docker-20-04" // ubuntu-22-04-x64
  name = "minitwit-swarm-leader"
  region = var.region
  size = "s-2vcpu-4gb"
  # add public ssh key so we can access the machine
  ssh_keys = [digitalocean_ssh_key.minitwit.fingerprint]

  # specify a ssh connection
  connection {
    user = "root"
    host = self.ipv4_address
    type = "ssh"
    private_key = file(var.pvt_key)
    timeout = "2m"
  }

  # provisioner "file" {
  #   source = "stack/minitwit_stack.yml"
  #   destination = "/root/minitwit_stack.yml"
  # }

  provisioner "remote-exec" {
    inline = [
      # allow ports for docker swarm
      "ufw allow 2377/tcp",
      "ufw allow 7946",
      "ufw allow 4789/udp",
      # ports for apps
      "ufw allow 80",
      "ufw allow 8080",
      "ufw allow 8881",
      "ufw allow 9090",
      "ufw allow 3000",
      "ufw allow 9100",
      "ufw allow 24244",
      "ufw allow 24244/udp",
      "ufw allow 9200",
      "ufw allow 5601",
      # SSH
      "ufw allow 22",

      # initialize docker swarm cluster
      "docker swarm init --advertise-addr ${self.ipv4_address}"
    ]
  }
}

resource "null_resource" "swarm-worker-token" {
  depends_on = [digitalocean_droplet.minitwit-swarm-leader]

  # save the worker join token
  provisioner "local-exec" {
    command = "ssh -o 'ConnectionAttempts 3600' -o 'StrictHostKeyChecking no' root@${digitalocean_droplet.minitwit-swarm-leader.ipv4_address} -i ssh_key/terraform 'docker swarm join-token worker -q' > worker_token"
  }
}

resource "null_resource" "swarm-manager-token" {
  depends_on = [digitalocean_droplet.minitwit-swarm-leader]
  # save the manager join token
  provisioner "local-exec" {
    command = "ssh -o 'ConnectionAttempts 3600' -o 'StrictHostKeyChecking no' root@${digitalocean_droplet.minitwit-swarm-leader.ipv4_address} -i ssh_key/terraform 'docker swarm join-token manager -q' > manager_token"
  }
}

# Managers
# create cloud vm
resource "digitalocean_droplet" "minitwit-swarm-manager" {
  # create managers after the leader
  depends_on = [null_resource.swarm-manager-token]

  # number of vms to create
  count = 1

  image = "docker-20-04"
  name = "minitwit-swarm-manager-${count.index}"
  region = var.region
  size = "s-2vcpu-4gb"
  # add public ssh key so we can access the machine
  ssh_keys = [digitalocean_ssh_key.minitwit.fingerprint]

  # specify a ssh connection
  connection {
    user = "root"
    host = self.ipv4_address
    type = "ssh"
    private_key = file(var.pvt_key)
    timeout = "2m"
  }

  provisioner "file" {
    source = "manager_token"
    destination = "/root/manager_token"
  }

  provisioner "remote-exec" {
    inline = [
      # allow ports for docker swarm
      "ufw allow 2377/tcp",
      "ufw allow 7946",
      "ufw allow 4789/udp",
      # ports for apps
      "ufw allow 80",
      "ufw allow 8080",
      "ufw allow 8881",
      "ufw allow 9090",
      "ufw allow 3000",
      "ufw allow 9100",
      "ufw allow 24244",
      "ufw allow 24244/udp",
      "ufw allow 9200",
      "ufw allow 5601",
      # SSH
      "ufw allow 22",

      # join swarm cluster as managers
      "docker swarm join --token $(cat manager_token) ${digitalocean_droplet.minitwit-swarm-leader.ipv4_address}"
    ]
  }
}


# Workers
# create cloud vm
resource "digitalocean_droplet" "minitwit-swarm-worker" {
  # create workers after the leader
  depends_on = [null_resource.swarm-worker-token]

  # number of vms to create
  count = 3

  image = "docker-20-04"
  name = "minitwit-swarm-worker-${count.index}"
  region = var.region
  size = "s-2vcpu-4gb"
  # add public ssh key so we can access the machine
  ssh_keys = [digitalocean_ssh_key.minitwit.fingerprint]

  # specify a ssh connection
  connection {
    user = "root"
    host = self.ipv4_address
    type = "ssh"
    private_key = file(var.pvt_key)
    timeout = "2m"
  }

  provisioner "file" {
    source = "worker_token"
    destination = "/root/worker_token"
  }

  provisioner "remote-exec" {
    inline = [
      # allow ports for docker swarm
      "ufw allow 2377/tcp",
      "ufw allow 7946",
      "ufw allow 4789/udp",
      # ports for apps
      "ufw allow 80",
      "ufw allow 8080",
      "ufw allow 8881",
      "ufw allow 9090",
      "ufw allow 3000",
      "ufw allow 9100",
      "ufw allow 24244",
      "ufw allow 24244/udp",
      "ufw allow 9200",
      "ufw allow 5601",
      # SSH
      "ufw allow 22",

      # join swarm cluster as workers
      "docker swarm join --token $(cat worker_token) ${digitalocean_droplet.minitwit-swarm-leader.ipv4_address}"
    ]
  }
}

output "minitwit-swarm-leader-ip-address" {
  value = digitalocean_droplet.minitwit-swarm-leader.ipv4_address
}

output "minitwit-swarm-manager-ip-address" {
  value = digitalocean_droplet.minitwit-swarm-manager.*.ipv4_address
}

output "minitwit-swarm-worker-ip-address" {
  value = digitalocean_droplet.minitwit-swarm-worker.*.ipv4_address
}