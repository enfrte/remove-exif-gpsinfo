# Docker & Azure Deployment Setup

This guide walks you through deploying the `exif-remover` service to Azure using Docker.

## Prerequisites

- Docker installed locally
- Azure CLI installed
- An Azure subscription
- Docker Hub or Azure Container Registry (ACR) account

## Local Development with Docker

### Build the Docker Image

```bash
docker build -t exif-remover:latest .
```

### Run Container Locally

```bash
docker run -p 8080:8080 -e PORT=8080 exif-remover:latest
```

Test with curl:

```bash
curl -v -F "image=@test-image.jpg" http://localhost:8080/remove-gps > output.jpg
```

### Using Docker Compose

```bash
docker-compose up --build
```

## Azure Deployment Options

### Option 1: Azure Container Registry (ACR) + App Service (Recommended)

#### 1.1 Create an Azure Container Registry

```bash
az acr create --resource-group <your-resource-group> \
  --name <your-registry-name> \
  --sku Basic
```

#### 1.2 Build and Push Image to ACR

```bash
az acr build --registry <your-registry-name> --image exif-remover:latest .
```

Or using Docker CLI:

```bash
# Login to ACR
az acr login --name <your-registry-name>

# Build and push
docker build -t <your-registry-name>.azurecr.io/exif-remover:latest .
docker push <your-registry-name>.azurecr.io/exif-remover:latest
```

#### 1.3 Create an App Service Plan

```bash
az appservice plan create \
  --name <your-plan-name> \
  --resource-group <your-resource-group> \
  --sku B1 \
  --is-linux
```

#### 1.4 Create Web App

```bash
az webapp create \
  --resource-group <your-resource-group> \
  --plan <your-plan-name> \
  --name <your-app-name> \
  --deployment-container-image-name <your-registry-name>.azurecr.io/exif-remover:latest
```

#### 1.5 Configure App Service Settings

```bash
az webapp config appsettings set \
  --resource-group <your-resource-group> \
  --name <your-app-name> \
  --settings DOCKER_REGISTRY_SERVER_URL=https://<your-registry-name>.azurecr.io \
  WEBSITES_PORT=8080 \
  PORT=8080

# Configure container credentials (if using Basic auth)
az webapp config container set \
  --resource-group <your-resource-group> \
  --name <your-app-name> \
  --docker-custom-image-name <your-registry-name>.azurecr.io/exif-remover:latest \
  --docker-registry-server-url https://<your-registry-name>.azurecr.io \
  --docker-registry-server-username <registry-username> \
  --docker-registry-server-password <registry-password>
```

### Option 2: Azure Container Instances (Quick Testing)

```bash
az container create \
  --resource-group <your-resource-group> \
  --name exif-remover \
  --image <your-registry-name>.azurecr.io/exif-remover:latest \
  --registry-login-server https://<your-registry-name>.azurecr.io \
  --registry-username <username> \
  --registry-password <password> \
  --ports 8080 \
  --environment-variables PORT=8080 \
  --cpu 1 \
  --memory 1
```

### Option 3: Azure Kubernetes Service (AKS)

Create a deployment manifest (deployment.yaml):

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: exif-remover
spec:
  replicas: 3
  selector:
    matchLabels:
      app: exif-remover
  template:
    metadata:
      labels:
        app: exif-remover
    spec:
      containers:
      - name: exif-remover
        image: <your-registry-name>.azurecr.io/exif-remover:latest
        ports:
        - containerPort: 8080
        env:
        - name: PORT
          value: "8080"
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "256Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 10
---
apiVersion: v1
kind: Service
metadata:
  name: exif-remover-service
spec:
  selector:
    app: exif-remover
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8080
  type: LoadBalancer
```

Deploy to AKS:

```bash
kubectl apply -f deployment.yaml
```

## CI/CD with Azure Pipelines

The `azure-pipelines.yml` file provides automated building and deployment:

1. Update the following placeholders in `azure-pipelines.yml`:
   - `your-acr-connection` - Your ACR service connection
   - `your-registry.azurecr.io` - Your ACR URL
   - `your-azure-subscription` - Your subscription
   - `your-app-service-name` - Your App Service name

2. Commit and push to trigger the pipeline:

```bash
git add .
git commit -m "Add Docker and Azure setup"
git push origin main
```

## Environment Variables

- `PORT` - Server port (default: 8080)

## Testing in Azure

Once deployed, test with:

```bash
curl -v -F "image=@test-image.jpg" https://<your-app-name>.azurewebsites.net/remove-gps > output.jpg
```

## Cleanup

Remove resources when no longer needed:

```bash
# Delete app service
az webapp delete --resource-group <your-resource-group> --name <your-app-name>

# Delete app service plan
az appservice plan delete --resource-group <your-resource-group> --name <your-plan-name>

# Delete ACR
az acr delete --resource-group <your-resource-group> --name <your-registry-name>
```

## Troubleshooting

### Container won't start
- Check logs: `az webapp log tail --resource-group <rg> --name <app-name>`
- Ensure `PORT` environment variable is set to 8080
- Verify `WEBSITES_PORT=8080` is configured

### Image pull failures
- Verify ACR credentials
- Check image name and tag
- Ensure ACR login server URL is correct

### Performance issues
- Monitor CPU/memory in Azure Portal
- Adjust App Service plan SKU
- Increase container resource limits in AKS

## Additional Resources

- [Azure App Service Documentation](https://docs.microsoft.com/en-us/azure/app-service/)
- [Azure Container Registry Docs](https://docs.microsoft.com/en-us/azure/container-registry/)
- [Docker Best Practices](https://docs.docker.com/develop/dev-best-practices/)
