## lys-ref MCP Analytics Server

This binary provides analytics tools over MCP for the Digmark campaign performance data in Postgres.

- Transport: stdio (meaning: accepts STDIN, outputs to STDOUT and STDERR)

The server is intended to be launched by an MCP host (for example VS Code Copilot Chat) and queried in natural language.

## Example Questions

- "What are the best performing verticals this week?"
- "Show me daily trend for the last 30 days"
- "Which campaigns had the highest ROI this month?"
- "Give me a summary of this week's performance"

## Tool Parameters (Quick Reference)

### `get_vertical_performance`

- `period`: `day|week|month|year` (default: `week`)
- `order_by`: `profit|revenue|roi|conversions` (default: `profit`)

### `get_top_campaigns`

- `period`: `day|week|month|year` (default: `week`)
- `metric`: `profit|revenue|roi` (default: `profit`)
- `limit`: integer `1..100` (default: `10`)

### `get_performance_summary`

- `period`: `day|week|month|year` (default: `week`)

### `get_daily_trend`

- `days`: integer `1..365` (default: `14`)

## AWS Deployment Options

The server currently uses stdio transport, so deploy it where the MCP host process can start it as a child process.

### Option 1: EC2 + systemd (simple and reliable)

Best when your MCP host runs on the same EC2 instance (for example a self-hosted coding environment).

1. Provision EC2 in the same VPC/subnet region as RDS.
2. Install Go (or ship prebuilt `bin/mcpsrv`).
3. Place `mcp_config.toml` with least-privileged DB credentials.
4. Configure security groups:
	 - EC2 egress to Postgres port
	 - RDS ingress from EC2 security group
5. Launch the MCP host and configure it to run `./bin/mcpsrv -config ...`.

Use AWS Systems Manager Session Manager for access instead of opening SSH broadly.

### Option 2: ECS/Fargate sidecar (containerized host + MCP)

Best when your MCP-capable app is containerized.

1. Build a container image containing `mcpsrv`.
2. Run MCP host and `mcpsrv` in the same task (or same container image) so stdio can be used locally.
3. Inject DB config via environment variables + generated TOML at startup, or mount a secret file.
4. Use Secrets Manager for DB credentials.
5. Place the task in private subnets with NAT as needed and allow DB access via security groups.

Important: stdio does not cross network boundaries by itself; the MCP host must be able to spawn `mcpsrv` directly.

### Option 3: Remote MCP over HTTP (requires an additional bridge)

Use this only if your MCP client cannot run the server as a child process.

1. Add an MCP transport bridge/gateway that converts HTTP/SSE/streamable transport to local tool execution.
2. Run `mcpsrv` behind that bridge.
3. Publish via ALB/API Gateway with TLS and auth.

This repository does not currently include a native HTTP MCP transport in `cmd/mcpsrv`.

## AWS Security Notes

- Prefer IAM role-based secret retrieval (Secrets Manager or Parameter Store).
- Restrict DB user permissions to read-only analytics access.
- Keep all workloads in private subnets where possible.
- Enable CloudWatch logs for host process output and MCP failures.
- Rotate DB credentials and avoid hardcoding them in images.

## Troubleshooting

- "config file not found": verify `-config` path and working directory.
- "failed to decode config": check TOML keys and section names.
- "Db.Database is empty" or "DbMcpUser.Name is empty": required keys are missing.
- "failed to open DB pool": verify network path, credentials, SSL settings, and DB availability.
- MCP client cannot see tools: ensure MCP server is enabled and launched by the client, not just run manually.
