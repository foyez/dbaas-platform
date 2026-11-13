# 8. Delivery Is GitOps, and CI Writes Image Tags Back to Git

**Why:** Git holds the full desired state, so a rollback is a `git revert`, and the cluster never needs inbound CI credentials.

**Trade-off:** CI needs write access to the repository. Child Argo apps are still synced manually, which slows down how quickly a change reaches the cluster.
