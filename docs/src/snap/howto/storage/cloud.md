# How to use cloud storage

{{product}} simplifies the process of integrating and managing cloud storage
solutions like Amazon EBS. This guide provides steps to configure IAM policies,
deploy the cloud controller manager, and set up the necessary drivers for you
to take advantage of cloud storage solutions in the context of Kubernetes.

## What you'll need

This guide assumes the following:

- You have root or sudo access to an Amazon EC2 instance
- You can create roles and policies in AWS


## Set IAM Policies

Your instance will need a few IAM policies to be able to communciate with the
AWS APIs. The policies provided here are quite open and should be scoped down
based on your security requirements.

You will most likely want to create a role for your instance. You can call this
role "k8s-control-plane" or "k8s-worker". Then, define and attach the following
policies to the role. Once the role is created with the required policies,
attach the role to the instance.

For a control plane node:

```{dropdown} Control Plane Policies
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "autoscaling:DescribeAutoScalingGroups",
        "autoscaling:DescribeLaunchConfigurations",
        "autoscaling:DescribeTags",
        "ec2:DescribeInstances",
        "ec2:DescribeRegions",
        "ec2:DescribeRouteTables",
        "ec2:DescribeSecurityGroups",
        "ec2:DescribeSubnets",
        "ec2:DescribeVolumes",
        "ec2:DescribeAvailabilityZones",
        "ec2:CreateSecurityGroup",
        "ec2:CreateTags",
        "ec2:CreateVolume",
        "ec2:ModifyInstanceAttribute",
        "ec2:ModifyVolume",
        "ec2:AttachVolume",
        "ec2:AuthorizeSecurityGroupIngress",
        "ec2:CreateRoute",
        "ec2:DeleteRoute",
        "ec2:DeleteSecurityGroup",
        "ec2:DeleteVolume",
        "ec2:DetachVolume",
        "ec2:RevokeSecurityGroupIngress",
        "ec2:DescribeVpcs",
        "ec2:DescribeInstanceTopology",
        "elasticloadbalancing:AddTags",
        "elasticloadbalancing:AttachLoadBalancerToSubnets",
        "elasticloadbalancing:ApplySecurityGroupsToLoadBalancer",
        "elasticloadbalancing:CreateLoadBalancer",
        "elasticloadbalancing:CreateLoadBalancerPolicy",
        "elasticloadbalancing:CreateLoadBalancerListeners",
        "elasticloadbalancing:ConfigureHealthCheck",
        "elasticloadbalancing:DeleteLoadBalancer",
        "elasticloadbalancing:DeleteLoadBalancerListeners",
        "elasticloadbalancing:DescribeLoadBalancers",
        "elasticloadbalancing:DescribeLoadBalancerAttributes",
        "elasticloadbalancing:DetachLoadBalancerFromSubnets",
        "elasticloadbalancing:DeregisterInstancesFromLoadBalancer",
        "elasticloadbalancing:ModifyLoadBalancerAttributes",
        "elasticloadbalancing:RegisterInstancesWithLoadBalancer",
        "elasticloadbalancing:SetLoadBalancerPoliciesForBackendServer",
        "elasticloadbalancing:AddTags",
        "elasticloadbalancing:CreateListener",
        "elasticloadbalancing:CreateTargetGroup",
        "elasticloadbalancing:DeleteListener",
        "elasticloadbalancing:DeleteTargetGroup",
        "elasticloadbalancing:DescribeListeners",
        "elasticloadbalancing:DescribeLoadBalancerPolicies",
        "elasticloadbalancing:DescribeTargetGroups",
        "elasticloadbalancing:DescribeTargetHealth",
        "elasticloadbalancing:ModifyListener",
        "elasticloadbalancing:ModifyTargetGroup",
        "elasticloadbalancing:RegisterTargets",
        "elasticloadbalancing:DeregisterTargets",
        "elasticloadbalancing:SetLoadBalancerPoliciesOfListener",
        "iam:CreateServiceLinkedRole",
        "kms:DescribeKey"
      ],
      "Resource": [
        "*"
      ]
    }
  ]
}
```

For a worker node:

```{dropdown} Worker Policies
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ec2:DescribeInstances",
        "ec2:DescribeRegions",
        "ecr:GetAuthorizationToken",
        "ecr:BatchCheckLayerAvailability",
        "ecr:GetDownloadUrlForLayer",
        "ecr:GetRepositoryPolicy",
        "ecr:DescribeRepositories",
        "ecr:ListImages",
        "ecr:BatchGetImage"
      ],
      "Resource": "*"
    }
  ]
}
```


## Set your host name

The cloud controller manager uses the node name to correctly associate the node
with an EC2 instance. In Canonical K8s, the node name is derived from the
hostname of the machine. Therefore, before bootstrapping the cluster, we must
first set an appropriate host name.

```bash
echo "$(sudo cloud-init query ds.meta_data.local-hostname)" | sudo tee /etc/hostname
```

Then, reboot the machine.

When the machine is up, use `hostname -f` to check the host name. It should
look like:

```bash
ip-172-31-11-86.us-east-2.compute.internal
```

This host name format is called IP-based naming and is specific to AWS.

```bash
{note} Don't rely on the PS1 prompt to know if your host name was changed successfully. The PS1 prompt only displays the hostname up to the first `.`.
```


## Bootstrap Canonical K8s

Now that your machine has an appropriate host name, you are ready to bootstrap
Canonical K8s.

First, create a bootstrap configuration file that sets the cloud-provider
configuration to "external".

```bash
echo "cluster-config:
  cloud-provider: external" > bootstrap-config.yaml
```

Then, bootstrap the cluster:

```bash
sudo k8s bootstrap --file ./bootstrap-config.yaml
sudo k8s status --wait-ready
```

## Deploy the cloud controller manager

Now that you have an appropriate host name, policies, and a Canonical K8s
cluster, you have everything you need to deploy the cloud controller manager.

Here is a YAML definition file that sets appropriate defaults for you, it
configures the necessary service accounts, roles, and daemonsets:

```{dropdown} CCM deployment manifest
---
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: aws-cloud-controller-manager
  namespace: kube-system
  labels:
    k8s-app: aws-cloud-controller-manager
spec:
  selector:
    matchLabels:
      k8s-app: aws-cloud-controller-manager
  updateStrategy:
    type: RollingUpdate
  template:
    metadata:
      labels:
        k8s-app: aws-cloud-controller-manager
    spec:
      nodeSelector:
        node-role.kubernetes.io/control-plane: ""
      tolerations:
        - key: node.cloudprovider.kubernetes.io/uninitialized
          value: "true"
          effect: NoSchedule
        - effect: NoSchedule
          key: node-role.kubernetes.io/control-plane
      affinity:
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
              - matchExpressions:
                  - key: node-role.kubernetes.io/control-plane
                    operator: Exists
      serviceAccountName: cloud-controller-manager
      containers:
        - name: aws-cloud-controller-manager
          image: registry.k8s.io/provider-aws/cloud-controller-manager:v1.28.3
          args:
            - --v=2
            - --cloud-provider=aws
            - --use-service-account-credentials=true
            - --configure-cloud-routes=false
          resources:
            requests:
              cpu: 200m
      hostNetwork: true
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: cloud-controller-manager
  namespace: kube-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: cloud-controller-manager:apiserver-authentication-reader
  namespace: kube-system
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: extension-apiserver-authentication-reader
subjects:
  - apiGroup: ""
    kind: ServiceAccount
    name: cloud-controller-manager
    namespace: kube-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: system:cloud-controller-manager
rules:
- apiGroups:
  - ""
  resources:
  - events
  verbs:
  - create
  - patch
  - update
- apiGroups:
  - ""
  resources:
  - nodes
  verbs:
  - '*'
- apiGroups:
  - ""
  resources:
  - nodes/status
  verbs:
  - patch
- apiGroups:
  - ""
  resources:
  - services
  verbs:
  - list
  - patch
  - update
  - watch
- apiGroups:
  - ""
  resources:
  - services/status
  verbs:
  - list
  - patch
  - update
  - watch
- apiGroups:
  - ""
  resources:
  - serviceaccounts
  verbs:
  - create
  - get
  - list
  - watch
- apiGroups:
  - ""
  resources:
  - persistentvolumes
  verbs:
  - get
  - list
  - update
  - watch
- apiGroups:
  - ""
  resources:
  - endpoints
  verbs:
  - create
  - get
  - list
  - watch
  - update
- apiGroups:
  - coordination.k8s.io
  resources:
  - leases
  verbs:
  - create
  - get
  - list
  - watch
  - update
- apiGroups:
  - ""
  resources:
  - serviceaccounts/token
  verbs:
  - create
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: system:cloud-controller-manager
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: system:cloud-controller-manager
subjects:
  - apiGroup: ""
    kind: ServiceAccount
    name: cloud-controller-manager
    namespace: kube-system
```

After a moment, you should see the cloud controller manager pod was
successfully deployed.

```bash
NAME                                 READY   STATUS    RESTARTS        AGE
aws-cloud-controller-manager-ndbtq   1/1     Running   1 (3h51m ago)   9h
```

## Deploy the EBS CSI Driver

Now that the cloud controller manager is deployed and can communicate with AWS,
you are ready to deploy the EBS CSI driver. The easiest way to deploy the
driver is with the Helm chart. Luckily, Canonical K8s has a built-in helm
command.

If you want to create encrypted drives, you need to add the statement to the
policy you are using for the instance.

```json
{
  "Effect": "Allow",
  "Action": [
      "kms:Decrypt",
      "kms:GenerateDataKeyWithoutPlaintext",
      "kms:CreateGrant"
  ],
  "Resource": "*"
}
```

Then, add the helm repo for the EBS CSI Driver.

```bash
sudo k8s helm repo add aws-ebs-csi-driver https://kubernetes-sigs.github.io/aws-ebs-csi-driver
sudo k8s helm repo update
```

Finally, install the Helm chart, making sure to set the correct region as an
argument.

```bash
sudo k8s helm upgrade --install aws-ebs-csi-driver \
    --namespace kube-system \
    aws-ebs-csi-driver/aws-ebs-csi-driver \
    --set controller.region=us-east-2
```

Once the command completes, you can verify the pods are successfully deployed:

```bash
kubectl get pods -n kube-system -l app.kubernetes.io/name=aws-ebs-csi-driver
```

```bash
NAME                                  READY   STATUS    RESTARTS        AGE
ebs-csi-controller-78bcd46cf8-5zk8q   5/5     Running   2 (3h48m ago)   8h
ebs-csi-controller-78bcd46cf8-g7l5h   5/5     Running   1 (3h48m ago)   8h
ebs-csi-node-nx6rg                    3/3     Running   0               9h
```

The status of all pods should be "Running".

## Deploy a workload

Everything is in place for you to deploy a workload that dynamically creates
and uses an EBS volume.

First, create a StorageClass and a PersistentVolumeClaim:

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: ebs-sc
provisioner: ebs.csi.aws.com
volumeBindingMode: WaitForFirstConsumer
parameters:
  type: gp3 # EBS volume type (gp3, gp2, etc.)
  fsType: ext4 # File system type
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: ebs-pvc
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
  storageClassName: ebs-sc
```

Then, you can deploy a pod that uses a volume. Because we used
`WaitForFirstConsumer`, you'll only see the volume in AWS once the pod is
deployed.

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: app-using-ebs
spec:
  containers:
  - name: app
    image: nginx
    volumeMounts:
    - mountPath: "/data"
      name: ebs-volume
  volumes:
  - name: ebs-volume
    persistentVolumeClaim:
      claimName: ebs-pvc
```

Congratulations! By following this guide, you've set up cloud storage
integration for your Kubernetes cluster. When you go to the `Elastic Block
Store > Volumes` page in AWS, you should see a 10Gi gp3 volume.


<!-- LINKS -->
[getting-started-guide]: /snap/tutorial/getting-started.md
