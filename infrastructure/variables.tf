variable "aws_region" {
    description = "AWS Region"
    default = "us-east-1"
    type = string
}

variable "aws_access_key" {
    description = "aws access key"
    type = string
}


variable "aws_secret_key" {
    description = "aws secret key"
    type = string
}


variable "project_name" {
    description = "Application Name"
    default = "pong-mania"
    type = string
}

variable "environment" {
    description = "current running environment"
    default = "dev"
    type = string
}

variable "vpc_id" {
    description = "VPS Id where rds will be"
    type = string
}

variable "vpc_cidr" {
    description = "CIDR block of VPC" 
    type = string
}

variable "subnet_id" {
    description = "Subnet Id of the VPC, same vpc with RDS" 
    type = string
}

variable "ec2_key_name" {
    description = "Key name of the ssh"
    type = string
}

variable "github_token" {
    description = "github token for cloning the project inside ec2"
    type = string
}

variable "ec2_deployment_user" {
    type = string
    description = "username of the new ec2 instance user"
}

variable "ec2_deployment_password" {
    type = string
    description = "password of the new ec2 instance user"
}
