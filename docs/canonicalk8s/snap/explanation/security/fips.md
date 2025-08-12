# FIPS compliance

[FIPS 140-3] (Federal Information Processing Standard) is a U.S. government
standard for cryptographic modules. In order to comply with FIPS standards,
each cryptographic module must meet specific security requirements and must
undergo testing and validation by the U.S. National Institute of Standards
and Technology ([NIST]).

## Why is FIPS important for Kubernetes?

<!-- TODO: Update if we choose to build from a separate track -->
{{ product }} can be configured to use FIPS compliant cryptographic
modules from the host system, ensuring that users can meet the security
requirements set forth for the use in federal and other regulated
environments. All of our components including the built-in features
such as networking or load-balancer can be configured to use the
host systems FIPS compliant libraries instead of the non-compliant
internal go cryptographic modules. By choosing a Kubernetes distribution,
such as {{ product }}, organizations can
minimize their security risks by enabling FIPS and adhere to compliance
requirements. When building workloads on top of {{ product }},
it is essential that organizations build these in a FIPS
compliant manner to comply with the FIPS security requirements. In addition,
[FIPS 140-3] has additional requirements to the system and hardware that have
to be met in order to be fully FIPS compliant.

## Use FIPS with {{product}}

If you would like to enable FIPS in your Kubernetes cluster see our [how-to guide].

<!-- LINKS -->
[FIPS 140-3]: https://csrc.nist.gov/pubs/fips/140-3/final
[how-to guide]: /snap/howto/security/fips.md
[NIST]: https://www.nist.gov/
