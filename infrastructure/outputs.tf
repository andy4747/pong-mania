output "PROFILE_IMAGE_S3_BUCKET_ID" {
    description = "ID of the S3 Bucket for storing user's profile images"
    value = aws_s3_bucket.profile_images.id
}

output "PROFILE_IMAGE_S3_BUCKET" {
    description = "Name of the S3 Bucket for storing user's profile images"
    value = aws_s3_bucket.profile_images.bucket
}

output "INSTANCE_PUBLIC_IPV4" {
    description = "Public IPV4 address of the pong mania ec2 instance"
    value = aws_eip.PM_ec2_eip.public_ip
}

output "ec2_instance_public_ip" {
    description = "Public IP address of the ec2 instance"
    value = aws_instance.pong-mania-compute.public_ip
}


output "ec2_instance_public_dns" {
    description = "Public dns name of the ec2 instance"
    value = aws_instance.pong-mania-compute.public_dns
}

output "ec2_instance_id" {
    description = "ID of the ec2 instance"
    value = aws_instance.pong-mania-compute.id
}

