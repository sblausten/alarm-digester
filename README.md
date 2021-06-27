## Nats Test Service

### Approach:

This is my first Golang service so I may well be breaking some common patterns and mis-using certain things. Please see
prioritised ToDo's at the bottom of this readme for things I'd have liked to do if I'd had more time on this project.

##### Nats
I used vanilla Nats rather than Nats streaming or Jetstream as it was not indicated/did not seem like the upstream 
producers and test server were using these. I would normally have sought clarification on this. 

Jetstream would have been my preferred mechanism and would have enabled a different approach. 

Likewise I did not implement Replies to guarantee delivery for my producer or subscriber as it was not indicated that 
the upstream producers would be using the reply-to field. I'm not sure if this was implied from the assertion that the 
broker could guarantee at-least-once delivery, so I would have asked for further clarification on this normally. 

I used a Nats subscriber queue group so that I did not need to implement de-duplication logic using the database.

##### **State management**
I used MongoDb for state, so that I could fetch historic alarms and keep track of the last digest request in order to 
only send new alarms downstream. 

I have not assumed that we want a remote database connection for simplicity as this was not specified. 

##### **Testing**
The application can be run for end to end testing using a docker-compose script but this is not intended for production.

I used gomock for more complex interfaces and hand wrote mocks for simpler ones in my unit tests. 

I did not try to test the daos as the go mongodriver does not use interfacs which meant I would have had to build them 
out myself. Due to time constraints I skipped this in favour of integration tests that exercised the dao's happy paths. 

I didn't get to test the Nats Subscriber and Publisher also due to time constraints, but I would have liked to test 
specifically the drain implementation on cancellation. 

### To Run:

Run dockerised application with a db and nats server:
```shell script
docker-compose -d up
```

### Automated Testing:

Run unit tests:
```shell script
go test
```



### Manual Testing:
Using nats-cli - https://github.com/nats-io/natscli

Subscribe to output:
```shell script
nats sub AlarmDigest
```

Send some test messages to the service in another terminal:
```shell script
nats pub AlarmStatusChanged "{ \"UserID\": \"1\", \"AlarmID\": \"{{ID}}\", \"Status\": \"CRITICAL\", \"ChangedAt\": \"{{TimeStamp}}\"  }"
nats pub AlarmStatusChanged "{ \"UserID\": \"2\", \"AlarmID\": \"{{ID}}\", \"Status\": \"CRITICAL\", \"ChangedAt\": \"{{TimeStamp}}\"  }"
nats pub AlarmStatusChanged "{ \"UserID\": \"2\", \"AlarmID\": \"{{ID}}\", \"Status\": \"CLEARED\", \"ChangedAt\": \"{{TimeStamp}}\"  }"
nats pub AlarmStatusChanged "{ \"UserID\": \"3\", \"AlarmID\": \"{{ID}}\", \"Status\": \"CRITICAL\", \"ChangedAt\": \"{{TimeStamp}}\"  }"
nats pub SendAlarmDigest "{ \"UserID\": \"2\" }"
```

### To Do:
1. Implement e2e tests using Go or a bash script against dockerised application
2. Improve unit test coverage
3. Implement retries with a backoff for the Publisher 
4. Implement replies or use Jetstream for producer to guarantee delivery
5. Implement replies for subscribers or switch to Jetstream / Streaming and acks.
6. Implement MongoDb transactions in SendAlarmDigestHandler to ensure graceful shutdown and processing of inflight messages
7. Remove use of log.Fatal and ensure cleanup at top level - https://dave.cheney.net/2015/11/05/lets-talk-about-logging 
