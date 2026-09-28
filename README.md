# LuaSkills Go SDK

English documentation is the default package documentation. For Chinese, see [README_cn.md](README_cn.md).

Main LuaSkills repository: [LuaSkills/luaskills](https://github.com/LuaSkills/luaskills)

Go SDK for integrating the LuaSkills runtime through the public JSON FFI surface.

`0.5.7` is the current release line. It retains the strict package-level skill configuration contract and defaults runtime assets to LuaSkills core `v0.5.7`, vldb-controller `v0.2.3`, and vldb-sqlite `v0.1.6`.

The SDK wraps cgo JSON FFI calls, engine lifecycle, formal skill roots, authority-aware management calls, skill config, provider callback boundaries, host-tool callback boundaries, and runtime manifest helpers.

## Embedded runtime development API

Explicit persistence is available through `InitializePersistent(ctx, options, budgets, persistence)`. Existing `Initialize` keeps memory-only behavior. Supply the generated configuration with a host-owned absolute `path`, complete `journal` retention budgets and `worker` receipt budgets; the SDK adds no storage defaults or fallback. Initialization acknowledges an attempt; inspect native status for the outcome. Status includes actual optional storage ownership, and coordinated shutdown waits for the core, writer and retained receipts.

Runtime methods `StorageStatus / RecoverStorage` expose worker ownership and explicit same-file recovery. `HistoryGet / HistoryNext / HistoryForget` read and remove original history by **core runtime namespace**, distinct from the FFI slot ID. History and recovery use the work lane; storage status and operation methods `PersistenceFailure / RetryCheckpoint` use the control lane. Recovery never retries a checkpoint automatically or replays business work. A retry without a failed checkpoint reports busy; false means an existing retry is already pending. History enumeration uses original-key cursors, not a multi-call snapshot.

`RecoverStorageWorker` independently reconstructs one supervised failed writer only after its actual thread exits and all queued/active attempts have been supervised. It returns false for a healthy running writer and rejects explicit writer closure, unproven exits or poisoned ownership. Original receipts, byte/count reservations and the same storage owner remain intact. It never replays old writes or business work; database recovery and original checkpoint retry remain separate explicit actions. Runtime closure can still recover a writer while the original core is draining, before permanent writer closure.

Historical records never become active handles. Before administrative reconciliation or removal, forget any retained live operation. `HistoryReconcile` attaches one final, bounded host attestation while preserving the original snapshot, caller and effect identities. The trusted host must authorize the resolver, prove all original execution and external owners stopped, and verify the whole operation and every recorded effect; a supplied resolver string is not authentication. Unknown ordinary Lua effects cannot be inferred from successful return. Use the original namespace, operation ID and predecessor revision; retry the identical resolution to acknowledge the same successor after storage recovery. The returned revision supports explicit removal. This API does not query external systems, replay callbacks, manufacture an interrupted result or recover a failed writer/execution stack. Development disk format 4 rejects unpublished formats 1–3 without rewriting them. Matching development core and SDK contracts are required.

The development source adds `NewEmbeddedTransport` over the independent version-one C ABI. Link a matching development core containing the new exports; this does not claim that published `0.5.7` libraries support them. Versions and default runtime assets will move together after ecosystem acceptance. With `CGO_ENABLED=0`, contracts and codecs remain available while native construction returns an explicit error.

Before native allocation, construction reads `luaskills_ffi_embedded_describe_v1`, bounds the borrowed pointer/length before copying, and validates exact core/protocol/ABI/description versions, the packaged contract digest, required commands and capabilities, supported backend, process OS and pointer width. Malformed or incompatible metadata returns `*EmbeddedCompatibilityError`; nonzero native status retains `*EmbeddedTransportError`. The library owns discovery bytes; they are never passed to a result-free function. `CoreDescription()` returns independently cloned typed evidence, also after successful release. Build hashes describe selected inputs and do not authenticate a binary or prove a hermetic build.

The cgo backend calls symbols resolved at link/load time; an old library missing the discovery symbol can fail linking or OS loading before Go code runs, so that case cannot always become an SDK error. There is no legacy fallback or SDK library unload operation. Driver goroutines reuse the exact linked backend retained by their transport instead of loading another path. Use immutable versioned native assets; replacing a core in a running process is outside this API's update contract.

Transport configuration explicitly bounds runtime count, response count, aggregate response bytes, per-response bytes and request bytes, then remains immutable. `Request(map[string]any)` synchronously freezes a command and returns actual delivery. Go integers retain all 64 bits; floats retain decimal/exponent markers. Decoded numbers are `json.Number`, never rounded through float64. Inputs support string-keyed maps, typed slices/arrays, scalars and explicitly JSON-tagged structs. Tags support names, `omitempty` and exclusion only; custom serialization hooks and implicit base64 are rejected. Duplicate decoded keys, invalid Unicode, cycles, extra envelope fields and numeric overflow are errors.

`EmbeddedTransportError` and `EmbeddedRuntimeError` distinguish ABI failures from delivered business rejections. `EmbeddedResultReleaseError` retains copied response evidence: `ResponseBytes()` returns a fresh copy and `DeliveredResult()` decodes the original result. Release failure never authorizes mutation replay. `ReleaseResults()` retries only retained allocation releases and rejects active readers to prevent racing a copy.

`Close()` requests admission closure. Close and remove actual runtimes and release buffers before `Free()` succeeds. Native execution never holds a global Go lock, so control calls can proceed concurrently. GC does not replace cleanup; `LiveEmbeddedTransports()` keeps owners discoverable. This synchronous low-level interface has no observer cancellation.

`NewEmbeddedCommandDriver(transport, config)` adds fixed business workers and one independent control worker over a borrowed transport. Configure positive `WorkWorkers`, `MaxWorkCommands` and `MaxControlCommands`; quotas include completed receipts until explicit `Forget()`. Construction requires no active requests, retained native results or existing driver, and enough count and aggregate-byte capacity for every worker's maximum response. The guarantee covers driver workers; callers must account separately for concurrent raw transport use. Lua VM concurrency remains core-owned.

`Submit(ctx, generatedCommand)` freezes and validates the command, derives its lane and publishes an `EmbeddedCommand` before execution. The context controls pre-admission cancellation only. `Result(ctx)` cancels only that observer's wait; accepted work and its receipt remain owned and discoverable through `Commands()`. Results, response bytes and retained public errors are copied for independent observers. Decode `ResponseBytes()` with the corresponding generated response decoder. Native `operation_wait` is rejected: poll `operation_status`, and request actual cancellation explicitly with `operation_cancel`.

A failed native result release pauses the owning worker, preserves its frame reservation and exposes `EmbeddedResultReleaseError`. `driver.ReleaseResults(ctx)` retries retained buffer releases without receipt quota or command replay; active readers return Busy. Once recovery starts it is synchronous and cannot be interrupted by context cancellation. A successful recovery resumes paused workers; their original receipt still records the release failure, and `DeliveredResult()` exposes the original delivery. Use driver-level recovery while a driver owns the transport so paused workers are notified.

`RequestClose()` fences submissions and drains accepted queues. `Close(ctx)` starts the same drainage even if its observer is already cancelled, and waits for actual worker exit; timeout leaves the prestarted close coordinator alive. Recover failed releases explicitly to let drainage finish. Closing the driver does not close borrowed runtimes or the transport. Close native runtimes through control commands before driver closure, or use the raw transport afterward, then close/free the transport separately. `LiveEmbeddedCommandDrivers()` retains closed drivers with receipts until every completed receipt is explicitly forgotten. Unexpected worker/recovery exits produce `EmbeddedDriverFailure`, retain uncertain ownership and forbid successful closure or replay; an uncertain command never reports `Done()`.

`NewEmbeddedCallbackPump(transport, runtimeID, config)` now owns automatic queued Go callbacks. Set positive `MaxConcurrentHandlers`, `MaxPendingCommands` and `PollIntervalMS`, then observe `Ready(ctx)` before registering. Each pump has one sequential native coordinator and fixed handler goroutines; returned handlers awaiting acknowledgement still occupy capacity. Driver and pump response reservations are checked together. Create owners at a quiescent transport boundary; one pump is permitted per exact runtime. Route all of that runtime's queued registrations through the pump.

Register `EmbeddedHostCapability` values with a generated queued descriptor and an `EmbeddedHostHandler`. `Register(ctx, capabilities)` freezes descriptors and retains the exact handlers before publication; cancellation detaches the observer, so consult `Status()` for registrations that were accepted. Handlers receive arguments separately from `EmbeddedHostCallbackContext`, which implements `context.Context`, exposes copied trusted caller metadata, and observes the first core cancellation. `RemainingMS()` is advisory, not a fabricated local deadline. Mutating handlers start with unknown effects; report actual evidence with `ReportEffects`, which seals when the handler returns. Panic, `runtime.Goexit`, invalid results and arbitrary errors become bounded generic failures; explicitly valid `EmbeddedRuntimeError` messages are host-authorized disclosure. Lua `vulcan.capabilities.call` receives an `ok/value/error/effects` envelope; returning that envelope normally does not turn a callback failure into a failed outer Lua operation.

Pass the supplied callback context, or a derived context, into downstream APIs. Driver submissions and SDK waits reject this marker until controlled callback dependencies are implemented. Deliberately discarding the context or directly calling the unmanaged synchronous transport bypasses this guard; it is not a goroutine sandbox. Nonblocking diagnostics and closure requests remain available.

`Unregister(ctx, registrationID)` waits for actual handler return, native drainage and metadata removal. `Close(ctx)` retires all accepted registrations and joins actual handler workers without terminating application code. Errors fence new admission and retain ownership; `RetryAcknowledgements(ctx)` explicitly releases retained native buffers and reconciles failed completions using exact request, registration and operation evidence. It never re-runs a handler. Only a request-parser `INVALID_ARGUMENT` proving non-dispatch permits one replacement failure acknowledgement with the original effect evidence. Lost registration, extraction or retirement responses are retained without replay or invented success. `LiveEmbeddedCallbackPumps()` preserves discovery, and transport `Free()` rejects an owned pump. Recovery drains a failed pump; it does not reopen admission. Runtime scopes below own ordered shutdown.

`NewEmbeddedClient(driver)` borrows an existing driver and exposes `EmbeddedRuntime`, `EmbeddedPlugin`, `EmbeddedPool`, `EmbeddedSession` and `EmbeddedOperation`. Command methods take an admission context and return `(*EmbeddedPending[T], error)`. `pending.Result(ctx)` observes delivery and validates the generated result shape; `pending.DeliveredResult()` projects only original delivery, recovering handles created before buffer-release failure without replaying mutations. `pending.Receipt()` exposes the exact driver receipt. `pending.Forget()` returns SDK quota only; a handle's `Forget(ctx)` or runtime `Free(ctx)` changes core records. Projection failure retains the receipt and original bytes. Each observation decodes fresh data rather than sharing mutable responses.

Use `client.Reserve(ctx)` to obtain the actual slot identity, then `runtime.Initialize(ctx, options, budgets)` with generated `EmbeddedInputLuaEngineOptions` and `EmbeddedInputEmbeddedRuntimeConfig`. `RuntimeID()` identifies the FFI slot; snapshot `CoreRuntimeId` belongs to a different namespace and is not interchangeable. `client.Runtime(id)` and runtime `Plugin/Pool/Session/Operation(id)` bind known nonempty exact identities without probing or selecting replacements. Hosts provide explicit configuration; the typed API does not change the map result of the existing `CreateEngineOptions` helper.

`RegisterPool(ctx, definition, policy, permissions, executionRevision)` registers an immutable execution domain. **Register required host capabilities before registering the pool**: the core captures a capability snapshot at pool registration, so later publications do not enter an existing pool. Create a new execution domain and drain the old one when adopting a new registration. Generated policy explicitly selects shared/dedicated ownership and reusable/single-call/session lifetime. `pool.Submit(ctx, export, arguments, invocation, timeoutMS)` returns an independent operation. `pool.OpenSession(ctx, timeoutMS)` returns `EmbeddedSessionOpen{Session, Initialization}`; observe the initialization operation separately before using the fixed session's `Submit` method to retain module state.

`operation.Wait(ctx)` uses `EmbeddedDefaultPollInterval`; `WaitInterval(ctx, positive time.Duration)` selects an explicit interval. It polls short status commands without blocking a native worker on a wait. Only `succeeded/failed/cancelled` are terminal; cleanup and cancellation intent are not completion. Failed and cancelled terminal snapshots retain effect evidence; Go errors represent admission, delivery, projection or observation failure. Successful read-only polls automatically forget their SDK receipts. Interrupted or failed status receipts remain in `client.Driver().Commands()` and must be observed and explicitly forgotten to return quota. An observer timeout does not send native cancellation; use `operation.Cancel(ctx)` for cooperative cancellation, then continue querying actual completion and late effects.

`NewEmbeddedRuntimeScope(runtime, pump, interval)` owns native admission closure, actual core drainage with callbacks available, callback pump shutdown and final slot removal. Pass `nil` without a callback pump and use `EmbeddedDefaultPollInterval` for the default interval. One prestarted control coordinator owns an independent worst-case response reservation, budgeted together with ordinary driver and pump frames. Scope cleanup remains available when ordinary receipt quota is full or the driver has closed. The driver and transport are borrowed and are not closed on behalf of other runtimes.

Single-call modules may declare `Finalizer` with a declared `Export`, fixed `Arguments` and finite `TimeoutMs`. The core retains the same VM and original operation through automatic closing. Generated `Finalization` fields retain independent business and closing outcome unions, preserving successful null versus absent evidence. A missing closing outcome does not prove the callback never ran. Scope shutdown preserves callback delivery until finalizers finish; explicit capability revocation still applies. Use matching generated contracts and core binaries; storage recovery never replays a closing callback.

Session modules also support this declaration. Opening reserves one additional operation slot under both runtime and plugin retention limits before initialization; plugin status exposes `ReservedOperations` separately from retained records. Closing an initialized session consumes that reservation, preserves earlier business results, and exposes the independent operation through session status `FinalizationOperation`. Its business baseline is successful null with no business effects; the actual closing result is `Finalization`'s outcome. Explicit session, pool, plugin and runtime close, or business failure/use exhaustion, all preserve the original VM through closing and actual retirement. The closing deadline starts only when closing execution is issued.

Reusable pools also support automatic closing. Each new VM reserves one closing operation slot before initialization; reuse of that VM needs no additional reservation. Idle expiry, pressure eviction, business failure/use exhaustion and scope closure trigger an independent closing operation after the business result is published. Pool/plugin closure drains admitted business calls; explicit runtime closure still cancels them. Use runtime `ListOperations` (wire command `operation_list`) with an optional exact pool filter, a retained `after_operation_id` cursor and explicit positive `limit` to discover these operations in publication order. Their module context carries `finalization_instance_id`, and the outcome remains queryable after pool removal. An empty page preserves its cursor; forgetting the cursor requires restarting enumeration. Pages and response bytes stay bounded by core budgets, and discovery neither executes callbacks nor replaces durable history.

Adopt after existing driver commands settle and the transport has no active calls or retained results; an existing pump's short poll may transiently report Busy. Create the pump first and adopt that exact instance. Missing or foreign pumps, duplicate scopes and late pump attachment are rejected. All cleanup frames are validated against the request budget before adoption. Driver admission and scope adoption use the same lock order, so queued but not yet entered removal also blocks adoption. After adoption, both typed `Free` and direct driver `runtime_free` are rejected. The synchronous low-level transport is unmanaged; callers must respect ownership themselves. Construct scopes through the constructor, never copy them, and do not rely on GC cleanup.

`scope.RequestClose()` starts cleanup without waiting. `scope.Close(ctx)` starts the same cleanup even when the observer context is already cancelled, then observes actual completion and coordinator exit. Nil or callback-marked contexts are rejected before starting. Observer timeout/cancellation does not release handlers, runtimes or response reservations and does not stop cleanup. `Status()` and `LiveEmbeddedRuntimeScopes()` expose retained owners. Repeated close observes the original attempt rather than implicitly retrying a failure.

`scope.RetryClose(ctx)` explicitly recovers retained allocations, failed acknowledgements of the adopted pump, or control rejection proven to precede mutation. Copied successful receipts advance checkpoints even when release fails; recovery never closes or removes that slot again. Request-entry capacity refusal and local transport-entry Busy have explicit non-execution evidence. Only the core's business `busy` on slot removal is automatically polled while leases remain. Missing delivery, invalid shape or identity, `faulted` construction and coordinator panic/`Goexit` retain uncertain ownership without fabricated success or mutation replay. An exited callback coordinator is not offered ordinary acknowledgement recovery. Other drivers or pumps paused on the shared transport still require their own recovery APIs to continue.

After successful scope closure, settle ordinary driver receipts and close the driver; close/free the shared transport after every scope and pump has drained. Without a scope, typed handles still require explicit native close, pump drainage, actual closed-state observation and slot release, preserving original delivery evidence on every failure.

The packaged contract also generates independent `EmbeddedInput*` and `EmbeddedOutput*` types, enum constants, sealed command alternatives and every declared response decoder. `EncodeEmbeddedRequest` freezes and validates a typed envelope; callers set `EmbeddedProtocolVersion` and the generated command discriminator explicitly. `DecodeEmbeddedOutput*Response` validates exact field names, required fields, enums, integer widths and declared set constraints without custom serialization hooks. This is structural wire validation; native business checks, such as host-completion success semantics and runtime budgets, remain authoritative.

Optional fields use an outer pointer for presence; nullable fields use an inner pointer. Thus `**T` can distinguish absent, present null and present value, and `*any` preserves successful null separately from a missing result. Use the generated response decoders to preserve these states: ordinary `encoding/json.Unmarshal` does not retain that distinction for null pointers. Where upstream permits additional object fields, typed projection accepts them but exposes only declared fields; retain original response bytes when extensions are needed. Closed shapes and exact root envelopes reject additional fields. The generator rejects unsupported future schema constraints, conflicting output definitions and generated-name collisions instead of weakening types.

`contracts/embedded/v1` and both root C headers are exact copies from the same upstream source. `go run ./scripts/generate-embedded-contract --check` verifies generated protocol constants, status codes, command metadata and all wire types from the packaged contract without an adjacent checkout. Native tests require explicit `LUASKILLS_NATIVE_E2E=1` and matching library configuration. `python scripts/verify_embedded_distribution.py` verifies a real Go module ZIP through a private file proxy and empty module cache, including packaged generation and embedded tests. Use `--go` to select an installed toolchain. This validation version is never published externally.

Windows cgo links `luaskills.dll` explicitly to avoid selecting an MSVC static archive in the same directory. Both the linker directory and runtime PATH must contain the matching DLL, following [GNU ld's Windows DLL rules](https://sourceware.org/binutils/docs/ld/WIN32.html). Go request bytes are borrowed only during the synchronous C call; responses are copied before release according to [cgo pointer rules](https://pkg.go.dev/cmd/cgo#hdr-Passing_pointers).

## Installation

```bash
go get github.com/LuaSkills/luaskills-sdk-go
```

Runtime calls require `CGO_ENABLED=1`, a C compiler compatible with Go cgo, and a discoverable LuaSkills dynamic library.

Windows example:

```powershell
$env:CGO_ENABLED = "1"
$env:CGO_LDFLAGS = "-LD:\runtime\luaskills\libs"
$env:PATH = "D:\runtime\luaskills\libs;$env:PATH"
```

Linux / macOS example:

```bash
export CGO_ENABLED=1
export CGO_LDFLAGS="-L/opt/luaskills-runtime/libs"
export LD_LIBRARY_PATH="/opt/luaskills-runtime/libs:${LD_LIBRARY_PATH}"
```

## Runtime Assets

The repository includes a unified synchronization script that directly downloads LuaSkills FFI, Lua runtime packages, and VLDB without requiring the Python or TypeScript installer:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/deps/sync_runtime_assets.ps1 -Target all -Database vldb-controller -RuntimeRoot D:\runtime\luaskills
```

```bash
RUNTIME_ROOT=/opt/luaskills scripts/deps/sync_runtime_assets.sh all vldb-controller
```

Supported targets are `all`, `luaskills`, `lua`, and `vldb`. VLDB presets are `none`, `vldb-controller`, `vldb-direct`, and `host-callback`. The scripts pin LuaSkills to `v0.5.7` by default and accept explicit release-version overrides.

The Go SDK plans and consumes the shared SDK runtime manifest, while the repository scripts above directly download release assets. The shared manifest now points at:

- `lua-runtime-packages-{platform}.tar.gz` from `LuaSkills/luaskills-packages`
- `luaskills-ffi-sdk-{platform}.tar.gz` from `LuaSkills/luaskills`
- optional managed Python, `uv`, Node.js, and `pnpm` paths under `runtime_root/dependencies/runtimes/...`

## Repository Structure

The module root is the public `luaskills` package and contains exported API domains plus the thin native bridge required by Go build tags. Private implementation is isolated under `internal`:

```text
internal/protocol/       JSON FFI response-envelope decoding
internal/runtimeassets/  Runtime-manifest path validation
internal/sessionwake/    Concurrent managed-session wake callback ownership
examples/                Independently buildable SDK examples
scripts/                 Runtime asset and managed-runtime bootstrap tools
```

The cgo and no-cgo bridge files remain in the root package because they implement the same private functions behind mutually exclusive build tags. Internal packages never expose C types and cannot be imported by SDK consumers.

Managed child runtimes support Windows x64, Linux x64/ARM64, and macOS x64/ARM64. Windows ARM is explicitly rejected before any download or target-directory creation. The repository also ships standalone fetch and layout-validation tools for hosts that prepare debug runtimes without a Python or TypeScript installer:

The current exact managed dependency versions are Python `3.14.6`, uv `0.11.28`, Node.js `24.18.0`, and pnpm `11.11.0`. Package `dependencies.yaml` files must declare the same exact runtime and package-manager versions unless the host deliberately installs another supported version.

LuaSkills 0.5.1 separates the LuaSkills data root, read-only interpreter distribution root, and writable managed-environment root. Both explicit managed roots must be absolute; when omitted, LuaSkills keeps the compatible `runtime_root/dependencies/runtimes` and `runtime_root/dependencies/envs` layout.

```go
invokeTimeoutMS := uint64(30_000)
hostOptions := map[string]any{
	"managed_runtime_distribution_root": "D:/VulcanCode/dependencies/runtimes",
	"managed_runtime_environment_root":  "D:/VulcanCodeData/managed-runtime-envs",
	"managed_runtime_config": luaskills.ManagedRuntimeConfig{
		WorkerPoolMaxSizePerEnvironment:                   8,
		WorkerIdleTTLSecs:                                 120,
		PersistentSessionLimitPerEngine:                   128,
		PersistentSessionDefaultBufferLimitBytesPerStream: 2 * 1024 * 1024,
		InvokeDefaultTimeoutMS:                            &invokeTimeoutMS,
	},
}

pythonInstall, err := luaskills.ResolveManagedRuntimeInstall(luaskills.ManagedRuntimeResolveOptions{
	DistributionRoot: "D:/VulcanCode/dependencies/runtimes",
	Runtime:          luaskills.ManagedRuntimeKindPython,
	Version:          "3.14.6",
	Platform:         "windows-x64",
})
if err != nil {
	panic(err)
}
```

`DefaultManagedRuntimeConfig()` returns the stable engine defaults: `4` Workers per exact environment/package-owner pool, `60` idle seconds, `256` persistent sessions, `1 MiB` per session output stream, and a nil default invoke timeout. Copy that value and set `InvokeDefaultTimeoutMS` to a positive `*uint64` when a finite engine default is required. Every configured number must be positive; per-call `invoke.timeout_ms` and per-session `session.open.buffer_limit_bytes` override only their matching engine defaults.

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/deps/fetch_managed_runtimes.ps1 -RuntimeRoot D:\runtime\luaskills -Target all
python scripts/debug-tools/managed_runtime_layout_check.py D:\runtime\luaskills

# Split host-managed roots
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/deps/fetch_managed_runtimes.ps1 -RuntimeRoot D:\VulcanCodeData\luaskills -DistributionRoot D:\VulcanCode\dependencies\runtimes -Target all
python scripts/debug-tools/managed_runtime_layout_check.py D:\VulcanCodeData\luaskills --distribution-root D:\VulcanCode\dependencies\runtimes --environment-root D:\VulcanCodeData\managed-runtime-envs
```

```bash
RUNTIME_ROOT=/opt/luaskills scripts/deps/fetch_managed_runtimes.sh all
python3 scripts/debug-tools/managed_runtime_layout_check.py /opt/luaskills
```

By default, the shared manifest keeps LuaSkills core aligned with the SDK release and resolves runtime packages from the compatible `0.1` series by selecting the newest published patch automatically.

## Version Alignment

- Keep the SDK and LuaSkills core on the same current release line whenever possible.
- The current SDK defaults to LuaSkills core tag `v0.5.7`.
- Runtime packages and native dependencies still come from the split `LuaSkills/luaskills-packages` and related release assets.
- SDK default host options pass `runtime_root`, null managed-root override slots, and the complete stable `managed_runtime_config`; LuaSkills derives the fixed data layout until the host explicitly overrides roots or policy.
- Host tools live directly under `runtime_root/bin`, not `runtime_root/bin/tools`.

```powershell
npx @luaskills/sdk install-runtime --database none --runtime-root D:\runtime\luaskills
npx @luaskills/sdk install-runtime --database none --managed-runtimes all --runtime-root D:\runtime\luaskills
```

```powershell
pip install luaskills-sdk
luaskills install-runtime --database vldb-direct --runtime-root D:\runtime\luaskills
```

Go hosts can inspect the same asset plan:

```go
manifest, err := luaskills.BuildRuntimeInstallManifest(luaskills.RuntimeInstallOptions{
	RuntimeRoot:      "D:/runtime/luaskills",
	Database:         luaskills.RuntimeDatabaseVldbDirect,
	SkipLuaRuntime:   false,
	ManagedRuntimes: luaskills.ManagedRuntimeAll,
})
if err != nil {
	panic(err)
}

hostOptions, err := luaskills.HostOptionsFromRuntimeManifest(manifest)
if err != nil {
	panic(err)
}
```

`DefaultHostOptions(runtimeRoot)` returns host options plus an error. `DefaultHostOptions` and `NewClient` automatically read `runtimeRoot/resources/luaskills-sdk-runtime-manifest.json` and merge `host_options_patch` when the manifest exists. A missing manifest keeps SDK base defaults; a malformed manifest or host path that escapes `runtimeRoot` fails with a path-aware error.

Database modes:

- `RuntimeDatabaseNone`: installs the Lua runtime archive and the LuaSkills FFI SDK archive, without database providers.
- `RuntimeDatabaseVldbController`: uses the `vldb-controller` executable through `space_controller` mode.
- `RuntimeDatabaseVldbDirect`: uses `vldb-sqlite-lib` and `vldb-lancedb-lib` dynamic libraries.
- `RuntimeDatabaseHostCallback`: expects the host to provide JSON callbacks.

## Basic Usage

Prepare `runtimeRoot`, then create a client:

```go
package main

import (
	"fmt"

	luaskills "github.com/LuaSkills/luaskills-sdk-go"
)

func main() {
	runtimeRoot := "D:/runtime/luaskills"
	roots := luaskills.StandardRoots(runtimeRoot)

	client, err := luaskills.NewClient(luaskills.ClientOptions{
		RuntimeRoot:         runtimeRoot,
		EnsureRuntimeLayout: true,
	})
	if err != nil {
		panic(err)
	}
	defer client.Close()

	if _, err := client.LoadFromRoots(roots); err != nil {
		panic(err)
	}

	entries, err := client.ListEntries(luaskills.AuthorityDelegatedTool)
	if err != nil {
		panic(err)
	}

	result, err := client.CallSkill("demo-standard-ffi-skill-ping", map[string]any{
		"note": "go-sdk",
	}, nil)
	if err != nil {
		panic(err)
	}

	fmt.Println(entries)
	fmt.Println(result.Content)
}
```

## Examples

Detailed source-tree examples live under `examples/`.

```powershell
go run .\examples\basic
go run .\examples\call
go run .\examples\query
go run .\examples\lifecycle
go run .\examples\runtime_lease
go run .\examples\provider_callback
```

`provider_callback` covers both JSON provider callbacks and the `vulcan.host.*` host-tool callback boundary. It returns bridge-required errors until a host-owned cgo callback bridge is installed.

The query, lifecycle, and persistent runtime-lease examples use the bundled fixture skill at `examples/fixture-runtime/user_skills/demo-standard-ffi-skill`. Prepare runtime assets with a TypeScript or Python installer first:

```powershell
npx @luaskills/sdk install-runtime --database none --runtime-root .\examples\fixture-runtime
```

See [examples/README.md](examples/README.md) for the full example index and runtime notes. The Chinese example guide is [examples/README_cn.md](examples/README_cn.md).

## Persistent Runtime Leases

Use `client.RuntimeLeases()` for the public lease endpoints, or `client.System(authority).RuntimeLeases()` when the host wants fixed authority injection through the dedicated system runtime-lease exports provided by the latest native library.

```go
client, err := luaskills.NewClient(luaskills.ClientOptions{RuntimeRoot: "D:/runtime/luaskills"})
if err != nil {
	log.Fatal(err)
}
defer client.Close()

leases := client.System(luaskills.AuthoritySystem).RuntimeLeases()
cwd := "D:/runtime/luaskills/system_lua_lib"
ttlSec := 600
session, err := leases.CreateHandleWithOptions("demo-session", true, &luaskills.RuntimeLeaseCreateOptions{
	TTLSec: &ttlSec,
	CWD:    &cwd,
	Mounts: map[string]any{"channel": "demo"},
	SystemPackage: &luaskills.SystemRuntimePackage{ID: "debug-plugin", Root: "D:/runtime/luaskills/system_lua_lib/debug-plugin", DependenciesFile: "dependencies.json"},
})
if err != nil {
	log.Fatal(err)
}

result, err := session.Eval("counter = (counter or 0) + 1; return { counter = counter }", nil, 60000)
if err != nil {
	log.Fatal(err)
}

fmt.Println(result["result"])
```

## Migration Notes

- Existing `client.System(authority)` lifecycle calls keep working; the returned wrapper now also exposes query helpers and `RuntimeLeases()`.
- `RuntimeLeaseHandle` persists `lease_id + sid + generation` and automatically reattaches identity guards on `Eval`, `Status`, and `Close`.
- Authority-bound runtime-lease helpers dispatch directly to dedicated `luaskills_ffi_system_runtime_lease_*` entrypoints.
- `CallSkill` now returns optional `HostResult` when the host enables `request_context.client_capabilities.host_result` and one Lua tool emits a fourth structured return value.
- When `HostResult.Kind == "change_set"`, hosts should decode `HostResult.Payload` into `RuntimeChangeSetPayload`.
- Canonical `change_set` payloads now use file lifecycle records plus hunk-level `before + delete[] + insert[] + after` blocks for `modify` changes.
- `create` and `delete` file records carry full-file `content`, while `rename` records carry `old_path` and `new_path`.
- Public leases accept `cwd`, `workspace_root`, `lua_roots`, `c_roots`, and `mounts`. System leases require `SystemPackage`, reject `lua_roots/c_roots`, and derive roots from the trusted package manifest.
- `PollManagedSessionEvents`, `WaitManagedSessionEvents`, and `SetManagedSessionWakeCallback` expose the 0.5.1 managed-session event surface.
- Go hosts should deploy the matching latest LuaSkills native library when using these APIs, because cgo links directly against the current exported symbol set.

## Authority And Management

Query APIs should use `AuthorityDelegatedTool` by default, so ROOT skills are hidden from delegated tools.

`AuthoritySystem` only means the host may manage ROOT. It does not bypass ROOT ownership or same-`skill_id` conflict rules.

`CallSkill` and `RunLua` are runtime execution surfaces. They are not ROOT visibility filters.

## Skill Package Config

Configuration belongs to the effective skill package identified by `skill_id`, not to an individual entry inside that package. Packages declare their accepted keys, types, descriptions, defaults, and constraints in `skill.yaml`; `Set` rejects undeclared keys and values that do not satisfy the declaration.

```go
schema, err := client.Config.Describe(luaskills.SkillPackageConfigDescribeOptions{
    SkillID: "example.settings",
})
if err != nil {
    log.Fatal(err)
}

status, err := client.Config.Validate("example.settings")
if err != nil {
    log.Fatal(err)
}
fmt.Println(schema, status.Complete)

write, err := client.Config.SetValues(
    "example.settings",
    map[string]any{"api_key": "value", "retry_count": 3},
    "",
)
if err != nil {
    log.Fatal(err)
}
_, err = client.Config.Set("example.settings", "retry_count", 4, write.Revision)
```

Set `skill_config_root` in `ClientOptions.HostOptions` to an absolute user-level directory. LuaSkills stores ordinary and ROOT-owned package configuration separately under `skills/config.json` and `system-skills/config.json`. Every raw `List` entry includes `StoreScope`, so retained records with the same package id remain unambiguous across both files. Strict versioned documents use decimal-string revisions, cross-process companion locks, atomic replacement, cached snapshots, and file-watch reloads. Old unversioned documents are rejected.

`Describe` exposes parameter names, stable types (`integer`, `string`, `float`, `enum`, and `boolean`), constraints, UI hints, package-authored enum metadata, effective-value state, and completeness. `DescribeInstalled` discovers every physical package without executing Lua. Package authors choose one language for human-readable fields; English is recommended but not enforced.

Values are excluded by default. `IncludeValues: true` returns unmasked effective values. The host must allow, deny, replace, or ask the user to authorize disclosure and mutations; LuaSkills and this SDK intentionally do not implement that policy. Lua code can modify only its own package configuration, while host-level SDK calls are intentionally unrestricted.

`SetValues` is the canonical atomic batch operation; `Set` wraps one key into that transaction. The revision argument enables compare-and-swap writes and deletes. `PollEvents`, `WaitEvents`, and `WatchEvents` expose ordered local-write and external-reload events. A missing configuration should be handled by showing `Describe` output and asking the user or an authorized AI tool for declared parameters. Configuration survives package uninstall; explicit cleanup belongs to the host.

## JSON Provider Callback

The Go SDK exposes the callback API boundary, but does not install a process-level cgo callback bridge by default.

```go
err := luaskills.SetSQLiteProviderJSONCallback(func(request any) (any, error) {
	return map[string]any{"ok": true, "request": request}, nil
})
```

Currently this returns `ErrProviderCallbacksRequireHostBridge`. Production Go hosts that need `host_callback + json` should implement a controlled cgo callback bridge in the host process, or use the TypeScript / Python SDK for JSON callbacks.

## Host Tool Callback

`vulcan.host.*` uses the fixed host-tool callback registered through `luaskills_ffi_set_host_tool_json_callback`. The Go SDK exposes the typed request shape and registration boundary:

```go
// Register the host-tool callback boundary in hosts that provide a cgo bridge.
// 在提供 cgo 桥的宿主中注册宿主工具 callback 边界。
err := luaskills.SetHostToolJSONCallback(func(request luaskills.HostToolJSONRequest) (any, error) {
	return map[string]any{"ok": true, "value": request.Args}, nil
})
```

Currently this returns `ErrHostToolCallbacksRequireHostBridge`. A production Go host that wants Lua skills to call host tools should implement the controlled cgo bridge in the host process. The callback request contains `action`, `tool_name`, and `args`; `list` returns metadata, `has` returns availability, and `call` returns one complete table-shaped result without streaming.

## Model Callback

`vulcan.models.*` uses fixed callbacks registered through `luaskills_ffi_set_model_embed_json_callback` and `luaskills_ffi_set_model_llm_json_callback`. The Go SDK exposes typed request, response, and error shapes, but still leaves the process-level cgo callback bridge to the host:

Use these types to keep the production bridge contract explicit:

- `ModelEmbedJSONRequest`: receives `Text` and `Caller`.
- `ModelLLMJSONRequest`: receives `System`, `User`, and `Caller`.
- `ModelEmbedJSONResponse`: returns `Vector`, `Dimensions`, and optional `Usage`.
- `ModelLLMJSONResponse`: returns `Assistant` and optional `Usage`.
- `ModelJSONErrorEnvelope`: preserves model errors and optional provider details.

```go
// Register the model callback boundary in hosts that provide a cgo bridge.
// 在提供 cgo 桥的宿主中注册模型 callback 边界。
err := luaskills.SetModelEmbedJSONCallback(func(request luaskills.ModelEmbedJSONRequest) (any, error) {
	return luaskills.ModelEmbedJSONResponse{
		Vector:     []float32{0.1, 0.2, 0.3},
		Dimensions: 3,
	}, nil
})
```

Provider failures should be returned as structured envelopes when callers need diagnostics:

```go
func ptr[T any](value T) *T {
	return &value
}

failure := luaskills.ModelJSONErrorEnvelope{
	OK: false,
	Error: luaskills.ModelJSONError{
		Code:            luaskills.ModelJSONErrorProviderError,
		Message:         "model provider rejected the request",
		ProviderMessage: ptr("raw provider message after host-side redaction"),
		ProviderCode:    ptr("model_not_found"),
		ProviderStatus:  ptr(uint16(404)),
	},
}
```

Currently `SetModelEmbedJSONCallback` and `SetModelLLMJSONCallback` return `ErrModelCallbacksRequireHostBridge`. A production Go host should implement the controlled cgo callback bridge in its own process, forward `{ text, caller }` for embeddings and `{ system, user, caller }` for LLM calls, and return either a bare success payload or `ModelJSONErrorEnvelope`. Lua does not receive or override model configuration.

Go host checklist:

- Keep provider settings in host configuration, not in Lua skill config.
- Redact API keys, Authorization headers, signatures, and request headers before filling provider error fields.
- Use `Caller` for cost attribution, rate limits, audit logs, and per-skill policy.
- Treat thrown Go errors from the bridge as internal bridge failures; use `ModelJSONErrorEnvelope` for provider failures that should reach Lua.

## Verification

Source-tree checks:

```powershell
$env:CGO_ENABLED = "0"
go test ./...
```

Full native FFI checks need `CGO_ENABLED=1` and a cgo-compatible compiler. On Windows, Visual Studio alone is usually not enough for Go cgo; install a MinGW-w64/UCRT64 toolchain or another Go-compatible GCC distribution.

## Publishing

The release version is stored in `VERSION`. Go users consume SDK versions through Go module tags such as `v0.5.7`.

For one unified ecosystem release, publish `LuaSkills/luaskills-packages` first, then publish `LuaSkills/luaskills`, and publish the TypeScript SDK before the Go examples release flow because the Go examples workflow installs runtime assets through the published TypeScript package.

Before publishing:

```powershell
$env:CGO_ENABLED = "0"
go test ./...
```

Publish the SDK by pushing the matching Go module tag:

```powershell
git tag v0.5.7
git push origin v0.5.7
```

After the Go module tag is available, run the GitHub Actions workflow **Examples Release** manually. It reads `VERSION`, verifies `github.com/LuaSkills/luaskills-sdk-go@v{VERSION}`, installs LuaSkills runtime assets through the published TypeScript installer, runs the Go examples, then creates or updates the `examples-v{VERSION}` GitHub Release with:

- `luaskills-sdk-go-examples-{VERSION}.zip`
- `luaskills-sdk-go-examples-{VERSION}.zip.sha256`

The examples release tag intentionally uses the `examples-v` prefix so it does not interfere with Go module semver tags.

Recommended unified publish order: `luaskills-packages` -> `luaskills` core release -> TypeScript SDK -> Python SDK -> Go SDK -> SDK examples releases.
