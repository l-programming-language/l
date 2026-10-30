*** Process-owned nouns. These layouts describe native ownership; their
    operations are declared separately as grammatical source Functions. ***
Publicly define Runtime's "Handle" as `<DEF>HY{|Key: ~Runtime~, Props: ~&[Prop{Key: inner, Type: std::sync::Arc<RuntimeInner>}Prop]&~|}<DEF>`.
Define Runtime's "Inner" as `<DEF>HY{|Key: ~RuntimeInner~, Props: ~&[Prop{Key: companion_directory, Type: Result<std::path::PathBuf>}Prop, Prop{Key: state, Type: std::sync::Mutex<RuntimeState>}Prop, Prop{Key: cancelled, Type: tokio::sync::watch::Sender<bool>}Prop, Prop{Key: shutdown_done, Type: tokio::sync::watch::Sender<Option<Result<()>>>}Prop, Prop{Key: signals, Type: std::sync::Mutex<Option<tokio::task::JoinHandle<()>>>}Prop, Prop{Key: stdout, Type: tokio::sync::Mutex<crate::runtime::host::Output>}Prop, Prop{Key: stderr, Type: tokio::sync::Mutex<crate::runtime::host::Output>}Prop, Prop{Key: dependencies, Type: std::sync::Arc<dyn crate::runtime::host::Dependencies>}Prop, Prop{Key: finish_gate, Type: tokio::sync::Mutex<()>}Prop]&~|}<DEF>`.
Define Runtime's "State" as `<DEF>HY{|Key: ~RuntimeState~, Props: ~&[Prop{Key: owned, Type: std::collections::BTreeMap<u64, std::sync::Arc<Invocation>>}Prop, Prop{Key: cleanup_errors, Type: std::collections::BTreeMap<u64, Error>}Prop, Prop{Key: next, Type: u64}Prop, Prop{Key: stopping, Type: bool}Prop, Prop{Key: terminal, Type: Option<Terminal>}Prop, Prop{Key: finalized, Type: bool}Prop, Prop{Key: final_status, Type: Option<u8>}Prop]&~|}<DEF>`.
Publicly define Runtime's "Terminal" as `<DEF>HY{|Key: ~Terminal~, Props: ~&[Prop{Key: status, Type: u8}Prop, Prop{Key: error, Type: Option<Error>}Prop]&~|}<DEF>`.
Publicly define Runtime's "App" as `<DEF>HY{|Key: ~App~, Props: ~&[Prop{Key: runtime, Type: std::sync::Weak<RuntimeInner>}Prop, Prop{Key: path, Type: std::path::PathBuf}Prop, Prop{Key: state, Type: std::sync::Mutex<AppState>}Prop]&~|}<DEF>`.
Define Runtime's App's "State" as `<DEF>HY{|Key: ~AppState~, Props: ~&[Prop{Key: source, Type: String}Prop, Prop{Key: invocation, Type: Option<std::sync::Arc<Invocation>>}Prop, Prop{Key: outstanding, Type: bool}Prop, Prop{Key: selected, Type: Option<std::sync::Arc<Invocation>>}Prop, Prop{Key: operation_error, Type: Option<Error>}Prop]&~|}<DEF>`.
Publicly define Runtime's "Snapshot" as `<DEF>HY{|Key: ~Snapshot~, Props: ~&[Prop{Key: pub stdout, Type: std::sync::Arc<[u8]>}Prop, Prop{Key: pub stderr, Type: std::sync::Arc<[u8]>}Prop, Prop{Key: pub error, Type: Option<Error>}Prop, Prop{Key: pub status, Type: Option<i32>}Prop]&~|}<DEF>`.
Publicly define Runtime's "Invocation" as `<DEF>HY{|Key: ~Invocation~, Props: ~&[Prop{Key: app, Type: std::sync::Weak<App>}Prop, Prop{Key: runtime, Type: std::sync::Weak<RuntimeInner>}Prop, Prop{Key: sequence, Type: u64}Prop, Prop{Key: dependencies, Type: std::sync::Arc<dyn crate::runtime::host::Dependencies>}Prop, Prop{Key: wait_delay, Type: tokio::time::Duration}Prop, Prop{Key: state, Type: std::sync::Mutex<InvocationState>}Prop, Prop{Key: done, Type: tokio::sync::watch::Sender<bool>}Prop, Prop{Key: group, Type: tokio::sync::Mutex<GroupState>}Prop, Prop{Key: drain_deadline, Type: tokio::sync::watch::Sender<Option<tokio::time::Instant>>}Prop, Prop{Key: leader_requests, Type: tokio::sync::mpsc::UnboundedSender<tokio::sync::oneshot::Sender<Result<()>>>}Prop]&~|}<DEF>`.
Define Runtime's Invocation's "State" as `<DEF>HY{|Key: ~InvocationState~, Props: ~&[Prop{Key: snapshot, Type: Option<Snapshot>}Prop, Prop{Key: claim, Type: u8}Prop]&~|}<DEF>`.
Define Runtime's Invocation's "Group" as `<DEF>HY{|Key: ~GroupState~, Props: ~&[Prop{Key: pid, Type: Option<u32>}Prop, Prop{Key: cleaned, Type: bool}Prop, Prop{Key: termination_deadline, Type: Option<tokio::time::Instant>}Prop]&~|}<DEF>`.

Define Runtime's "Future" as `<DEF>COB{|Key: ~<@Name|Label@>~|}^CV{|Capture: ~Move~|}^(<@Process@>)<DEF>`.
Define Runtime's "Pin|-s|-ned|-ning|-ner" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~Box::pin~, Inputs: ~&[<@Value@>]&~|}<DEF>`.
Define Runtime's "Output" as a New Adjective describing any Runtime's Pinner.
It describes any Runtime's Pinner as `<DEF>MOB{|Key: ~<@Name|Label@>~, Type: ~crate::runtime::host::Output~|}^KY{|Key: ~Box::pin~, Inputs: ~&[<@Value@>]&~|}<DEF>`.
Define "Pending" as a New Adjective describing any Function.
It describes a Generic Function of a New Process of a New Function's Returner of a New Runtime's Pinner of a New Runtime's Future of any Process.
Define Runtime's "Completion" as `<DEF>COB{|Key: ~<@Name|Label@>~|}^DV{|Key: ~<@Future@>~|}<DEF>`.
Define "Dispose|-s|-d|--ing|--er" to `<DEF>MOB{|Key: ~_~|}^KY{|Key: ~std::mem::drop~, Inputs: ~&[<@Value@>]&~|}<DEF>`.
Use Storage.Success as "Success".
Use Storage.Rejection as "Failure".
Use Storage.Nothing as "Nothing".
Define "Lock|-s|-ed|-ing|-er" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~crate::runtime::host::lock~, Inputs: ~&[&<@Value@>]&~|}<DEF>`.

Define Runtime's "Reference" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~&<@Value@>.inner~|}<DEF>`.
Define Runtime's "Admission" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~&<@Value@>.state~|}<DEF>`.
Define Runtime's "Stopping" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~<@Value@>.stopping~|}<DEF>`.
Define Runtime's "Finalized" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~<@Value@>.finalized~|}<DEF>`.
Define Runtime's "Selected" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~is_some~, Receiver: ~<@Value@>.terminal~|}<DEF>`.
Define Runtime's "Shared" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~std::sync::Arc::new~, Inputs: ~&[<@Value@>]&~|}<DEF>`.
Define Runtime's "Weak" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~std::sync::Arc::downgrade~, Inputs: ~&[<@Value@>]&~|}<DEF>`.
Define Runtime's "Mutex" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~std::sync::Mutex::new~, Inputs: ~&[<@Value@>]&~|}<DEF>`.
Define Runtime's "Ascend|-s|-ed|-ing|-er" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~upgrade~, Receiver: ~<@Value@>~|}<DEF>`.
Use Storage.App.Retain as Runtime's "Retain".
Define Runtime's Handle's "Content" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^Y{|Key: ~Runtime~, Props: ~&[Prop{Key: inner, Value: <@Inner@>}Prop]&~|}<DEF>`.
Define Runtime's App's "Parent" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~&<@Value@>.runtime~|}<DEF>`.
Define Runtime's App's "Admission" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~&<@Value@>.state~|}<DEF>`.
Define Runtime's App's State's "Content" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^Y{|Key: ~AppState~, Props: ~&[Prop{Key: source, Value: <@Source@>}Prop, Prop{Key: invocation, Value: None}Prop, Prop{Key: outstanding, Value: false}Prop, Prop{Key: selected, Value: None}Prop, Prop{Key: operation_error, Value: None}Prop]&~|}<DEF>`.
Define Runtime's App's "Content" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^Y{|Key: ~App~, Props: ~&[Prop{Key: runtime, Value: <@Runtime@>}Prop, Prop{Key: path, Value: <@Path@>}Prop, Prop{Key: state, Value: <@State@>}Prop]&~|}<DEF>`.
Define Runtime's App's "Source" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~&mut <@Value@>.source~|}<DEF>`.
Define Runtime's "Replace|-s|-d|--ing|--er" to `<DEF>O{|Key: ~*<@Target@>~, Value: ~<@Value@>~|}<DEF>`.
Define Runtime's "Empty" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~String::new~|}<DEF>`.

    Create Runtime's "Admit|-s|-ted|-ting|-ter" as a Generic `runtime_admitted` Function with Inputs of "state" as `&RuntimeState` and Outputs of `bool`.
Create "Stopping" as a New Runtime's Stopping of State.
Create "Selected" as a New Runtime's Selected of State.
Create "Finalized" as a New Runtime's Finalized of State.
Create "Terminating" as a New Storage's Alternative of Stopping and Selected.
Create "Stopped" as a New Storage's Alternative of Terminating and Finalized.
Return a New Equality of Stopped and false.

    Publicly create Runtime's "Check|-s|-ed|-ing|-er" as a Generic `runtime_check` Function with Inputs of "value" as `&Runtime` and Outputs of `Result<()>`.
Create "Guard" as a New Locker of a New Runtime's Admission of a New Runtime's Reference of Value.
Create "Allowed" as a New Runtime's Admitter of a New Storage's Borrower of Guard.
If Allowed, Create "Unit" as a New Nothing and then Create "Result" as a New Success of Unit and then Return Result.
Return a New Failure of a New Runtime's Error's Canceller.

    Publicly create Runtime's "Appoint|-s|-ed|-ing|-er" as a Generic `runtime_app` Function with Inputs of "value" as `&Runtime` and "path" as `std::path::PathBuf` and Outputs of `std::sync::Arc<App>`.
Create "Owner" as a New Runtime's Weak of a New Runtime's Reference of Value.
Create "Mutex" as a New Runtime's Mutex of a New Runtime's App's State's Content of a New Runtime's Empty.
Return a New Runtime's Shared of a New Runtime's App's Content of Owner and Path and Mutex.

    Publicly create Runtime's App's "Set|-s|-t|-ting|-ter" as a Generic `app_set_source` Function with Inputs of "value" as `&App` and "source" as `String` and Outputs of `()`.
Create "State" as a New Mutable Binding of a New Locker of a New Runtime's App's Admission of Value.
Create "Target" as a New Runtime's App's Source of State.
Runtime.Replace Target and Source.
Return a New Nothing.

Define Runtime's Invocation's "Parent" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~&<@Value@>.runtime~|}<DEF>`.
Define Runtime's Invocation's "Admission" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~&<@Value@>.state~|}<DEF>`.
Define Runtime's Invocation's "Available" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~is_some~, Receiver: ~<@Value@>.snapshot~|}<DEF>`.
Define Runtime's Invocation's "Phase" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~<@Value@>.claim~|}<DEF>`.
Define Runtime's Invocation's "Mark" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~&mut <@Value@>.claim~|}<DEF>`.
Define Runtime's "Byte" as `<DEF>DOB{|Key: ~<@Name|Label@>~|}^DE{|Ops: ~&[Op{Type: Calculate, Operator: as, Operands: &[<@Value@>, u8]&}Op]&~|}<DEF>`.
Define Runtime's "Fault" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~is_err~, Receiver: ~<@Value@>~|}<DEF>`.

    Create Runtime's "Reject|-s|-ed|-ing|-er" as a Generic Storage.Fallible `reject_result` Storage.Routine with Inputs of "value" as `std::result::Result<T, E>` and Outputs of `E`.
Create "Optional" as a New Runtime's Failure of Value.
Return a New Contents of Optional and `checked error result`.

    Publicly create Runtime's "Copy|--ies|--ied|-ing|--ier" as a Generic `runtime_copy` Function with Inputs of "value" as `&Runtime` and Outputs of Runtime's Handle.
Return a New Runtime's Handle's Content of a New Runtime's Retainer of a New Runtime's Reference of Value.

    Create Runtime's Invocation's "Own|-s|-ed|-ing|-er" as a Generic `invocation_owner` Function with Inputs of "value" as `&Invocation` and Outputs of `Result<Runtime>`.
Create "Optional" as a New Runtime's Ascender of a New Runtime's Invocation's Parent of Value.
Create "Available" as a New Presence of Optional.
If Available, Create "Inner" as a New Contents of Optional and `present invocation runtime` and then Create "Handle" as a New Runtime's Handle's Content of Inner and then Create "Found" as a New Success of Handle and then Return Found.
Return a New Failure of a New Runtime's Error's Describer of `invocation runtime is unavailable`.

    Publicly create Runtime's Invocation's "Claim|-s|-ed|-ing|-er" as a Generic `invocation_claim` Function with Inputs of "value" as `&Invocation` and Outputs of `Result<bool>`.
Create "Optional" as a New Runtime's Invocation's Owner of Value.
Create "Failed" as a New Runtime's Fault of Optional.
If Failed, Create "Reason" as a New Runtime's Rejecter of Optional and then Create "Rejected" as a New Failure of Reason and then Return Rejected.
Create "State" as a New Runtime's Admission of a New Runtime's Reference of a New Contents of Optional and `present invocation runtime`.
Create "Admission" as a New Locker of State.
Create "Allowed" as a New Runtime's Admitter of a New Storage's Borrower of Admission.
If Allowed is false, Create "Cancellation" as a New Runtime's Error's Canceller and then Create "Cancelled" as a New Failure of Cancellation and then Return Cancelled.
Create "Guard" as a New Mutable Binding of a New Locker of a New Runtime's Invocation's Admission of Value.
Create "Available" as a New Runtime's Invocation's Available of Guard.
Create "Claim" as a New Runtime's Invocation's Phase of Guard.
Create "Ready" as a New Agreement of Available and a New Equality of Claim and a New Runtime's Byte of 0.
If Ready is false, Create "Absent" as a New Success of false and then Return Absent.
Create "One" as a New Runtime's Byte of 1.
Create "Target" as a New Runtime's Invocation's Mark of Guard.
Runtime.Replace Target and One.
Return a New Success of true.

Define Runtime's Invocation's "App" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~&<@Value@>.app~|}<DEF>`.
Define Runtime's App's "Current" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~as_ref~, Receiver: ~<@Value@>.invocation~|}<DEF>`.
Define Runtime's App's "Outstanding" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~&mut <@Value@>.outstanding~|}<DEF>`.
Define Runtime's "Identical" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~std::sync::Arc::ptr_eq~, Inputs: ~&[<@Left@>, <@Right@>]&~|}<DEF>`.

    Create Runtime's App's "Release|-s|-d|--ing|-r" as a Generic `app_release` Function with Inputs of "value" as `&App` and "invocation" as `&std::sync::Arc<Invocation>` and Outputs of `()`.
Create "Unit" as a New Nothing.
Create "Guard" as a New Mutable Binding of a New Locker of a New Runtime's App's Admission of Value.
Create "Optional" as a New Runtime's App's Current of Guard.
Create "Available" as a New Presence of Optional.
If Available is false, Return Unit.
Create "Current" as a New Contents of Optional and `present invocation`.
Create "Same" as a New Runtime's Identical of Current and Invocation.
If Same, Create "Target" as a New Runtime's App's Outstanding of Guard and then Runtime.Replace Target and false.
Return Unit.

    Publicly create Runtime's Invocation's "Finish|-es|-ed|-ing|-er" as a Generic `invocation_finish` Function with Inputs of "value" as `&std::sync::Arc<Invocation>` and Outputs of `()`.
Create "Unit" as a New Nothing.
Create "Guard" as a New Mutable Binding of a New Locker of a New Runtime's Invocation's Admission of Value.
Create "Phase" as a New Runtime's Invocation's Phase of Guard.
Create "One" as a New Runtime's Byte of 1.
If Phase isnt One, Return Unit.
Create "Two" as a New Runtime's Byte of 2.
Create "Target" as a New Runtime's Invocation's Mark of Guard.
Runtime.Replace Target and Two.
Dispose Guard.
Create "Optional" as a New Runtime's Ascender of a New Runtime's Invocation's App of Value.
Create "Available" as a New Presence of Optional.
If Available, Create "App" as a New Contents of Optional and `present invocation App` and then Create "Borrowed" as a New Storage's Borrower of App and then Create "Released" as a New Runtime's App's Releaser of Borrowed and Value.
Return Unit.

Define Runtime's Snapshot's "Output" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~&<@Value@>.stdout~|}<DEF>`.
Define Runtime's Snapshot's "Error" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~&<@Value@>.stderr~|}<DEF>`.
Define Runtime's Snapshot's "Fault" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~&<@Value@>.error~|}<DEF>`.
Define Runtime's Snapshot's "Status" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~<@Value@>.status~|}<DEF>`.
Define Runtime's Snapshot's "Content" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^Y{|Key: ~Snapshot~, Props: ~&[Prop{Key: stdout, Value: <@Output@>}Prop, Prop{Key: stderr, Value: <@Error@>}Prop, Prop{Key: error, Value: <@Fault@>}Prop, Prop{Key: status, Value: <@Status@>}Prop]&~|}<DEF>`.

    Publicly create Runtime's Snapshot's "Copy|--ies|--ied|-ing|--ier" as a Generic `snapshot_copy` Function with Inputs of "value" as `&Snapshot` and Outputs of Runtime's Snapshot.
Create "Stdout" as a New Runtime's Retainer of a New Runtime's Snapshot's Output of Value.
Create "Stderr" as a New Runtime's Retainer of a New Runtime's Snapshot's Error of Value.
Create "Copied" as a New Storage's Copier of a New Runtime's Snapshot's Fault of Value.
Create "Status" as a New Runtime's Snapshot's Status of Value.
Return a New Runtime's Snapshot's Content of Stdout and Stderr and Copied and Status.

Define Runtime's Invocation's "Snapshot" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~&<@Value@>.snapshot~|}<DEF>`.

    Publicly create Runtime's Invocation's "Read|-s|-d|-ing|-er" as a Generic `invocation_snapshot` Function with Inputs of "value" as `&Invocation` and Outputs of `Option<Snapshot>`.
Create "Snapshot" as a New Runtime's Invocation's Snapshot of a New Locker of a New Runtime's Invocation's Admission of Value.
Return a New Storage's Copier of Snapshot.

Define Runtime's "Subscribe|-s|-d|--ing|--er" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~subscribe~, Receiver: ~<@Value@>~|}<DEF>`.
Define Runtime's "Current" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~borrow~, Receiver: ~<@Value@>~|}<DEF>`.
Define Runtime's "Value" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~*<@Reference@>~|}<DEF>`.
Define Runtime's "Change|-s|-d|--ing|-r" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~changed~, Receiver: ~<@Value@>~|}<DEF>`.
Create Runtime's "Repeat|-s|-ed|-ing|-er" for Processing a New Storage's Loop of any Process.
Define Runtime's "Halt|-s|-ed|-ing|-er" to `<DEF>BOG^GI^(E{|Statement: ~<@Condition@>~|}*W^(BOD^E{|Statement: ~break;~|})*E{|Statement: ~Void~|}*W)<DEF>`.
Define Runtime's Invocation's "Done" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~&<@Value@>.done~|}<DEF>`.

    Create Runtime's "Wait|-s|-ed|-ing|-er" as a Generic Pending `wait_true` Function with Inputs of "receiver" as `&mut tokio::sync::watch::Receiver<bool>` and Outputs of `std::pin::Pin<Box<dyn std::future::Future<Output = ()> + Send + '_>>`.
Runtime.Repeat a New Process for Runtime.Halting a New Runtime's Observer of Receiver and then
Runtime.Halting a New Runtime's Fault of a New Runtime's Completion of a New Runtime's Changer of Receiver.
Return a New Nothing.

    Publicly create Runtime's Invocation's "Complete|-s|-d|--ing|--er" as a Generic Pending `invocation_completed` Function with Inputs of "value" as `&Invocation` and Outputs of `std::pin::Pin<Box<dyn std::future::Future<Output = ()> + Send + '_>>`.
Create "Receiver" as a New Mutable Binding of a New Runtime's Subscriber of a New Runtime's Invocation's Done of Value.
Return a New Runtime's Completion of a New Runtime's Waiter of a New Storage.Editable Storage's Borrower of Receiver.

Define Runtime's "Asynchronous" as a New Adjective describing any Runtime's Mutex.
It describes any Runtime's Mutex as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~tokio::sync::Mutex::new~, Inputs: ~&[<@Value@>]&~|}<DEF>`.
Define Runtime's "Watch" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~tokio::sync::watch::channel~, Inputs: ~&[<@Value@>]&~|}<DEF>`.
Define Runtime's "Channel" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~tokio::sync::mpsc::unbounded_channel~|}<DEF>`.
Define Runtime's Invocation's State's "Content" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^Y{|Key: ~InvocationState~, Props: ~&[Prop{Key: snapshot, Value: None}Prop, Prop{Key: claim, Value: 0}Prop]&~|}<DEF>`.
Define Runtime's Invocation's Group's "Content" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^Y{|Key: ~GroupState~, Props: ~&[Prop{Key: pid, Value: <@Pid@>}Prop, Prop{Key: cleaned, Value: false}Prop, Prop{Key: termination_deadline, Value: None}Prop]&~|}<DEF>`.
Define Runtime's Invocation's "Content" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^Y{|Key: ~Invocation~, Props: ~&[Prop{Key: app, Value: <@App@>}Prop, Prop{Key: runtime, Value: <@Runtime@>}Prop, Prop{Key: sequence, Value: <@Sequence@>}Prop, Prop{Key: dependencies, Value: <@Dependencies@>}Prop, Prop{Key: wait_delay, Value: <@Delay@>}Prop, Prop{Key: state, Value: <@State@>}Prop, Prop{Key: done, Value: <@Done@>}Prop, Prop{Key: group, Value: <@Group@>}Prop, Prop{Key: drain_deadline, Value: <@Deadline@>}Prop, Prop{Key: leader_requests, Value: <@Requests@>}Prop]&~|}<DEF>`.

    Publicly create Runtime's Invocation's "Allocate|-s|-d|--ing|--or" as a Generic `invocation_allocate` Function with Inputs of "app" as `std::sync::Weak<App>` and "owner" as `std::sync::Weak<RuntimeInner>` and "sequence" as `u64` and "dependencies" as `std::sync::Arc<dyn crate::runtime::host::Dependencies>` and "delay" as `tokio::time::Duration` and "pid" as `Option<u32>` and Outputs of `(std::sync::Arc<Invocation>, tokio::sync::mpsc::UnboundedReceiver<tokio::sync::oneshot::Sender<Result<()>>>)`.
Create "Completion" as a New Runtime's Watch of false.
Create "Done" as a New First of Completion.
Create "Unused" as a New Second of Completion.
Dispose Unused.
Create "Deadline" as a New Runtime's Watch of a New Absent.
Create "Bound" as a New First of Deadline.
Create "Unobserved" as a New Second of Deadline.
Dispose Unobserved.
Create "Channel" as a New Runtime's Channel.
Create "Requests" as a New First of Channel.
Create "Receiver" as a New Second of Channel.
Create "State" as a New Runtime's Mutex of a New Runtime's Invocation's State's Content.
Create "Gate" as a New Runtime.Asynchronous Runtime's Mutex of a New Runtime's Invocation's Group's Content of Pid.
Create "Shared" as a New Runtime's Shared of a New Runtime's Invocation's Content of App, Owner, Sequence, Dependencies, Delay, State, Done, Gate, Bound, and Requests.
Return a New Pair of Shared and Receiver.

Define Runtime's "Signals" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~crate::process::Signals::new~|}<DEF>`.
Define Runtime's "Executable" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~std::env::current_exe~|}<DEF>`.
Define Runtime's "Canonical" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~std::fs::canonicalize~, Inputs: ~&[<@Path@>]&~|}<DEF>`.
Define Runtime's "Parent" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~parent~, Receiver: ~<@Path@>~|}<DEF>`.
Define Runtime's "Path" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~to_path_buf~, Receiver: ~<@Path@>~|}<DEF>`.
Define Runtime's "Explain|-s|-ed|-ing|-er" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~map_err~, Receiver: ~<@Result@>~, Inputs: ~&[<@Visit@>]&~|}<DEF>`.
Define Runtime's "Initial" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^Y{|Key: ~RuntimeState~, Props: ~&[Prop{Key: owned, Value: <@Owned@>}Prop, Prop{Key: cleanup_errors, Value: <@Errors@>}Prop, Prop{Key: next, Value: 0}Prop, Prop{Key: stopping, Value: false}Prop, Prop{Key: terminal, Value: None}Prop, Prop{Key: finalized, Value: false}Prop, Prop{Key: final_status, Value: None}Prop]&~|}<DEF>`.
Define Runtime's "Map" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~std::collections::BTreeMap::new~|}<DEF>`.
Define Runtime's Inner's "Content" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^Y{|Key: ~RuntimeInner~, Props: ~&[Prop{Key: companion_directory, Value: <@Directory@>}Prop, Prop{Key: state, Value: <@State@>}Prop, Prop{Key: cancelled, Value: <@Cancelled@>}Prop, Prop{Key: shutdown_done, Value: <@Done@>}Prop, Prop{Key: signals, Value: <@Signals@>}Prop, Prop{Key: stdout, Value: <@Output@>}Prop, Prop{Key: stderr, Value: <@Error@>}Prop, Prop{Key: dependencies, Value: <@Dependencies@>}Prop, Prop{Key: finish_gate, Value: <@Gate@>}Prop]&~|}<DEF>`.

    Create Runtime's Path's "Explain|-s|-ed|-ing|-er" as a Generic `runtime_location_error` Function with Inputs of "error" as `std::io::Error` and Outputs of Runtime's Error.
Return a New Runtime's Error's Explainer of `locate orchestrator` and Error.

    Create Runtime's "Locate|-s|-d|--ing|--or" as a Generic `runtime_locate` Function with No Inputs and Outputs of `Result<std::path::PathBuf>`.
Create "Executable" as a New Runtime's Executable.
Create "Canonical" as a New Runtime's Canonical of a New Storage.Checked Storage's Binding of a New Runtime's Explainer of Executable and Runtime's Path's Explainer.
Create "Path" as a New Storage.Checked Storage's Binding of a New Runtime's Explainer of Canonical and Runtime's Path's Explainer.
Create "Parent" as a New Runtime's Parent of Path.
Create "Available" as a New Presence of Parent.
If Available, Create "Directory" as a New Contents of Parent and `present executable directory` and then Create "Owned" as a New Runtime's Path of Directory and then Create "Result" as a New Success of Owned and then Return Result.
Return a New Failure of a New Runtime's Error's Describer of `orchestrator has no parent directory`.

    Create Runtime's "Initialize|-s|-d|--ing|-r" as a Generic `runtime_state` Function with No Inputs and Outputs of Runtime's State.
Create "Owned" as a New Runtime's Map.
Return a New Runtime's Initial of Owned and a New Runtime's Map.


    Create Runtime's "Observe|-s|-d|--ing|-r" as a Generic Storage.Cloneable `observe` Storage.Routine with Inputs of "receiver" as `&tokio::sync::watch::Receiver<T>` and Outputs of `T`.
Return a New Storage's Copier of a New Storage's Reference of a New Runtime's Current of Receiver.

Define Runtime's "Native" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~crate::runtime::host::NativeDependencies~|}<DEF>`.
Define Runtime's "Dependencies" as `<DEF>MOB{|Key: ~<@Name|Label@>~, Type: ~std::sync::Arc<dyn crate::runtime::host::Dependencies>~|}^KY{|Key: ~std::sync::Arc::new~, Inputs: ~&[<@Value@>]&~|}<DEF>`.
Define Runtime's "Stdout" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~tokio::io::stdout~|}<DEF>`.
Define Runtime's "Stderr" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~tokio::io::stderr~|}<DEF>`.
Define Runtime's "Spawn|-s|-ed|-ing|-er" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~tokio::spawn~, Inputs: ~&[<@Value@>]&~|}<DEF>`.
Define Runtime's "Receive|-s|-d|--ing|-r" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~recv~, Receiver: ~<@Value@>~|}<DEF>`.
Define Runtime's "Task" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~&<@Value@>.signals~|}<DEF>`.
Define Runtime's "Authority" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~&<@Value@>.dependencies~|}<DEF>`.
Define Runtime's "Signal" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~crate::runtime::host::TerminalRequest::Signal~, Inputs: ~&[<@Value@>]&~|}<DEF>`.
Define Runtime's "Enter|-s|-ed|-ing|-er" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~terminal_entered~, Receiver: ~<@Dependencies@>~, Inputs: ~&[<@Request@>]&~|}<DEF>`.

    Create Runtime's Signals' "Explain|-s|-ed|-ing|-er" as a Generic `runtime_signal_error` Function with Inputs of "error" as `std::io::Error` and Outputs of Runtime's Error.
Return a New Runtime's Error's Explainer of `install application signal subscriptions` and Error.

    Create Runtime's "Assemble|-s|-d|--ing|-r" as a Generic `runtime_assemble` Function with Inputs of "dependencies" as `std::sync::Arc<dyn crate::runtime::host::Dependencies>` and "output" as `crate::runtime::host::Output` and "error" as `crate::runtime::host::Output` and Outputs of Runtime's Handle.
Create "Directory" as a New Runtime's Locator.
Create "State" as a New Runtime's Mutex of a New Runtime's Initializer.
Create "Cancellation" as a New Runtime's Watch of false.
Create "Cancel" as a New First of Cancellation.
Create "Unobserved" as a New Second of Cancellation.
Dispose Unobserved.
Create "Completion" as a New Runtime's Watch of a New Absent.
Create "Done" as a New First of Completion.
Create "Unused" as a New Second of Completion.
Dispose Unused.
Create "Task" as a New Runtime's Mutex of a New Absent.
Create "Stdout" as a New Runtime.Asynchronous Runtime's Mutex of Output.
Create "Stderr" as a New Runtime.Asynchronous Runtime's Mutex of Error.
Create "Gate" as a New Runtime.Asynchronous Runtime's Mutex of a New Nothing.
Return a New Runtime's Handle's Content of a New Runtime's Shared of a New Runtime's Inner's Content of Directory, State, Cancel, Done, Task, Stdout, Stderr, Dependencies, and Gate.

Define Runtime's Invocation's "Gate" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~&<@Value@>.group~|}<DEF>`.
Define Runtime's "Await|-s|-ed|-ing|-er" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~lock~, Receiver: ~<@Value@>~|}<DEF>`.
Define Runtime's Invocation's "Cleaned" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~<@Value@>.cleaned~|}<DEF>`.
Define Runtime's Invocation's "Pid" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~<@Value@>.pid~|}<DEF>`.
Define Runtime's Invocation's "Deadline" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~<@Value@>.termination_deadline~|}<DEF>`.
Define Runtime's Invocation's "Limit" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~&mut <@Value@>.termination_deadline~|}<DEF>`.
Define Runtime's Invocation's "Seal" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~&mut <@Value@>.cleaned~|}<DEF>`.
Define Runtime's Invocation's "Delay" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~<@Value@>.wait_delay~|}<DEF>`.
Define Runtime's Invocation's "Dependencies" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~&<@Value@>.dependencies~|}<DEF>`.
Define Runtime's Invocation's "Notification" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~&<@Value@>.drain_deadline~|}<DEF>`.
Define Runtime's "Now" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~tokio::time::Instant::now~|}<DEF>`.
Define Runtime's "Group" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~crate::runtime::host::SignalTarget::Group~, Inputs: ~&[<@Pid@>]&~|}<DEF>`.
Define Runtime's "Kill" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~nix::sys::signal::Signal::SIGKILL~|}<DEF>`.
Define Runtime's "Send|-s|--t|-ing|-er" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~signal~, Receiver: ~<@Dependencies@>~, Inputs: ~&[<@Target@>, <@Signal@>]&~|}<DEF>`.
Define Runtime's "Code" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~raw_os_error~, Receiver: ~<@Error@>~|}<DEF>`.
Define Runtime's "Missing" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~nix::libc::ESRCH~|}<DEF>`.
Define Runtime's "Valid" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~is_ok~, Receiver: ~<@Value@>~|}<DEF>`.
Define Runtime's "Minimum" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~min~, Receiver: ~<@Left@>~, Inputs: ~&[<@Right@>]&~|}<DEF>`.
Define Runtime's "Bound|-s|-ed|-ing|-er" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~crate::process::final_group_result~, Inputs: ~&[<@Pid@>, <@Result@>, <@Deadline@>]&~|}<DEF>`.
Define Runtime's "Modifier" as `<DEF>KOB{|Key: ~<@Name|Label@>~|}^BW{|Inputs: ~&[Prop{Key: current, Type: &mut Option<tokio::time::Instant>}Prop]&~|}^W^MOB{|Key: ~_~|}^KY{|Key: ~<@Visit@>~, Inputs: ~&[current, <@Deadline@>]&~|}<DEF>`.
Define Runtime's "Amend|-s|-ed|-ing|-er" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~send_modify~, Receiver: ~<@Sender@>~, Inputs: ~&[<@Modifier@>]&~|}<DEF>`.

    Create Runtime's "Revise|-s|-d|--ing|-r" as a Generic `revise_deadline` Function with Inputs of "sender" as `&tokio::sync::watch::Sender<Option<tokio::time::Instant>>` and "deadline" as `tokio::time::Instant` and Outputs of `()`.
Return a New Runtime's Amender of Sender and a New Runtime's Modifier of Runtime's Limiter and Deadline.

    Create Runtime's "Normalize|-s|-d|--ing|-r" as a Generic `normalize_signal` Function with Inputs of "value" as `std::io::Result<()>` and Outputs of `std::io::Result<()>`.
Create "Valid" as a New Runtime's Valid of Value.
If Valid, Return Value.
Create "Error" as a New Runtime's Rejecter of Value.
Create "Code" as a New Runtime's Code of Error.
Create "Expected" as a New Present of a New Runtime's Missing.
If Code is Expected, Create "Unit" as a New Nothing and then Create "Accepted" as a New Success of Unit and then Return Accepted.
Return a New Failure of Error.

    Create Runtime's "Limit|-s|-ed|-ing|-er" as a Generic `limit_deadline` Function with Inputs of "current" as `&mut Option<tokio::time::Instant>` and "deadline" as `tokio::time::Instant` and Outputs of `()`.
Create "Value" as a New Runtime's Value of Current.
Create "Available" as a New Presence of Value.
If Available, Create "Old" as a New Contents of Value and `present cleanup deadline` and then Create "Earlier" as a New Runtime's Minimum of Old and Deadline and then Create "Selected" as a New Present of Earlier and then Runtime.Replace Current and Selected.
Otherwise, Create "Wrapped" as a New Present of Deadline and then Runtime.Replace Current and Wrapped.
Return a New Nothing.

    Create Runtime's Invocation's "Bound|-s|-ed|-ing|-er" as a Generic `invocation_bound` Function with Inputs of "value" as `&Invocation` and "deadline" as `tokio::time::Instant` and Outputs of `()`.
Create "Sender" as a New Runtime's Invocation's Notification of Value.
Create "Revised" as a New Runtime's Reviser of Sender and Deadline.
Return a New Nothing.

    Create Runtime's Invocation's "Signal|-s|-ed|-ing|-er" as a Generic Pending `invocation_signal` Function with Inputs of "value" as `&Invocation` and "last" as `bool` and Outputs of `std::pin::Pin<Box<dyn std::future::Future<Output = std::io::Result<()>> + Send + '_>>`.
Create "Unit" as a New Nothing.
Create "Guard" as a New Mutable Binding of a New Runtime's Completion of a New Runtime's Awaiter of a New Runtime's Invocation's Gate of Value.
Create "Cleaned" as a New Runtime's Invocation's Cleaned of Guard.
Create "Done" as a New Storage's Alternative of Cleaned and a New Presence of a New Runtime's Invocation's Reader of Value.
If Done, Create "Acceptedvalue" as a New Success of Unit and then Return Acceptedvalue.
Create "Optional" as a New Runtime's Invocation's Pid of Guard.
Create "Available" as a New Presence of Optional.
If Available is false, Create "Absent" as a New Success of Unit and then Return Absent.
Create "Pid" as a New Contents of Optional and `present invocation process`.
Create "Attempted" as a New Runtime's Now.
Create "Dependencies" as a New Runtime's Invocation's Dependencies of Value.
Create "Target" as a New Runtime's Group of Pid.
Create "Signal" as a New Runtime's Kill.
Create "Result" as a New Runtime's Normalizer of a New Runtime's Sender of Dependencies, Target, and Signal.
Create "Valid" as a New Runtime's Valid of Result.
Create "Accepted" as a New Agreement of Valid and a New Equality of Last and false.
Create "Set" as a New Presence of a New Runtime's Invocation's Deadline of Guard.
Create "Needed" as a New Agreement of Accepted and a New Equality of Set and false.
If Needed, Create "Delay" as a New Runtime's Invocation's Delay of Value and then Create "Deadline" as a New Sum of Attempted and Delay and then Create "Wrapped" as a New Present of Deadline and then Create "Limit" as a New Runtime's Invocation's Limit of Guard and then Runtime.Replace Limit and Wrapped and then Create "Bounded" as a New Runtime's Invocation's Bounder of Value and Deadline.
If Last, Create "Finaldeadline" as a New Runtime's Invocation's Deadline of Guard and then Create "Finaltask" as a New Runtime's Bounder of Pid, Result, and Finaldeadline and then Create "Finalresult" as a New Runtime's Completion of Finaltask and then Create "Seal" as a New Runtime's Invocation's Seal of Guard and then Runtime.Replace Seal and true and then Return Finalresult.
Return Result.

    Create Runtime's Invocation's "Explain|-s|-ed|-ing|-er" as a Generic `invocation_signal_error` Function with Inputs of "error" as `std::io::Error` and Outputs of Runtime's Error.
Return a New Runtime's Error's Explainer of `terminate application process group` and Error.

    Publicly create Runtime's Invocation's "Terminate|-s|-d|--ing|--or" as a Generic Pending `invocation_terminate` Function with Inputs of "value" as `&Invocation` and "last" as `bool` and Outputs of `std::pin::Pin<Box<dyn std::future::Future<Output = Result<()>> + Send + '_>>`.
Create "Result" as a New Runtime's Completion of a New Runtime's Invocation's Signaler of Value and Last.
Return a New Runtime's Explainer of Result and Runtime's Invocation's Explainer.

Define Runtime's "Reply" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~tokio::sync::oneshot::channel~|}<DEF>`.
Define Runtime's Invocation's "Requests" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~&<@Value@>.leader_requests~|}<DEF>`.
Define Runtime's "Submit|-s|-ted|-ting|-ter" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~send~, Receiver: ~<@Sender@>~, Inputs: ~&[<@Value@>]&~|}<DEF>`.
Define Runtime's Invocation's "Sequence" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~<@Value@>.sequence~|}<DEF>`.
Define Runtime's "Failures" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~&mut <@Value@>.cleanup_errors~|}<DEF>`.
Define Runtime's "Entry" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~entry~, Receiver: ~<@Values@>~, Inputs: ~&[<@Key@>]&~|}<DEF>`.
Define Runtime's "Insert|-s|-ed|-ing|-er" to `<DEF>MOB{|Key: ~_~|}^KY{|Key: ~or_insert~, Receiver: ~<@Entry@>~, Inputs: ~&[<@Value@>]&~|}<DEF>`.

    Publicly create Runtime's Invocation's "Kill|-s|-ed|-ing|-er" as a Generic Pending `invocation_kill` Function with Inputs of "value" as `&Invocation` and Outputs of `std::pin::Pin<Box<dyn std::future::Future<Output = Result<()>> + Send + '_>>`.
Create "Pair" as a New Runtime's Reply.
Create "Reply" as a New First of Pair.
Create "Receiver" as a New Second of Pair.
Create "Requests" as a New Runtime's Invocation's Requests of Value.
Create "Closed" as a New Runtime's Fault of a New Runtime's Submitter of Requests and Reply.
Create "Accepted" as a New Success of a New Nothing.
If Closed, Return Accepted.
Create "Completed" as a New Runtime's Completion of Receiver.
Return a New Alternative of Completed and Accepted.

    Publicly create Runtime's Invocation's "Record|-s|-ed|-ing|-er" as a Generic `invocation_record` Function with Inputs of "value" as `&Invocation` and "error" as `Error` and Outputs of `()`.
Create "Optional" as a New Runtime's Ascender of a New Runtime's Invocation's Parent of Value.
Create "Available" as a New Presence of Optional.
Create "Unit" as a New Nothing.
If Available is false, Return Unit.
Create "Locked" as a New Locker of a New Runtime's Admission of a New Contents of Optional and `present invocation runtime`.
Create "Failures" as a New Runtime's Failures of a New Mutable Binding of Locked.
Create "Sequence" as a New Runtime's Invocation's Sequence of Value.
Create "Entry" as a New Runtime's Entry of Failures and Sequence.
Runtime.Insert Entry and Error.
Return Unit.

Define Runtime's Invocation's "Capture" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~&mut <@Value@>.snapshot~|}<DEF>`.
Define Runtime's Invocation's "Publication" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~crate::runtime::host::Checkpoint::PrePublication~|}<DEF>`.
Use Runtime.Notify as Runtime's Invocation's "Notify".

    Publicly create Runtime's Invocation's "Publish|-es|-ed|-ing|-er" as a Generic `invocation_publish` Function with Inputs of "value" as `&std::sync::Arc<Invocation>` and "snapshot" as `Snapshot` and Outputs of `()`.
Create "Dependencies" as a New Runtime's Invocation's Dependencies of Value.
Create "Event" as a New Runtime's Invocation's Publication.
Create "Retained" as a New Runtime's Retainer of Value.
Runtime.Checkpoint.Enter Dependencies, Event, and Retained.
Create "Guard" as a New Mutable Binding of a New Locker of a New Runtime's Invocation's Admission of Value.
Create "Target" as a New Runtime's Invocation's Capture of Guard.
Create "Wrapped" as a New Present of Snapshot.
Runtime.Replace Target and Wrapped.
Dispose Guard.
Create "Done" as a New Runtime's Invocation's Done of Value.
Runtime.Invocation.Notify Done and true.
Return a New Nothing.
