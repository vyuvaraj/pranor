# Complete Pranor Ecosystem Error Code Reference

This document provides an exhaustive, centralized index of all **140+ unique error codes** used across Pranor microservices and language tools.

Every HTTP error response follows the standard format:

```json
{
  "error": "Detailed description",
  "code": "ERR_EXAMPLE_CODE",
  "status": 400
}
```

## Pranor Auth (9 Error Codes)

| Error Code | Category | Description / Typical Trigger |
|------------|----------|-------------------------------|
| `ERR_BAD_REQUEST` | Client Error | Error triggered in `pranor/auth` service layer. |
| `ERR_CONFLICT` | Domain Policy | Error triggered in `pranor/auth` service layer. |
| `ERR_FORBIDDEN` | Client Error | Error triggered in `pranor/auth` service layer. |
| `ERR_INTERNAL_SERVER_ERROR` | Server Error | Error triggered in `pranor/auth` service layer. |
| `ERR_METHOD_NOT_ALLOWED` | Domain Policy | Error triggered in `pranor/auth` service layer. |
| `ERR_NOT_FOUND` | Domain Policy | Error triggered in `pranor/auth` service layer. |
| `ERR_NOT_IMPLEMENTED` | Domain Policy | Error triggered in `pranor/auth` service layer. |
| `ERR_SESSION_REVOKED` | Domain Policy | Error triggered in `pranor/auth` service layer. |
| `ERR_UNAUTHORIZED` | Client Error | Error triggered in `pranor/auth` service layer. |

---

## Pranor Cache (6 Error Codes)

| Error Code | Category | Description / Typical Trigger |
|------------|----------|-------------------------------|
| `ERR_BAD_REQUEST` | Client Error | Error triggered in `pranor/cache` service layer. |
| `ERR_BAD_REQUEST_BODY` | Client Error | Error triggered in `pranor/cache` service layer. |
| `ERR_INTERNAL_SERVER_ERROR` | Server Error | Error triggered in `pranor/cache` service layer. |
| `ERR_INVALID_PAYLOAD` | Client Error | Error triggered in `pranor/cache` service layer. |
| `ERR_METHOD_NOT_ALLOWED` | Domain Policy | Error triggered in `pranor/cache` service layer. |
| `ERR_NOT_FOUND` | Domain Policy | Error triggered in `pranor/cache` service layer. |

---

## Pranor Chrono (7 Error Codes)

| Error Code | Category | Description / Typical Trigger |
|------------|----------|-------------------------------|
| `ERR_ADD_JOB_FAILED` | Server Error | Error triggered in `pranor/chrono` service layer. |
| `ERR_BAD_REQUEST` | Client Error | Error triggered in `pranor/chrono` service layer. |
| `ERR_BAD_REQUEST_BODY` | Client Error | Error triggered in `pranor/chrono` service layer. |
| `ERR_FORBIDDEN` | Client Error | Error triggered in `pranor/chrono` service layer. |
| `ERR_JOB_NOT_FOUND` | Domain Policy | Error triggered in `pranor/chrono` service layer. |
| `ERR_METHOD_NOT_ALLOWED` | Domain Policy | Error triggered in `pranor/chrono` service layer. |
| `ERR_TRIGGER_JOB_FAILED` | Server Error | Error triggered in `pranor/chrono` service layer. |

---

## Pranor Console (40 Error Codes)

| Error Code | Category | Description / Typical Trigger |
|------------|----------|-------------------------------|
| `ERR_ALERT_NOT_FOUND` | Domain Policy | Error triggered in `pranor/console` service layer. |
| `ERR_BAD_REQUEST` | Client Error | Error triggered in `pranor/console` service layer. |
| `ERR_BAD_REQUEST_BODY` | Client Error | Error triggered in `pranor/console` service layer. |
| `ERR_CACHE_UNREACHABLE` | Server Error | Error triggered in `pranor/console` service layer. |
| `ERR_CLOUD_UNREACHABLE` | Server Error | Error triggered in `pranor/console` service layer. |
| `ERR_CONFIG_LOAD_FAILED` | Server Error | Error triggered in `pranor/console` service layer. |
| `ERR_CONFIG_SAVE_FAILED` | Server Error | Error triggered in `pranor/console` service layer. |
| `ERR_CREATE_REQUEST_FAILED` | Server Error | Error triggered in `pranor/console` service layer. |
| `ERR_CRON_UNREACHABLE` | Server Error | Error triggered in `pranor/console` service layer. |
| `ERR_DEPLOYMENT_NOT_FOUND` | Domain Policy | Error triggered in `pranor/console` service layer. |
| `ERR_EE_REQUIRED` | Domain Policy | Error triggered in `pranor/console` service layer. |
| `ERR_ENTERPRISE_REQUIRED` | Domain Policy | Error triggered in `pranor/console` service layer. |
| `ERR_FETCH_TRACE_FAILED` | Server Error | Error triggered in `pranor/console` service layer. |
| `ERR_FORBIDDEN` | Client Error | Error triggered in `pranor/console` service layer. |
| `ERR_INTERNAL` | Server Error | Error triggered in `pranor/console` service layer. |
| `ERR_INTERNAL_ERROR` | Server Error | Error triggered in `pranor/console` service layer. |
| `ERR_INVALID_BODY` | Client Error | Error triggered in `pranor/console` service layer. |
| `ERR_INVALID_ENVIRONMENT` | Client Error | Error triggered in `pranor/console` service layer. |
| `ERR_INVALID_PAYLOAD` | Client Error | Error triggered in `pranor/console` service layer. |
| `ERR_INVALID_ROUTE_PAYLOAD` | Client Error | Error triggered in `pranor/console` service layer. |
| `ERR_INVALID_SPAN_FORMAT` | Client Error | Error triggered in `pranor/console` service layer. |
| `ERR_LOCK_CONNECT` | Domain Policy | Error triggered in `pranor/console` service layer. |
| `ERR_MESH_UNREACHABLE` | Server Error | Error triggered in `pranor/console` service layer. |
| `ERR_METHOD_NOT_ALLOWED` | Domain Policy | Error triggered in `pranor/console` service layer. |
| `ERR_MISSING_FIELDS` | Client Error | Error triggered in `pranor/console` service layer. |
| `ERR_MISSING_ID` | Client Error | Error triggered in `pranor/console` service layer. |
| `ERR_MISSING_PARAM` | Client Error | Error triggered in `pranor/console` service layer. |
| `ERR_MISSING_TRACE_ID` | Client Error | Error triggered in `pranor/console` service layer. |
| `ERR_NOT_FOUND` | Domain Policy | Error triggered in `pranor/console` service layer. |
| `ERR_PARSE_TRACE_FAILED` | Server Error | Error triggered in `pranor/console` service layer. |
| `ERR_REGISTRY_UNREACHABLE` | Server Error | Error triggered in `pranor/console` service layer. |
| `ERR_ROUTE_NOT_FOUND` | Domain Policy | Error triggered in `pranor/console` service layer. |
| `ERR_RUNBOOK_NOT_FOUND` | Domain Policy | Error triggered in `pranor/console` service layer. |
| `ERR_SECRET_CONNECT` | Domain Policy | Error triggered in `pranor/console` service layer. |
| `ERR_TENANT_ID_REQUIRED` | Domain Policy | Error triggered in `pranor/console` service layer. |
| `ERR_TEST` | Domain Policy | Error triggered in `pranor/console` service layer. |
| `ERR_TRACE_ID_REQUIRED` | Domain Policy | Error triggered in `pranor/console` service layer. |
| `ERR_TRACE_NOT_FOUND` | Domain Policy | Error triggered in `pranor/console` service layer. |
| `ERR_UNAUTHORIZED` | Client Error | Error triggered in `pranor/console` service layer. |
| `ERR_UNSUPPORTED_DRIVER` | Domain Policy | Error triggered in `pranor/console` service layer. |

---

## Pranor Core (10 Error Codes)

| Error Code | Category | Description / Typical Trigger |
|------------|----------|-------------------------------|
| `ERR_API_KEY_REQUIRED` | Domain Policy | Error triggered in `pranor/core` service layer. |
| `ERR_BAD_REQUEST` | Client Error | Error triggered in `pranor/core` service layer. |
| `ERR_CHAOS_DROPPED` | Domain Policy | Error triggered in `pranor/core` service layer. |
| `ERR_FORBIDDEN` | Client Error | Error triggered in `pranor/core` service layer. |
| `ERR_INVALID_TOKEN` | Client Error | Error triggered in `pranor/core` service layer. |
| `ERR_MISSING_AUTH` | Client Error | Error triggered in `pranor/core` service layer. |
| `ERR_RATE_LIMIT_EXCEEDED` | Domain Policy | Error triggered in `pranor/core` service layer. |
| `ERR_SCOPE_REQUIRED` | Domain Policy | Error triggered in `pranor/core` service layer. |
| `ERR_TENANT_MISMATCH` | Domain Policy | Error triggered in `pranor/core` service layer. |
| `ERR_VALIDATION_FAILED` | Server Error | Error triggered in `pranor/core` service layer. |

---

## Pranor Deploy (8 Error Codes)

| Error Code | Category | Description / Typical Trigger |
|------------|----------|-------------------------------|
| `ERR_BAD_REQUEST` | Client Error | Error triggered in `pranor/deploy` service layer. |
| `ERR_CONFLICT` | Domain Policy | Error triggered in `pranor/deploy` service layer. |
| `ERR_FORBIDDEN` | Client Error | Error triggered in `pranor/deploy` service layer. |
| `ERR_INTERNAL_SERVER_ERROR` | Server Error | Error triggered in `pranor/deploy` service layer. |
| `ERR_METHOD_NOT_ALLOWED` | Domain Policy | Error triggered in `pranor/deploy` service layer. |
| `ERR_NOT_FOUND` | Domain Policy | Error triggered in `pranor/deploy` service layer. |
| `ERR_NOT_IMPLEMENTED` | Domain Policy | Error triggered in `pranor/deploy` service layer. |
| `ERR_UNAUTHORIZED` | Client Error | Error triggered in `pranor/deploy` service layer. |

---

## Pranor Flow (8 Error Codes)

| Error Code | Category | Description / Typical Trigger |
|------------|----------|-------------------------------|
| `ERR_BAD_REQUEST` | Client Error | Error triggered in `pranor/flow` service layer. |
| `ERR_CONFLICT` | Domain Policy | Error triggered in `pranor/flow` service layer. |
| `ERR_FORBIDDEN` | Client Error | Error triggered in `pranor/flow` service layer. |
| `ERR_INTERNAL_SERVER_ERROR` | Server Error | Error triggered in `pranor/flow` service layer. |
| `ERR_METHOD_NOT_ALLOWED` | Domain Policy | Error triggered in `pranor/flow` service layer. |
| `ERR_NOT_FOUND` | Domain Policy | Error triggered in `pranor/flow` service layer. |
| `ERR_NOT_IMPLEMENTED` | Domain Policy | Error triggered in `pranor/flow` service layer. |
| `ERR_UNAUTHORIZED` | Client Error | Error triggered in `pranor/flow` service layer. |

---

## Pranor Gate (36 Error Codes)

| Error Code | Category | Description / Typical Trigger |
|------------|----------|-------------------------------|
| `ERR_ACCESS_DENIED` | Domain Policy | Error triggered in `pranor/gate` service layer. |
| `ERR_AI_WAF_BLOCKED` | Domain Policy | Error triggered in `pranor/gate` service layer. |
| `ERR_BACKPRESSURE_TIMEOUT` | Domain Policy | Error triggered in `pranor/gate` service layer. |
| `ERR_BAD_GATEWAY` | Client Error | Error triggered in `pranor/gate` service layer. |
| `ERR_BAD_GATEWAY_TARGET` | Client Error | Error triggered in `pranor/gate` service layer. |
| `ERR_BAD_REQUEST_BODY` | Client Error | Error triggered in `pranor/gate` service layer. |
| `ERR_CIRCUIT_OPEN` | Domain Policy | Error triggered in `pranor/gate` service layer. |
| `ERR_CONFIG_LOAD_FAILED` | Server Error | Error triggered in `pranor/gate` service layer. |
| `ERR_CONFIG_SAVE_FAILED` | Server Error | Error triggered in `pranor/gate` service layer. |
| `ERR_EE_REQUIRED` | Domain Policy | Error triggered in `pranor/gate` service layer. |
| `ERR_FORBIDDEN_ROUTE` | Client Error | Error triggered in `pranor/gate` service layer. |
| `ERR_GO_PLUGIN_FAILED` | Server Error | Error triggered in `pranor/gate` service layer. |
| `ERR_INTERNAL_SERVER_ERROR` | Server Error | Error triggered in `pranor/gate` service layer. |
| `ERR_INVALID_API_KEY` | Client Error | Error triggered in `pranor/gate` service layer. |
| `ERR_INVALID_PATH` | Client Error | Error triggered in `pranor/gate` service layer. |
| `ERR_INVALID_PAYLOAD` | Client Error | Error triggered in `pranor/gate` service layer. |
| `ERR_INVALID_ROUTE_PAYLOAD` | Client Error | Error triggered in `pranor/gate` service layer. |
| `ERR_IP_ACCESS_DENIED` | Domain Policy | Error triggered in `pranor/gate` service layer. |
| `ERR_METHOD_NOT_ALLOWED` | Domain Policy | Error triggered in `pranor/gate` service layer. |
| `ERR_MISSING_API_KEY` | Client Error | Error triggered in `pranor/gate` service layer. |
| `ERR_POLICY_DENIED` | Domain Policy | Error triggered in `pranor/gate` service layer. |
| `ERR_PROMPT_INJECTION_DETECTED` | Domain Policy | Error triggered in `pranor/gate` service layer. |
| `ERR_QUEUE_BRIDGE_FAILED` | Server Error | Error triggered in `pranor/gate` service layer. |
| `ERR_QUEUE_FULL` | Domain Policy | Error triggered in `pranor/gate` service layer. |
| `ERR_QUEUE_RESPONSE_ERROR` | Domain Policy | Error triggered in `pranor/gate` service layer. |
| `ERR_RATE_LIMIT_EXCEEDED` | Domain Policy | Error triggered in `pranor/gate` service layer. |
| `ERR_ROUTE_NOT_FOUND` | Domain Policy | Error triggered in `pranor/gate` service layer. |
| `ERR_SCHEMA_VALIDATION_FAILED` | Server Error | Error triggered in `pranor/gate` service layer. |
| `ERR_TENANT_ACCESS_DENIED` | Domain Policy | Error triggered in `pranor/gate` service layer. |
| `ERR_TENANT_POLICY_VIOLATION` | Domain Policy | Error triggered in `pranor/gate` service layer. |
| `ERR_UNAUTHORIZED` | Client Error | Error triggered in `pranor/gate` service layer. |
| `ERR_VALIDATION_FAILED` | Server Error | Error triggered in `pranor/gate` service layer. |
| `ERR_WASM_COMPILATION_FAILED` | Server Error | Error triggered in `pranor/gate` service layer. |
| `ERR_WASM_MIDDLEWARE_FAILED` | Server Error | Error triggered in `pranor/gate` service layer. |
| `ERR_WS_HIJACK_FAILED` | Server Error | Error triggered in `pranor/gate` service layer. |
| `ERR_WS_HIJACK_NOT_SUPPORTED` | Domain Policy | Error triggered in `pranor/gate` service layer. |

---

## Pranor Hub (26 Error Codes)

| Error Code | Category | Description / Typical Trigger |
|------------|----------|-------------------------------|
| `ERR_BAD_REQUEST` | Client Error | Error triggered in `pranor/hub` service layer. |
| `ERR_BAD_REQUEST_BODY` | Client Error | Error triggered in `pranor/hub` service layer. |
| `ERR_FORBIDDEN` | Client Error | Error triggered in `pranor/hub` service layer. |
| `ERR_INTERNAL_SERVER_ERROR` | Server Error | Error triggered in `pranor/hub` service layer. |
| `ERR_INVALID_JWT` | Client Error | Error triggered in `pranor/hub` service layer. |
| `ERR_INVALID_PACKAGE_VERSION` | Client Error | Error triggered in `pranor/hub` service layer. |
| `ERR_INVALID_PATH` | Client Error | Error triggered in `pranor/hub` service layer. |
| `ERR_INVALID_PUBLIC_KEY` | Client Error | Error triggered in `pranor/hub` service layer. |
| `ERR_INVALID_SCHEMA` | Client Error | Error triggered in `pranor/hub` service layer. |
| `ERR_INVALID_SIGNATURE` | Client Error | Error triggered in `pranor/hub` service layer. |
| `ERR_METADATA_UPLOAD_FAILED` | Server Error | Error triggered in `pranor/hub` service layer. |
| `ERR_METHOD_NOT_ALLOWED` | Domain Policy | Error triggered in `pranor/hub` service layer. |
| `ERR_MISSING_FILENAME` | Client Error | Error triggered in `pranor/hub` service layer. |
| `ERR_MISSING_NAME_PARAMETER` | Client Error | Error triggered in `pranor/hub` service layer. |
| `ERR_MISSING_SIGNATURE` | Client Error | Error triggered in `pranor/hub` service layer. |
| `ERR_NAME_REQUIRED` | Domain Policy | Error triggered in `pranor/hub` service layer. |
| `ERR_NOT_FOUND` | Domain Policy | Error triggered in `pranor/hub` service layer. |
| `ERR_PACKAGE_NOT_FOUND` | Domain Policy | Error triggered in `pranor/hub` service layer. |
| `ERR_PACKAGE_UPLOAD_FAILED` | Server Error | Error triggered in `pranor/hub` service layer. |
| `ERR_PROVENANCE_NOT_FOUND` | Domain Policy | Error triggered in `pranor/hub` service layer. |
| `ERR_SCHEMA_NOT_FOUND` | Domain Policy | Error triggered in `pranor/hub` service layer. |
| `ERR_SIGNATURE_UPLOAD_FAILED` | Server Error | Error triggered in `pranor/hub` service layer. |
| `ERR_SIGNATURE_VERIFICATION_FAILED` | Server Error | Error triggered in `pranor/hub` service layer. |
| `ERR_UNAUTHORIZED` | Client Error | Error triggered in `pranor/hub` service layer. |
| `ERR_VERSION_CONFLICT` | Domain Policy | Error triggered in `pranor/hub` service layer. |
| `ERR_VERSION_NOT_FOUND` | Domain Policy | Error triggered in `pranor/hub` service layer. |

---

## Pranor Lang (14 Error Codes)

| Error Code | Category | Description / Typical Trigger |
|------------|----------|-------------------------------|
| `ERR_FORBIDDEN` | Client Error | Error triggered in `pranor/lang` service layer. |
| `ERR_RATE_LIMIT_EXCEEDED` | Domain Policy | Error triggered in `pranor/lang` service layer. |
| `ERR_ROUTE_NOT_FOUND` | Domain Policy | Error triggered in `pranor/lang` service layer. |
| `ERR_UNAUTHORIZED` | Client Error | Error triggered in `pranor/lang` service layer. |
| `SRV-E001` | Domain Policy | Error triggered in `pranor/lang` service layer. |
| `SRV-E002` | Domain Policy | Error triggered in `pranor/lang` service layer. |
| `SRV-E003` | Domain Policy | Error triggered in `pranor/lang` service layer. |
| `SRV-E004` | Domain Policy | Error triggered in `pranor/lang` service layer. |
| `SRV-E005` | Domain Policy | Error triggered in `pranor/lang` service layer. |
| `SRV-E006` | Domain Policy | Error triggered in `pranor/lang` service layer. |
| `SRV-E007` | Domain Policy | Error triggered in `pranor/lang` service layer. |
| `SRV-E008` | Domain Policy | Error triggered in `pranor/lang` service layer. |
| `SRV-E009` | Domain Policy | Error triggered in `pranor/lang` service layer. |
| `SRV-E010` | Domain Policy | Error triggered in `pranor/lang` service layer. |

---

## Pranor Mesh (8 Error Codes)

| Error Code | Category | Description / Typical Trigger |
|------------|----------|-------------------------------|
| `ERR_BAD_REQUEST` | Client Error | Error triggered in `pranor/mesh` service layer. |
| `ERR_CONFLICT` | Domain Policy | Error triggered in `pranor/mesh` service layer. |
| `ERR_FORBIDDEN` | Client Error | Error triggered in `pranor/mesh` service layer. |
| `ERR_INTERNAL_SERVER_ERROR` | Server Error | Error triggered in `pranor/mesh` service layer. |
| `ERR_METHOD_NOT_ALLOWED` | Domain Policy | Error triggered in `pranor/mesh` service layer. |
| `ERR_NOT_FOUND` | Domain Policy | Error triggered in `pranor/mesh` service layer. |
| `ERR_NOT_IMPLEMENTED` | Domain Policy | Error triggered in `pranor/mesh` service layer. |
| `ERR_UNAUTHORIZED` | Client Error | Error triggered in `pranor/mesh` service layer. |

---

## Pranor Notify (6 Error Codes)

| Error Code | Category | Description / Typical Trigger |
|------------|----------|-------------------------------|
| `ERR_BAD_REQUEST` | Client Error | Error triggered in `pranor/notify` service layer. |
| `ERR_BAD_REQUEST_BODY` | Client Error | Error triggered in `pranor/notify` service layer. |
| `ERR_METHOD_NOT_ALLOWED` | Domain Policy | Error triggered in `pranor/notify` service layer. |
| `ERR_NOT_FOUND` | Domain Policy | Error triggered in `pranor/notify` service layer. |
| `ERR_TEMPLATE_COMPILE_ERROR` | Domain Policy | Error triggered in `pranor/notify` service layer. |
| `ERR_UNSUPPORTED_CHANNEL` | Domain Policy | Error triggered in `pranor/notify` service layer. |

---

## Pranor Pool (9 Error Codes)

| Error Code | Category | Description / Typical Trigger |
|------------|----------|-------------------------------|
| `ERR_BAD_REQUEST` | Client Error | Error triggered in `pranor/pool` service layer. |
| `ERR_CONFLICT` | Domain Policy | Error triggered in `pranor/pool` service layer. |
| `ERR_FORBIDDEN` | Client Error | Error triggered in `pranor/pool` service layer. |
| `ERR_INTERNAL_SERVER_ERROR` | Server Error | Error triggered in `pranor/pool` service layer. |
| `ERR_METHOD_NOT_ALLOWED` | Domain Policy | Error triggered in `pranor/pool` service layer. |
| `ERR_NOT_FOUND` | Domain Policy | Error triggered in `pranor/pool` service layer. |
| `ERR_NOT_IMPLEMENTED` | Domain Policy | Error triggered in `pranor/pool` service layer. |
| `ERR_SERVICE_UNAVAILABLE` | Domain Policy | Error triggered in `pranor/pool` service layer. |
| `ERR_UNAUTHORIZED` | Client Error | Error triggered in `pranor/pool` service layer. |

---

## Pranor Pulse (22 Error Codes)

| Error Code | Category | Description / Typical Trigger |
|------------|----------|-------------------------------|
| `ERR_BAD_REQUEST` | Client Error | Error triggered in `pranor/pulse` service layer. |
| `ERR_BAD_REQUEST_BODY` | Client Error | Error triggered in `pranor/pulse` service layer. |
| `ERR_FORBIDDEN` | Client Error | Error triggered in `pranor/pulse` service layer. |
| `ERR_INTERNAL_SERVER_ERROR` | Server Error | Error triggered in `pranor/pulse` service layer. |
| `ERR_INVALID_PATH` | Client Error | Error triggered in `pranor/pulse` service layer. |
| `ERR_INVALID_TOKEN` | Client Error | Error triggered in `pranor/pulse` service layer. |
| `ERR_METHOD_NOT_ALLOWED` | Domain Policy | Error triggered in `pranor/pulse` service layer. |
| `ERR_MISSING_AUTH_HEADER` | Client Error | Error triggered in `pranor/pulse` service layer. |
| `ERR_MISSING_DLQ_TOPIC` | Client Error | Error triggered in `pranor/pulse` service layer. |
| `ERR_MISSING_FIELDS` | Client Error | Error triggered in `pranor/pulse` service layer. |
| `ERR_MISSING_PARAMETERS` | Client Error | Error triggered in `pranor/pulse` service layer. |
| `ERR_MISSING_TOPIC` | Client Error | Error triggered in `pranor/pulse` service layer. |
| `ERR_MISSING_TOPIC_PARAMETER` | Client Error | Error triggered in `pranor/pulse` service layer. |
| `ERR_NOT_FOUND` | Domain Policy | Error triggered in `pranor/pulse` service layer. |
| `ERR_QUERY_FAILED` | Server Error | Error triggered in `pranor/pulse` service layer. |
| `ERR_RATE_LIMIT_EXCEEDED` | Domain Policy | Error triggered in `pranor/pulse` service layer. |
| `ERR_REPLAY_FAILED` | Server Error | Error triggered in `pranor/pulse` service layer. |
| `ERR_SEEK_FAILED` | Server Error | Error triggered in `pranor/pulse` service layer. |
| `ERR_SQLITE_UNAVAILABLE` | Domain Policy | Error triggered in `pranor/pulse` service layer. |
| `ERR_STREAMING_UNSUPPORTED` | Domain Policy | Error triggered in `pranor/pulse` service layer. |
| `ERR_WASM_COMPILATION_FAILED` | Server Error | Error triggered in `pranor/pulse` service layer. |
| `ERR_WASM_TRANSFORM_FAILED` | Server Error | Error triggered in `pranor/pulse` service layer. |

---

## Pranor Secret (5 Error Codes)

| Error Code | Category | Description / Typical Trigger |
|------------|----------|-------------------------------|
| `ERR_BAD_REQUEST` | Client Error | Error triggered in `pranor/secret` service layer. |
| `ERR_FORBIDDEN` | Client Error | Error triggered in `pranor/secret` service layer. |
| `ERR_INTERNAL` | Server Error | Error triggered in `pranor/secret` service layer. |
| `ERR_METHOD_NOT_ALLOWED` | Domain Policy | Error triggered in `pranor/secret` service layer. |
| `ERR_NOT_FOUND` | Domain Policy | Error triggered in `pranor/secret` service layer. |

---

## Pranor Trace (8 Error Codes)

| Error Code | Category | Description / Typical Trigger |
|------------|----------|-------------------------------|
| `ERR_BAD_REQUEST` | Client Error | Error triggered in `pranor/trace` service layer. |
| `ERR_CONFLICT` | Domain Policy | Error triggered in `pranor/trace` service layer. |
| `ERR_FORBIDDEN` | Client Error | Error triggered in `pranor/trace` service layer. |
| `ERR_INTERNAL_SERVER_ERROR` | Server Error | Error triggered in `pranor/trace` service layer. |
| `ERR_METHOD_NOT_ALLOWED` | Domain Policy | Error triggered in `pranor/trace` service layer. |
| `ERR_NOT_FOUND` | Domain Policy | Error triggered in `pranor/trace` service layer. |
| `ERR_NOT_IMPLEMENTED` | Domain Policy | Error triggered in `pranor/trace` service layer. |
| `ERR_UNAUTHORIZED` | Client Error | Error triggered in `pranor/trace` service layer. |

---

## Pranor Tunnel (10 Error Codes)

| Error Code | Category | Description / Typical Trigger |
|------------|----------|-------------------------------|
| `ERR_BAD_GATEWAY` | Client Error | Error triggered in `pranor/tunnel` service layer. |
| `ERR_BAD_REQUEST` | Client Error | Error triggered in `pranor/tunnel` service layer. |
| `ERR_CONFLICT` | Domain Policy | Error triggered in `pranor/tunnel` service layer. |
| `ERR_FORBIDDEN` | Client Error | Error triggered in `pranor/tunnel` service layer. |
| `ERR_GATEWAY_TIMEOUT` | Domain Policy | Error triggered in `pranor/tunnel` service layer. |
| `ERR_INTERNAL_SERVER_ERROR` | Server Error | Error triggered in `pranor/tunnel` service layer. |
| `ERR_METHOD_NOT_ALLOWED` | Domain Policy | Error triggered in `pranor/tunnel` service layer. |
| `ERR_NOT_FOUND` | Domain Policy | Error triggered in `pranor/tunnel` service layer. |
| `ERR_NOT_IMPLEMENTED` | Domain Policy | Error triggered in `pranor/tunnel` service layer. |
| `ERR_UNAUTHORIZED` | Client Error | Error triggered in `pranor/tunnel` service layer. |

---

## Pranor Vault (29 Error Codes)

| Error Code | Category | Description / Typical Trigger |
|------------|----------|-------------------------------|
| `ERR_BAD_REQUEST` | Client Error | Error triggered in `pranor/vault` service layer. |
| `ERR_CLUSTER_NOT_ENABLED` | Domain Policy | Error triggered in `pranor/vault` service layer. |
| `ERR_INVALID_PATH` | Client Error | Error triggered in `pranor/vault` service layer. |
| `ERR_INVALID_POLICY` | Client Error | Error triggered in `pranor/vault` service layer. |
| `ERR_INVALID_REQUEST_BODY` | Client Error | Error triggered in `pranor/vault` service layer. |
| `ERR_INVALID_STATE` | Client Error | Error triggered in `pranor/vault` service layer. |
| `ERR_METHOD_NOT_ALLOWED` | Domain Policy | Error triggered in `pranor/vault` service layer. |
| `ERR_MISSING_CODE` | Client Error | Error triggered in `pranor/vault` service layer. |
| `ERR_MISSING_PARAMETER` | Client Error | Error triggered in `pranor/vault` service layer. |
| `ERR_NOT_FOUND` | Domain Policy | Error triggered in `pranor/vault` service layer. |
| `ERR_NOT_LEADER` | Domain Policy | Error triggered in `pranor/vault` service layer. |
| `ERR_OIDC_AUTH_URL_FAILED` | Server Error | Error triggered in `pranor/vault` service layer. |
| `ERR_OIDC_NOT_CONFIGURED` | Domain Policy | Error triggered in `pranor/vault` service layer. |
| `ERR_PLACEMENT_LOOKUP_FAILED` | Server Error | Error triggered in `pranor/vault` service layer. |
| `ERR_POLICY_DELETE_FAILED` | Server Error | Error triggered in `pranor/vault` service layer. |
| `ERR_POLICY_GET_FAILED` | Server Error | Error triggered in `pranor/vault` service layer. |
| `ERR_POLICY_PUT_FAILED` | Server Error | Error triggered in `pranor/vault` service layer. |
| `ERR_PRESIGNED_URL_FAILED` | Server Error | Error triggered in `pranor/vault` service layer. |
| `ERR_RAFT_JOIN_FAILED` | Server Error | Error triggered in `pranor/vault` service layer. |
| `ERR_RAFT_PROPOSE_FAILED` | Server Error | Error triggered in `pranor/vault` service layer. |
| `ERR_SCHEMA_GET_FAILED` | Server Error | Error triggered in `pranor/vault` service layer. |
| `ERR_SCHEMA_LIST_FAILED` | Server Error | Error triggered in `pranor/vault` service layer. |
| `ERR_SCHEMA_NOT_FOUND` | Domain Policy | Error triggered in `pranor/vault` service layer. |
| `ERR_SCHEMA_PUT_FAILED` | Server Error | Error triggered in `pranor/vault` service layer. |
| `ERR_STORE_PUT_FAILED` | Server Error | Error triggered in `pranor/vault` service layer. |
| `ERR_TOKEN_EXCHANGE_FAILED` | Server Error | Error triggered in `pranor/vault` service layer. |
| `ERR_TOKEN_GENERATION_FAILED` | Server Error | Error triggered in `pranor/vault` service layer. |
| `ERR_UNAUTHORIZED` | Client Error | Error triggered in `pranor/vault` service layer. |
| `ERR_USER_INFO_FAILED` | Server Error | Error triggered in `pranor/vault` service layer. |

---
