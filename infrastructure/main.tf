resource "aws_s3_bucket" "profile_images" {
    bucket = "${var.project_name}-${var.environment}-profile-images"
    tags = {
        Name = "${var.project_name} profile image"
        Environment = var.environment
    }
}

resource "aws_s3_bucket_cors_configuration" "profile_images" {
    bucket = aws_s3_bucket.profile_images.id 
    cors_rule {
        allowed_headers = ["*"]
        allowed_methods = ["GET", "PUT", "POST"]
        allowed_origins = ["*"]
        max_age_seconds = 3000
    }
}

resource "aws_s3_bucket_ownership_controls" "profile_images" {
    bucket = aws_s3_bucket.profile_images.id
    rule {
      object_ownership = "BucketOwnerPreferred"
    }
}

resource "aws_s3_bucket_acl" "profile_images" {
    depends_on = [ aws_s3_bucket_ownership_controls.profile_images ] 
    bucket = aws_s3_bucket.profile_images.id
    acl = "private"
}

data "aws_ami" "amazon_linux_2" {
    most_recent = true
    owners = ["amazon"]

    filter {
        name = "name"
        values = ["amzn2-ami-hvm-*-x86_64-gp2"]
    }

    filter {
        name   = "virtualization-type"
        values = ["hvm"]
    }
}


resource "aws_security_group" "ec2_sg" {
    name = "${var.project_name}-${var.environment}-ec2-sg" 
    description = "Security Group for Pong Mania EC2 Instance"
    vpc_id = var.vpc_id
    tags = {
        Name = "EC2 Instance Security Group"
    }
}

resource "aws_vpc_security_group_ingress_rule" "ec2_allow_http" {
    security_group_id = aws_security_group.ec2_sg.id
    ip_protocol = "tcp"
    from_port = 80
    cidr_ipv4 = var.vpc_cidr
    to_port = 80
}

resource "aws_vpc_security_group_ingress_rule" "ec2_allow_https" {
    security_group_id = aws_security_group.ec2_sg.id
    ip_protocol = "tcp"
    from_port = 443
    cidr_ipv4 = var.vpc_cidr
    to_port = 443
}

resource "aws_vpc_security_group_ingress_rule" "ec2_allow_tls" {
    security_group_id = aws_security_group.ec2_sg.id
    ip_protocol = "tcp"
    cidr_ipv4 = var.vpc_cidr
    from_port = 8080
    to_port = 8080
}

resource "aws_vpc_security_group_ingress_rule" "ec2_allow_ssh" {
    security_group_id = aws_security_group.ec2_sg.id
    ip_protocol = "tcp"
    cidr_ipv4 = var.vpc_cidr
    from_port = 22
    to_port = 22
}

resource "aws_vpc_security_group_egress_rule" "ec2_outbound_rule" {
    security_group_id = aws_security_group.ec2_sg.id
    ip_protocol = "-1"
    cidr_ipv4 = var.vpc_cidr
    from_port = 0
    to_port = 0
}

resource "aws_eip" "PM_ec2_eip" {
    domain = "vpc"
    tags = {
        Name = "PM-EC2-EIP"
    }
}

resource "aws_instance" "pong-mania-compute" {
    ami = data.aws_ami.amazon_linux_2.id
    instance_type = "t2.micro"
    vpc_security_group_ids = [aws_security_group.ec2_sg.id]
    subnet_id = var.subnet_id //subnet from same vpc as RDS
    associate_public_ip_address = true
    key_name = aws_key_pair.PM_KEY.key_name
    user_data = <<-EOF
                    #!/bin/bash
                    set -euo pipefail

                    # Logging function 
                    log() {
                        echo "[$(date +'%Y-%m-%d %H:%M:%S')] $1" | tee -a /var/log/user-data.log
                    }

                    # Error handling function
                    error_exit() {
                        log "ERROR: $1" >&2
                        exit 1
                    }

                    # Check if a command exists
                    command_exists() {
                        command -v "$1" >/dev/null 2>&1
                    }

                    # Create log file
                    touch /var/log/user-data.log || error_exit "Unable to create log file"
                    log "Starting user data script execution"

                    # Update the system
                    log "Updating the system"
                    sudo yum update -y || error_exit "System update failed"

                    # Install and configure Docker
                    if ! command_exists docker; then
                        log "Installing Docker"
                        sudo amazon-linux-extras install docker -y || error_exit "Docker installation failed"
                        sudo systemctl enable docker || error_exit "Failed to enable Docker service"
                        sudo systemctl start docker || error_exit "Failed to start Docker service"
                    else
                        log "Docker is already installed"
                    fi

                    # Install Docker Compose
                    if ! command_exists docker-compose; then
                        log "Installing Docker Compose"
                        sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$$(uname -s)-$$(uname -m)" -o /usr/local/bin/docker-compose || error_exit "Failed to download Docker Compose"
                        sudo chmod +x /usr/local/bin/docker-compose || error_exit "Failed to make Docker Compose executable"
                    else
                        log "Docker Compose is already installed"
                    fi
                    # Install Git
                    log "Installing Git"
                    sudo yum install -y git || error_exit "Git installation failed"

                    # Install jq
                    log "Installing jq"
                    sudo yum install -y jq || error_exit "jq installation failed"

                    log "Tools setup completed successfully"

                    # Setup for application service
                    log "Setting up application service"
                    sudo mkdir -p /app || error_exit "Failed to create /app directory"

                    # Clone the private GitHub repository
                    log "Cloning the private GitHub repository"
                    GITHUB_REPO="https://github.com/andy4747/pong-mania.git"
                    # TODO change the token
                    GITHUB_TOKEN="xxxxxxxxx"

                    cd /app || error_exit "Failed to change directory to /app"
                    git clone https://\$${GITHUB_TOKEN}@github.com/andy4747/pong-htmx.git . || error_exit "Failed to clone the repository"

                    # Create the service file
                    cat <<EOT | sudo tee /etc/systemd/system/mania.service || error_exit "Failed to create service file"
                    [Unit]
                    Description=Pong Mania Service
                    After=docker.service
                    Requires=docker.service

                    [Service]
                    WorkingDirectory=/app
                    ExecStart=/usr/local/bin/docker-compose up
                    ExecStop=/usr/local/bin/docker-compose down
                    User=ec2-user
                    Restart=always

                    [Install]
                    WantedBy=multi-user.target
                    EOT

                    # Enable and start the service
                    log "Enabling and starting the application service"
                    sudo systemctl enable mania.service || error_exit "Failed to enable application service"
                    sudo systemctl start mania.service || error_exit "Failed to start application service"

                    log "User data script execution completed successfully"

                    EOF
    tags = {
        Name = "${var.project_name}-${var.environment}-app-server"
    }
}

resource "aws_eip_association" "PM_eip_association" {
    instance_id = aws_instance.pong-mania-compute.id
    allocation_id = aws_eip.PM_ec2_eip.id
}

resource "aws_key_pair" "PM_KEY" {
    key_name = var.ec2_key_name
    public_key = tls_private_key.PM_rsa.public_key_openssh
}

resource "tls_private_key" "PM_rsa" {
    algorithm = "RSA"
    rsa_bits = 4096
}

resource "local_file" "PM_PRIVATE_KEY_OPEN_SSH" {
    filename = "${var.ec2_key_name}_open_ssh"
    content = tls_private_key.PM_rsa.private_key_openssh
}

resource "local_file" "PM_PRIVATE_KEY_PEM" {
    filename = "${var.ec2_key_name}_pem"
    content = tls_private_key.PM_rsa.private_key_pem
}

resource "null_resource" "set_key_perms" {
    depends_on = [ local_file.PM_PRIVATE_KEY_PEM, local_file.PM_PRIVATE_KEY_OPEN_SSH ]
    provisioner "local-exec" {
        command = <<-EOT
            chmod 400 ${local_file.PM_PRIVATE_KEY_OPEN_SSH.filename}
            chmod 400 ${local_file.PM_PRIVATE_KEY_PEM.filename}
        EOT

        interpreter = [ "bash", "-c" ]
    }
}

