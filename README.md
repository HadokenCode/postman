# Postman

Postman is a HTTP to AMQP reverse proxy that combines the ease of
implementing an HTTP API with the benefits of async inter-service communication.

Here is a simple diagram of how Postman works


<img src="./assets/process1.png" align="left" alt="process" style="margin: 10px 0 20px 0" />


Most of the HTTP services use a reverse proxy (nginx, Apache, etc) already, you can
think of **postman** as the async equivalent to those reverse proxies.

REST over HTTP might not be the ideal protocol for internal microservice
communication, but it is certainly the easiest and fastest to implement.

AMQP or any other async protocol might be the best, but it is certainly
not as easy and straight forward to implement as regular HTTP. Plus, many
of our existing services use HTTP already.

**What are the benefits?**

- No load balancing needed.
- No service discovery needed.
- Forget about the complexities of implementing async calls.

**Here's how a typical `Service A -> Service B` request looks like using postman:**

```
+---+---------------+-------------+-------------+-------------+---+
    |               |             |             |             |
    +--------------->             |             |             |
    |1. HTTP request|             |             |             |
    |to service B   +------------->             |             |
    |               |2. Serialize |             |             |
    |               |& send async +------------->             |
    |               |             |3. Send to   |             |
    |               |             |a service B  +------------->
    |               |             |instance     |4. Send HTTP |
    |               |             |             |request to   |
    |               |             |             |service B    |
    |               |             |             |             |
    |               |             |             <-------------+
    |               |             <-------------+5. Send HTTP |
    |               <-------------+6. Serialize |response.    |
    <---------------+7. Send to   |& send async |             |
    |8. Send HTTP   |response     |             |             |
    |response.      |queue        |             |             |
    +               +             +             +             +

Service A        Service A     RabbitMQ      Service B     Service B
                 Postman                     Postman
```

In detail...

1. **Service A** sends a regular HTTP request to **Service B** via postman.
2. Postman gets the message and sends it over RabbitMQ.
3. RabbitMQ sends the message through to the next available **Service B** instance.
4. Postman on **Service B** gets the RabbitMQ message and creates a regular HTTP request to **Service B**
5. **Service B** processes the request and send a response back.
6. Postman on **Service B** sends the response through RabbitMQ.
7. Postman on **Service A** gets the response.
8. Postman on **Service A** send the response back to **Service A** as a regular HTTP response. 

# Installation

To install, just download the appropriate binary from the releases, extract and execute
directly or, preferrably, copy to one dir inside your `PATH`. No other dependencies required.

## Getting started

```bash
postman -config /path/to/config.toml
```

# Compile from source

```bash
go get github.com/rgamba/postman
cd $GOPATH/src/github.com/rgamba/postman
make
```



