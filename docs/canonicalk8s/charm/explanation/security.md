```{include} /snap/explanation/security.md
:start-after: <!-- Start -->
:end-before: <!-- First charm end here -->
````

## Security of the charm

There are several security considerations that must be taken into account when
deploying any charm as outlined in the [Juju security documentation]. With
regards to the `k8s` and `k8s-worker` charms, there must be particular care
given to ensuring the principle of least privilege is observed and users only
have access to alter cluster resources they are entitled to. For more
information on creating users, assigning access levels and what access these
levels bestow, please check the following pages of Juju documentation:

- [Juju user types] - describes the different types of users supported by Juju
and their abilities.
- [Working with multiple users] - A how-to guide on sharing control of a cluster
with multiple Juju users.
- [Machine authentication] - describes how SSH keys are stored and used by Juju.

```{include} /snap/explanation/security.md
:start-after: <!-- First charm end here -->
````

<!-- LINKS -->
[Juju security documentation]:https://canonical-juju.readthedocs-hosted.com/en/latest/user/explanation/juju-security/
[Machine authentication]: https://canonical-juju.readthedocs-hosted.com/en/latest/user/reference/ssh-key/
[Working with multiple users]: https://canonical-juju.readthedocs-hosted.com/en/latest/user/howto/manage-users
[Juju user types]: https://canonical-juju.readthedocs-hosted.com/en/latest/user/reference/user/
[Snapcraft documentation]: https://snapcraft.io/docs/security-policies
