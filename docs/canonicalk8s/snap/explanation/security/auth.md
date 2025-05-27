# Authentication and authorization in {{product}}

{{product}} comes with a well-defined set of authentication and authorization
mechanisms enabled by default. This guide provides an overview of these built-in
security features, helping users understand what to expect out of the box and
where to find further information for custom configuration.

## Overview

{{product}} leverages upstream Kubernetes security primitives for both
authentication (identity verification) and authorization (determining
permissions). By default, {{product}} includes:

- TLS certificates for secure communication between components
- Service accounts and service account tokens for workloads
- Optional integration paths for external identity providers
- Role-Based Access Control (RBAC)
- Admission controllers

## Authentication

Authentication in {{product}} ensures that users and components are who they
claim to be. The following methods are included by default:

### Client certificates

- The Kubernetes API server is configured to trust specific client certificates.
- Certificates are issued to admin users during cluster creation.
- These can be seen by running `k8s config` on a control plane node.

### Service accounts

- Every pod in Kubernetes is automatically assigned a service account, unless
  specified otherwise.
- Service account tokens are mounted into pods, enabling them to authenticate
  with the API server securely, unless specified otherwise.
- A policy engine (if added) may restrict auto-mounting the default service
  account token into Pods if used.
- These are managed in the namespace where the pod is deployed.

The Kubernetes API server can be configured to accept OpenID Connect (OIDC)
tokens for authentication from external identity providers.

In {{product}}, anonymous API access is disabled by default.

For more details, see [Kubernetes Authentication](
https://kubernetes.io/docs/reference/access-authn-authz/authentication/)
upstream docs.

## Authorization

After authentication, Kubernetes checks whether the user or service account is
authorized to perform a requested action. In {{product}}, this is done through
Role-Based Access Control.

### Role-Based Access Control (RBAC)

RBAC Authorization in {{product}} is done through the following types of
Kubernetes objects:

- **Roles** define a set of _allow_ permissions within a namespace.
- **ClusterRoles** define cluster-wide _allow_ permissions.
- **RoleBindings** and **ClusterRoleBindings** assign these roles to users or
  service accounts.

Example use cases:

- Granting users or service accounts read-only access to logs in a namespace.
- Granting read-only access to nodes, services, and pods to a monitoring
  operator.

Kubernetes defines a set of `ClusterRoles` and `ClusterRoleBindings` that apply
to all users and service accounts, including the `default` service accounts
mounted in pods:

- `system:basic-user`: Allows a user read-only access to basic information about
  themselves.
- `system:discovery`: Allows read-only access to API discovery endpoints needed
  to discover and negotiate an API level.
- `system:public-info-viewer`: Allows read-only access to non-sensitive
  information about the cluster.

For more details, see:

- [Kubernetes RBAC](https://kubernetes.io/docs/reference/access-authn-authz/rbac/)
- [Default roles and role bindings](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#default-roles-and-role-bindings)

## Admission controllers

After a request has been authenticated and authorized, admission controllers
are used to **validate** and / or **mutate** requests to the Kubernetes API that
create, modify, or delete resources in Kubernetes. Admission controllers cannot
block **get**, **list**, or **watch** requests.

Kubernetes has a default list of admission controllers, which can be expanded
or contracted. Keep in mind that several important features of Kubernetes
require admission controller to be enabled in order to properly support them.
As a result, a Kubernetes API server that is not properly configured with the
right set of admission controllers is an incomplete server and will not support
all the features you expect.

In addition to the default list of admission controllers, {{product}} enables
the `NodeRestriction` admission controller. This controller limits the `Node`
and `Pod` objects a kubelet can modify, allowing them to modify only their
own `Node` object and `Pod` objects bound to their node.

For more details, see [Kubernetes Admission Control](
https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/)
upstream docs.

## Extending authentication and authorization

For advanced authentication and authorization, users may want to configure:

- OpenID Connect (OIDC) for integrating with external identity providers
  (e.g.: LDAP, Google, Azure AD).
- Webhook token authentication for custom authentication.
- Custom RBAC roles.
- Pod Security Admission Controller, which can enforce the Pod Security
  Standards.

For further information on how to extend authentication and authorization in
your cluster, see:

- [Kubernetes authentication strategies](https://kubernetes.io/docs/reference/access-authn-authz/authentication/#authentication-strategies)
- [Webhook token authentication](https://kubernetes.io/docs/reference/access-authn-authz/authentication/#webhook-token-authentication)
- [Managing RBAC roles](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#user-facing-roles)
- [Pod security admission](https://kubernetes.io/docs/concepts/security/pod-security-admission/)
