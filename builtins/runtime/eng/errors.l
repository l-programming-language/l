Use Storage from `storage`.
Publicly define Runtime's "Result" as `<DEF>E{|Statement: ~pub type Result<T> = std::result::Result<T, Error>;~|}<DEF>`.
Publicly define Runtime's "Error" as `<DEF>HY{|Key: ~Error~, Props: ~&[Prop{Key: data, Type: std::sync::Arc<ErrorData>}Prop]&~|}<DEF>`.
Define Runtime's Error's "Data" as `<DEF>HY{|Key: ~ErrorData~, Props: ~&[Prop{Key: message, Type: Vec<u8>}Prop, Prop{Key: causes, Type: Vec<Error>}Prop, Prop{Key: cancelled, Type: bool}Prop, Prop{Key: native, Type: Option<std::sync::Arc<dyn std::error::Error + Send + Sync>>}Prop]&~|}<DEF>`.
Define Runtime's Error's Data's "Content" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^Y{|Key: ~ErrorData~, Props: ~&[Prop{Key: message, Value: <@Message@>}Prop, Prop{Key: causes, Value: <@Causes@>}Prop, Prop{Key: cancelled, Value: <@Cancelled@>}Prop, Prop{Key: native, Value: <@Native@>}Prop]&~|}<DEF>`.
Define Runtime's Error's "Fault" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^Y{|Key: ~Error~, Props: ~&[Prop{Key: data, Value: <@Data@>}Prop]&~|}<DEF>`.
Define Runtime's Error's "Causes" as `<DEF>MOB{|Key: ~<@Name|Label@>~, Type: ~Vec<Error>~|}^KY{|Key: ~Vec::new~|}<DEF>`.
Define Runtime's Error's "Message" as `<DEF>MOB{|Key: ~<@Name|Label@>~, Type: ~Vec<u8>~|}^KY{|Key: ~Vec::new~|}<DEF>`.
Define Runtime's Error's "Cancellation" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~<@Value|Truth@>~|}<DEF>`.
Define Runtime's Error's "Status" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~<@Value@>.data.cancelled~|}<DEF>`.
Define Runtime's Error's "Iterator" as `<DEF>MOB{|Key: ~mut <@Name|Label@>~|}^KY{|Key: ~into_iter~, Receiver: ~<@Value@>~|}<DEF>`.
Define Runtime's Error's "First" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~next~, Receiver: ~<@Value@>~|}<DEF>`.
Define Runtime's Error's "Bytes" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~as_slice~, Receiver: ~<@Value@>.data.message~|}<DEF>`.
Define Runtime's Error's "Datum" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~&<@Value@>.data~|}<DEF>`.
Define Runtime's Error's "Address" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~std::sync::Arc::as_ptr~, Inputs: ~&[&<@Value@>.data]&~|}<DEF>`.
Define Runtime's Error's "Children" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~as_slice~, Receiver: ~<@Value@>.data.causes~|}<DEF>`.

    Create Runtime's Error's "Wrap|-s|-ped|-ping|-per" as a Generic `error_wrap` Function with Inputs of "message" as `Vec<u8>` and "causes" as `Vec<Error>` and "cancelled" as `bool` and "native" as `Option<std::sync::Arc<dyn std::error::Error + Send + Sync>>` and Outputs of Runtime's Error.
Return a New Runtime's Error's Fault of a New Runtime's Shared of a New Runtime's Error's Data's Content of Message and Causes and Cancelled and Native.

Define Runtime's Error's "Caused" as a New Adjective describing any Storage's Routine.
It describes a Generic Storage's Routine of `E: std::error::Error + Send + Sync + 'static`.
Define Runtime's Error's "Interface" as `<DEF>DOB{|Key: ~<@Name|Label@>~|}^DE{|Ops: ~&[Op{Type: Calculate, Operator: as, Operands: &[<@Value@>, std::sync::Arc<dyn std::error::Error + Send + Sync>]&}Op]&~|}<DEF>`.
Define Runtime's Error's "Context" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~format!~, Inputs: ~&["{}: {}", <@Message@>, <@Error@>]&~|}<DEF>`.
Define Runtime's Error's "Search|-es|-ed|-ing|-er" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~find_map~, Receiver: ~<@Iterator@>~, Inputs: ~&[<@Action@>]&~|}<DEF>`.
Define Runtime's Error's "Native" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~as_ref~, Receiver: ~<@Value@>.data.native~|}<DEF>`.
Define Runtime's Error's "Source" as `<DEF>DOB{|Key: ~<@Name|Label@>~|}^DE{|Ops: ~&[Op{Type: Calculate, Operator: as, Operands: &[<@Value@>, &(dyn std::error::Error + 'static)]&}Op]&~|}<DEF>`.
Define Runtime's Error's "Lossy" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~String::from_utf8_lossy~, Inputs: ~&[&<@Value@>]&~|}<DEF>`.
Define Runtime's Error's "Punctuation" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~b": "~|}<DEF>`.
Define Runtime's Error's "Parent" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~Vec::from~, Inputs: ~&[[<@Value@>]]&~|}<DEF>`.
Define Runtime's Error's "Write|-s|----ote|--ing|--er" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~write_str~, Receiver: ~<@Formatter@>~, Inputs: ~&[&<@Text@>]&~|}<DEF>`.

    Publicly create Runtime's Error's "Explain|-s|-ed|-ing|-er" as a Generic Runtime.Error.Caused `error_context` Storage.Routine with Inputs of "message" as `&str` and "error" as `E` and Outputs of Runtime's Error.
Create "Bytes" as a New Encoding of a New Runtime's Error's Context of Message and Error.
Create "Native" as a New Present of a New Runtime's Error's Interface of a New Runtime's Shared of Error.
Create "Causes" as a New Runtime's Error's Causes.
Create "Cancelled" as a New Runtime's Error's Cancellation of false.
Return a New Runtime's Error's Wrapper of Bytes and Causes and Cancelled and Native.

    Publicly create Runtime's Error's "Import|-s|-ed|-ing|-er" as a Generic `error_from_io` Function with Inputs of "error" as `std::io::Error` and Outputs of Runtime's Error.
Create "Bytes" as a New Encoding of a New Description of Error.
Create "Native" as a New Present of a New Runtime's Error's Interface of a New Runtime's Shared of Error.
Create "Causes" as a New Runtime's Error's Causes.
Create "Cancelled" as a New Runtime's Error's Cancellation of false.
Return a New Runtime's Error's Wrapper of Bytes and Causes and Cancelled and Native.

    Publicly create Runtime's Error's "Trace|-s|-d|--ing|--er" as a Generic `error_native_cause` Function with Inputs of "value" as `&Error` and Outputs of `Option<&(dyn std::error::Error + 'static)>`.
Create "Native" as a New Runtime's Error's Native of Value.
Create "Available" as a New Presence of Native.
If Available, Create "Cause" as a New Contents of Native and `present native cause` and then Create "Borrowed" as a New Storage's Referrer of Cause and then Create "Source" as a New Runtime's Error's Source of Borrowed and then Create "Wrapped" as a New Present of Source and then Return Wrapped.
Create "Iterator" as a New Runtime's Error's Iterator of a New Runtime's Error's Inspector of Value.
Return a New Runtime's Error's Searcher of Iterator and Runtime's Error's Tracer.

    Publicly create Runtime's Error's "Attach|-es|-ed|-ing|-er" as a Generic `error_with_stderr` Function with Inputs of "value" as `Error` and "bytes" as `&[u8]` and Outputs of Runtime's Error.
Create "Size" as a New Length of Bytes.
Create "Zero" as a New Indexed Integer of 0.
If Size is Zero, Return Value.
Create "Description" as a New Runtime's Error's Lossy of a New Runtime's Error's Renderer of a New Storage's Borrower of Value.
Create "Message" as a New Mutable Binding of a New Storage's Owner of a New Storage's Text's Bytes of Description.
Create "Punctuation" as a New Runtime's Error's Punctuation.
Create "Separated" as a New Runtime's Error's Extender of Message and Punctuation.
Create "Extended" as a New Runtime's Error's Extender of Message and Bytes.
Create "Causes" as a New Runtime's Error's Parent of Value.
Create "Cancelled" as a New Runtime's Error's Cancellation of false.
Return a New Runtime's Error's Wrapper of Message and Causes and Cancelled and a New Absent.

    Publicly create Runtime's Error's "Display|-s|-ed|-ing|-er" as a Generic `error_display` Function with Inputs of "value" as `&Error` and "formatter" as `&mut std::fmt::Formatter<'_>` and Outputs of `std::fmt::Result`.
Return a New Runtime's Error's Writer of Formatter and a New Runtime's Error's Lossy of a New Runtime's Error's Renderer of Value.

*** Error traversal is composed from the same field access, collection and
    control operations as other Functions. Message presence is distinct from
    byte length: an empty leaf still contributes a line to joined diagnostics. ***
Define Runtime's Error's "Seen" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~insert~, Receiver: ~<@Set@>~, Inputs: ~&[<@Identity@>]&~|}<DEF>`.
Define Runtime's Error's "Set" as `<DEF>MOB{|Key: ~mut <@Name|Label@>~, Type: ~std::collections::HashSet<usize>~|}^KY{|Key: ~std::collections::HashSet::new~|}<DEF>`.
Define Runtime's Error's "Buffer" as `<DEF>MOB{|Key: ~mut <@Name|Label@>~, Type: ~Vec<u8>~|}^KY{|Key: ~Vec::new~|}<DEF>`.
Define Runtime's Error's "Flag" as `<DEF>OB{|Key: ~mut <@Name|Label@>~, Value: ~false~|}<DEF>`.
Define Runtime's Error's "Value" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~*<@Reference@>~|}<DEF>`.
Define Runtime's Error's "Mark|-s|-ed|-ing|-er" to `<DEF>O{|Key: ~*<@Reference@>~, Value: ~true~|}<DEF>`.
Define Runtime's Error's "Extend|-s|-ed|-ing|-er" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~extend_from_slice~, Receiver: ~<@Output@>~, Inputs: ~&[<@Bytes@>]&~|}<DEF>`.
Define Runtime's Error's "Separate|-s|-d|--ing|--or" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~push~, Receiver: ~<@Output@>~, Inputs: ~&[10_u8]&~|}<DEF>`.

Define Runtime's Error's "Collection" as `<DEF>OB{|Key: ~mut <@Name|Label@>~, Value: ~(<@Output@>, <@Seen@>, <@Written@>)~|}<DEF>`.
Define Runtime's Error's "Destination" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~&mut *<@Context@>.0~|}<DEF>`.
Define Runtime's Error's "Identities" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~&mut *<@Context@>.1~|}<DEF>`.
Define Runtime's Error's "Presence" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~&mut *<@Context@>.2~|}<DEF>`.

    Create Runtime's Error's "Remember|-s|-ed|-ing|-er" as a Generic `error_mark` Function with Inputs of "value" as `&Error` and "seen" as `&mut std::collections::HashSet<usize>` and Outputs of `()`.
Create "Recorded" as a New Runtime's Error's Seen of Seen and a New Runtime's Error's Identifier of Value.
Create "Children" as a New Runtime's Error's Inspector of Value.
Storage.Traverse Children and Runtime's Error's Rememberer and Seen.
Return a New Nothing.

    Create Runtime's Error's "Accumulate|-s|-d|--ing|--or" as a Generic `error_accumulate` Function with Inputs of "value" as `&Error` and "context" as `&mut (&mut Vec<u8>, &mut std::collections::HashSet<usize>, &mut bool)` and Outputs of `()`.
Create "Output" as a New Runtime's Error's Destination of Context.
Create "Seen" as a New Runtime's Error's Identities of Context.
Return a New Runtime's Error's Appender of Value and Output and Seen and a New Runtime's Error's Presence of Context.

    Create Runtime's Error's "Collect|-s|-ed|-ing|-or" as a Generic `error_append_children` Function with Inputs of "values" as `&[Error]` and "output" as `&mut Vec<u8>` and "seen" as `&mut std::collections::HashSet<usize>` and "written" as `&mut bool` and Outputs of `()`.
Create "Context" as a New Storage.Editable Storage's Borrower of a New Runtime's Error's Collection of Output and Seen and Written.
Storage.Traverse Values and Runtime's Error's Accumulator and Context.
Return a New Nothing.

    Create Runtime's Error's "Append|-s|-ed|-ing|-er" as a Generic `error_append` Function with Inputs of "value" as `&Error` and "output" as `&mut Vec<u8>` and "seen" as `&mut std::collections::HashSet<usize>` and "written" as `&mut bool` and Outputs of `()`.
Create "Unit" as a New Nothing.
Create "Fresh" as a New Runtime's Error's Seen of Seen and a New Runtime's Error's Identifier of Value.
If Fresh is false, Return Unit.
Create "Message" as a New Runtime's Error's Reader of Value.
Create "Children" as a New Runtime's Error's Inspector of Value.
Create "Size" as a New Length of Message.
Create "Count" as a New Length of Children.
Create "Zero" as a New Indexed Integer of 0.
Create "Empty" as a New Equality of Size and Zero.
Create "Childless" as a New Equality of Count and Zero.
Create "Joined" as a New Agreement of Empty and a New Equality of Childless and false.
If Joined, Create "Collected" as a New Runtime's Error's Collector of Children and Output and Seen and Written and then Return Collected.
Storage.Traverse Children and Runtime's Error's Rememberer and Seen.
Create "Present" as a New Runtime's Error's Value of Written.
If Present, Create "Separated" as a New Runtime's Error's Separator of Output.
Create "Extended" as a New Runtime's Error's Extender of Output and Message.
Runtime.Error.Mark Written.
Return Unit.

    Publicly create Runtime's Error's "Render|-s|-ed|-ing|-er" as a Generic `error_render` Function with Inputs of "value" as `&Error` and Outputs of `Vec<u8>`.
Create "Buffer" as a New Runtime's Error's Buffer.
Create "Set" as a New Runtime's Error's Set.
Create "Flag" as a New Runtime's Error's Flag.
Create "Output" as a New Storage.Editable Storage's Borrower of Buffer.
Create "Seen" as a New Storage.Editable Storage's Borrower of Set.
Create "Written" as a New Storage.Editable Storage's Borrower of Flag.
Create "Appended" as a New Runtime's Error's Appender of Value and Output and Seen and Written.
Return Buffer.

    Publicly create Runtime's Error's "Describe|-s|-d|--ing|--er" as a Generic `error_new_text` Function with Inputs of "message" as `&str` and Outputs of Runtime's Error.
Return a New Runtime's Error's Former of a New Storage's Text's Owner of Message.

    Publicly create Runtime's Error's "Cancel|-s|-led|-ling|-ler" as a Generic `error_cancelled` Function with No Inputs and Outputs of Runtime's Error.
Create "Message" as a New Storage's Text's Owner of `source execution stopped`.
Create "Causes" as a New Runtime's Error's Causes.
Create "Cancelled" as a New Runtime's Error's Cancellation of true.
Return a New Runtime's Error's Wrapper of Message and Causes and Cancelled and a New Absent.

    Publicly create Runtime's Error's "Identify|--ies|--ied|-ing|--ier" as a Generic `error_identity` Function with Inputs of "value" as `&Error` and Outputs of `usize`.
Return a New Indexed Integer of a New Runtime's Error's Address of Value.

    Publicly create Runtime's Error's "Check|-s|-ed|-ing|-er" as a Generic `error_is_cancelled` Function with Inputs of "value" as `&Error` and Outputs of `bool`.
Return a New Runtime's Error's Status of Value.

    Publicly create Runtime's Error's "Inspect|-s|-ed|-ing|-or" as a Generic `error_causes` Function with Inputs of "value" as `&Error` and Outputs of `&[Error]`.
Return a New Runtime's Error's Children of Value.

    Publicly create Runtime's Error's "Form|-s|-ed|-ing|-er" as a Generic `error_new` Function with Inputs of "message" as `Vec<u8>` and Outputs of Runtime's Error.
Create "Causes" as a New Runtime's Error's Causes.
Create "Cancelled" as a New Runtime's Error's Cancellation of false.
Return a New Runtime's Error's Wrapper of Message and Causes and Cancelled and a New Absent.

    Publicly create Runtime's Error's "Join|-s|-ed|-ing|-er" as a Generic `error_join` Function with Inputs of "causes" as `Vec<Error>` and Outputs of `Option<Error>`.
Create "Count" as a New Storage's Length of Causes.
Create "Zero" as a New Indexed Integer of 0.
If Count is Zero, Create "Missing" as a New Absent and then Return Missing.
Create "One" as a New Indexed Integer of 1.
If Count is One, Create "Iterator" as a New Runtime's Error's Iterator of Causes and then Create "Sole" as a New Runtime's Error's First of Iterator and then Return Sole.
Create "Message" as a New Runtime's Error's Message.
Create "Cancelled" as a New Runtime's Error's Cancellation of false.
Return a New Present of a New Runtime's Error's Wrapper of Message and Causes and Cancelled and a New Absent.

    Publicly create Runtime's Error's "Read|-s|-|-ing|-er" as a Generic `error_bytes` Function with Inputs of "value" as `&Error` and Outputs of `&[u8]`.
Return a New Runtime's Error's Bytes of Value.

    Publicly create Runtime's Error's "Copy|--ies|--ied|-ing|--ier" as a Generic `error_copy` Function with Inputs of "value" as `&Error` and Outputs of Runtime's Error.
Return a New Runtime's Error's Fault of a New Retainer of a New Runtime's Error's Datum of Value.
