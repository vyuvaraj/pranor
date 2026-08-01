# Error Code Reference

This document serves as the centralized reference for all standardized error codes across the **Pranor** platform.

Every HTTP API error response in Pranor uses a unified JSON structure provided by `pranor/core`:

```json
{
  "error": "Human readable message describing the error",
  "code": "ERR_STANDARD_CODE",
  "status": 400
}
```

---

## 1. Core & Authentication Errors

| Error Code | HTTP Status | Description & Cause | Recommended Solution |
|------------|-------------|---------------------|----------------------|
| `ERR_UNAUTHORIZED` | 401 | Missing or invalid authentication token. | Include a valid `Authorization: Bearer <token>` header. |
| `ERR_INVALID_TOKEN` | 401 | Expired, malformed, or untrusted JWT/API token. | Refresh credentials via Pranor Auth. |
| `ERR_FORBIDDEN` | 403 | Authenticated user lacks permission or scope. | Check RBAC role assignments or tenant headers (`X-Tenant-ID`). |
| `ERR_ACCESS_DENIED` | 403 | IP or WAF security rule blocked the request. | Check Gateway WAF logs or IP whitelist. |
| `ERR_RATE_LIMIT_EXCEEDED` | 429 | Request rate exceeded tier limits. | Back off and retry; check rate limiting headers. |

---

## 2. API Gateway & Ingress Errors (`Pranor Gate`)

| Error Code | HTTP Status | Description & Cause | Recommended Solution |
|------------|-------------|---------------------|----------------------|
| `ERR_BAD_GATEWAY` | 502 | Target backend service is down or unreachable. | Check upstream container health with `pranor doctor`. |
| `ERR_BAD_GATEWAY_TARGET` | 502 | Invalid target URL configured for route. | Verify upstream URI configuration in Pranor Gate. |
| `ERR_CIRCUIT_OPEN` | 503 | Target service circuit breaker opened due to failure spikes. | Allow service health recovery window before retrying. |
| `ERR_AI_WAF_BLOCKED` | 403 | Prompt injection or LLM guardrail violation detected. | Sanitize client input text payload. |

---

## 3. Storage & Database Errors (`Pranor Vault` & `Pranor Pool`)

| Error Code | HTTP Status | Description & Cause | Recommended Solution |
|------------|-------------|---------------------|----------------------|
| `ERR_NOT_FOUND` | 404 | Bucket, object, database record, or key does not exist. | Verify object key or entity ID. |
| `ERR_CONFLICT` | 409 | Entity or object with the same key already exists. | Use overwrite flag or unique identifier. |
| `ERR_SERVICE_UNAVAILABLE` | 503 | Database connection pool exhausted. | Adjust pool size in `Pranor Pool` configuration. |

---

## 4. Messaging & Workflow Errors (`Pranor Pulse` & `Pranor Flow`)

| Error Code | HTTP Status | Description & Cause | Recommended Solution |
|------------|-------------|---------------------|----------------------|
| `ERR_MISSING_DLQ_TOPIC` | 400 | DLQ action attempted on topic with no DLQ registered. | Register a Dead Letter Queue before requesting summary. |
| `ERR_WASM_TRANSFORM_FAILED` | 500 | WASM message transformation crashed or timed out. | Debug transform WASM code in sandbox. |
| `ERR_BACKPRESSURE_TIMEOUT` | 503 | Queue consumer channel buffer full. | Scale out consumer instances. |

---

## 5. System & Common Errors

| Error Code | HTTP Status | Description & Cause | Recommended Solution |
|------------|-------------|---------------------|----------------------|
| `ERR_BAD_REQUEST` | 400 | Request parameters missing or invalid type. | Verify query parameters against documentation. |
| `ERR_BAD_REQUEST_BODY` | 400 | Malformed JSON body payload. | Ensure JSON request body is valid. |
| `ERR_METHOD_NOT_ALLOWED` | 405 | HTTP verb (GET, POST, PUT, DELETE) not supported. | Check endpoint documentation for supported verbs. |
| `ERR_INTERNAL_SERVER_ERROR` | 500 | Unexpected server error encountered. | Inspect service logs with `pranor trace`. |
