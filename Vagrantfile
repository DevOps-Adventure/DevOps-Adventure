# -*- mode: ruby -*-
# vi: set ft=ruby :

Vagrant.configure("2") do |config|
    config.vm.box = 'digital_ocean'
    config.vm.box_url = "https://github.com/devopsgroup-io/vagrant-digitalocean/raw/master/box/digital_ocean.box"
    config.ssh.private_key_path = '~/.ssh/do_ssh_key'
  
   # config.vm.synced_folder "remote_files", "/minitwit", type: "rsync"
    config.vm.synced_folder '.', '/vagrant', disabled: true
  
    config.vm.define "minitwit", primary: true do |server|
  
      server.vm.provider :digital_ocean do |provider|
        provider.ssh_key_name = "do_ssh_key"
        provider.token = ENV['DIGITAL_OCEAN_TOKEN']
        provider.image = 'ubuntu-22-04-x64'
        provider.region = 'fra1'
        provider.size = 's-1vcpu-1gb'
      end
  
      server.vm.hostname = "minitwit-go-server"
  
      server.vm.provision "shell", inline: <<-SHELL
      sudo apt-get update
      sudo apt-get install -y git
  
      # Install Go
      wget https://dl.google.com/go/go1.18.linux-amd64.tar.gz
      sudo tar -xvf go1.18.linux-amd64.tar.gz
      sudo mv go /usr/local
  
      # Setting up Go environment variables
      echo "export GOROOT=/usr/local/go" >> $HOME/.profile
      echo "export GOPATH=$HOME/go" >> $HOME/.profile
      echo "export PATH=$GOPATH/bin:$GOROOT/bin:$PATH" >> $HOME/.profile
      source $HOME/.profile
  
      # Verify Go installation
      go version
  
      # Your Go project setup can be added here
  
      echo -e "\nGo environment setup is complete."
      SHELL
    end
  end
  