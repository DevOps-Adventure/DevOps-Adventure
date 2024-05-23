#!/bin/bash

set -e

echo -e "\n--> Bootstrapping Minitwit\n"

echo -e "\n--> Loading environment variables from secrets file\n"
source secrets

echo -e "\n--> Checking that environment variables are set\n"
# check that all variables are set
[ -z "$TF_VAR_do_token" ] && echo "TF_VAR_do_token is not set" && exit

echo -e "\n--> Initializing terraform\n"
# initialize terraform
terraform init

# check that everything looks good
echo -e "\n--> Validating terraform configuration\n"
terraform validate

# create infrastructure
echo -e "\n--> Creating Infrastructure\n"
terraform apply -auto-approve

#copying files to the proxy droplet
echo -e "\n--> copying the files from swarm_files directory to the leader"
scp -r -i ~/.ssh/terraform -o StrictHostKeyChecking=no remote_files/* root@$(terraform output -raw proxy-ip-address)

#starting the proxy node with a backup service
echo -e "\n--> Starting proxy\n"
ssh \
    -o 'StrictHostKeyChecking no' \
    root@$(terraform output -raw proxy-ip-address) \
    -i ~/.ssh/terraform \
    'docker compose -f docker-compose.yml up -d --remove-orphans'

#copying files to the swarm leader
echo -e "\n--> copying the files from swarm_files directory to the leader"
scp -r -i ~/.ssh/terraform -o StrictHostKeyChecking=no swarm_files/* root@$(terraform output -raw minitwit-swarm-leader-ip-address)

# deploy the stack to the cluster
echo -e "\n--> Deploying the Minitwit stack to the cluster\n"
ssh \
    -o 'StrictHostKeyChecking no' \
    root@$(terraform output -raw minitwit-swarm-leader-ip-address) \
    -i ~/.ssh/terraform \
    'docker stack deploy minitwit -c docker-compose.yml'

echo -e "\n--> Done bootstrapping Minitwit"
echo -e "--> The dbs will need a moment to initialize, this can take up to a couple of minutes..."
echo -e "--> Site will be avilable @ http://$(terraform output -raw proxy-ip-address)"
echo -e "--> ssh to swarm leader with 'ssh root@\$(terraform output -raw minitwit-swarm-leader-ip-address) -i ~/.ssh/terraform'"
echo -e "--> To remove the infrastructure run: terraform destroy -auto-approve"
echo -e "!!! Remember to set manually the migration of the IP from the proxy to the swarm leader in the nginx config file !!!"
echo -e "!!! Remember to change the DNS ip redirection of gemst.dk !!!"