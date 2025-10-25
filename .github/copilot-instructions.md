# GitHub Copilot Instructions for terraform-provider-pbs
## Project snapshot
- Terraform provider targeting Proxmox Backup Server 4.0+, using HashiCorp terraform-plugin-framework (see `fwprovider/provider.go`).
- Provider wiring hands a shared `pbs.Client` (built in `pbs/client.go`) into resources via `fwprovider/config`.
- HTTP interactions are wrapped by `pbs/api`, then domain clients under `pbs/datastores`, `pbs/endpoints`, `pbs/jobs`, etc.
## Key directories
- `fwprovider/resources/**`: Terraform resource implementations; each file holds schema, CRUD, and helper translators.
- `pbs/**`: Thin Go SDK mirroring PBS API endpoints, including async task handling and S3 backend parsing.
- `test/integration`: Acceptance-style tests driven by Terraform CLI (`tfexec`), requires a built provider binary.
- `scripts/docker-compose.test.yml` + `test/run_docker_tests.sh`: spins up PBS + auxiliaries (InfluxDB, Gotify, webhook, NFS, CIFS).
## Implementation patterns
- Resources follow a plan↔API converter pair (e.g., `planToDatastore`, `datastoreToState`) that lowercases enums and preserve credentials.
- Updates must carry the PBS `digest` for optimistic locking and send `Delete` slices when Terraform nulls optional fields.
- Long-running PBS tasks (datastore creation) return a UPID; `pbs/datastores.CreateDatastore` waits and sleeps 3s post-task—reuse that flow for new async operations.
- Notification resources normalize list/string attributes and default booleans via explicit plan modifiers; copy this style to avoid diffs.
## PBS API schema quick reference
- Source: `https://pbs.proxmox.com/docs/api-viewer/apidoc.js` exposes a tree of paths under `apiSchema`; each leaf lists HTTP methods with parameter/return schemas, permission checks, and doc strings.
- Authentication & access control lives under `/access` (users, tokens, ACLs, realms). Provider currently consumes `/access/users` + `/access/acl` indirectly via login; respect regex constraints on IDs (`^[A-Za-z0-9_][A-Za-z0-9._\-]*$`).
- Datastores:
	- Config endpoints live under `/config/datastore` (GET/POST/PUT/DELETE require `Datastore.Modify` & carry `digest`). Optional tuning is encoded as strings like `gc-*/sync-level`; mirror `Delete` semantics when Terraform unsets them.
	- Admin/runtime operations reside at `/admin/datastore/{store}` (garbage collection, status, notes, namespaces, snapshots, verify, sync). Many writes return `UPID:` strings; always wait via `WaitForTask` and pass namespaces when supported.
- Jobs:
	- `/config/prune`, `/config/sync`, `/config/verify` expose CRUD for scheduled jobs with repeated patterns: `id`, optional boolean `disable`, namespace filters (`ns`, `max-depth`), target mapping arrays, and `schedule` strings (`<calendar-event>`). Deletions use `delete` arrays in PUT.
- Remotes & S3:
	- `/config/remote` handles pull targets with fields `remote-host`, `auth-id`, `fingerprint`; nested `/scan` endpoints enumerate stores/groups/namespaces.
	- `/config/s3` (alias `/config/system/s3-endpoint`) requires `provider`, `endpoint`, `access-key`, `secret-key`; optional `provider-quirks` array includes `skip-if-none-match-header` used in code.
- Metrics & notifications:
	- `/config/metrics/influxdb-http|udp` share structs (`bucket`, `organization`, `token`, `verify-tls`). Default booleans and `comment` fields align with provider schemas—ensure we propagate defaults to avoid diff noise.
	- `/config/notifications` splits into `endpoints/*` (gotify, sendmail, smtp, webhook) and `matchers`. Endpoints use base64 `secret`/`header` arrays with `name` + optional `value`; matchers reference `targets` and `filter` arrays of string expressions.
- Tape & system namespaces expose many paths (e.g., `/tape/drive`, `/nodes/{node}/network`). Provider currently ignores them; if future work adds resources, follow existing regex/enum constraints and permission requirements from the spec.
- Common patterns: most update endpoints expect `digest`, booleans default to `true`/`false` instead of omitting, and many parameters allow optional namespace strings up to 256 chars using PBS namespace regex. Preserve these constraints in validation.
## Provider configuration quirks
- Terraform config expects `endpoint`, but the provider also reads env vars (`PBS_ENDPOINT`, `PBS_API_TOKEN`, etc.); tests still export `PBS_ADDRESS`, so set both when scripting.
- The `insecure` flag flows into TLS skip-verify—respect it when adding new clients.
- `PBS_DESTROY_DATA_ON_DELETE=true` forces destructive cleanup during tests; honor it in delete paths.
## Adding or updating resources
- Register new resources in `fwprovider/provider.go` so Terraform can discover them.
- Compose schema with plugin-framework validators and `planmodifier.UseStateForUnknown()` for computed-but-optional fields.
- Translate Terraform types to API structs with helper converters (see `fwprovider/resources/jobs/helpers.go`) instead of re-implementing pointer plumbing.
- For nested config encoded as strings by PBS (notify, tuning, maintenance), reuse the formatter/parsers in `pbs/datastores`.
## Testing workflow
- Run `make test` for Go unit coverage; `make testacc` toggles `TF_ACC` against the framework-level resources.
- Integration suite: `scripts/run-integration-tests.sh` (runs docker services) or `test/run_docker_tests.sh` (local PBS container) after `go build .` produces `terraform-provider-pbs`.
- When using the docker harness, Terraform state requires `NODE_PATH` adjustments already handled in `test/integration/setup.go`; avoid removing that logic.
- Multi-provider S3 tests only execute if AWS/B2/Scaleway env vars are present; keep skips in place to allow partial credentials.
## Debugging tips
- Enable `TF_LOG=DEBUG` to surface `tflog` traces baked into each resource.
- API errors often include a UPID; use `pbs/api.Client.WaitForTask` helpers instead of polling manually.
- Lock contention errors from PBS are retried with exponential backoff (see `createDatastoreWithRetry`); match that behavior for other write-heavy endpoints.
## Reference snippets
- Example Terraform acceptance fixture lives in `examples/provider/provider.tf`; mirror its provider block when synthesizing configs in tests.
- Metrics server tests rely on InfluxDB ports (8086/8089) exported via `TEST_INFLUXDB_*`; ensure new metrics resources respect those variables.
- Notification endpoints treat `targets` as optional lists and preserve `origin` from the API—maintain those semantics during updates.
## Deployment
- `make build` emits a local provider binary; `make install` copies it into `~/.terraform.d/plugins/registry.terraform.io/micah/pbs/${VERSION}/darwin_amd64/` for manual Terraform runs.
- Release builds are managed through GoReleaser (`make release-test`), so avoid introducing extra build steps without updating the workflow.
