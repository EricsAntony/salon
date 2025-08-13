# user-service Kubernetes deployments

This directory uses Kustomize to manage environments.

Structure:
- `base/`: shared Deployment, Service, ConfigMap
- `overlays/dev|stage|prod/`: environment-specific patches and tags

Deploy per environment:
```bash
kubectl apply -k deployments/k8s/overlays/dev
kubectl apply -k deployments/k8s/overlays/stage
kubectl apply -k deployments/k8s/overlays/prod
```

Notes:
- Create the Secret `user-service-secrets` in each target namespace with keys: `db_url`, `jwt_access_secret`, `jwt_refresh_secret`.
- Overlays set namespace to `salon-<env>`. Create namespaces if they don't exist:
  ```bash
  kubectl create ns salon-dev
  kubectl create ns salon-stage
  kubectl create ns salon-prod
  ```
- Overlays set image tag to `dev|stage|prod`. Adjust registry/repo in `images` if needed.
- Dev runs 1 replica, Stage 2, Prod 3. Prod uses `imagePullPolicy: Always`.

## Legacy (archived)

The previous flat manifests (`deployments/k8s/deployment.yaml`, `service.yaml`, `configmap.yaml`) have been archived to `deployments/k8s/archive/` and should no longer be applied. Use the Kustomize overlays under `deployments/k8s/overlays/` instead.
