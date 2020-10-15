GitHub Top Contributors
======================

## Problem Analysis
 Our goal is to get the top 50/100/150 github contributors by location. 3 paths has been analyzed:
  - GraphQL API V4. Max rate limit is 5,000 points per hour, and, after digging some time I didn't find a way to order contributors results... in a graph database is really hard to answer our required query, too many nodes must be transversed... so this approach does not work for us, discarded.
  - REST API V3.(Search) Really easy to find the proper query, search API rate limit is 10 requests/minute on unauthorized users, 30 on authorized. Adding a cache layer here will reduce the total number of request to real github API
  - REST API V3.(users) Users rate limit is 5000 request / minute, looking around about how many users has github I've found 10 Millions approx... so at 300.000 req/hour we get short on pull the whole dataset considering Github V3 API data refreshed on 24h basis... so discarded.

## App structure
 App has been built using decoupled layers, following ports and adapters, considering microservices as just exposed services, that get plugged to transport.

Main app layers:
 - Service layer: Bussines logic
 - Endpoints: command handlers exposed by service layer
 - Middlewares: wrap endpoints building chains
 - Encoders
 - Server Handlers: transport layer

 App Core based on Service layer, all incoming server are being considered commands handled by services.  Those service handlers are the base of exposed endpoints, enabling middleware design, so instrumentation, metrics, rate-limits... and many more components can be plugged easy in a decoupled way.

 Endpoints and middlewares are finally wrapped by encoders and plugged to the transport layer. In our case, json is used as encoders and are plugged as http.Handler in our http Server.

 ![alt text](https://raw.githubusercontent.com/mycodesmells/gokit-example/master/res/onion.png "App architecture")

 Service errors translated properly to transport error codes, in our case, as http.StatusCode embedding errors as json payloads.

### Design Trade-offs
As our challenge goal could be achieved with a regular http server, github client and a cache layer, the approach could be seen as little over-complex, but, with that in mind, endpoints and middleware architecture enables us to replace all project components where required, example: expose our service using grpc it's just a layer to wrap service layer... to add persistent connections, just need to add new socket handlers... and so on.

Apart from that, go-kit components give us really valuable services that can be added as just another middleware layers: request tracing, abuse control...

### Service Layer
 In that scheme, our service exposes GetTopContributors command as an endpoint, wrapped by Instrumentation & Logging middlewares, and with an auth layer that intercepts all non authorized requests.

 Same model has been applied to outside services, in our case, we pick up google/go-github client, that client becomes an endpoint too, wrapped by some middlewares, as:
 - Instrumentation
 - Logging
 - RateLimit: mirrors github api rate limit, so once achieved max rate, rate limit middleware disconnect new requests to the endpoint until release policy is done.

 From the service point of view, we just have a repository, in charge of loading results, this is implemented as an HttpRepository that consumes github client endpoint.

### Github Client
Github client includes now timeout and retry policy, retrying on different failure scenarios, including (202 responses). Rate limit middleware is just a mirror from real github api behaviour, wrapping original endpoint offering a limit where requests must be stopped up to free rate-limit max request.

### Cache Layer
HttpRepository is wrapped by a cache layer, so each request becomes a real request to github api if we have a cache miss.
Cache has been implemented in top of an LRU structure, adding a worker in charge of entry expiration.

### Auth Layer
 An authentication service has been built, using JWT and cookies (i don't like cookies too, but have been great for testing :) ). As explained, auth layer wraps service layer, so that, credentials are required to access final services, those credentials (user / pass) are validated using right now a static validator, but it's decoupled, so can be easy replaced.

 To allow jwt real scheme, me need a seperated authentication service... on the challenge has been implemented as another endpoint too.
 For testing purposes, **token expiration fixed to just one minute**, it's quite easy to get expired during tests.

 Required keys provided in config folder, keys have been generated as:
  ```
  openssl genrsa -out app.rsa 1024
  openssl rsa -in app.rsa -pubout > app.rsa.pub
  ```

### Endpoints
 - /top-contributors/v1: unprotected access
 - /auth/top-contributors/v1: auth restricted
 - /auth: authenticate credentials
 - /metrics: expose prometheus metrics (GithubTop namespace key)

## Docker
 App published in docker registry:
   https://hub.docker.com/r/marcosquesada/github-top/

## App launch
 From docker:
 ```
make docker-run
```

 On Cli, from root project:
```
go run main.go http --oauth  XXXXXXXXXXXXXX (Github Personal Token)
```

## Use cases

#### TopContributors Unprotected access enabled calling:
```
curl -X GET "http://localhost:8000/top-contributors/v1?city=barcelona&size=50"

{"Top":[{"ID":125005,"Name":"kristianmandrup","Url":"https://api.github.com/users/kristianmandrup"......

```
#### TopContributors authenticated endpoint without credentials:
 ```
 curl -X GET http://localhost:8000/auth/top-contributors/v1?city=barcelona&size=100
{"error":"Forbiden access"}
```
To guarantee request credentials, we need to request it to auth endpoint:
 ```
 curl -d '{"user":"test", "pass":"known"}' -H "Content-Type: application/json" -X POST http://localhost:8000/auth --cookie-jar ./cookies.text
{"Token":"eyJhbGciO.....
```

After that, we pass our cookie on each request:
```
curl -v --cookie ./cookies.text --cookie-jar ./cookies.text "http://localhost:8000/auth/top-contributors/v1?size=100&city=barcelona"
{"Top":[{"ID":125005.....
```

### TopContributors V2 Unprotected access:
 Included user details on contributor response (email, company and bio), if they are available in github api response (take in mind that those values can be private)
```
curl -X GET "http://localhost:8000/top-contributors/v2?city=barcelona&size=50"
```

### TopSearchedLocations
 Each top contributors request is tracked in a location ranking (inMemory / Redis)
```
curl -X GET "http://localhost:8000/top-searched-locations/v1?size=50"
{"Top":[{"name":"barcelona","score":4,"index":0},{"name":"madrid","score":1,"index":1},{"name":"london","score":1,"index":2}]}
```

### Cache implementation details
 Two available implementations:
 - InMemory: LRU based with expiration worker
 - Redis: Using key/value entries with expiration (HashMaps were considered, but, they don't offer single key expiration)
 
 By default, InMemory is available. Using --redis flag specifies a redis host to enable redis cache, replacing inMemory one.
 
### Ranking implementation details
  Two available implementations:
  - InMemory: Priority queue based in top of heap (heap.Interface), offering high performance (volatile data will not survive application restarts). Priority Queue has a maximum size, once achieved old entries purged from bottom.
  - Redis: Implemented in top of regular Sorted Sets, no bounded size
  
By default, InMemory is available. Using --redis-ranking flag specifies a redis host to enable redis ranking, replacing inMemory one.

### Check Metrics on:
```
http://localhost:8000/metrics
```
## Run Tests
```
make test
```
