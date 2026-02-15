## Instructions

### Prepare Credentials

Retrieve you temporary credentials from Learner's Lab.
Enter your configuration when prompted:
```
aws configure
```

Enter your AWS Access Key ID, Secret Access Key, region (e.g., us-east-1), and output format (json).


### Apply Infrastructure
```
cd terraform
terraform init
terraform apply -auto-approve 
```

### Authenticate Docker to ECR
```
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin <ECRlink>
```

### Build the image
```
cd src
docker build -t cs6650-product-api .
```

# Tag it for ECR
```
docker tag cs6650-product-api:latest <ECR>/cs6650-product-api:latest
```
# Push
```
docker push <ECR>/cs6650-product-api:latest
```

Get public ip address


Send some requests:
```
curl http://<PUBLIC-IP-ADDRESS>:8080/albums
```

To test using locust:
```
python -m locust -f locustfile.py --host=http://<PUBLIC-IP-ADDRESS>:8080 --class-picker
```

## Clean Up
```
terraform destroy -auto-approve
```
