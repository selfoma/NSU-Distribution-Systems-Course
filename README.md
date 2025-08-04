# CrackHash

## Description

CrackHash is a distributed system for cracking MD5 hashes using brute-force. The system consists of a manager and multiple workers communicating via HTTP and RabbitMQ.

### Main components:

* **Manager** receives hash cracking requests, distributes tasks among workers, aggregates results, and returns them to the user.
* **Workers** perform brute-force within a given range and send any found matches back to the manager.
* **RabbitMQ** ensures fault-tolerant communication between services.
* **(Replicated) MongoDB** stores request data to guarantee persistence.

## Technologies

* **Language**: Go
* **Containerization**: Docker, Docker Compose
* **Message Queue**: RabbitMQ
* **Database**: Replicated MongoDB

## Fault Tolerance

* Request data is stored in MongoDB.
* Tasks are sent through RabbitMQ with delivery acknowledgment.
* If one of the components fails, the system continues to operate.

## Run

```sh
docker compose up --build
```

## API

### Submit a hash cracking request

```http
POST /api/hash/crack
Content-Type: application/json
{
    "hash": "e2fc714c4727ee9395f324cd2e7f331f",
    "maxLength": 4
}
```

Response:

```json
{
    "requestId": "730a04e6-4de9-41f9-9d5b-53b88b17afac"
}
```

### Check status

```http
GET /api/hash/status?requestId=730a04e6-4de9-41f9-9d5b-53b88b17afac
```

Response (if still processing):

```json
{
    "status": "IN_PROGRESS",
    "data": null
}
```

Response (if finished):

```json
{
    "status": "READY",
    "data": ["abcd"]
}
```
