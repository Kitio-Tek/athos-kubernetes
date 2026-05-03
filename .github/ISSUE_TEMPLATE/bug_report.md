---
name: Bug report
about: Report something that is not working as expected
title: "[bug] "
labels: ["bug", "triage"]
---

## Description

<!-- A clear and concise description of the bug. -->

## Steps to reproduce

1.
2.
3.

## Expected behaviour

<!-- What did you expect to happen? -->

## Actual behaviour

<!-- What actually happened? Include controller logs and `kubectl describe`
output where relevant. -->

```text
# operator logs
```

```text
# kubectl describe pgc <name> -n <namespace>
```

## Environment

- Athos version (`kubectl get deploy athos-... -o jsonpath='{.spec.template.spec.containers[0].image}'`):
- Kubernetes version (`kubectl version`):
- Helm version (`helm version --short`):
- Cluster type (kind / EKS / GKE / AKS / on-prem):
- PostgreSQL major version requested:

## Additional context

<!-- Workarounds you have tried, related issues, etc. -->
