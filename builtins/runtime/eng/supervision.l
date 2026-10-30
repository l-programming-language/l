Define Runtime's "Supervision" as a New Dictionary.
Define Runtime's Supervision's "Waiter" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~wait~, Receiver: ~<@Dependencies@>~, Inputs: ~&[<@Child@>]&~|}<DEF>`.
Define Runtime's Supervision's "Close|-s|-d|--ing|-r" to `<DEF>MOB{|Key: ~_~|}^KY{|Key: ~close~, Receiver: ~<@Value@>~|}<DEF>`.
Define Runtime's Supervision's "Read|-s|-|-ing|-er" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~try_recv~, Receiver: ~<@Value@>~|}<DEF>`.
Define Runtime's Supervision's "Signal|-s|-ed|-ing|-er" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~crate::process::terminate_leader~, Inputs: ~&[<@Pid@>]&~|}<DEF>`.
Define Runtime's Supervision's "Discard|-s|-ed|-ing|-er" to `<DEF>MOB{|Key: ~_~|}^KY{|Key: ~send~, Receiver: ~<@Reply@>~, Inputs: ~&[<@Result@>]&~|}<DEF>`.

    Create Runtime's Supervision's "Explain|-s|-ed|-ing|-er" as a Generic `supervision_signal_error` Function with Inputs of "error" as `std::io::Error` and Outputs of Runtime's Error.
Return a New Runtime's Error's Explainer of `terminate application leader` and Error.

    Create Runtime's Supervision's "Terminate|-s|-d|--ing|--or" as a Generic `supervision_terminate_leader` Function with Inputs of "pid" as `Option<u32>` and Outputs of `Result<()>`.
Create "Available" as a New Presence of Pid.
If Available, Create "Leader" as a New Contents of Pid and `supervised child pid` and then Create "Signalled" as a New Runtime's Supervision's Signaler of Leader and then Create "Result" as a New Runtime's Explainer of Signalled and Runtime's Supervision's Explainer and then Return Result.
Return a New Success of a New Nothing.

    Create Runtime's Supervision's "Poll|-s|-ed|-ing|-er" as a Generic Pending `supervision_poll` Function with Inputs of "wait" as `&mut crate::runtime::host::Wait<'_>` and "requests" as `&mut tokio::sync::mpsc::UnboundedReceiver<tokio::sync::oneshot::Sender<Result<()>>>` and "pid" as `Option<u32>` and "result" as `&mut Option<std::io::Result<std::process::ExitStatus>>` and Outputs of `impl std::future::Future<Output = bool> + Send`.
Create "Incoming" as a New Runtime's Receiver of Requests.
Create "Borrowed" as a New Storage.Editable Storage's Reference of Wait.
Create "Ready" as a New Runtime's Completion of a New Runtime's Racer of Borrowed and Incoming.
Create "Waited" as a New Runtime's Valid of Ready.
If Waited, Create "Primary" as a New Storage's Optional of Ready and then Runtime.Replace Result and Primary and then Return true.
Create "Received" as a New Runtime's Rejecter of Ready.
Create "Available" as a New Presence of Received.
If Available, Create "Reply" as a New Contents of Received and `pending leader request` and then Create "Replyresult" as a New Runtime's Supervision's Terminator of Pid and then Runtime.Supervision.Discard Reply and Replyresult and then Return false.
Create "Wrapped" as a New Present of a New Runtime's Completion of Wait.
Runtime.Replace Result and Wrapped.
Return true.

    Create Runtime's Supervision's "Release|-s|-d|--ing|-r" as a Generic `supervision_release_reply` Function with Inputs of "requests" as `&mut tokio::sync::mpsc::UnboundedReceiver<tokio::sync::oneshot::Sender<Result<()>>>` and Outputs of `bool`.
Create "Received" as a New Runtime's Supervision's Reader of Requests.
Create "Failed" as a New Runtime's Fault of Received.
If Failed, Return true.
Create "Reply" as a New Contents of Received and `queued leader request`.
Create "Accepted" as a New Success of a New Nothing.
Runtime.Supervision.Discard Reply and Accepted.
Return false.

    Create Runtime's Supervision's "Await|-s|-ed|-ing|-er" as a Generic Pending `supervision_wait` Function with Inputs of "wait" as `crate::runtime::host::Wait<'_>` and "requests" as `tokio::sync::mpsc::UnboundedReceiver<tokio::sync::oneshot::Sender<Result<()>>>` and "pid" as `Option<u32>` and Outputs of `impl std::future::Future<Output = std::io::Result<std::process::ExitStatus>> + Send`.
Create "Waiting" as a New Mutable Binding of Wait.
Create "Replies" as a New Mutable Binding of Requests.
Create "Outcome" as a New Mutable Binding of a New Absent.
Create "Waitingref" as a New Storage.Editable Storage's Borrower of Waiting.
Create "Incoming" as a New Storage.Editable Storage's Borrower of Replies.
Create "Result" as a New Storage.Editable Storage's Borrower of Outcome.
Runtime.Repeat a New Process for Runtime.Halting a New Runtime's Completion of
 a New Runtime's Supervision's Poller of Waitingref, Incoming, Pid, and Result.
Runtime.Supervision.Close Incoming.
Runtime.Repeat a New Process for Runtime.Halting a New Runtime's Supervision's Releaser of Incoming.
Create "Optional" as a New Runtime's Taker of Result.
Return a New Contents of Optional and `completed supervisor wait`.

Define Runtime's Supervision's "Stdout" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~take~, Receiver: ~<@Child@>.stdout~|}<DEF>`.
Define Runtime's Supervision's "Stderr" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~take~, Receiver: ~<@Child@>.stderr~|}<DEF>`.
Define Runtime's Supervision's "Output" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~crate::runtime::host::Stream::Stdout~|}<DEF>`.
Define Runtime's Supervision's "Error" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~crate::runtime::host::Stream::Stderr~|}<DEF>`.
Define Runtime's Supervision's "Identity" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~id~, Receiver: ~<@Child@>~|}<DEF>`.
Define Runtime's Supervision's "Success" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~success~, Receiver: ~<@Status@>~|}<DEF>`.
Define Runtime's Supervision's "Description" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~format!~, Inputs: ~&["application {}", <@Status@>]&~|}<DEF>`.
Define Runtime's Supervision's "Code" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~crate::process::status_code~, Inputs: ~&[<@Status@>]&~|}<DEF>`.
Define Runtime's Supervision's "Extend|-s|-ed|-ing|-er" to `<DEF>MOB{|Key: ~_~|}^KY{|Key: ~extend~, Receiver: ~<@Values@>~, Inputs: ~&[<@Value@>]&~|}<DEF>`.
Define Runtime's Supervision's "Owned" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~&mut <@Value@>.owned~|}<DEF>`.
Define Runtime's Supervision's "Remove|-s|-d|--ing|-r" to `<DEF>MOB{|Key: ~_~|}^KY{|Key: ~remove~, Receiver: ~<@Values@>~, Inputs: ~&[<@Sequence@>]&~|}<DEF>`.

    Create Runtime's Supervision's "Inspect|-s|-ed|-ing|-or" as a Generic `supervision_status` Function with Inputs of "waited" as `std::io::Result<std::process::ExitStatus>` and "errors" as `&mut Vec<Error>` and Outputs of `Option<i32>`.
Create "Failed" as a New Runtime's Fault of Waited.
If Failed, Create "Cause" as a New Runtime's Rejecter of Waited and then Create "Fault" as a New Runtime's Error's Explainer of `wait for application` and Cause and then Runtime.Push Errors and Fault and then Create "Missing" as a New Absent and then Return Missing.
Create "Status" as a New Contents of Waited and `completed application status`.
Create "Success" as a New Runtime's Supervision's Success of Status.
If Success is false, Create "Description" as a New Runtime's Supervision's Description of Status and then Create "Text" as a New Runtime's Text of Description and then Create "Failure" as a New Runtime's Error's Describer of Text and then Runtime.Push Errors and Failure.
Return a New Runtime's Supervision's Code of Status.

    Create Runtime's Supervision's "Complete|-s|-d|--ing|--er" as a Generic `supervision_finalize` Function with Inputs of "value" as `&std::sync::Arc<Invocation>` and "owner" as `&Runtime` and "waited" as `std::io::Result<std::process::ExitStatus>` and "stdout" as `std::sync::Arc<[u8]>` and "stderr" as `std::sync::Arc<[u8]>` and "capture" as `Option<Error>` and "stream" as `Option<Error>` and "cleanup" as `Option<Error>` and Outputs of `()`.
Create "Errors" as a New Runtime's Faults.
Create "Reference" as a New Storage.Editable Storage's Borrower of Errors.
Create "Status" as a New Runtime's Supervision's Inspector of Waited and Reference.
Runtime.Supervision.Extend Errors and Capture.
Runtime.Supervision.Extend Errors and Stream.
Runtime.Supervision.Extend Errors and Cleanup.
Create "Joined" as a New Runtime's Error's Joiner of Errors.
Create "Borrowedstderr" as a New Storage's Borrower of Stderr.
Create "Error" as a New Runtime's Supervision's Decorator of Joined and Borrowedstderr.
Create "Snapshot" as a New Runtime's Snapshot's Content of Stdout, Stderr, Error, and Status.
Create "Published" as a New Runtime's Invocation's Publisher of Value and Snapshot.
Create "Locked" as a New Locker of a New Runtime's Admission of a New Runtime's Reference of Owner.
Create "Owned" as a New Runtime's Supervision's Owned of a New Mutable Binding of Locked.
Create "Borrowed" as a New Storage's Borrower of a New Runtime's Invocation's Sequence of Value.
Runtime.Supervision.Remove Owned and Borrowed.
Return a New Nothing.

    Create Runtime's Supervision's "Decorate|-s|-d|--ing|--or" as a Generic `supervision_stderr` Function with Inputs of "error" as `Option<Error>` and "bytes" as `&[u8]` and Outputs of `Option<Error>`.
Create "Available" as a New Presence of Error.
If Available is false, Return Error.
Create "Fault" as a New Contents of Error and `joined invocation error`.
Return a New Present of a New Runtime's Error's Attacher of Fault and Bytes.

    Publicly create Runtime's Invocation's "Supervise|-s|-d|--ing|-r" as a Generic Pending `invocation_supervise` Function with Inputs of "value" as `std::sync::Arc<Invocation>` and "process" as `tokio::process::Child` and "requests" as `tokio::sync::mpsc::UnboundedReceiver<tokio::sync::oneshot::Sender<Result<()>>>` and "owner" as `Runtime` and Outputs of `impl std::future::Future<Output = ()> + Send`.
Create "Child" as a New Mutable Binding of Process.
Create "Stdout" as a New Runtime's Supervision's Stdout of Child.
Create "Stderr" as a New Runtime's Supervision's Stderr of Child.
Create "Reference" as a New Storage's Borrower of Value.
Create "Out" as a New Runtime's Capture's Follower of Reference, Stdout, and a New Runtime's Supervision's Output.
Create "Err" as a New Runtime's Capture's Follower of Reference, Stderr, and a New Runtime's Supervision's Error.
Create "Dependencies" as a New Runtime's Invocation's Dependencies of Value.
Create "Pid" as a New Runtime's Supervision's Identity of Child.
Create "Mutableref" as a New Storage.Editable Storage's Borrower of Child.
Create "Wait" as a New Runtime's Supervision's Waiter of Dependencies and Mutableref.
Create "Waited" as a New Runtime's Completion of a New Runtime's Supervision's Awaiter of Wait, Requests, and Pid.
Create "Now" as a New Runtime's Now.
Create "Bounded" as a New Runtime's Invocation's Bounder of Reference and a New Sum of Now and a New Runtime's Duration of a New Unsigned Integer of 2.
Create "Outcompleted" as a New Runtime's Completion of Out.
Create "Errcompleted" as a New Runtime's Completion of Err.
Create "Outputcapture" as a New Runtime's Capture's Joiner of Outcompleted and `stdout`.
Create "Errorcapture" as a New Runtime's Capture's Joiner of Errcompleted and `stderr`.
Create "Bytes" as a New First of Outputcapture.
Create "Outerror" as a New Second of Outputcapture.
Create "Errors" as a New First of Errorcapture.
Create "Errerror" as a New Second of Errorcapture.
Create "Cleanup" as a New Runtime's Failure of a New Runtime's Completion of a New Runtime's Invocation's Terminator of Reference and true.
Create "Available" as a New Presence of Cleanup.
If Available, Create "Cleanupref" as a New Storage's Referrer of Cleanup and then Create "Fault" as a New Contents of Cleanupref and `present cleanup error` and then Create "Copied" as a New Runtime's Error's Copier of Fault and then Create "Recorded" as a New Runtime's Invocation's Recorder of Reference and Copied.
Create "Retained" as a New Storage's Borrower of Owner.
Create "Finished" as a New Runtime's Supervision's Completer of Reference, Retained, Waited, Bytes, Errors, Outerror, Errerror, and Cleanup.
Return a New Nothing.
