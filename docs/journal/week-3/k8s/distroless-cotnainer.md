# Distroless Cotnainer

## To test inside distroless container

```sh
kubectl run curl-test \
  --rm -it \
  --image=curlimages/curl \
  -n dbaas-api \
  -- sh
```