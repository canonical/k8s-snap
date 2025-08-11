# FIPS compliance

[FIPS 140-3] (Federal Information Processing Standard) is a U.S. government
standard for cryptographic modules. In order to comply with FIPS standards,
each cryptographic module must meet specific security requirements and must
undergo testing and validation by the U.S. National Institute of Standards
and Technology ([NIST]).

## Why is FIPS important for Kubernetes?

<!-- TODO: Update if we choose to build from a separate track -->
{{ product }} is built with FIPS compliant cryptographic
modules, ensuring that users can meet the security requirements set forth
for the use in federal and other regulated environments. All of our components
including the built-in features such as networking or load-balancer are built
with FIPS compliant libraries. By choosing a compliant Kubernetes distribution,
such as {{ product }}, organizations can minimize their security
risks and adhere to compliance requirements. When building workloads on top of
{{ product }}, it is essential that organizations build these in a FIPS
compliant manner to comply with the FIPS security requirements.

## Use FIPS with {{product}}

If you would like to use FIPS in your cluster see our [how-to guide].

<!-- LINKS -->
[FIPS 140-3]: https://csrc.nist.gov/pubs/fips/140-3/final
[how-to guide]: /snap/howto/security/fips.md
[NIST]: https://www.nist.gov/
