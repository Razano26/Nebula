# TP Kubernetes - Date limite: 12 juin

## Présentation

- Présentation orale en séance le 12 juin
- 20 minutes maximum par groupe
- Possibilité de travailler en binôme (ou trinôme si nombre impair de participants)
- Prévoir quelques slides pour la présentation

## Pré-requis

### Prérequis communs aux TP

Avant de commencer l'un de ces TP, les étudiants devront avoir mis en place un cluster Kubernetes fonctionnel en utilisant l'un des outils d'Infrastructure as Code suivants (au choix) :

#### Option 1 : Terraform

- Installation de Terraform
- Création d'un module Terraform pour déployer un cluster Kubernetes
- Documentation du processus et des variables utilisées

#### Option 2 : Ansible

- Installation d'Ansible
- Création de playbooks pour le provisionnement d'un cluster Kubernetes
- Configuration des inventaires et des rôles nécessaires
- Test et validation du déploiement

#### Option 3 : Cluster API

- Installation de clusterctl
- Configuration du provider d'infrastructure approprié
- Déploiement d'un cluster de management
- Utilisation des CRDs pour créer un cluster cible
- Documentation des manifestes utilisés

### Exigences minimales pour le cluster

- Version Kubernetes : 1.29+
- Minimum 3 nœuds worker (2 vCPU, 4 Go RAM chacun)
- Système de stockage persistant configuré (CSI)
- Accès administrateur au cluster
- kubectl configuré et fonctionnel
- Metrics Server installé
- Réseau CNI fonctionnel avec support pour les Network Policies

Le tout sera déployé sur l'infra OpenStack de Polytech.

### Documentation requise

Les étudiants devront fournir la documentation de leur processus d'installation, y compris :

- Code source IaC utilisé
- Journal des problèmes rencontrés et solutions appliquées
- Commandes de vérification démontrant que le cluster est opérationnel
- Architecture du cluster (schéma)

## TP 1: Auto Scaling avancé avec Karpenter et KEDA

### Objectif

Mettre en œuvre un système d'auto scaling à deux niveaux combinant provisionnement de nœuds intelligents et scaling basé sur les métriques applicatives.

### Contenu

- Configuration de Karpenter pour le provisionnement dynamique de nœuds
- Implémentation de provisioners Karpenter personnalisés avec contraintes
- Déploiement de KEDA et configuration de Scaled Objects
- Fournir une application (Node, Java, Rust…) visuelle pour expliciter le scaling
- Mise en place d'auto scaling basé sur diverses sources (files de messages, métriques Prometheus)
- Tests de charge pour observer le comportement combiné de Karpenter et KEDA
- Optimisation des coûts et performance

## TP 2: Kubernetes Gateway API et Service Mesh

### Objectif

Explorer la nouvelle Gateway API de Kubernetes et son intégration avec un service mesh.

### Contenu

- Mise en place d'une Gateway API OSS
- Configuration avancée de routing avec HTTPRoute, TCPRoute
- Implémentation de stratégies de trafic et sécurité via les Policy Attachments
- Demo d'une application maison pour expliciter le gain d'un service mesh
- Observation et debugging avec les outils du service mesh

## TP 3: Kubernetes Operator et GitOps avancé

### Objectif

Développer un opérateur Kubernetes personnalisé et l'intégrer dans un pipeline GitOps.

### Contenu

- Développement d'un opérateur Kubernetes avec Operator SDK (obligatoirement)
- Création de custom resources et controllers
- Implémentation de la réconciliation et des webhooks d'admission
- Intégration de l'opérateur dans un workflow GitOps avec Gitlab, Flux ou ArgoCD
- Mise en place d'un pipeline CI/CD complet pour l'opérateur
- Surveillance et observabilité des ressources personnalisées
