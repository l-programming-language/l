Use Storage from `storage`.
Use Runtime from `runtime`.
Create "GlossaryExamples" as a Dictionary.

    Publicly create GlossaryExamples's "StorageAppRender" as a Generic Function with Inputs of "data" as `&dyn std::any::Any` and Outputs of `runtime::Result<Vec<u8>>`.
Return a New Storage's App's Renderer of Data.

    Publicly create GlossaryExamples's "StorageBytesRead" as a Generic Function with Inputs of "bytes" as `&[u8]` and "index" as `f64` and Outputs of `runtime::Result<f64>`.
Return a New Storage's Bytes's Reader of Bytes and Index.

    Publicly create GlossaryExamples's "StorageBytesRender" as a Generic Function with Inputs of "data" as `&dyn std::any::Any` and Outputs of `runtime::Result<Vec<u8>>`.
Return a New Storage's Bytes's Renderer of Data.

    Publicly create GlossaryExamples's "StorageClone" as a Generic Storage.Cloneable Storage.Routine with Inputs of "value" as `&T` and Outputs of `T`.
Return a New Storage's Cloner of Value.

    Publicly create GlossaryExamples's "StorageEnvironmentCreate" as a Generic Function with Inputs of "types" as `Vec<Vec<u64>>` and "values" as `Vec<LValue>` and Outputs of `runtime::Result<LEnvironment>`.
Return a New Storage's Environment's Creator of Types and Values.

    Publicly create GlossaryExamples's "StorageEnvironmentRead" as a Generic Function with Inputs of "environment" as `&LEnvironment` and "index" as `usize` and Outputs of `runtime::Result<LValue>`.
Return a New Storage's Environment's Reader of Environment and Index.

    Publicly create GlossaryExamples's "StorageEnvironmentRender" as a Generic Function with Inputs of "data" as `&dyn std::any::Any` and Outputs of `runtime::Result<Vec<u8>>`.
Return a New Storage's Environment's Renderer of Data.

    Publicly create GlossaryExamples's "StorageEnvironmentRequire" as a Generic Function with Inputs of "value" as `&Option<LEnvironment>` and "message" as `&str` and Outputs of `runtime::Result<LEnvironment>`.
Return a New Storage's Environment's Requirer of Value and Message.

    Publicly create GlossaryExamples's "StorageEnvironmentUpdate" as a Generic Function with Inputs of "environment" as `&LEnvironment` and "index" as `usize` and "value" as `LValue` and Outputs of `runtime::Result<()>`.
Return a New Storage's Environment's Updater of Environment and Index and Value.

    Publicly create GlossaryExamples's "StorageEnvironmentVisit" as a Generic Function with Inputs of "data" as `&dyn std::any::Any` and "target" as `usize` and "seen" as `&mut std::collections::BTreeSet<usize>` and Outputs of `runtime::Result<bool>`.
Return a New Storage's Environment's Visitor of Data and Target and Seen.

    Publicly create GlossaryExamples's "StorageErrorRender" as a Generic Function with Inputs of "data" as `&dyn std::any::Any` and Outputs of `runtime::Result<Vec<u8>>`.
Return a New Storage's Error's Renderer of Data.

    Publicly create GlossaryExamples's "StorageFunctionCopy" as a Generic Storage.Copyable Storage.Routine with Inputs of "value" as `&(LEnvironment, T)` and Outputs of `(LEnvironment, T)`.
Return a New Storage's Function's Copier of Value.

    Publicly create GlossaryExamples's "StorageFunctionReach" as a Generic Storage.Permanent Storage.Routine with Inputs of "data" as `&dyn std::any::Any` and "target" as `usize` and "seen" as `&mut std::collections::BTreeSet<usize>` and Outputs of `runtime::Result<bool>`.
Return a New Storage's Function's Reacher of Data and Target and Seen.

    Publicly create GlossaryExamples's "StorageFunctionRender" as a Generic Function with Inputs of "data" as `&dyn std::any::Any` and Outputs of `runtime::Result<Vec<u8>>`.
Return a New Storage's Function's Renderer of Data.

    Publicly create GlossaryExamples's "StorageIgnore" as a Generic Function with Inputs of "value" as `&dyn std::any::Any` and "target" as `usize` and "seen" as `&mut std::collections::BTreeSet<usize>` and Outputs of `runtime::Result<bool>`.
Return a New Storage's Ignorer of Value and Target and Seen.

    Publicly create GlossaryExamples's "StorageListAppend" as a Generic Function with Inputs of "list" as `&mut LList` and "value" as `LValue` and Outputs of `runtime::Result<()>`.
Return a New Storage's List's Appender of List and Value.

    Publicly create GlossaryExamples's "StorageListCopy" as a Generic Function with Inputs of "value" as `&LList` and Outputs of Storage's List.
Return a New Storage's List's Copier of Value.

    Publicly create GlossaryExamples's "StorageListCreate" as a Generic Function with Inputs of "element" as `Vec<u64>` and "values" as `Vec<LValue>` and Outputs of `runtime::Result<LList>`.
Return a New Storage's List's Creator of Element and Values.

    Publicly create GlossaryExamples's "StorageListReach" as a Generic Function with Inputs of "data" as `&dyn std::any::Any` and "target" as `usize` and "seen" as `&mut std::collections::BTreeSet<usize>` and Outputs of `runtime::Result<bool>`.
Return a New Storage's List's Reacher of Data and Target and Seen.

    Publicly create GlossaryExamples's "StorageListRead" as a Generic Function with Inputs of "list" as `&LList` and "index" as `f64` and Outputs of `runtime::Result<LValue>`.
Return a New Storage's List's Reader of List and Index.

    Publicly create GlossaryExamples's "StorageListRender" as a Generic Function with Inputs of "data" as `&dyn std::any::Any` and Outputs of `runtime::Result<Vec<u8>>`.
Return a New Storage's List's Renderer of Data.

    Publicly create GlossaryExamples's "StorageListUpdate" as a Generic Function with Inputs of "list" as `&mut LList` and "index" as `f64` and "value" as `LValue` and Outputs of `runtime::Result<()>`.
Return a New Storage's List's Updater of List and Index and Value.

    Publicly create GlossaryExamples's "StorageMapCopy" as a Generic Function with Inputs of "value" as `&LMap` and Outputs of Storage's Map.
Return a New Storage's Map's Copier of Value.

    Publicly create GlossaryExamples's "StorageMapCreate" as a Generic Function with Inputs of "element" as `Vec<u64>` and Outputs of Storage's Map.
Return a New Storage's Map's Creator of Element.

    Publicly create GlossaryExamples's "StorageMapReach" as a Generic Function with Inputs of "data" as `&dyn std::any::Any` and "target" as `usize` and "seen" as `&mut std::collections::BTreeSet<usize>` and Outputs of `runtime::Result<bool>`.
Return a New Storage's Map's Reacher of Data and Target and Seen.

    Publicly create GlossaryExamples's "StorageMapRead" as a Generic Function with Inputs of "map" as `&LMap` and "key" as `&[u8]` and Outputs of `runtime::Result<LValue>`.
Return a New Storage's Map's Reader of Map and Key.

    Publicly create GlossaryExamples's "StorageMapRender" as a Generic Function with Inputs of "data" as `&dyn std::any::Any` and Outputs of `runtime::Result<Vec<u8>>`.
Return a New Storage's Map's Renderer of Data.

    Publicly create GlossaryExamples's "StorageMapUpdate" as a Generic Function with Inputs of "map" as `&mut LMap` and "key" as `Vec<u8>` and "value" as `LValue` and Outputs of `runtime::Result<()>`.
Return a New Storage's Map's Updater of Map and Key and Value.

    Publicly create GlossaryExamples's "StorageNullRender" as a Generic Function with Inputs of "data" as `&dyn std::any::Any` and Outputs of `runtime::Result<Vec<u8>>`.
Return a New Storage's Null's Renderer of Data.

    Publicly create GlossaryExamples's "StorageNumberRender" as a Generic Function with Inputs of "data" as `&dyn std::any::Any` and Outputs of `runtime::Result<Vec<u8>>`.
Return a New Storage's Number's Renderer of Data.

    Publicly create GlossaryExamples's "StoragePointerCopy" as a Generic Function with Inputs of "value" as `&LPointer` and Outputs of Storage's Pointer.
Return a New Storage's Pointer's Copier of Value.

    Publicly create GlossaryExamples's "StoragePointerCreate" as a Generic Function with Inputs of "element" as `Vec<u64>` and "value" as `LValue` and Outputs of `runtime::Result<LPointer>`.
Return a New Storage's Pointer's Creator of Element and Value.

    Publicly create GlossaryExamples's "StoragePointerRender" as a Generic Function with Inputs of "data" as `&dyn std::any::Any` and Outputs of `runtime::Result<Vec<u8>>`.
Return a New Storage's Pointer's Renderer of Data.

    Publicly create GlossaryExamples's "StoragePointerUpdate" as a Generic Function with Inputs of "pointer" as `&LPointer` and "value" as `LValue` and Outputs of `runtime::Result<()>`.
Return a New Storage's Pointer's Updater of Pointer and Value.

    Publicly create GlossaryExamples's "StoragePointerVisit" as a Generic Function with Inputs of "data" as `&dyn std::any::Any` and "target" as `usize` and "seen" as `&mut std::collections::BTreeSet<usize>` and Outputs of `runtime::Result<bool>`.
Return a New Storage's Pointer's Visitor of Data and Target and Seen.

    Publicly create GlossaryExamples's "StorageProgramActionsRead" as a Generic Function with Inputs of "program" as `&LProgram` and "index" as `usize` and Outputs of `runtime::Result<LEnvironment>`.
Return a New Storage's Program's Actions's Reader of Program and Index.

    Publicly create GlossaryExamples's "StorageProgramActionsWrite" as a Generic Function with Inputs of "program" as `&LProgram` and "index" as `usize` and "value" as `LEnvironment` and Outputs of `runtime::Result<()>`.
Return a New Storage's Program's Actions's Writer of Program and Index and Value.

    Publicly create GlossaryExamples's "StorageProgramCreate" as a Generic Function with Inputs of "apps" as `Vec<std::sync::Arc<runtime::App>>` and "types" as `Vec<Vec<u64>>` and "action_count" as `usize` and Outputs of Storage's Program.
Return a New Storage's Program's Creator of Apps and Types and Action_count.

    Publicly create GlossaryExamples's "StorageProgramSelect" as a Generic Function with Inputs of "program" as `&LProgram` and "index" as `usize` and Outputs of `runtime::Result<std::sync::Arc<runtime::App>>`.
Return a New Storage's Program's Selector of Program and Index.

    Publicly create GlossaryExamples's "StorageProgramValuesRead" as a Generic Function with Inputs of "program" as `&LProgram` and "index" as `usize` and Outputs of `runtime::Result<LValue>`.
Return a New Storage's Program's Values's Reader of Program and Index.

    Publicly create GlossaryExamples's "StorageProgramValuesWrite" as a Generic Function with Inputs of "program" as `&LProgram` and "index" as `usize` and "value" as `LValue` and Outputs of `runtime::Result<()>`.
Return a New Storage's Program's Values's Writer of Program and Index and Value.

    Publicly create GlossaryExamples's "StorageTextRender" as a Generic Function with Inputs of "data" as `&dyn std::any::Any` and Outputs of `runtime::Result<Vec<u8>>`.
Return a New Storage's Text's Renderer of Data.

    Publicly create GlossaryExamples's "StorageTruthRender" as a Generic Function with Inputs of "data" as `&dyn std::any::Any` and Outputs of `runtime::Result<Vec<u8>>`.
Return a New Storage's Truth's Renderer of Data.

    Publicly create GlossaryExamples's "StorageValueExpect" as a Generic Function with Inputs of "value" as `LValue` and "expected" as `Vec<u64>` and Outputs of `runtime::Result<LValue>`.
Return a New Storage's Value's Expecter of Value and Expected.

    Publicly create GlossaryExamples's "StorageValueRestore" as a Generic Storage.Permanent Storage.Routine with Inputs of "value" as `LValue` and "message" as `&str` and "snapshot" as `fn(&T) -> T` and Outputs of `runtime::Result<T>`.
Return a New Storage's Value's Restorer of Value and Message and Snapshot.

    Publicly create GlossaryExamples's "StorageValueWrap" as a Generic Storage.Permanent Storage.Routine with Inputs of "value" as `T` and "signature" as `Vec<u64>` and "renderer" as `fn(&dyn std::any::Any) -> runtime::Result<Vec<u8>>` and "visitor" as `fn(&dyn std::any::Any, usize, &mut std::collections::BTreeSet<usize>) -> runtime::Result<bool>` and Outputs of Storage's Value.
Return a New Storage's Value's Wrapper of Value and Signature and Renderer and Visitor.

    Publicly create GlossaryExamples's "RuntimeAppBuild" as a Generic Function with Inputs of "value" as `&App` and Outputs of `Result<()>`.
Return a New Runtime's App's Builder of Value.

    Publicly create GlossaryExamples's "RuntimeAppInspect" as a Generic Function with Inputs of "value" as `&App` and Outputs of `Result<Option<Error>>`.
Return a New Runtime's App's Inspector of Value.

    Publicly create GlossaryExamples's "RuntimeAppRead" as a Generic Function with Inputs of "value" as `&App` and Outputs of `Result<std::sync::Arc<[u8]>>`.
Return a New Runtime's App's Reader of Value.

    Publicly create GlossaryExamples's "RuntimeAppRun" as a Generic Function with Inputs of "value" as `&std::sync::Arc<App>` and Outputs of `Result<std::sync::Arc<Invocation>>`.
Return a New Runtime's App's Runner of Value.

    Publicly create GlossaryExamples's "RuntimeAppSet" as a Generic Function with Inputs of "value" as `&App` and "source" as `String` and Outputs of `()`.
Return a New Runtime's App's Setter of Value and Source.

    Publicly create GlossaryExamples's "RuntimeAppStop" as a Generic Function with Inputs of "value" as `&App` and Outputs of `std::pin::Pin<Box<dyn std::future::Future<Output = Result<()>> + Send + '_>>`.
Return a New Runtime's App's Stopper of Value.

    Publicly create GlossaryExamples's "RuntimeAppWait" as a Generic Function with Inputs of "value" as `&App` and Outputs of `std::pin::Pin<Box<dyn std::future::Future<Output = Result<std::sync::Arc<Invocation>>> + Send + '_>>`.
Return a New Runtime's App's Waiter of Value.

    Publicly create GlossaryExamples's "RuntimeAppoint" as a Generic Function with Inputs of "value" as `&Runtime` and "path" as `std::path::PathBuf` and Outputs of `std::sync::Arc<App>`.
Return a New Runtime's Appointer of Value and Path.

    Publicly create GlossaryExamples's "RuntimeErrorDisplay" as a Generic Function with Inputs of "value" as `&Error` and "formatter" as `&mut std::fmt::Formatter<'_>` and Outputs of `std::fmt::Result`.
Return a New Runtime's Error's Displayer of Value and Formatter.

    Publicly create GlossaryExamples's "RuntimeErrorTrace" as a Generic Function with Inputs of "value" as `&Error` and Outputs of `Option<&(dyn std::error::Error + 'static)>`.
Return a New Runtime's Error's Tracer of Value.

    Publicly create GlossaryExamples's "RuntimeInvocationClaim" as a Generic Function with Inputs of "value" as `&Invocation` and Outputs of `Result<bool>`.
Return a New Runtime's Invocation's Claimer of Value.

    Publicly create GlossaryExamples's "RuntimeInvocationFinish" as a Generic Function with Inputs of "value" as `&std::sync::Arc<Invocation>` and Outputs of `()`.
Return a New Runtime's Invocation's Finisher of Value.

    Publicly create GlossaryExamples's "RuntimeLeave" as a Generic Function with Inputs of "value" as `&Runtime` and Outputs of Runtime's Error.
Return a New Runtime's Leaver of Value.

    Publicly create GlossaryExamples's "RuntimeOpen" as a Generic Function with No Inputs and Outputs of `Result<Runtime>`.
Return a New Runtime's Opener.

    Publicly create GlossaryExamples's "RuntimeRun" as a Generic Function with Inputs of "value" as `&Runtime` and "source" as `impl std::future::Future<Output = Result<()>>` and Outputs of `impl std::future::Future<Output = u8>`.
Return a New Runtime's Runner of Value and Source.

    Publicly create GlossaryExamples's "RuntimeShow" as a Generic Function with Inputs of "value" as `&Runtime` and "bytes" as `&[u8]` and Outputs of `impl std::future::Future<Output = Result<()>> + Send`.
Return a New Runtime's Shower of Value and Bytes.

    Publicly create GlossaryExamples's "RuntimeSnapshotCopy" as a Generic Function with Inputs of "value" as `&Snapshot` and Outputs of Runtime's Snapshot.
Return a New Runtime's Snapshot's Copier of Value.

    Publicly create GlossaryExamples's "RuntimeTerminalCopy" as a Generic Function with Inputs of "value" as `&Terminal` and Outputs of Runtime's Terminal.
Return a New Runtime's Terminal's Copier of Value.
