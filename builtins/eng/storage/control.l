Define Storage's "Polymorphic" as a New Adjective describing any Storage's Routine.
It describes a Generic Storage's Routine of `T`.
Define Storage's "Permanent" as a New Adjective describing any Storage's Routine.
It describes a Generic Storage's Routine of `T: 'static`.
Define Storage's "Referenced" as a New Adjective describing any Storage's Routine.
It describes a Generic Storage.Lending Storage's Routine of `T: 'static` and `'a`.
Define Storage's "Failure" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~runtime::error_new_text~, Imports: ~&[runtime::error_new_text as _]&~, Inputs: ~&[<@Message@>]&~|}<DEF>`.
Define Storage's "Extract|-s|-ed|-ing|-or" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~ok~, Receiver: ~<@Value@>~|}<DEF>`.
Define Storage's "Projection" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~downcast_ref~, Receiver: ~<@Value@>~|}<DEF>`.
Define Storage's "Share|-s|-d|--ing|-r" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~std::rc::Rc::new~, Inputs: ~&[<@Value@>]&~|}<DEF>`.
Define Storage's "Success" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~Ok~, Inputs: ~&[<@Value@>]&~|}<DEF>`.
Define Storage's "Release|-s|-d|--ing|--er" to `<DEF>MOB{|Key: ~_~|}^KY{|Key: ~std::mem::drop~, Inputs: ~&[<@Value@>]&~|}<DEF>`.
Define "Presence" as `<DEF>E{|Statement: ~Some(<@Payload|Label@>)~|}*<@Body@><DEF>`.
Define "Absence" as `<DEF>E{|Statement: ~None~|}*<@Body@><DEF>`.
Define "Success" as `<DEF>E{|Statement: ~Ok(<@Payload|Label@>)~|}*<@Body@><DEF>`.
Define "Failure" as `<DEF>E{|Statement: ~Err(<@Failure|Label@>)~|}*<@Body@><DEF>`.
Define "Selection" as `<DEF>BOG^GI{|Key: ~<@Value@>~|}^(<@Present@>*<@Absent@>)<DEF>`.
Create Storage's "Option" as a Dictionary.
Create Storage's Option's "Select|-s|-ed|-ing|-or" for Processing a New Selection of any Thing, a New Presence of any Process, and a New Absence of any Process.
Create Storage's "Result" as a Dictionary.
Create Storage's Result's "Select|-s|-ed|-ing|-or" for Processing a New Selection of any Thing, a New Success of any Process, and a New Failure of any Process.
Define Storage's "Expect|-s|-ed|-ing|-er" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~expect~, Receiver: ~<@Value@>~, Inputs: ~&[<@Message@>]&~|}<DEF>`.

    Publicly create Storage's "Require|-s|-d|--ing|--er" as a Generic Storage.Polymorphic `l_required` Storage.Routine with Inputs of "value" as `Option<T>` and "message" as `&str` and Outputs of `runtime::Result<T>`.
Storage.Option.Select Value, a New Process with Inputs of "Payload" as `T` for Returning a New Storage's Success of Payload, and a New Process for Creating "Reason" as a New Storage's Failure of Message and then Returning a New Storage's Rejection of Reason.

    Publicly create Storage's "Accept|-s|-ed|-ing|-er" as a Generic `l_borrowed` Storage.Routine of `T, E` with Inputs of "value" as `Result<T, E>` and "message" as `&str` and Outputs of `runtime::Result<T>`.
Storage.Result.Select Value, a New Process with Inputs of "Payload" as `T` for Returning a New Storage's Success of Payload, and a New Process with Inputs of "Previous" as `E` for Storage.Releasing Previous and then Creating "Reason" as a New Storage's Failure of Message and then Returning a New Storage's Rejection of Reason.

    Publicly create Storage's "Project|-s|-ed|-ing|-or" as a Generic Storage.Referenced `l_project` Storage.Routine with Inputs of "value" as `&'a dyn std::any::Any` and "message" as `&str` and Outputs of `runtime::Result<&'a T>`.
Create "Projected" as a New Storage's Projection of Value.
Return a New Storage's Requirer of Projected and Message.

    Publicly create Storage's "Resolve|-s|-d|--ing|--er" as a Generic Storage.Polymorphic Storage.Routine with Inputs of "value" as `runtime::Result<&T>` and "copy" as `fn(&T) -> T` and Outputs of `runtime::Result<T>`.
Create "Checked" as a New Storage.Checked Binding of Value.
Create "Copied" as a New Storage.Produced Binding of a New Function's Invoker of Copy and Checked.
Return a New Storage's Success of Copied.

    Publicly create Storage's Value's "Wrap|-s|-ped|-ping|-per" as a Generic Storage.Permanent `l_value_wrap` Storage.Routine with Inputs of "value" as `T` and "signature" as `Vec<u64>` and "renderer" as `fn(&dyn std::any::Any) -> runtime::Result<Vec<u8>>` and "visitor" as `fn(&dyn std::any::Any, usize, &mut std::collections::BTreeSet<usize>) -> runtime::Result<bool>` and Outputs of Storage's Value.
Create "Data" as a New Storage's Sharer of Value.
Return a New Storage's Value's Content of Signature, Data, Renderer, and Visitor.

    Publicly create Storage's Value's "Restore|-s|-d|--ing|--er" as a Generic Storage.Permanent `l_restore` Storage.Routine with Inputs of "value" as `LValue` and "message" as `&str` and "snapshot" as `fn(&T) -> T` and Outputs of `runtime::Result<T>`.
Create "Data" as a New Storage's Value's Data of Value.
Create "Borrowed" as a New Storage's Referrer of Data.
Create "Projected" as a New Storage's Projector of Borrowed and Message.
Return a New Storage's Resolver of Projected and Snapshot.

Create Storage's "Signature" as a Dictionary.
Define Storage's Signature's "Universal" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~&[1]~|}<DEF>`.
Define Storage's "Equality" as `<DEF>DOB{|Key: ~<@Name|Label@>~|}^DE{|Ops: ~&[Op{Type: Calculate, Operator: ==, Operands: &[<@Left@>, <@Right@>]&}Op]&~|}<DEF>`.
Define Storage's "Alternative" as `<DEF>DOB{|Key: ~<@Name|Label@>~|}^DE{|Ops: ~&[Op{Type: Calculate, Operator: ||, Operands: &[<@Left@>, <@Right@>]&}Op]&~|}<DEF>`.
Define Storage's "Rejection" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~Err~, Inputs: ~&[<@Value@>]&~|}<DEF>`.

    Publicly create Storage's Signature's "Match|-es|-ed|-ing|-er" as a Generic `l_type_accepts` Function with Inputs of "expected" as `&[u64]` and "actual" as `&[u64]` and Outputs of `bool`.
Create "Any" as a New Storage's Signature's Universal.
Create "Broad" as a New Storage's Equality of Expected and Any.
Create "Same" as a New Storage's Equality of Expected and Actual.
Return a New Storage's Alternative of Broad and Same.

    Publicly create Storage's Value's "Expect|-s|-ed|-ing|-er" as a Generic `l_expect` Function with Inputs of "value" as `LValue` and "expected" as `Vec<u64>` and Outputs of `runtime::Result<LValue>`.
Create "Borrowed" as a New Storage's Borrower of Value.
Create "Signature" as a New Storage's Borrower of Expected.
Create "Checked" as a New Storage's Value's Checker of Borrowed, Signature, and `runtime value type mismatch`.
Storage.Verify Checked.
Return a New Storage's Success of Value.

    Publicly create Storage's Value's "Reach|-es|-ed|-ing|-er" as a Generic `l_reaches` Function with Inputs of "value" as `&LValue` and "target" as `usize` and "seen" as `&mut std::collections::BTreeSet<usize>` and Outputs of `runtime::Result<bool>`.
Create "Data" as a New Storage's Value's Data of Value.
Create "Borrowed" as a New Storage's Referrer of Data.
Create "Visitor" as a New Storage's Value's Visitor of Value.
Return a New Storage.Produced Binding of a New Function's Invoker of Visitor, Borrowed, Target, and Seen.

    Publicly create Storage's "Ignore|-s|-d|--ing|--er" as a Generic `l_reach_scalar` Function with Inputs of "value" as `&dyn std::any::Any` and "target" as `usize` and "seen" as `&mut std::collections::BTreeSet<usize>` and Outputs of `runtime::Result<bool>`.
Return a New Storage's Success of false.

Define Storage's "Sum" as `<DEF>DOB{|Key: ~<@Name|Label@>~|}^DE{|Ops: ~&[Op{Type: Calculate, Operator: +, Operands: &[<@Left@>, <@Right@>]&}Op]&~|}<DEF>`.
Define Storage's "Choice" as `<DEF>GOB{|Key: ~<@Name|Label@>~|}^GI^(E{|Statement: ~<@Condition@>~|}*W^BOD^E{|Statement: ~<@Then@>~|}*E{|Statement: ~Void~|}*W^BOD^E{|Statement: ~<@Else@>~|})<DEF>`.
Define Storage's "Agreement" as `<DEF>DOB{|Key: ~<@Name|Label@>~|}^DE{|Ops: ~&[Op{Type: Calculate, Operator: &&, Operands: &[<@Left@>, <@Right@>]&}Op]&~|}<DEF>`.
Define Storage's "Unwrap|-s|-ped|-ping|-per" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~unwrap_or~, Receiver: ~<@Value@>~, Inputs: ~&[<@Fallback@>]&~|}<DEF>`.
