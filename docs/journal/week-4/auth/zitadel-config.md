# Zitadel configuration

Create zitadel key secret for api:

```
kubectl create secret generic dbaas-zitadel-key \
  -n dbaas-api \
  --from-file=zitadel-key.json=./zitadel-api-application-key.json
```