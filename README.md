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
