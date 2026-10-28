Publicly define Runtime's "Capture" as `<DEF>HY{|Key: ~Capture~, Props: ~&[Prop{Key: bytes, Type: Vec<u8>}Prop, Prop{Key: error, Type: Option<Error>}Prop]&~|}<DEF>`.
Define Runtime's Capture's "Content" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^Y{|Key: ~Capture~, Props: ~&[Prop{Key: bytes, Value: <@Bytes@>}Prop, Prop{Key: error, Value: <@Error@>}Prop]&~|}<DEF>`.
Define Runtime's Capture's "Readable" as a New Adjective describing any Function.
It describes any Function as `<DEF>FW{|Func{Key: <@Name|Text@>, Inputs: <@Inputs|Props@>, Outputs: <@Outputs|Texts@>}Func, GenericTypes: ~R: tokio::io::AsyncRead + Unpin + Send + 'static~|}^(<@Process@>)<DEF>`.
Define Runtime's Capture's "Channel" as `<DEF>OB{|Key: ~<@Name|Label@>: Box<dyn tokio::io::AsyncRead + Unpin + Send>~, Value: ~<@Value@>~|}<DEF>`.
Define Runtime's Capture's "Box" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~Box::new~, Inputs: ~&[<@Value@>]&~|}<DEF>`.
Define Runtime's Capture's "Destination" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~capture~, Receiver: ~<@Dependencies@>~, Inputs: ~&[<@Stream@>, <@Bytes@>]&~|}<DEF>`.
Define Runtime's Capture's "Borrow" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~Box::new~, Inputs: ~&[&mut <@Value@>]&~|}<DEF>`.
Define Runtime's Capture's "Buffer" as `<DEF>OB{|Key: ~mut <@Name|Label@>~, Value: ~[0; 8192]~|}<DEF>`.
Define Runtime's Capture's "Fill|-s|-ed|-ing|-er" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~tokio::io::AsyncReadExt::read~, Inputs: ~&[<@Reader@>, <@Buffer@>]&~|}<DEF>`.
Define Runtime's Capture's "Write|-s|----ote|--ing|-r" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~std::io::Write::write_all~, Inputs: ~&[<@Destination@>, &<@Buffer@>[..<@Count@>]]&~|}<DEF>`.
Define Runtime's Capture's "Sleep" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~tokio::time::sleep_until~, Inputs: ~&[<@Until@>]&~|}<DEF>`.
Define Runtime's Capture's "Wait" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~std::future::pending~, GenericTypes: ~()~|}<DEF>`.
Define Runtime's Capture's "Fault" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~is_err~, Receiver: ~<@Value@>~|}<DEF>`.
Define Runtime's Capture's "Success" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~ok~, Receiver: ~<@Value@>~|}<DEF>`.
Define Runtime's Capture's "Bytes" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~<@Value@>.bytes~|}<DEF>`.
Define Runtime's Capture's "Error" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~<@Value@>.error~|}<DEF>`.
Define Runtime's Capture's "Shared" as `<DEF>MOB{|Key: ~<@Name|Label@>~, Type: ~std::sync::Arc<[u8]>~|}^KY{|Key: ~std::sync::Arc::from~, Inputs: ~&[<@Value@>]&~|}<DEF>`.
Define Runtime's Capture's "Description" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~format!~, Inputs: ~&["capture application {} task: {}", <@Stream@>, <@Error@>]&~|}<DEF>`.

    Create Runtime's Capture's "Adapt|-s|-ed|-ing|-er" as a Generic Runtime.Capture.Readable Function with Inputs of "reader" as `Option<R>` and Outputs of `Option<Box<dyn tokio::io::AsyncRead + Unpin + Send>>`.
Create "Available" as a New Presence of Reader.
If Available, Create "Value" as a New Contents of Reader and `present capture reader` and then Create "Boxed" as a New Runtime's Capture's Box of Value and then Create "Interface" as a New Runtime's Capture's Channel of Boxed and then Create "Wrapped" as a New Present of Interface and then Return Wrapped.
Return a New Absent.

    Create Runtime's Capture's "Fail|-s|-ed|-ing|-er" as a Generic Function with Inputs of "errors" as `&mut Option<Error>` and "cause" as `Error` and Outputs of `bool`.
Create "Wrapped" as a New Present of Cause.
Runtime.Replace Errors and Wrapped.
Return true.

    Create Runtime's Capture's "Reject|-s|-ed|-ing|-er" as a Generic Function with Inputs of "errors" as `&mut Option<Error>` and "message" as `&str` and Outputs of `bool`.
Create "Error" as a New Runtime's Error's Describer of Message.
Return a New Runtime's Capture's Failer of Errors and Error.

    Create Runtime's Capture's "Receive|-s|-d|--ing|-r" as a Generic Function with Inputs of "destination" as `&mut (dyn std::io::Write + Send)` and "buffer" as `&[u8]` and "errors" as `&mut Option<Error>` and "read" as `std::io::Result<usize>` and Outputs of `bool`.
Create "Readfailed" as a New Runtime's Capture's Fault of Read.
If Readfailed, Create "Failure" as a New Runtime's Failure of Read and then Create "Error" as a New Contents of Failure and `failed stream read` and then Create "Context" as a New Runtime's Error's Explainer of `capture application stream` and Error and then Create "Complete" as a New Runtime's Capture's Failer of Errors and Context and then Return Complete.
Create "Count" as a New Runtime's Capture's Success of Read.
Create "Size" as a New Contents of Count and `successful stream read`.
Create "Zero" as a New Indexed Integer of 0.
If Size is Zero, Return true.
Create "Written" as a New Runtime's Capture's Writer of Destination, Buffer, and Size.
Create "Writefailed" as a New Runtime's Capture's Fault of Written.
If Writefailed, Create "Writefailure" as a New Runtime's Failure of Written and then Create "Writeerror" as a New Contents of Writefailure and `failed stream write` and then Create "Writecontext" as a New Runtime's Error's Explainer of `capture application stream` and Writeerror and then Create "Finished" as a New Runtime's Capture's Failer of Errors and Writecontext and then Return Finished.
Return false.

    Create Runtime's Capture's "Timeout" as a Generic Pending Function with Inputs of "until" as `Option<tokio::time::Instant>` and Outputs of `impl std::future::Future<Output=()> + Send`.
Create "Available" as a New Presence of Until.
If Available, Create "Instant" as a New Contents of Until and `present cleanup deadline` and then Create "Sleep" as a New Runtime's Capture's Sleep of Instant and then Create "Elapsed" as a New Runtime's Completion of Sleep and then Return Elapsed.
Return a New Runtime's Completion of a New Runtime's Capture's Wait.

    Create Runtime's Capture's "Observe|-s|-d|--ing|--er" as a Generic Function with Inputs of "errors" as `&mut Option<Error>` and "notice" as `std::result::Result<(), tokio::sync::watch::error::RecvError>` and Outputs of `bool`.
Create "Failed" as a New Runtime's Capture's Fault of Notice.
If Failed, Create "Complete" as a New Runtime's Capture's Rejecter of Errors and `capture deadline notification closed` and then Return Complete.
Return false.

    Create Runtime's Capture's "Advance|-s|-d|--ing|--er" as a Generic Pending Function with Inputs of "reader" as `&mut (dyn tokio::io::AsyncRead + Unpin + Send)` and "deadline" as `&mut tokio::sync::watch::Receiver<Option<tokio::time::Instant>>` and "destination" as `&mut (dyn std::io::Write + Send)` and "buffer" as `&mut [u8; 8192]` and "errors" as `&mut Option<Error>` and Outputs of `impl std::future::Future<Output=bool> + Send`.
Create "Timeout" as a New Runtime's Capture's Timeout of a New Runtime's Observer of Deadline.
Create "Notification" as a New Runtime's Changer of Deadline.
Create "Reading" as a New Runtime's Capture's Filler of Reader and Buffer.
Create "Pendingupdates" as a New Runtime's Racer of Notification and Reading.
Create "Outcome" as a New Runtime's Completion of a New Runtime's Racer of Timeout and Pendingupdates.
Create "Expired" as a New Runtime's Valid of Outcome.
If Expired, Create "Expiredresult" as a New Runtime's Capture's Rejecter of Errors and `application output pipe remained open beyond cleanup deadline` and then Return Expiredresult.
Create "Changed" as a New Runtime's Rejecter of Outcome.
Create "Notified" as a New Runtime's Valid of Changed.
If Notified, Create "Notice" as a New Runtime's Capture's Success of Changed and then Create "Noticeoutcome" as a New Contents of Notice and `ready deadline change` and then Create "Observed" as a New Runtime's Capture's Observer of Errors and Noticeoutcome and then Return Observed.
Create "Result" as a New Runtime's Rejecter of Changed.
Return a New Runtime's Capture's Receiver of Destination, Buffer, Errors, and Result.

    Create Runtime's Capture's "Start|-s|-ed|-ing|-er" as a Generic Pending Function with Inputs of "reader" as `Option<Box<dyn tokio::io::AsyncRead + Unpin + Send>>` and "deadline" as `tokio::sync::watch::Receiver<Option<tokio::time::Instant>>` and "dependencies" as `std::sync::Arc<dyn crate::runtime::host::Dependencies>` and "stream" as `crate::runtime::host::Stream` and Outputs of `impl std::future::Future<Output=Capture> + Send`.
Create "Available" as a New Presence of Reader.
If Available is false, Create "Unavailable" as a New Runtime's Error's Message and then Create "Faulttext" as a New Runtime's Error's Describer of `application capture pipe is unavailable` and then Create "Fault" as a New Present of Faulttext and then Create "Missing" as a New Runtime's Capture's Content of Unavailable and Fault and then Return Missing.
Create "Input" as a New Mutable Binding of a New Contents of Reader and `present capture reader`.
Create "Notifications" as a New Mutable Binding of Deadline.
Create "Bytes" as a New Mutable Binding of a New Runtime's Error's Message.
Create "Borrowed" as a New Runtime's Capture's Borrow of Bytes.
Create "Destination" as a New Mutable Binding of a New Runtime's Capture's Destination of Dependencies, Stream, and Borrowed.
Create "Buffer" as a New Runtime's Capture's Buffer.
Create "Failure" as a New Mutable Binding of a New Absent.
Create "Source" as a New Runtime's Mutable of Input.
Create "Schedule" as a New Storage.Editable Storage's Borrower of Notifications.
Create "Writer" as a New Runtime's Mutable of Destination.
Create "Chunk" as a New Storage.Editable Storage's Borrower of Buffer.
Create "Errors" as a New Storage.Editable Storage's Borrower of Failure.
Runtime.Repeat a New Process for Runtime.Halting a New Runtime's Completion of
 a New Runtime's Capture's Advancer of Source, Schedule, Writer, Chunk, and Errors.
Dispose Input.
Dispose Notifications.
Dispose Destination.
Return a New Runtime's Capture's Content of Bytes and Failure.

    Publicly create Runtime's Capture's "Read|-s||-ing|-er" as a Generic Runtime.Capture.Readable `capture` Function with Inputs of "reader" as `Option<R>` and "deadline" as `tokio::sync::watch::Receiver<Option<tokio::time::Instant>>` and "dependencies" as `std::sync::Arc<dyn crate::runtime::host::Dependencies>` and "stream" as `crate::runtime::host::Stream` and Outputs of `impl std::future::Future<Output=Capture> + Send`.
Create "Adapted" as a New Runtime's Capture's Adapter of Reader.
Return a New Runtime's Capture's Starter of Adapted, Deadline, Dependencies, and Stream.

    Publicly create Runtime's Capture's "Join|-s|-ed|-ing|-er" as a Generic `captured` Function with Inputs of "result" as `std::result::Result<Capture, tokio::task::JoinError>` and "stream" as `&str` and Outputs of `(std::sync::Arc<[u8]>, Option<Error>)`.
Create "Failed" as a New Runtime's Capture's Fault of Result.
If Failed, Create "Failure" as a New Runtime's Failure of Result and then Create "Cause" as a New Contents of Failure and `failed capture task` and then Create "Description" as a New Runtime's Capture's Description of Stream and Cause and then Create "Message" as a New Encoding of Description and then Create "Fault" as a New Runtime's Error's Former of Message and then Create "Optionalfault" as a New Present of Fault and then Create "Empty" as a New Runtime's Error's Message and then Create "Missing" as a New Runtime's Capture's Shared of Empty and then Create "Rejected" as a New Pair of Missing and Optionalfault and then Return Rejected.
Create "Optional" as a New Runtime's Capture's Success of Result.
Create "Capture" as a New Contents of Optional and `successful capture task`.
Create "Bytes" as a New Runtime's Capture's Bytes of Capture.
Create "Error" as a New Runtime's Capture's Error of Capture.
Create "Shared" as a New Runtime's Capture's Shared of Bytes.
Return a New Pair of Shared and Error.
