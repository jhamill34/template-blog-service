# Checklist:
- [ ] Create EC2 instances and register them as docker swarm nodes (see below)
- [ ] Prepare your email sender with SES by creating SMTP credentials
- [ ] Clone down the repository (mainly just the deployments and configs directories)
- [ ] Prepare secrets (see below)
- [ ] Prepare environment variables (see below)
- [ ] Load configs with `./deployments/load_configs.sh`
- [ ] Load secrets with `./deployments/load_secrets.sh`
- [ ] Run `./deployments/deploy.sh`


## Creating docker swarm nodes

Make sure that the EC2 instances have docker and git installed. Add the following to 
the userdata or manually execute.


```bash
#!/bin/bash
sudo yum update
sudo yum -y install docker
sudo yum -y install git
service docker start
usermod -a -G docker ec2-user
chkconfig docker on
pip3 install docker-compose
```

Also, make sure that you attach a role to your instances so that you have at least:

```json
{
   "Version": "..." 
   "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "ecr:*"
            ],
            "Resource": "*"
        },
        {
            "Effect": "Allow",
            "Action": [
                "cloudwatch:PutMetricData",
                "logs:PutLogEvents",
                "logs:CreateLogStream",
                "logs:CreateLogGroup"
            ],
            "Resource": "*"
        }
   ]
}

```

These permissions will allow you to fetch your published containers and write logs to CloudWatch.

Tag one of the instances as `Main` and the rest as `Workers`

Make sure to also setup the cloudwatch agent to monitor the EC2 resource 

```bash 
sudo yum install amazon-cloudwatch-agent
sudo /opt/aws/amazon-cloudwatch-agent/bin/amazon-cloudwatch-agent-config-wizard 
sudo /opt/aws/amazon-cloudwatch-agent/bin/amazon-cloudwatch-agent-ctl -a fetch-config -m ec2 -s -c file:/opt/aws/amazon-cloudwatch-agent/bin/config.json
```

Use the following to get the IP addresses for your nodes

```bash
aws ec2 describe-instances \
        --filters "Name=instance-state-name,Values=running"  \
        --profile personal \
        --region us-east-1 | jq -r '.Reservations[].Instances[].PublicIpAddress'
```

Add those IP Addresses as "A Records" under your chosen domain names DNS records for both root 
and wildcard.

SSH into the Main node. 

```bash
IP=$(aws ec2 describe-instances \
        --filters "Name=instance-state-name,Values=running"  "Name=tag:Name,Values=Main" \
        --profile personal \
        --region us-east-1 | jq -r '.Reservations[].Instances[].PublicIpAddress')

ssh -i ~/.ssh/ec2-keys.pem ec2-user@${IP}
```

Initialize a docker swarm manager node with the IP address of the main node:

```bash
docker swarm init --advertise-addr ${IP}
```

This will spit out a command that you need to run on all the other nodes in your cluster.

## Prepare your secrets

Create a directory called `./secrets` 

```bash
mkdir ./secrets
```

Copy over the environment variable template and fill in the values in a file called `./secrets/env`

```
cp ./deployments/templates/secrets ./secrets/env
vim ./secrets/env 
```

Generate some key pairs:

```
./deployments/generate_key_pair.sh dkim
./deployments/generate_key_pair.sh token
```


Fetch your certificate:

```
mkdir .letsencrypt
cp ./deployments/templates/letsencrypt-owner.yaml ./.letsencrypt/owner.yaml
vim .letsencrypt/owner.yaml

./deployments/generate_key_pair.sh account
mv ./secrets/account-key.pem ./.letsencrypt/account-key.pem
rm ./secrets/account.pem

./deployments/generate_ssl_cert.sh 
```

The last command will require you updating a DNS record to fulfill the ACME challenge. 



## Prepare Environment Variables

Copy a template dotenv file into deployments to set values required for the deployment script:

```
cp ./deployments/templates/env ./deployments/.env 
# Or you can set it in the root directory 
# cp ./deployments/templates/env ./.env 

vim ./deployments/.env
```

