# Standalone Service retirement

The finite amendment approved October 14, 2026 is implemented. The
[October 15 macOS ARM64 validation](../../testdata/service-execution/validation-2026-10-15.md)
passed canonical acceptance, all 4,383 named outcomes, source generation,
runtime, full lint and tagged native compilation. This documentation update did
not rerun language validation.

## Current boundary

Only Apps own native process identity. Services retain their structural role in
App source construction.

| Construct | Retained behavior |
| --- | --- |
| App | Executable target with Build/Run/Stop/Await, captured Output/Error and once-only continuations |
| App-constructed Service | Source chapter; the primary Service identifies the App binary and additional Services may produce separate library packages |
| Standalone Service declaration | Structural semantic record with no automatic package, executable or continuation |
| Class / Object | Recursive source organization / source file; no new runtime scope |
| Raw AML library target | Supported backend source bundle, including independently authored native consumers |

Main and Basic keep their existing composition. AML `Services`, `CHAPTER` and
`StructureVersion: "2"` remain supported. Publication manifests contain App
companions and use version 1; existing generations are retained.

Standalone Service bodies reject with contextual unsupported-body diagnostics.
Native lifecycle operations and implicit runtime Output/Error require an actual
App after resolving canonical values, aliases and source-bound pronouns.
Ordinary source-defined actions and members named Run, Output or Error keep
lexical precedence. Service declarations and aliases cannot acquire continuations
or take ownership from a preceding App. Structural-only programs still receive
the normal top-level orchestrator.

## Finite amendment and retained observers

[service-executable-amendment.json](service-executable-amendment.json) records the
approved preimage hashes, action and retained coverage for 120 protected paths:
53 modifications and 67 deletions, plus 31 new App expectation files. The four
resource-retirement changes are also listed. The five typed process definitions
remain selected from the same resource filesystem and instantiated through
private compiler operands; the Service-only execution-role resource is removed.

The resulting inventory has 386 protected files; the other 302 records, all
sixteen original required outcomes and every required-absent entry are preserved.
Frozen glyphs, retained locks, native assets, tool pins, runtime watchdogs, budgets
and concurrency remain intact. The approval grants no standing permission to
change protected inputs.

The fourteen CLI scenario paths retain literal output and lifecycle obligations
on explicit Apps. The imported case keeps its provider-relative `worker.l`
lookup. Stop-without-joining retains its late-output pipe observer. Publication,
relocation, concurrent rebuilding and running-generation-survival observers keep
their protocols and cumulative runtime allowances.

Stored App expectations use independent App-control transformations and retained
manifest/configuration oracles. Expected-source and generated-source compilation
remain separate. App-owned workspace and raw backend library consumers preserve
native-linking coverage. Standalone bodies remain negative witnesses;
declaration-only cases establish no package emission.

## Assertion migration

The following maps removed test roots to retained observers or retired contracts.
Names in the first column omit the `TestRunnableService` prefix unless qualified.
Assertions about standalone emission, entry eligibility and adapter identity
retire with those features; unrelated behavior remains observable.

| Former root | Retained observer or retired feature |
| --- | --- |
| Construction | `TestServiceDeclarationsAreStructural`; fresh App/library projections remain in `TestWorkspaceConstructedServices` and `TestAppProcessArtifactsIgnoreStructuralServices` |
| LibraryRust; BodiesAndImports | Both original body fixtures and imported attribution remain in `TestStandaloneServiceBodiesAreUnsupported`; positive App body composition remains in `TestCheckedApplicationBodyUsesCanonicalAliasOwner`, `TestPhase1Examples` and `TestWorkspaceConstructedServices`; standalone Rust emission retires |
| LifecycleInstructions; MixedTargetOrder | Service rejection in `TestServiceLifecycleAndImplicitMembersReject`; App forms, ordering and typed effects in `TestPhase1NativeLifecycleCoordination`, `TestPhase1PassiveLifecycleTargets` and `TestBuiltinProcessLogemeAppParity` |
| RuntimeMembers | Service rejection in `TestServiceLifecycleAndImplicitMembersReject`; ordinary lexical names in `TestServiceOrdinaryMembersAndActionsRemainLexical`; App provenance in `TestNativeMemberOwnersResolveFromSource` |
| Continuations | `TestRunnableAppContinuations` retains distinct activation and storage assertions |
| UnsupportedTargetsAndOwners | `TestServiceUnsupportedTargetsAndOwners` retains original negative inputs, with the new standalone-body diagnostic |
| SourceActionShadowing | `TestServiceSourceActionAliasAndParameter` retains alias and parameter assertions; `TestServiceOrdinaryMembersAndActionsRemainLexical` covers all four lifecycle names |
| EntrypointIdentity; EntrypointValidation | Standalone entry eligibility, spoofing and entry-signature contracts retire; ordinary descriptor/parameter contracts remain in `TestConstructionDescriptorAliasesPreserveExpansion`, `TestConstructionParametersRequireProvidedNominalBody`, `TestConstructionFunctionSlotsRequireDeclaredParameters` and `TestBuiltinProcessLogemeCatalog` |
| ForwardConstructionOrder | Declaration line provenance in `TestServiceForwardDeclarationProvenance`; no-package inventory in `TestServiceDeclarationsAreStructural`; standalone package order retires |
| ImportedContinuation; MixedNestedAwait | `TestRetainedAppImportedContinuation` and `TestRetainedAppNestedAwaitWithStructuralService` preserve ownership and storage |
| DefiniteAssignment; SnapshotMutability; TypedCapture | Corresponding `TestRetainedAppDefiniteAssignment`, `TestRetainedAppSnapshotMutability` and `TestRetainedAppTypedCapture` |
| ProcessArtifacts; AdapterLocations | App identity/path validation, detached output and fresh buffers in `TestAppProcessArtifactsIgnoreStructuralServices`; App-owned dependency paths in `TestWorkspaceConstructedServices`; adapter locations retire |
| PackagesAvoidPrivateAppBuildNames; PackagesAvoidCasefoldAppNames; PackagesAvoidAppOwnedLibraries | Corresponding `TestStructuralServicesPreservePrivateAppBuildNames`, `TestStructuralServicesPreserveCasefoldAppNames` and `TestStructuralServicesPreserveAppOwnedLibraries` |
| WorkspacePreservesCompletedAppGraphAPI | `TestWorkspacePreservesCompletedAppGraphAPI` |
| NonrunnableParagraphOwnership | `TestStructuralServiceParagraphOwnership` retains traces and adds Main Services, aliases and later Apps |
| `TestServiceAdapterProjectionIsIsolated`; `TestServiceAdapterRejectsPackageAndLocationCollisions` | Adapter-specific APIs retire; structural no-artifact behavior and App artifact isolation/validation remain in the observers above |
| `TestBuiltinProcessLogemeEntryRoleSchema` and catalog entry-role subcases | Role JSON schema/selection retire; selected typed definitions and composition validation remain in `TestBuiltinProcessLogemeCatalog` and `TestBuiltinProcessLogemeCompositionOrder` |
| CLI ExactLibrarySource; LibraryOnly | `TestStandaloneServiceProcessSource` and `TestStandaloneServiceBodyUnsupported`; native library consumption moves to the unchanged App-owned fixture through `TestRunnableServiceLibraryConsumer` |
| CLI ExpectedCorpusShapeControls; ExpectedOrchestratorShapeControl | Independent conversions are materialized as stored App expectations and compared by `TestRunnableServiceExactRustWorkspace` and `TestRunnableServiceBuildAndRuntime`; separate expected compilation remains |
| `TestServiceAdapterSelectionKeepsIndependentLibraryIndices` | Adapter/library index selection retires; `TestStandaloneServiceDoesNotAcquireBuildArtifacts` checks the complete real Cargo inventory and App execution |

CLI lifecycle, invocation-result, Stop-without-joining, relocation, concurrent
publication and generation-survival roots retain their names and protocols on
explicit Apps. Raw backend AML library, visibility and native-consumer observers
are unchanged.
