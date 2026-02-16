## Instructions

### For local
```
cd src
go run main.go
```

### For test using docker
```
cd src
docker build -t product-api .
docker run -p 8080:8080 product-api
```

### FOR AWS - Prepare Credentials

Retrieve you temporary credentials from Learner's Lab.
Enter your configuration when prompted:
```
aws configure
```

Enter your AWS Access Key ID, Secret Access Key, region (e.g., us-east-1), and output format (json).


# Apply Infrastructure
```
cd terraform
terraform init
terraform apply -auto-approve 
```

# Get public ip address
```
aws ec2 describe-network-interfaces --network-interface-ids $(aws ecs describe-tasks --cluster $(terraform output -raw ecs_cluster_name) --tasks $(aws ecs list-tasks --cluster $(terraform output -raw ecs_cluster_name) --service-name $(terraform output -raw ecs_service_name) --query "taskArns[0]" --output text) --query "tasks[0].attachments[0].details[?name=='networkInterfaceId'].value" --output text) --query "NetworkInterfaces[0].Association.PublicIp" --output text
```


### Send some requests:
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
