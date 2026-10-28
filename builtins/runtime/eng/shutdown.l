*** Terminal selection and shutdown are source policy. Small access words expose
    owned state; synchronous Functions delimit locks before asynchronous work. ***
Define Runtime's Inner's "Cancellation" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~&<@Value@>.cancelled~|}<DEF>`.
Define Runtime's Inner's "Notification" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~&<@Value@>.shutdown_done~|}<DEF>`.
Define Runtime's Inner's "Signals" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~&<@Value@>.signals~|}<DEF>`.
Define Runtime's Inner's "Gate" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~&<@Value@>.finish_gate~|}<DEF>`.
Define Runtime's Inner's "Stdout" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~&<@Value@>.stdout~|}<DEF>`.
Define Runtime's Inner's "Stderr" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~&<@Value@>.stderr~|}<DEF>`.
Define Runtime's State's "Terminal" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~&mut <@Value@>.terminal~|}<DEF>`.
Define Runtime's State's "Stop" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~&mut <@Value@>.stopping~|}<DEF>`.
Define Runtime's State's "Finalize" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~&mut <@Value@>.finalized~|}<DEF>`.
Define Runtime's State's "Status" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~<@Value@>.final_status~|}<DEF>`.
Define Runtime's State's "Remember" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~&mut <@Value@>.final_status~|}<DEF>`.
Define Runtime's State's "Owned" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~values~, Receiver: ~<@Value@>.owned~|}<DEF>`.
Define Runtime's State's "Errors" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~values~, Receiver: ~<@Value@>.cleanup_errors~|}<DEF>`.
Define Runtime's Terminal's "Status" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~<@Value@>.status~|}<DEF>`.
Define Runtime's Terminal's "Error" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~&<@Value@>.error~|}<DEF>`.
Define Runtime's Terminal's "Content" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^Y{|Key: ~Terminal~, Props: ~&[Prop{Key: status, Value: <@Status@>}Prop, Prop{Key: error, Value: <@Error@>}Prop]&~|}<DEF>`.
Define Runtime's "Notify|--ies|--ied|-ing|--ier" to `<DEF>MOB{|Key: ~_~|}^KY{|Key: ~send_replace~, Receiver: ~<@Sender@>~, Inputs: ~&[<@Value@>]&~|}<DEF>`.
Define Runtime's "Cloned" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~cloned~, Receiver: ~<@Iterator@>~|}<DEF>`.
Define Runtime's "Collection" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~collect~, Receiver: ~<@Iterator@>~, GenericTypes: ~Vec<_>~|}<DEF>`.
Define Runtime's "Take|-s|----ook|--ing|-r" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~take~, Receiver: ~<@Value@>~|}<DEF>`.
Define Runtime's "Abort|-s|-ed|-ing|-er" to `<DEF>MOB{|Key: ~_~|}^KY{|Key: ~abort~, Receiver: ~<@Task@>~|}<DEF>`.
Define Runtime's "Join|-s|-ed|-ing|-er" to `<DEF>BOP^X{|Iterable: ~<@Values@>~, Key: ~item~, RangeArity: ~1~|}^W^(MOB{|Key: ~pending~|}^KY{|Key: ~<@Action@>~, Inputs: ~&[item]&~|}*COB{|Key: ~_~|}^DV{|Key: ~pending~|})<DEF>`.
Define Runtime's "Flush|-es|-ed|-ing|-er" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~tokio::io::AsyncWriteExt::flush~, Inputs: ~&[<@Output@>]&~|}<DEF>`.
Define Runtime's "Write|-s|----ote|--ing|-r" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~tokio::io::AsyncWriteExt::write_all~, Inputs: ~&[<@Output@>, <@Bytes@>]&~|}<DEF>`.
Define Runtime's "Failure" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~err~, Receiver: ~<@Value@>~|}<DEF>`.
Define Runtime's "Faults" as `<DEF>MOB{|Key: ~mut <@Name|Label@>: Vec<Error>~|}^KY{|Key: ~Vec::new~|}<DEF>`.
Define Runtime's "Push|-es|-ed|-ing|-er" to `<DEF>MOB{|Key: ~_~|}^KY{|Key: ~push~, Receiver: ~<@Values@>~, Inputs: ~&[<@Value@>]&~|}<DEF>`.
Define Runtime's "Acquire|-s|-d|--ing|-r" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~lock~, Receiver: ~<@Mutex@>~|}<DEF>`.
Define Runtime's "Output" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~&mut *<@Value@>~|}<DEF>`.
Define Runtime's Terminal's "Record|-s|-ed|-ing|-er" to `<DEF>MOB{|Key: ~_~|}^KY{|Key: ~terminal_entered~, Receiver: ~<@Value@>.inner.dependencies~, Inputs: ~&[<@Request@>]&~|}<DEF>`.
Define Runtime's "Fatal" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~crate::runtime::host::TerminalRequest::Failure~|}<DEF>`.
Define Runtime's "Exit" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~crate::runtime::host::TerminalRequest::Exit~|}<DEF>`.

    Publicly create Runtime's "Fail|-s|-ed|-ing|-er" as a Generic `runtime_fail` Function with Inputs of "value" as `&Runtime` and "error" as `Option<Error>` and Outputs of Runtime's Error.
Create "Available" as a New Presence of Error.
If Available, Create "Fault" as a New Contents of Error and `present failure` and then Create "Selected" as a New Runtime's Refuser of Value and Fault and then Return Selected.
Create "Missing" as a New Runtime's Error's Describer of `cannot exit with an absent Error`.
Return a New Runtime's Refuser of Value and Missing.

    Create Runtime's "Refuse|-s|-d|--ing|-r" as a Generic `runtime_fail_present` Function with Inputs of "value" as `&Runtime` and "error" as `Error` and Outputs of Runtime's Error.
Create "Request" as a New Runtime's Fatal.
Runtime.Terminal.Record Value and Request.
Create "Wrapped" as a New Present of a New Runtime's Error's Copier of a New Storage's Borrower of Error.
Create "One" as a New Runtime's Byte of 1.
Create "Requested" as a New Runtime's Requester of Value and One and Wrapped.
Return Error.

    Publicly create Runtime's "Leave|-s|-----eft|--ing|-r" as a Generic `runtime_exit` Function with Inputs of "value" as `&Runtime` and Outputs of Runtime's Error.
Create "Request" as a New Runtime's Exit.
Runtime.Terminal.Record Value and Request.
Create "Missing" as a New Absent.
Create "Zero" as a New Runtime's Byte of 0.
Create "Requested" as a New Runtime's Requester of Value and Zero and Missing.
Return a New Runtime's Error's Canceller.

    Publicly create Runtime's "Show|-s|-ed|-ing|-er" as a Generic Pending `runtime_show` Function with Inputs of "value" as `&Runtime` and "bytes" as `&[u8]` and Outputs of `impl std::future::Future<Output = Result<()>> + Send`.
Create "Initial" as a New Runtime's Checker of Value.
Create "Rejected" as a New Runtime's Fault of Initial.
If Rejected, Return Initial.
Create "Inner" as a New Runtime's Reference of Value.
Create "Locked" as a New Runtime's Completion of a New Runtime's Acquirer of a New Runtime's Inner's Stdout of Inner.
Create "Guard" as a New Mutable Binding of Locked.
Create "Current" as a New Runtime's Checker of Value.
Create "Cancelled" as a New Runtime's Fault of Current.
If Cancelled, Return Current.
Create "Output" as a New Runtime's Output of Guard.
Create "Written" as a New Runtime's Completion of a New Runtime's Writer of Output and Bytes.
Create "Failed" as a New Runtime's Fault of Written.
If Failed, Create "Cause" as a New Runtime's Rejecter of Written and then Create "Reason" as a New Runtime's Error's Explainer of `write application stdout` and Cause and then Create "Fault" as a New Runtime's Refuser of Value and Reason and then Create "Failedresult" as a New Failure of Fault and then Return Failedresult.
Create "Flushed" as a New Runtime's Completion of a New Runtime's Flusher of Output.
Create "Incomplete" as a New Runtime's Fault of Flushed.
If Incomplete, Create "Later" as a New Runtime's Rejecter of Flushed and then Create "Explanation" as a New Runtime's Error's Explainer of `flush application stdout` and Later and then Create "Fatal" as a New Runtime's Refuser of Value and Explanation and then Create "Erroneous" as a New Failure of Fatal and then Return Erroneous.
Return a New Success of a New Nothing.

    Create Runtime's "Report|-s|-ed|-ing|-er" as a Generic Pending `runtime_report` Function with Inputs of "value" as `&Runtime` and "bytes" as `Vec<u8>` and Outputs of `std::pin::Pin<Box<dyn std::future::Future<Output = bool> + Send + '_>>`.
Create "Inner" as a New Runtime's Reference of Value.
Create "Locked" as a New Runtime's Completion of a New Runtime's Acquirer of a New Runtime's Inner's Stderr of Inner.
Create "Output" as a New Runtime's Output of a New Mutable Binding of Locked.
Create "Reference" as a New Storage's Borrower of Bytes.
Create "Failed" as a New Runtime's Fault of a New Runtime's Completion of a New Runtime's Writer of Output and Reference.
If Failed, Return false.
Create "Incomplete" as a New Runtime's Fault of a New Runtime's Completion of a New Runtime's Flusher of Output.
Return a New Equality of Incomplete and false.

    Publicly create Runtime's Terminal's "Copy|--ies|--ied|-ing|--ier" as a Generic `terminal_copy` Function with Inputs of "value" as `&Terminal` and Outputs of Runtime's Terminal.
Create "Status" as a New Runtime's Terminal's Status of Value.
Create "Copied" as a New Storage's Copier of a New Runtime's Terminal's Error of Value.
Return a New Runtime's Terminal's Content of Status and Copied.

    Create Runtime's "Choose|-s|-----ose|--ing|-r" as a Generic `runtime_choose` Function with Inputs of "value" as `&Runtime` and "status" as `u8` and "error" as `Option<Error>` and Outputs of `bool`.
Create "Inner" as a New Runtime's Reference of Value.
Create "Guard" as a New Mutable Binding of a New Locker of a New Runtime's Admission of Inner.
Create "Finalized" as a New Runtime's Finalized of Guard.
Create "Selected" as a New Runtime's Selected of Guard.
Create "Chosen" as a New Runtime's Either of Finalized and Selected.
If Chosen, Return false.
Create "Terminal" as a New Runtime's Terminal's Content of Status and Error.
Create "Wrapped" as a New Present of Terminal.
Create "Target" as a New Runtime's State's Terminal of Guard.
Runtime.Replace Target and Wrapped.
Create "Cancellation" as a New Runtime's Inner's Cancellation of Inner.
Runtime.Notify Cancellation and true.
Return true.

    Publicly create Runtime's "Request|-s|-ed|-ing|-er" as a Generic `runtime_request_exit` Function with Inputs of "value" as `&Runtime` and "status" as `u8` and "error" as `Option<Error>` and Outputs of `()`.
Create "Selected" as a New Runtime's Chooser of Value and Status and Error.
If Selected, Create "Begun" as a New Runtime's Beginner of Value.
Return a New Nothing.

Define Runtime's "Contents" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~&*<@Value@>~|}<DEF>`.

    Create Runtime's "Listen|-s|-ed|-ing|-er" as a Generic Pending `runtime_notice_shutdown` Function with Inputs of "receiver" as `&mut tokio::sync::watch::Receiver<Option<Result<()>>>` and "outcome" as `&mut Option<Result<()>>` and Outputs of `impl std::future::Future<Output = bool> + Send`.
Create "Available" as a New Runtime's Observer of Receiver.
Create "Ready" as a New Presence of Available.
If Ready, Runtime.Replace Outcome and Available and then Return true.
Create "Closed" as a New Runtime's Fault of a New Runtime's Completion of a New Runtime's Changer of Receiver.
If Closed is false, Return false.
Create "Result" as a New Present of a New Failure of a New Runtime's Error's Describer of `runtime shutdown notification closed`.
Runtime.Replace Outcome and Result.
Return true.

    Publicly create Runtime's "Clean|-s|-ed|-ing|-er" as a Generic Pending `runtime_shutdown` Function with Inputs of "value" as `&Runtime` and Outputs of `std::pin::Pin<Box<dyn std::future::Future<Output = Result<()>> + Send + '_>>`.
Create "Begun" as a New Runtime's Beginner of Value.
Create "Inner" as a New Runtime's Reference of Value.
Create "Receiver" as a New Mutable Binding of a New Runtime's Subscriber of a New Runtime's Inner's Notification of Inner.
Create "Outcome" as a New Mutable Binding of a New Absent.
Create "Notification" as a New Storage.Editable Storage's Borrower of Receiver.
Create "Destination" as a New Storage.Editable Storage's Borrower of Outcome.
Runtime.Repeat a New Process for Runtime.Halting a New Runtime's Completion of a New Runtime's Listener of Notification and Destination.
Dispose Receiver.
Create "Optional" as a New Runtime's Taker of Destination.
Return a New Contents of Optional and `completed shutdown`.

    Create Runtime's "Read|-s|-|-ing|-er" as a Generic `runtime_status` Function with Inputs of "value" as `&Runtime` and Outputs of `Option<u8>`.
Create "Guard" as a New Locker of a New Runtime's Admission of a New Runtime's Reference of Value.
Return a New Runtime's State's Status of Guard.

    Create Runtime's "Conclude|-s|-d|--ing|-r" as a Generic `runtime_terminal` Function with Inputs of "value" as `&Runtime` and Outputs of Runtime's Terminal.
Create "Locked" as a New Locker of a New Runtime's Admission of a New Runtime's Reference of Value.
Create "Guard" as a New Mutable Binding of Locked.
Create "Target" as a New Runtime's State's Finalize of Guard.
Runtime.Replace Target and true.
Create "Selected" as a New Runtime's State's Terminal of Guard.
Create "Optional" as a New Storage's Copier of Selected.
Create "Present" as a New Presence of Optional.
If Present, Create "Terminal" as a New Contents of Optional and `selected terminal` and then Return Terminal.
Create "Zero" as a New Runtime's Byte of 0.
Create "Missing" as a New Absent.
Return a New Runtime's Terminal's Content of Zero and Missing.

    Create Runtime's "Remember|-s|-ed|-ing|-er" as a Generic `runtime_final_status` Function with Inputs of "value" as `&Runtime` and "status" as `u8` and Outputs of `u8`.
Create "Locked" as a New Locker of a New Runtime's Admission of a New Runtime's Reference of Value.
Create "Target" as a New Runtime's State's Remember of a New Mutable Binding of Locked.
Create "Wrapped" as a New Present of Status.
Runtime.Replace Target and Wrapped.
Return Status.

    Create Runtime's "Resolve|-s|-d|--ing|-r" as a Generic `runtime_handle_result` Function with Inputs of "value" as `&Runtime` and "result" as `Result<()>` and Outputs of `()`.
Create "Failed" as a New Runtime's Fault of Result.
Create "Unit" as a New Nothing.
If Failed is false, Return Unit.
Create "Error" as a New Runtime's Rejecter of Result.
Create "Cancelled" as a New Runtime's Error's Checker of a New Storage's Borrower of Error.
If Cancelled, Return Unit.
Create "One" as a New Runtime's Byte of 1.
Create "Wrapped" as a New Present of Error.
Create "Requested" as a New Runtime's Requester of Value and One and Wrapped.
Return Unit.

    Create Runtime's "Diagnose|-s|-d|--ing|-r" as a Generic `runtime_diagnostics` Function with Inputs of "primary" as `Option<Error>` and "cleanup" as `Option<Error>` and Outputs of `Option<Vec<u8>>`.
Create "Errors" as a New Runtime's Faults.
Create "First" as a New Presence of Primary.
If First, Create "Cause" as a New Contents of Primary and `present primary error` and then Runtime.Push Errors and Cause.
Create "Second" as a New Presence of Cleanup.
If Second, Create "Failure" as a New Contents of Cleanup and `present cleanup error` and then Runtime.Push Errors and Failure.
Create "Joined" as a New Runtime's Error's Joiner of Errors.
Create "Available" as a New Presence of Joined.
If Available is false, Create "Missing" as a New Absent and then Return Missing.
Create "Rendered" as a New Runtime's Error's Renderer of a New Storage's Borrower of a New Contents of Joined and `present diagnostic`.
Create "Buffer" as a New Mutable Binding of Rendered.
Create "Separated" as a New Runtime's Error's Separator of Buffer.
Return a New Present of Buffer.

    Publicly create Runtime's "Finish|-es|-ed|-ing|-er" as a Generic Pending `runtime_finish` Function with Inputs of "value" as `&Runtime` and "result" as `Result<()>` and Outputs of `std::pin::Pin<Box<dyn std::future::Future<Output = u8> + Send + '_>>`.
Create "Inner" as a New Runtime's Reference of Value.
Create "Guard" as a New Runtime's Completion of a New Runtime's Acquirer of a New Runtime's Inner's Gate of Inner.
Create "Previous" as a New Runtime's Reader of Value.
Create "Finalized" as a New Presence of Previous.
If Finalized, Create "Recorded" as a New Contents of Previous and `finalized status` and then Return Recorded.
Create "Handled" as a New Runtime's Resolver of Value and Result.
Create "Cleanup" as a New Runtime's Failure of a New Runtime's Completion of a New Runtime's Cleaner of Value.
Create "Failed" as a New Presence of Cleanup.
Create "Terminal" as a New Runtime's Concluder of Value.
Create "Status" as a New Mutable Binding of a New Runtime's Terminal's Status of Terminal.
Create "Zero" as a New Runtime's Byte of 0.
Create "One" as a New Runtime's Byte of 1.
Create "Successful" as a New Equality of Status and Zero.
Create "Incomplete" as a New Agreement of Failed and Successful.
If Incomplete, Create "Target" as a New Storage.Editable Storage's Borrower of Status and then Runtime.Replace Target and One.
Create "Stopped" as a New Runtime's Completion of a New Runtime's Queller of Value.
Create "Primary" as a New Storage's Copier of a New Runtime's Terminal's Error of Terminal.
Create "Optional" as a New Runtime's Diagnoser of Primary and Cleanup.
Create "Supplied" as a New Presence of Optional.
If Supplied, Create "Bytes" as a New Contents of Optional and `supplied diagnostic` and then Create "Reporting" as a New Runtime's Reporter of Value and Bytes and then Create "Reported" as a New Runtime's Completion of Reporting and then Create "Outcome" as a New Choice of Reported and Status and One and then Create "Destination" as a New Storage.Editable Storage's Borrower of Status and then Runtime.Replace Destination and Outcome.
Return a New Runtime's Rememberer of Value and Status.

Define Runtime's "Catch|-es|----ught|-ing|-er" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~crate::runtime::host::catch~, Inputs: ~&[<@Future@>]&~|}<DEF>`.
Define Runtime's "Resume|-s|-d|--ing|-r" to `<DEF>MOB{|Key: ~_~|}^KY{|Key: ~std::panic::resume_unwind~, Inputs: ~&[<@Panic@>]&~|}<DEF>`.

    Create Runtime's "Decide|-s|-d|--ing|-r" as a Generic `runtime_selected` Function with Inputs of "value" as `&Runtime` and Outputs of `bool`.
Create "Guard" as a New Locker of a New Runtime's Admission of a New Runtime's Reference of Value.
Return a New Runtime's Selected of Guard.

    Publicly create Runtime's "Run|-s|---an|-ning|-ner" as a Generic Pending `runtime_run` Function with Inputs of "value" as `&Runtime` and "source" as `impl std::future::Future<Output = Result<()>>` and Outputs of `impl std::future::Future<Output = u8>`.
Create "Inner" as a New Runtime's Reference of Value.
Create "Receiver" as a New Mutable Binding of a New Runtime's Subscriber of a New Runtime's Inner's Cancellation of Inner.
Create "Cancelling" as a New Runtime's Waiter of a New Storage.Editable Storage's Borrower of Receiver.
Create "Executing" as a New Runtime's Catcher of Source.
Create "Finished" as a New Runtime's Completion of a New Runtime's Racer of Cancelling and Executing.
Create "Cancelled" as a New Runtime's Valid of Finished.
If Cancelled, Create "Reason" as a New Runtime's Error's Canceller and then Create "Rejected" as a New Failure of Reason and then Create "Cleaning" as a New Runtime's Finisher of Value and Rejected and then Create "Interrupted" as a New Runtime's Completion of Cleaning and then Return Interrupted.
Create "Result" as a New Runtime's Rejecter of Finished.
Create "Panicked" as a New Runtime's Fault of Result.
If Panicked is false, Create "Completed" as a New Contents of Result and `completed source execution` and then Create "Settling" as a New Runtime's Finisher of Value and Completed and then Create "Settled" as a New Runtime's Completion of Settling and then Return Settled.
Create "Panic" as a New Runtime's Rejecter of Result.
Create "Succeeded" as a New Success of a New Nothing.
Create "Status" as a New Runtime's Completion of a New Runtime's Finisher of Value and Succeeded.
Create "Zero" as a New Runtime's Byte of 0.
If Status isnt Zero, Return Status.
Create "Selected" as a New Runtime's Decider of Value.
If Selected, Return Status.
Runtime.Resume Panic.
Return Status.

    Create Runtime's "Gather|-s|-ed|-ing|-er" as a Generic `runtime_owned` Function with Inputs of "value" as `&Runtime` and Outputs of `Option<Vec<std::sync::Arc<Invocation>>>`.
Create "Inner" as a New Runtime's Reference of Value.
Create "Guard" as a New Mutable Binding of a New Locker of a New Runtime's Admission of Inner.
Create "Stopping" as a New Runtime's Stopping of Guard.
If Stopping, Create "Missing" as a New Absent and then Return Missing.
Create "Target" as a New Runtime's State's Stop of Guard.
Runtime.Replace Target and true.
Create "Signal" as a New Runtime's Inner's Cancellation of Inner.
Runtime.Notify Signal and true.
Create "Owned" as a New Runtime's Collection of a New Runtime's Cloned of a New Runtime's State's Owned of Guard.
Return a New Present of Owned.

    Create Runtime's "Unsubscribe|-s|-d|--ing|-r" as a Generic `runtime_signals` Function with Inputs of "value" as `&Runtime` and Outputs of `Option<tokio::task::JoinHandle<()>>`.
Create "Inner" as a New Runtime's Reference of Value.
Create "Guard" as a New Mutable Binding of a New Locker of a New Runtime's Inner's Signals of Inner.
Return a New Runtime's Taker of Guard.

    Create Runtime's "Quell|-s|-ed|-ing|-er" as a Generic Pending `runtime_quell` Function with Inputs of "value" as `&Runtime` and Outputs of `std::pin::Pin<Box<dyn std::future::Future<Output = ()> + Send + '_>>`.
Create "Optional" as a New Runtime's Unsubscriber of Value.
Create "Available" as a New Presence of Optional.
If Available, Create "Task" as a New Contents of Optional and `present signal task` and then Runtime.Abort Task and then Create "Stopped" as a New Runtime's Completion of Task.
Return a New Nothing.

    Create Runtime's "Assess|-es|-ed|-ing|-or" as a Generic `runtime_cleanup` Function with Inputs of "value" as `&Runtime` and Outputs of `Result<()>`.
Create "Guard" as a New Locker of a New Runtime's Admission of a New Runtime's Reference of Value.
Create "Errors" as a New Runtime's Collection of a New Runtime's Cloned of a New Runtime's State's Errors of Guard.
Create "Joined" as a New Runtime's Error's Joiner of Errors.
Create "Available" as a New Presence of Joined.
If Available, Create "Reason" as a New Contents of Joined and `present cleanup error` and then Create "Failed" as a New Failure of Reason and then Return Failed.
Return a New Success of a New Nothing.

    Create Runtime's Invocation's "Settle|-s|-d|--ing|-r" as a Generic Pending `invocation_settle` Function with Inputs of "value" as `&std::sync::Arc<Invocation>` and Outputs of `std::pin::Pin<Box<dyn std::future::Future<Output = ()> + Send + '_>>`.
Create "Terminated" as a New Runtime's Completion of a New Runtime's Invocation's Terminator of Value and false.
Create "Failed" as a New Runtime's Fault of Terminated.
Create "Unit" as a New Nothing.
If Failed is false, Return Unit.
Create "Reason" as a New Runtime's Rejecter of Terminated.
Create "Errors" as a New Runtime's Faults.
Runtime.Push Errors and Reason.
Create "Killed" as a New Runtime's Completion of a New Runtime's Invocation's Killer of Value.
Create "Remaining" as a New Runtime's Fault of Killed.
If Remaining, Create "Second" as a New Runtime's Rejecter of Killed and then Runtime.Push Errors and Second.
Create "Joined" as a New Runtime's Error's Joiner of Errors.
Create "Available" as a New Presence of Joined.
If Available, Create "Combined" as a New Contents of Joined and `present cleanup error` and then Create "Recorded" as a New Runtime's Invocation's Recorder of Value and Combined.
Return Unit.

    Create Runtime's "Reap|-s|-ed|-ing|-er" as a Generic Pending `runtime_reap` Function with Inputs of "value" as `Runtime` and "owned" as `Vec<std::sync::Arc<Invocation>>` and Outputs of `std::pin::Pin<Box<dyn std::future::Future<Output = ()> + Send>>`.
Create "Invocations" as a New Storage's Borrower of Owned.
Runtime.Join Invocations and Runtime's Invocation's Settler.
Runtime.Join Invocations and Runtime's Invocation's Completer.
Create "Borrowed" as a New Storage's Borrower of Value.
Create "Result" as a New Runtime's Assessor of Borrowed.
Create "Stopped" as a New Runtime's Completion of a New Runtime's Queller of Borrowed.
Create "Inner" as a New Runtime's Reference of Value.
Create "Notification" as a New Runtime's Inner's Notification of Inner.
Create "Wrapped" as a New Present of Result.
Runtime.Notify Notification and Wrapped.
Return a New Nothing.

    Publicly create Runtime's "Begin|-s|---an|-ning|-ner" as a Generic `runtime_begin_shutdown` Function with Inputs of "value" as `&Runtime` and Outputs of `()`.
Create "Optional" as a New Runtime's Gatherer of Value.
Create "Available" as a New Presence of Optional.
Create "Unit" as a New Nothing.
If Available is false, Return Unit.
Create "Owned" as a New Contents of Optional and `selected shutdown owner`.
Create "Handle" as a New Runtime's Copier of Value.
Create "Spawned" as a New Runtime's Spawner of a New Runtime's Reaper of Handle and Owned.
Return Unit.
