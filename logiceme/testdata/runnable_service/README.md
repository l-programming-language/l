# Raw library and native-consumer composition controls

These independently authored inputs and exact source/manifest expectations
establish that the existing frozen A/FW/W/BOD/E logemes can express the
raw AML library and explicit binary-consumer boundary. These controls retain
their historical fixture and test names; they do not grant standalone L Services
a process adapter, library package or runnable entry.

`library.aml.txt` declares a sole `Worker` library with its public `main` entry
in `src/main.rs`. The entry writes exactly `work completed\n`. The separate
`l_service_worker` adapter calls `service::main()` once through an explicit path
dependency. The Service receives no runtime, executable configuration, primary
binary, or start/stop artifacts. The adapter contains the ordinary executable
support and no supervisor. These tests do not exercise an orchestrator.

`TestRunnableServiceComposition` compares all seven returned artifacts, ordered
paths, dependency metadata, and nil distinctions byte-for-byte. It checks both
embedded and filesystem resources, input immutability, A/B/A service reuse, and
previous bundle stability. `TestRunnableServiceLibraryBoundaries` retains the
rejections for runtime, executable settings, and a Primary library target.

`TestRunnableServiceCompositionNative` separately compiles Expected and Generated
workspaces with frozen resolution and the retained local-only lock. It verifies
that only the adapter has a native executable and that Cargo leaves every input
byte unchanged. It runs neither binaries nor generated deployment scripts.
Compilation success is not App lifecycle or runtime evidence.

Private-entry and private-ancestor-module controls compare exact source without
automatically exporting either declaration. Their independent Expected and
Generated adapter builds must both report Rust E0603 at the adapter call site;
unrelated Cargo or compiler failures do not satisfy the assertion. These are
backend visibility controls, not claims of new L private-entry syntax.

The expectations were authored before evaluating these inputs. Adapter config
and stop-script expectations reuse the unchanged independent fixtures under
`../lucid/common/`; all other bytes are stored here. The existing eighteen
segment freeze assertions remain unchanged.
