# Drupal Operator

Drupal Operator generated via KubeBuilder to enable managing multiple Drupal installs.

## Goals

The main goals of the operator are:

1. Ability to deploy Drupal sites on top of Kubernetes
2. Provide best practices for application lifecycle
3. Facilitate proper devops (backups, monitoring and high-availability)

> Project is currently under active development.

## Components

1. Drupal Operator (this project)
2. Drupal Container Image (https://github.com/drupalwxt/site-wxt)

## Installation of Controller (CRD)

```sh
helm repo add sylus https://sylus.github.io/charts
helm --name drupal-operator install sylus/drupal-operator
```

## Usage

First we need to install the mysql-operator as well as default role bindings.

```sh
# Create our namespace
kubectl create ns mysql-operator

# Install via Helm
helm install --name mysql-operator -f values.yaml --namespace mysql-operator .

# Install RoleBindings for appropriate namespace
cat <<EOF | kubectl create -f -
kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: mysql-agent
  namespace: default
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: mysql-agent
subjects:
- kind: ServiceAccount
  name: mysql-agent
  namespace: default
EOF
```

Next we can start to utilize the Drupal operator!

```sh

# Create initial database
kubectl run mysql-client --image=mysql:5.7 -it --rm --restart=Never -- mysql -h mysite-mysql -uroot -pmy-super-secret-pass -e 'create database drupal;'

# Deploy the operator (helm chart still being tested)
make deploy

# Leverage our example spec
kubectl apply -f config/samples/drupal_v1beta1_droplet.yaml

# Run Drush and install our site
export POD_NAME=$(kubectl get pods --namespace default -l "app.kubernetes.io/component=drupal" -o jsonpath="{.items[0].metadata.name}")
kubectl exec -it $POD_NAME -n default -- drush si wxt \
    --sites-subdir=default \
    --account-name=admin \
    --account-pass=Drupal@2019 \
    --site-mail=admin@example.com \
    --site-name="Drupal Install Profile (WxT)" \
    install_configure_form.update_status_module='array(FALSE,FALSE)' \
    --yes
```
