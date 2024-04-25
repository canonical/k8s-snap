name: Bug Report
description: File a bug report
labels: ["Type: Bug"]
body:
  - type: markdown
    attributes:
      value: >
        Thanks for taking the time to fill out this bug report! 
  - type: textarea
    id: bug-description
    attributes:
      label: Bug Description
      description: >
         Please explain the bug in a few short sentences.
    validations:
      required: true
  - type: textarea
    id: reproduction
    attributes:
      label: Reproduction steps 
      description: >
         Are you able to consistently reproduce the issue? Please add a list of steps that lead to the bug.
    validations:
      required: true
  - type: textarea
    id: environment
    attributes:
      label: System information
      description: >
        We need to know a bit more about the context in which you run the snap.
        Please provide an overview of your setup (e.g. number of nodes) and the output of:
         `snap version`
         `uname -a`
         `snap list k8s`
         `snap services k8s`
         `snap logs k8s -n 10000`
         `k8s status`
    validations:
      required: true
  - type: textarea
    id: fix
    attributes:
      label: Can you suggest a fix? 
      description: >
         This section is optional. How do you propose that the issue be fixed?
  - type: textarea
    id: contribution 
    attributes:
      label: yes/no, or @mention maintainers. Community contributions are welcome.
      description: >
         Are you interested in contributing a fix?