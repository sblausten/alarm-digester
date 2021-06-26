### Nats Test Service

#### Start application with db:

```shell script
docker run -p 27017:27017 mongo:4.4 

go run
```

ToDo:
- Implement transactions in the DigestHandler
- Ensure graceful shutdown and draining of Subscribers