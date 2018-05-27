# regsecret-operator
## Kubernetes imagePull secrets operator

`regsecret-operator` automates the creation of image pull secrets for one or more private registries in your namespaces.

It works watching namespaces events (optionally matching a selector) and creating the configured `kubernetes.io/dockerconfigjson` secrets for you.

### Quick start

Create a configuration file (ie. `config.json`):

``` json
{
  "secrets": [
    {
      "secretName": "regsecret",
      "credentials": {
        "https://index.docker.io/v1/": {
          "username": "my-username",
          "password": "my-password",
          "email": "my-email"
        }
      }
    }
  ]
}
```

Upload it as a secret in kubernetes:

```
kubectl -n kube-system create secret generic regsecret-operator-config --from-file=config=./config.json
```

Finally apply the deployment.yaml file contained in this repo:

```
kubectl apply -f https://raw.githubusercontent.com/mcasimir/regsecret-operator/master/deployment.yaml
```

### Configuration options

| Option                              | Type     | Description                                                                                          | Required | Default  |
|-------------------------------------|----------|------------------------------------------------------------------------------------------------------|----------|----------|
| logger.level                        | `string` | Minimum allowed level for log messages. One of: `"debug"`, `"info"`, `"warn"`, `"error"`, `"fatal"`. | false    | "info"   |
| logger.format                       | `string` | Log format. One of: `"pretty"`, `"json"`.                                                            | false    | "pretty" |
| secrets[].namespaceSelector         | `string` | A namespace label selector. ie. `foo==bar`. Leaving it empty will match any namespace.                                                          | false    |          |
| secrets[].secretName                | `string` | The name of the secret to be created.                                                                | true     |          |
| secrets[].credentials[uri]          | `string` | The url of the registry.                                                                             | true     |          |
| secrets[].credentials[uri].username | `string` | Username for authentication with the registry.                                                                           | true     |          |
| secrets[].credentials[uri].password | `string` | Password for authentication with the registry.                                                                           | true     |          |
| secrets[].credentials[uri].email    | `string` | Email for authentication with the registry.                                                                              | true     |          |
### Caveats

If you plan to use a `namespaceSelector` be aware that labeling a namespace with `kubectl label` will not trigger any event. In this case, the chosen selector may not match the namespace immediately but only after the next resync (which will eventually happen but not so immediately).

Adding/changing labels by editing the namespace resource directly (ie. with `kubectl edit` or `kubectl apply`) does not have the same issue.
