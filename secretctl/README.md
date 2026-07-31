# pranor-secretctl

`pranor-secretctl` is the standalone, open-source command-line interface (CLI) administration client for **Pranor Secret** (Secret & Credential Manager).

It provides terminal controls to set, get, delete, rollback, and list secrets, as well as running child commands with secrets injected directly into their environment.

## Installation

```bash
go build -o pranor-secretctl.exe .
```

## Configuration

You can configure the client using flags or environment variables:

| Flag | Env Variable | Description | Default |
|---|---|---|---|
| `--url` | `PRANOR_SECRET_URL` | Pranor Secret server URL | `http://localhost:8091` |
| `--api-key` | `PRANOR_SECRET_API_KEY` | API Key for authentication | (None) |
| `--tenant` | - | Tenant ID for request isolation | `default` |

## Usage Commands

### 1. Set a secret
```bash
pranor-secretctl set --key db.password --value "supersecretpass"
```

### 2. Retrieve a secret
```bash
pranor-secretctl get --key db.password
```

### 3. Delete a secret
```bash
pranor-secretctl delete --key db.password
```

### 4. List secret keys in tenant scope
```bash
pranor-secretctl list
```

### 5. Rotate master encryption key
```bash
pranor-secretctl rotate --new-key "32_byte_hex_key_here"
```

### 6. Rollback a secret to previous version
```bash
pranor-secretctl rollback --key db.password
```

### 7. Run command with secrets injected into environment
```bash
pranor-secretctl run --cmd "npm run start"
```

### 8. Check Health
```bash
pranor-secretctl health
```

## License
Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0). See [LICENSE](./LICENSE) for details.
