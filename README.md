# zms-devops-auth

Build and push to DockerHub
```shell
docker build -t devopszms2024/zms-devops-auth-service:latest .
docker push devopszms2024/zms-devops-auth-service:latest
```

Create namespace & setup keycloak & auth-service infrastructure

```shell
minikube addons enable ingress
istioctl install --set profile=demo -y
```
First time you can use apply
```shell
kubectl apply -R -f auth-k8s 
kubectl apply -R -f auth-istio
kubectl apply -R -f keycloak-k8s 
```
When you want to replace existing pod, svc... you should use this command 
```shell
kubectl replace --force -f auth-k8s
kubectl replace --force -f auth-istio
kubectl replace --force -f keycloak-k8s
```

```shell
kubectl get pods -n backend
kubectl describe pods POD -n backend
```