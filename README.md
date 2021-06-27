### Nats Test Service

#### Start application with db:

```shell script
docker run -p 27017:27017 mongo:4.2 

go run .

nats sub AlarmDigest

nats pub AlarmStatusChanged "{ \"UserID\": \"1\", \"AlarmID\": \"{{ID}}\", \"Status\": \"CRITICAL\", \"ChangedAt\": \"{{TimeStamp}}\"  }" --count=1
nats pub AlarmStatusChanged "{ \"UserID\": \"2\", \"AlarmID\": \"{{ID}}\", \"Status\": \"CRITICAL\", \"ChangedAt\": \"{{TimeStamp}}\"  }" --count=1
nats pub AlarmStatusChanged "{ \"UserID\": \"2\", \"AlarmID\": \"{{ID}}\", \"Status\": \"CLEARED\", \"ChangedAt\": \"{{TimeStamp}}\"  }" --count=1
nats pub AlarmStatusChanged "{ \"UserID\": \"3\", \"AlarmID\": \"{{ID}}\", \"Status\": \"CRITICAL\", \"ChangedAt\": \"{{TimeStamp}}\"  }" --count=1
nats pub SendAlarmDigest "{ \"UserID\": \"2\" }" --count=1
```

#### ToDo:
- Implement transactions in the DigestHandler
- Ensure graceful shutdown and draining of Subscribers
- 