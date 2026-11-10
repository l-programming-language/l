Create Storage's "Decide|-s|-d|--ing|--er" for Processing a New Process for Creating "Accepted" as a New Storage.Checked Storage's Binding of any Thing and then Processing a New Storage's Selector of Accepted, a New Process using any Process, and a New Storage.Empty Process.
Create Storage's "Probe|-s|-d|--ing|--er" for Processing a New Process for Creating "Outcome" as a New Storage's Applier of any Thing, any Thing, and any Thing and then Storage.Deciding Outcome and any Process.
Create Storage's "Search|-es|-ed|-ing|-er" for Processing a New Storage's Iteration of any Thing and a New Process with Inputs of "Item" as `_` for Storage.Probing any Thing, Item, any Thing, and any Process.
Define Storage's "Iterative" as a New Adjective describing any Storage's Routine.
It describes a Generic Storage.Lending Storage's Routine of `I: IntoIterator<Item = &'a LValue>` and `'a`.
Define Storage's Value's "Target" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~<@Context@>.0~|}<DEF>`.
Define Storage's Value's "History" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~&mut *<@Context@>.1~|}<DEF>`.
Define Storage's Map's "Children" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~values~, Receiver: ~<@Value@>.values~|}<DEF>`.
Define Storage's "Iterate|-s|-d|--ing|--or" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~iter~, Receiver: ~<@Value@>~|}<DEF>`.

    Publicly create Storage's Value's "Follow|-s|-ed|-ing|-er" as a Generic Function with Inputs of "value" as `&LValue` and "context" as `&mut (usize, &mut std::collections::BTreeSet<usize>)` and Outputs of `runtime::Result<bool>`.
Create "Target" as a New Storage's Value's Target of Context.
Create "Seen" as a New Storage's Value's History of Context.
Return a New Storage's Value's Reacher of Value, Target, and Seen.

    Publicly create Storage's Value's "Search|-es|-ed|-ing|-er" as a Generic Storage.Iterative Storage.Routine with Inputs of "values" as `I` and "target" as `usize` and "seen" as `&mut std::collections::BTreeSet<usize>` and Outputs of `runtime::Result<bool>`.
Create "Pair" as a New Storage's Tuple of Target and Seen.
Create "State" as a New Storage.Mutable Binding of Pair.
Create "Context" as a New Storage.Editable Storage's Borrower of State.
Create "Success" as a New Storage's Success of true.
Storage.Search Values, Storage's Value's Follower, Context, and a New Function's Returner of Success.
Return a New Storage's Success of false.

    Publicly create Storage's List's "Reach|-es|-ed|-ing|-er" as a Generic `l_reach_list` Function with Inputs of "data" as `&dyn std::any::Any` and "target" as `usize` and "seen" as `&mut std::collections::BTreeSet<usize>` and Outputs of `runtime::Result<bool>`.
Create "Value" as a New Storage.Checked Binding of a New Storage's Annotation of `runtime::Result<&LList>` and a New Storage's Projector of Data and `expected List`.
Create "Children" as a New Storage's List's Values of Value.
Return a New Storage's Value's Searcher of Children, Target, and Seen.

    Publicly create Storage's Map's "Reach|-es|-ed|-ing|-er" as a Generic `l_reach_map` Function with Inputs of "data" as `&dyn std::any::Any` and "target" as `usize` and "seen" as `&mut std::collections::BTreeSet<usize>` and Outputs of `runtime::Result<bool>`.
Create "Value" as a New Storage.Checked Binding of a New Storage's Annotation of `runtime::Result<&LMap>` and a New Storage's Projector of Data and `expected Map`.
Create "Children" as a New Storage's Map's Children of Value.
Return a New Storage's Value's Searcher of Children, Target, and Seen.

    Publicly create Storage's Pointer's "Visit|-s|-ed|-ing|-or" as a Generic `l_reach_pointer` Function with Inputs of "data" as `&dyn std::any::Any` and "target" as `usize` and "seen" as `&mut std::collections::BTreeSet<usize>` and Outputs of `runtime::Result<bool>`.
Create "Value" as a New Storage.Checked Binding of a New Storage's Annotation of `runtime::Result<&LPointer>` and a New Storage's Projector of Data and `expected Pointer`.
Return a New Storage's Pointer's Reacher of Value, Target, and Seen.

    Publicly create Storage's Environment's "Reach|-es|-ed|-ing|-er" as a Generic `l_environment_reaches` Function with Inputs of "environment" as `&LEnvironment` and "target" as `usize` and "seen" as `&mut std::collections::BTreeSet<usize>` and Outputs of `runtime::Result<bool>`.
Create "Values" as a New Storage's Environment's Values of Environment.
Create "Identity" as a New Storage's Identifier of Values.
If Identity is Target, Create "Success" as a New Storage's Success of true and then Return Success.
Create "Fresh" as a New Storage's Rememberer of Seen and Identity.
If Fresh is false, Create "Missing" as a New Storage's Success of false and then Return Missing.
Create "Borrowed" as a New Storage's Reader of Values.
Create "Guard" as a New Storage.Checked Binding of a New Storage's Accepter of Borrowed and `capture environment is already borrowed`.
Create "Children" as a New Storage's Iterator of Guard.
Return a New Storage's Value's Searcher of Children, Target, and Seen.

    Publicly create Storage's Environment's "Visit|-s|-ed|-ing|-or" as a Generic `l_reach_environment` Function with Inputs of "data" as `&dyn std::any::Any` and "target" as `usize` and "seen" as `&mut std::collections::BTreeSet<usize>` and Outputs of `runtime::Result<bool>`.
Create "Value" as a New Storage.Checked Binding of a New Storage's Annotation of `runtime::Result<&LEnvironment>` and a New Storage's Projector of Data and `expected Environment`.
Return a New Storage's Environment's Reacher of Value, Target, and Seen.

    Publicly create Storage's Function's "Reach|-es|-ed|-ing|-er" as a Generic Storage.Permanent `l_reach_function` Storage.Routine with Inputs of "data" as `&dyn std::any::Any` and "target" as `usize` and "seen" as `&mut std::collections::BTreeSet<usize>` and Outputs of `runtime::Result<bool>`.
Create "Value" as a New Storage.Checked Binding of a New Storage's Annotation of `runtime::Result<&(LEnvironment, T)>` and a New Storage's Projector of Data and `Function type mismatch`.
Create "Environment" as a New Storage's Function's Environment of Value.
Return a New Storage's Environment's Reacher of Environment, Target, and Seen.

Define Storage's "Pair|-s|-ed|-ing|-er" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~zip~, Receiver: ~<@Left@>~, Inputs: ~&[<@Right@>]&~|}<DEF>`.
Create Storage's "Review|-s|-ed|-ing|-er" for Processing a New Process for Creating "Outcome" as a New Storage.Paired Storage's Applier of any Thing, any Thing, any Thing, and any Thing and then Storage.Verifying Outcome.
Create Storage's "Audit|-s|-ed|-ing|-or" for Processing a New Storage.Paired Storage's Iteration of any Thing and a New Process with Inputs of "First" as `_` and "Second" as `_` for Storage.Reviewing any Thing, First, Second, and any Thing.

    Publicly create Storage's Environment's "Check|-s|-ed|-ing|-er" as a Generic Function with Inputs of "expected" as `&Vec<u64>` and "value" as `&LValue` and "message" as `&str` and Outputs of `runtime::Result<()>`.
Return a New Storage's Value's Checker of Value, Expected, and Message.

    Publicly create Storage's Environment's "Create|-s|-d|--ing|--or" as a Generic `l_environment_new` Function with Inputs of "types" as `Vec<Vec<u64>>` and "values" as `Vec<LValue>` and Outputs of `runtime::Result<LEnvironment>`.
Create "Signatures" as a New Storage's Length of Types.
Create "Contents" as a New Storage's Length of Values.
Create "Equal" as a New Storage's Equality of Signatures and Contents.
Create "Valid" as a New Storage's Validator of Equal and `capture initialization type mismatch`.
Storage.Verify Valid.
Create "Iterator" as a New Storage's Iterator of Types.
Create "Borrowed" as a New Storage's Borrower of Values.
Create "Pairs" as a New Storage's Pairer of Iterator and Borrowed.
Storage.Audit Pairs, Storage's Environment's Checker, and `capture initialization type mismatch`.
Create "Cell" as a New Storage's Cell of Values.
Create "Shared" as a New Storage's Sharer of Cell.
Create "Signature" as a New Storage's Sharer of Types.
Create "Environment" as a New Storage's Environment's Content of Signature and Shared.
Return a New Storage's Success of Environment.

    Publicly create Storage's Environment's "Read|-s||-ing|-er" as a Generic `l_environment_get` Function with Inputs of "environment" as `&LEnvironment` and "index" as `usize` and Outputs of `runtime::Result<LValue>`.
Create "Values" as a New Storage's Environment's Values of Environment.
Create "Borrowed" as a New Storage's Reader of Values.
Create "Guard" as a New Storage.Checked Binding of a New Storage's Accepter of Borrowed and `capture environment is already borrowed`.
Create "Found" as a New Storage's Getter of Guard and Index.
Create "Required" as a New Storage's Requirer of Found and `capture index out of bounds`.
Return a New Storage's Resolver of Required and Storage's Value's Copier.

    Publicly create Storage's Environment's "Update|-s|-d|--ing|--er" as a Generic `l_environment_set` Function with Inputs of "environment" as `&LEnvironment` and "index" as `usize` and "value" as `LValue` and Outputs of `runtime::Result<()>`.
Create "Values" as a New Storage's Environment's Values of Environment.
Create "Identity" as a New Storage's Identifier of Values.
Create "Borrowed" as a New Storage's Borrower of Value.
Create "Protected" as a New Storage's Value's Protector of Borrowed, Identity, and `capture write would create an ownership cycle`.
Storage.Verify Protected.
Create "Editable" as a New Storage.Editable Storage's Reader of Values.
Create "Temporary" as a New Storage.Checked Binding of a New Storage's Accepter of Editable and `capture environment is already borrowed`.
Create "Guard" as a New Storage.Mutable Binding of Temporary.
Create "Found" as a New Storage's Locator of Guard and Index.
Create "Target" as a New Storage.Checked Binding of a New Storage's Requirer of Found and `capture index out of bounds`.
Create "Types" as a New Storage's Environment's Types of Environment.
Create "Signature" as a New Storage's Getter of Types and Index.
Create "Expected" as a New Storage.Checked Binding of a New Storage's Requirer of Signature and `capture index out of bounds`.
Create "Checked" as a New Storage's Value's Checker of Borrowed, Expected, and `capture type mismatch`.
Storage.Verify Checked.
Create "Previous" as a New Storage's Replacer of Target and Value.
Create "Unit" as a New Storage's Nothing.
Return a New Storage's Success of Unit.

Define Storage's "Join|-s|-ed|-ing|-er" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~join~, Receiver: ~<@Values@>~, Inputs: ~&[<@Separator@>]&~|}<DEF>`.

    Publicly create Storage's Text's "Join|-s|-ed|-ing|-er" as a Generic Function with Inputs of "values" as `Vec<Vec<u8>>` and "separator" as `&str` and Outputs of `Vec<u8>`.
Create "Bytes" as a New Storage's Text's Bytes of Separator.
Return a New Storage's Joiner of Values and Bytes.

    Publicly create Storage's Text's "Enclose|-s|-d|--ing|--er" as a Generic Function with Inputs of "value" as `Vec<u8>` and "opening" as `&str` and "closing" as `&str` and Outputs of `Vec<u8>`.
Create "Buffer" as a New Storage.Mutable Binding of a New Storage's Text's Owner of Opening.
Create "Target" as a New Storage.Editable Storage's Borrower of Buffer.
Create "Bytes" as a New Storage's Borrower of Value.
Create "Addedvalue" as a New Storage's Bytes's Extender of Target and Bytes.
Create "Closingbytes" as a New Storage's Text's Bytes of Closing.
Create "Addedclosing" as a New Storage's Bytes's Extender of Target and Closingbytes.
Return Buffer.

    Publicly create Storage's List's "Describe|-s|-d|--ing|--er" as a Generic Function with Inputs of "value" as `&LValue` and "parts" as `&mut Vec<Vec<u8>>` and Outputs of `runtime::Result<()>`.
Create "Bytes" as a New Storage.Checked Binding of a New Storage's Value's Formatter of Value.
Create "Added" as a New Storage's Appender of Parts and Bytes.
Return a New Storage's Success of Added.

    Publicly create Storage's Map's "Describe|-s|-d|--ing|--er" as a Generic Function with Inputs of "key" as `&Vec<u8>` and "value" as `&LValue` and "parts" as `&mut Vec<Vec<u8>>` and Outputs of `runtime::Result<()>`.
Create "Copied" as a New Storage's Copier of Key.
Create "Separator" as a New Storage's Text's Owner of `: `.
Create "Bytes" as a New Storage.Checked Binding of a New Storage's Value's Formatter of Value.
Create "Entry" as a New Storage's Vector of Copied, Separator, and Bytes.
Create "Joined" as a New Storage's Text's Joiner of Entry and ``.
Create "Added" as a New Storage's Appender of Parts and Joined.
Return a New Storage's Success of Added.

    Publicly create Storage's List's "Render|-s|-ed|-ing|-er" as a Generic `l_show_list` Function with Inputs of "data" as `&dyn std::any::Any` and Outputs of `runtime::Result<Vec<u8>>`.
Create "Value" as a New Storage.Checked Binding of a New Storage's Annotation of `runtime::Result<&LList>` and a New Storage's Projector of Data and `expected List`.
Create "Parts" as a New Storage's Vector's Buffer of `Vec<u8>`.
Create "Context" as a New Storage.Editable Storage's Borrower of Parts.
Create "Children" as a New Storage's List's Values of Value.
Storage.Inspect Children and Storage's List's Describer and Context.
Create "Joined" as a New Storage's Text's Joiner of Parts and `, `.
Create "Bytes" as a New Storage's Text's Encloser of Joined, `[`, and `]`.
Return a New Storage's Success of Bytes.

    Publicly create Storage's Map's "Render|-s|-ed|-ing|-er" as a Generic `l_show_map` Function with Inputs of "data" as `&dyn std::any::Any` and Outputs of `runtime::Result<Vec<u8>>`.
Create "Value" as a New Storage.Checked Binding of a New Storage's Annotation of `runtime::Result<&LMap>` and a New Storage's Projector of Data and `expected Map`.
Create "Parts" as a New Storage's Vector's Buffer of `Vec<u8>`.
Create "Context" as a New Storage.Editable Storage's Borrower of Parts.
Create "Entries" as a New Storage's Map's Entries of Value.
Storage.Audit Entries and Storage's Map's Describer and Context.
Create "Joined" as a New Storage's Text's Joiner of Parts and `, `.
Create "Bytes" as a New Storage's Text's Encloser of Joined, `{`, and `}`.
Return a New Storage's Success of Bytes.

    Publicly create Storage's Pointer's "Render|-s|-ed|-ing|-er" as a Generic `l_show_pointer` Function with Inputs of "data" as `&dyn std::any::Any` and Outputs of `runtime::Result<Vec<u8>>`.
Create "Pointer" as a New Storage.Checked Binding of a New Storage's Annotation of `runtime::Result<&LPointer>` and a New Storage's Projector of Data and `expected Pointer`.
Create "Value" as a New Storage.Checked Binding of a New Storage's Pointer's Reader of Pointer.
Create "Borrowed" as a New Storage's Borrower of Value.
Return a New Storage's Value's Formatter of Borrowed.

