# Ratelimit

**ratelimit** is a burstable rate-limiter microservice written in Go. It uses gRPC for receiving commands, as per the protocol definition in `pkg/proto/ratelimit.proto`.

The **ratelimit** implementation is heavily inspired by the [smyte ratelimiter](https://medium.com/smyte/rate-limiter-df3408325846) and implements the [token bucket](https://en.wikipedia.org/wiki/Token_bucket) algorithm. However, unlike the smyte limiter, it uses gRPC for communication and badger-db for persistence (smyte uses rocksdb). Roc ksDB is a great project, but I wanted this one to be pure Go.

There is a single command, `reduce`, which takes a key and various refill parameters (see below). A single tokens is decreases for each request but also refilled when appropriate. The `reduce` command returns a result specifying whether the request should be allowed and the number of remaining tokens.


## Usage

```
$ PORT=8080 DATABASE_DIR=db make run
```

As you can see in the .proto file (pkg/proto/ratelimit.go), the following request fields are used:

```
message ReduceRequest {
  string key = 1;
  uint32 maxAmount = 2;
  uint32 refillAmount = 3;
  uint32 refillDurationSec = 4;
}
```

The ratelimiter creates a token bucket the first time it receives a specific `key` and it initially has `maxAmount` tokens. Each time you call the `ReduceRequest` method, 1 token is decreased. Every `refillDurationSec` seconds, `refillAmount` tokens are added back to the bucket. If `refillAmount` is not specified, it defaults to `maxAmount`. See the command-line usage below for a full example.


### Command-line Usage (for testing purposes)

As an example, let's say we have a service where we want to limit the number of sms a user can send per time interval. In the following example, the user id=321, so let's make the key = 'sms:321'. Let's set the max amount of tokens to 3 and refill them at a rate of 1 every minute. So this is burstable, allowing 3 to be sent immediately and then the user must wait another minute to send another sms.

1) Run **ratelimit** service with reflection on

```
$ REFLECTION=true make run
```

2) Install grpc_cli

3) List Services

```
$ grpc_cli ls localhost:8080
pkg.proto.RateLimitService
```

4) List Methods

```
$ grpc_cli ls localhost:8080 pkg.proto.RateLimitService -l
filename: pkg/proto/ratelimit.proto
package: pkg.proto;
service RateLimitService {
  rpc Reduce(pkg.proto.ReduceRequest) returns (pkg.proto.ReduceResponse) {}
}
```

5) Call ratelimit reduce method

```
$ grpc_cli call 127.0.0.1:8080 Reduce "key: 'sms:321', maxAmount: 3, refillAmount: 1 refillDurationSec: 60"
status: OK
remaining: 2

$ grpc_cli call 127.0.0.1:8080 Reduce "key: 'sms:321', maxAmount: 3, refillAmount: 1 refillDurationSec: 60"
status: OK
remaining: 1

$ grpc_cli call 127.0.0.1:8080 Reduce "key: 'sms:321', maxAmount: 3, refillAmount: 1 refillDurationSec: 60"
status: OK
(remaining: 0)
```

3 reduce operations were performed so at this point 0 tokens are left, so let's see what happens when we try to reduce the tokens again:

```
$ grpc_cli call 127.0.0.1:8080 Reduce "key: 'sms:321', maxAmount: 3, refillAmount: 1 refillDurationSec: 60"
status: NG
(remaining: 0)
```

No more tokens and the status=NG is returned meaning we should not allow the request. If we were to wait another minute (`refillDurationsec`=60), the token could will have increased from 0 to 1 (note that the `maxAmount`=3, so even after 3+ minutes, it will be 3).


### Programmatic Usage

```
conn, err := grpc.Dial("localhost:8080", grpc.WithInsecure()) // assume behind firewall service-to-service
if err != nil {
	fmt.Printf("error establishing connection: %s\n", err)
	os.Exit(-1)
}

defer conn.Close()
client := proto.NewRateLimitServiceClient(conn)
req := &proto.ReduceRequest{
	Key:               "sms:543",
	MaxAmount:         3,
	RefillAmount:      1,
	RefillDurationSec: 60,
}
resp, err := client.Reduce(context.Background(), req)
if err != nil {
	fmt.Printf("error executing reduce command %s\n", err)
	os.Exit(-1)
}

fmt.Println("got resp", resp)
```


Note: not yet recommended for production use. Has only been in development for one day.
