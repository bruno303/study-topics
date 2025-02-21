# K8s-Test-Java

Simple application to demonstrate how broadcast calls for another pods inside kubernetes cluster.

## Curls

### Broadcast change loglevel:
```shell
curl -X POST http://localhost:8080/broadcast \
    --header "Content-type: application/json" \
    --header "Authorization: Bearer XPTO" \
    --data '{"urls":["http://localhost:8080/actuator/loggers/com.bso.k8stest"],"method":"POST","addHeaders":{"test":"value","Content-type":"application/json"},"bypassHeaders":["authorization"],"body":"{\"configuredLevel\":\"INFO\"}"}'
```

### Broadcast get loglevel (all responses will be logged in the pod that made the requests):
```shell
curl -X POST http://localhost:8080/broadcast \
    --header "Content-type: application/json" \
    --header "Authorization: Bearer XPTO" \
    --data '{"urls":["http://localhost:8080/actuator/loggers/com.bso.k8stest"],"method":"GET"}'
```

### Get loglevel:
```shell
curl -X POST http://localhost:8080/actuator/loggers/com.bso.k8stest \
    --header 'content-type: application/json' \
    --data '{"configuredLevel": "INFO"}'
```

