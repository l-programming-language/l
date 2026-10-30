    Create Runtime's Signals' "Interpret|-s|-ed|-ing|-er" as a Generic `runtime_signal_fault` Function with Inputs of "status" as `u8` and Outputs of `Option<Error>`.
Create "One" as a New Runtime's Byte of 1.
If Status is One, Create "Fault" as a New Runtime's Error's Describer of `application output pipe closed` and then Create "Wrapped" as a New Present of Fault and then Return Wrapped.
Return a New Absent.

    Create Runtime's Signals' "Watch|-es|-ed|-ing|-er" as a Generic Pending `runtime_signal_watch` Function with Inputs of "signals" as `crate::process::Signals` and "owner" as `std::sync::Weak<RuntimeInner>` and Outputs of `std::pin::Pin<Box<dyn std::future::Future<Output = ()> + Send>>`.
Create "Receiver" as a New Mutable Binding of Signals.
Create "Event" as a New Runtime's Completion of a New Runtime's Receiver of Receiver.
Create "Optional" as a New Runtime's Ascender of Owner.
Create "Available" as a New Presence of Optional.
Create "Unit" as a New Nothing.
If Available is false, Return Unit.
Create "Handle" as a New Runtime's Handle's Content of a New Contents of Optional and `present signal runtime`.
Create "Reference" as a New Storage's Borrower of Handle.
Create "Failed" as a New Runtime's Fault of Event.
If Failed, Create "Cause" as a New Runtime's Rejecter of Event and then Create "Fault" as a New Runtime's Error's Explainer of `receive application signal` and Cause and then Create "Wrapped" as a New Present of Fault and then Create "One" as a New Runtime's Byte of 1 and then Create "Requested" as a New Runtime's Requester of Reference, One, and Wrapped and then Return Unit.
Create "Status" as a New Contents of Event and `received application signal`.
Create "Error" as a New Runtime's Signals' Interpreter of Status.
Create "Request" as a New Runtime's Signal of Status.
Create "Dependencies" as a New Runtime's Authority of a New Runtime's Reference of Handle.
Create "Entered" as a New Runtime's Enterer of Dependencies and Request.
Create "Selected" as a New Runtime's Requester of Reference, Status, and Error.
Return Unit.

    Create Runtime's "Attach|-es|-ed|-ing|-er" as a Generic `runtime_attach` Function with Inputs of "value" as `&Runtime` and "task" as `tokio::task::JoinHandle<()>` and Outputs of `()`.
Create "Locked" as a New Locker of a New Runtime's Task of a New Runtime's Reference of Value.
Create "Reference" as a New Storage.Editable Storage's Reference of a New Mutable Binding of Locked.
Create "Retained" as a New Present of Task.
Runtime.Replace Reference and Retained.
Return a New Nothing.

    Publicly create Runtime's "Install|-s|-ed|-ing|-er" as a Generic `runtime_observed` Function with Inputs of "dependencies" as `std::sync::Arc<dyn crate::runtime::host::Dependencies>` and "output" as `crate::runtime::host::Output` and "error" as `crate::runtime::host::Output` and Outputs of `Result<Runtime>`.
Create "Signals" as a New Runtime's Signals.
Create "Receiver" as a New Storage.Checked Storage's Binding of a New Runtime's Explainer of Signals and Runtime's Signals' Explainer.
Create "Handle" as a New Runtime's Assembler of Dependencies, Output, and Error.
Create "Weak" as a New Runtime's Weak of a New Runtime's Reference of Handle.
Create "Spawned" as a New Runtime's Spawner of a New Runtime's Signals' Watcher of Receiver and Weak.
Create "Borrowed" as a New Storage's Borrower of Handle.
Create "Attached" as a New Runtime's Attacher of Borrowed and Spawned.
Return a New Success of Handle.

    Publicly create Runtime's "Open|-s|-ed|-ing|-er" as a Generic `runtime_new` Function with No Inputs and Outputs of `Result<Runtime>`.
Create "Dependencies" as a New Runtime's Dependencies of a New Runtime's Native.
Create "Output" as a New Runtime.Output Runtime's Pinner of a New Runtime's Stdout.
Create "Error" as a New Runtime.Output Runtime's Pinner of a New Runtime's Stderr.
Return a New Runtime's Installer of Dependencies, Output, and Error.
