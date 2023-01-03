## User microservice

This repository is a generic user microservice written in Go. It defines and implement the
business logic to handle users within the GRPC protocol. The service comes with scripts to
easy running and testing of it. MongoDB has been chosen to implement the database layer while Kakfa has been
chosen for inter services messaging.

## Table of Contents
- [User microservice](#User-microservice)
  - [Table of Contents](#Table-of-Contents)
  - [Project Structure](#Project-Structure)
  - [Getting Started](#Getting-Started)
    - [Prerequisites](#Prerequisites)
    - [Api explanation](#Api-model)
    - [Install and Run](#Install-and-Run)
    - [Docker compose local deployment](#Docker-compose-local-deployment)
    - [Running the tests](#Running-the-tests)
        - [Unit tests](#Unit-tests)
        - [Integration test](#Integration-tests)
    - [Using Postman to import .proto](#Using-Postman-to-test-api-collection)
  - [Future developments](#Future-developments)
  - [Built With](#Built-With)
  - [Versioning](#Versioning)
  - [Contributing](#Contributing)
  - [Authors](#Authors)
  
## Project Structure
```javascript
user-service/
├── api/ // .proto file defining application GRPC interface, and Google protobuf dependencies    
├── client/ // integration tests
├── scripts/ // scripts
    ├── generate_api.sh // generate /service/api classes from .proto files in /api
    ├── generate_mocks.sh // generate /service/mocks classes using golang mockery
    ├── run_integration_test.sh // run integration test launching docker-compose instance
    ├── run_unitTest.sh // run all unit tests locally and generate code coverage report
├── service/ //golang service
    ├── api/    // autogenerated golang classes from ../api .proto files
    ├── config/ // parsing of environment variables to golang config
    ├── mocks/ // autogenerated mocks for unit tests. generated from generate_mocks.sh
    ├── model/ // contains model definition for the domain's objects
    ├── producer/ // contains the logic to push messages to 'users_topic' in Kafka 
    ├── repository/ // contains the logic to call mongdo-db
    ├── utility/ // contains the logic to create Hmac256 hash for passwords
    ├── service.go // service business logic and gRPC server implementation
    ├── service_test.go // service unit tests
├── .env // local environment variables
├── .gitignore
├── go.mod
├── main.go
├── Dockerfile
├── docker-compose.yml // Docker compose file to launch the service locally
├── README.md
```

## Getting Started 
### Prerequisites
* Go environment setup
   * Go - v1.18+
   * GOPATH set
* Docker local setup (Docker Desktop or bare metal Docker environment)
* BloomRPC or Postman (recommended) for calling gRPC api
* OPTIONAL: Protoc (command line to generate updated api) with Go plugins\
  You can download the protobuf and protoc-gen-go libraries with *brew* running the following commands:\
  `brew install protobuf`\
  `brew install protoc-gen-go`\
  Do not install the protoc-gen-go library with `go install` as it may cause errors when compiling the proto files
  (e.g. *--go_out: protoc-gen-go: plugins are not supported; use 'protoc --go-grpc_out=...' to generate gRPC*)
  * *Only required if you wish to change protobuf - by regenerating /service/api using /scripts/generate-api.sh*

### Api model

The apis implemented allow crud operations on `User` object, the business logic implementation can be found int `service/service.go` file. <br>
A user will be stored with the following properties:
```json
{
  "id":         string, #(uuid v4 model)
  "firstname":  string,
  "lastname":   string,
  "nickname":   string,
  "password":   string,
  "email":      string,
  "country":    string,
  "created_at": string,
  "updated_at": string
}
```
The `id` will be generated during the creation flow and will be immutable with uuid v4 format. 
The `created_at` and the `updated_at` dates are in the format RFC3339 (2022-01-02T15:04:05Z07:00).

The service exposes the crud api to create, update, delete and retrieve a list of users.<br>
The proto files for the `user_service` defines four main rpc apis:
- CreateUser, that is used to create a new user.
- GetUsers, that is used to retrieve a paginated list of users (filter can be applied) 
- UpdateUser, that is used to update a user
- DeleteUser, that is used to delete a user based on its id
- GetStatus, that is used to check if the service is up and running

#### Create User

The CreateUser api use the POST Http method and requires the following body:
```json
{
    "country": "IT", 
    "email": "user@email.com",
    "firstname": "Test",
    "lastname": "User",
    "nickname": "TestUserNickname",
    "password": "StrongUserPassword"
}
```
While most properties are free text string, the country is based on the following gRPC enum:
```
enum Country {
  UNKNOWN = 0;
  EN = 1;
  IT = 2;
  FR = 3;
  DE = 4;
}
```
When a user is created the `created_at` and `updated_at` properties are set with the same value.
A message with format `("Created user " + user.ID)` is sent to the users_topic to notify all topic subscribers.

Example of successful response:
```json
{
    "user": {
        "id": "aa68af4f-05e2-47e9-b318-0ecc9e28bd8c",
        "firstname": "Federico",
        "lastname": "Boarelli",
        "nickname": "RedBoa",
        "email": "fboarelli@email.com",
        "country": "IT",
        "created_at": "2022-07-21T10:56:08Z",
        "updated_at": "2022-07-21T10:56:08Z"
    }
}
```

#### Get Users

The GetUsers api use the Http GET method to retrieve a paginated list of users from the database. Filter on the country can be applied, in case
the filter is defined only user of the specific country will be return. Pagination has been implemented adding to parameters to the request,
`page` and `page_size` to allow iteration between pages. The pages value must be specified and cannot be null in the request.

An example of request is the following:
```json
{
  "page": "0",
  "page_size": "5",
  "filter_country": "EN"
}
```

Example of successful response:
```json
{
  "page": "0",
  "page_size": "5",
  "total_count": "3",
  "results": [
    {
      "id": "aa68af4f-05e2-47e9-b318-0ecc9e28bd8c",
      "firstname": "Federico",
      "lastname": "Boarelli",
      "nickname": "RedBoa",
      "email": "fboarelli@email.com",
      "country": "IT",
      "created_at": "2022-07-21T10:56:08Z",
      "updated_at": "2022-07-21T10:56:08Z"
    },
    {
      "id": "f1d25213-3d61-4a23-bb61-0a37a512716a",
      "firstname": "Mario",
      "lastname": "Verdi",
      "nickname": "GreenBoa",
      "email": "mverdi@email.com",
      "country": "IT",
      "created_at": "2022-07-21T10:58:00Z",
      "updated_at": "2022-07-21T10:58:00Z"
    },
    {
      "id": "f855b301-9807-46b7-a096-dfb6f376da77",
      "firstname": "Luigi",
      "lastname": "Gialli",
      "nickname": "YellowBoa",
      "email": "lgialli@email.com",
      "country": "IT",
      "created_at": "2022-07-21T10:58:12Z",
      "updated_at": "2022-07-21T10:58:12Z"
    }
  ]
}
```

More precise pagination's example:
```json
{
    "page": "2",
    "page_size": "1",
    "total_count": "1",
    "results": [
        {
            "id": "f855b301-9807-46b7-a096-dfb6f376da77",
            "firstname": "Luigi",
            "lastname": "Gialli",
            "nickname": "YellowBoa",
            "email": "lgialli@email.com",
            "country": "IT",
            "created_at": "2022-07-21T10:58:12Z",
            "updated_at": "2022-07-21T10:58:12Z"
        }
    ]
}
```

#### Update User

The UpdateUser api use the PUT method and requires a body similar to the CreateUser but it needs the `id` of the user. All editable properties 
are optional and will be updated only if set in the request's body. For example:
```json
{
  "id": "22b42028-0796-491b-971f-148198b67f1c",
  "firstname": "New Test",
  "lastname": "User",
  "nickname": "NewTestUserNickname",
  "email": "user-new@email",
  "password": "NewStrongUserPassword",
  "country": "EN"
}
```

All the fields in the body will be updated. `updated_at` property will be updated too in the User's document stored in Mongo DB.
A message with format `("Updated user " + user.ID)` is sent to the users_topic to notify all topic subscribers.

An empty response is returned if operation was successful, gRPC error will be return otherwise.

#### Delete User

The DeleteUser api use the Http DELETE method taking as input parameter the user id in the uuid v4 format. If id exists, the user will be
deleted, otherwise a `not_found` gRPC error will be returned.

A message with format `("Deleted user " + user.ID)` is sent to the users_topic to notify all topic subscribers.

An empty response is returned if operation was successful, gRPC error will be return otherwise.

### Install and Run

#### Docker compose local deployment

The most convenient way to run the service is to use the ``docker-compose.yml`` script
that can be found in the root folder of the project. <br>
Two main components are part of the script:
- User service
- Mongo DB service to store data
- Kafka and ZooKeeper to handle async messages

The ``docker-compose.yml`` will look for the `user-service-local` image, therefore before running it you'll need to
create a Docker image for the `user-service` with the following command:
```shell
docker build -t user-service-local .  
```

Once done you can run the service with the following command executed in the root folder:
```shell
docker-compose up
```
No need of specific setup for Mongo DB.

Load `api/v1/user_service.proto` into Postman or BloomRPC (don't forget to import `api/v1/` path as well to load .proto dependencies) and test calls

### Running the tests

For testing ``github.com/stretchr/testify`` is used assertions.

#### Unit tests
To run unit tests it has been created a specific shell script that is `scripts/run_unit_test.sh`. <br>
The script will run the tests and will also generate a testing report (only for files covered by unit tests). 

For unit tests it has been used ``github.com/vektra/mockery`` for generating mock objects.

If existing interfaces are updated or a new one is created it is needed to update or generate mocks
running the script `scripts/generate_mocks.sh`, the generated mocks will be put in `service/mocks/` folder.

#### Integration tests
To run integration tests it has been created a specific shell script that is `scripts/run_integration_test.sh`. <br>
The script will build a Docker image of the file, start the docker-compose environment and the run a full crud cycle that is 
defined in the `client/user_crud_flow_test.go` and `client/status_test.go` that we'll run actual grpc with a running instance of
the service replying.


The specific goal to check code stylings may be run with the command ``golint .``. Sonar checks code stylings in CI pipeline.

### Using Postman to test Api collection

To create a gRPC collection, you just need to create a new collection selecting "gRPC request", then on the "Import proto file" dropdown
you can select the "import .proto" file. Select the `user_service.proto` and after selecting the correct file remember to select the
"Add an import path" and select the `/api/v1/` path so .proto dependencies get resolved.

Then it is just about to select the gRPC that the service expose, generate a dummy body and then update it with relevant data.
Unfortunately Postman still does not allow to export gRPC collection, so you wont find it in this repo.

### Future developments

Improvement handling readness of Kafka service with health probes (for now just put manual sleep in Go service was the fastest workaround for DEV mode)

Improvement of messaging management to add fallbacks on data layer.

To be prod ready the service must implement probes, that could be done handling health and readness prob at deployment level and by integrating with
some monitoring tools for analytics like Prometheus, building on top of that a Grafana dashboard if advanced metrics is needed.

### Built With
- [Golang](https://golang.org/)
- [GRPC protocol](https://developers.google.com/protocol-buffers/docs/gotutorial)
- [Mongo DB](https://www.mongodb.com/)
- [Apache Kafka](https://kafka.apache.org/)

### Authors
- **Federico Boarelli** - *Initial work*