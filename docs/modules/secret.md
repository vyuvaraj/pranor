# Pranor Secret — Secret & Credential Management

`Pranor Secret` is the centralized secrets, credentials, and configuration protection engine for the **Pranor** ecosystem. It provides tenant-isolated secret storage encrypted at rest using AES-GCM (Galois/Counter Mode).

## Features

- **Centralized Encrypted Storage**: Encrypts all stored secrets using a 32-byte master key.
- **Tenant Isolation**: Organizes secrets dynamically per tenant context.
- **Microservice Ready**: Plugs directly into `Pranor Core` middleware for authentication, tracing, and rate limiting.
- **Graceful Shutdown**: Stops safely without corrupting the encrypted local storage file.

---

## Getting Started

### Local Development

1. **Provide a Master Key**: Define the 32-byte master key as a hex-encoded string in the environment:
   ```bash
   # Example hex key (32 bytes)
   export PRANOR_SECRET_MASTER_KEY="000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f"
   ```
   *Note: If no master key is supplied, a temporary random key will be generated at startup, and stored secrets will not persist across restarts.*

2. **Run the Service**:
   ```bash
   go run main.go --port 8091 --file secrets.enc
   ```

---

## API Documentation

All endpoints support standard header authentication and `X-Tenant-ID` routing (integrated with `Pranor Core`).

### 1. Set or Update a Secret
* **Endpoint**: `POST /api/v1/secrets`
* **Headers**:
  * `X-Tenant-ID`: tenant-a
  * `Authorization`: Bearer `<token>`
* **Request Body**:
  ```json
  {
    "key": "database-password",
    "value": "super-secret-passphrase"
  }
  ```
* **Response (201 Created)**:
  ```json
  {
    "key": "database-password",
    "value": "super-secret-passphrase"
  }
  ```

### 2. Get a Secret
* **Endpoint**: `GET /api/v1/secrets/{key}`
* **Response (200 OK)**:
  ```json
  {
    "key": "database-password",
    "value": "super-secret-passphrase"
  }
  ```

### 3. List Stored Secret Keys
* **Endpoint**: `GET /api/v1/secrets`
* **Response (200 OK)**:
  ```json
  {
    "keys": ["database-password"]
  }
  ```

### 4. Delete a Secret
* **Endpoint**: `DELETE /api/v1/secrets/{key}`
* **Response (200 OK)**:
  ```json
  {
    "status": "deleted",
    "key": "database-password"
  }
  ```

---

## License

This project is licensed under Apache 2.0 - see the [LICENSE](https://github.com/vyuvaraj/pranor/blob/main/LICENSE) file for details.
