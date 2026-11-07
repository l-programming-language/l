Define Storage's "Verify|--ies|--ied|-ing|--ier" to `<DEF>OB{|Key: ~_~, Value: ~<@Value@>?~|}<DEF>`.
Define Storage's "Get|-s|--ot|-ting|-ter" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~get~, Receiver: ~<@Collection@>~, Inputs: ~&[<@Index@>]&~|}<DEF>`.
Define Storage's "Locate|-s|-d|--ing|--or" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~get_mut~, Receiver: ~<@Collection@>~, Inputs: ~&[<@Index@>]&~|}<DEF>`.
Define Storage's "Replace|-s|-d|--ing|--er" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~std::mem::replace~, Inputs: ~&[<@Target@>, <@Value@>]&~|}<DEF>`.
Create Storage's "Examine|-s|-d|--ing|--er" for Processing a New Process for Creating "Outcome" as a New Storage's Applier of any Thing, any Thing, and any Thing and then Storage.Verifying Outcome.
Create Storage's "Inspect|-s|-ed|-ing|-or" for Processing a New Storage's Iteration of any Thing and a New Process with Inputs of "Item" as `_` for Storage.Examining any Thing, Item, and any Thing.

    Publicly create Storage's "Validate|-s|-d|--ing|--or" as a Generic Function with Inputs of "valid" as `bool` and "message" as `&str` and Outputs of `runtime::Result<()>`.
If Valid, Create "Unit" as a New Storage's Nothing and then Create "Success" as a New Storage's Success of Unit and then Return Success.
Create "Failure" as a New Storage's Failure of Message.
Return a New Storage's Rejection of Failure.

Create Storage's "Index" as a Dictionary.
Define Storage's Index's "Limit" as `<DEF>DOB{|Key: ~<@Name|Label@>~|}^DE{|Ops: ~&[Op{Type: Calculate, Operator: as, Operands: &[usize::MAX, f64]&}Op]&~|}<DEF>`.
Define Storage's "Integral" as `<DEF>DOB{|Key: ~<@Name|Label@>~|}^DE{|Ops: ~&[Op{Type: Calculate, Operator: as, Operands: &[<@Value@>, <@Type|Type@>]&}Op]&~|}<DEF>`.
Define Storage's "Dereference" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~*<@Value@>~|}<DEF>`.
Define Storage's Number's "Fraction" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~fract~, Receiver: ~<@Value@>~|}<DEF>`.
Define Storage's Number's "Finite" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~is_finite~, Receiver: ~<@Value@>~|}<DEF>`.

    Publicly create Storage's Index's "Check|-s|-ed|-ing|-er" as a Generic `l_index` Function with Inputs of "index" as `f64` and Outputs of `runtime::Result<usize>`.
Create "Finite" as a New Storage's Number's Finite of Index.
If Finite is false, Storage.Reject `index must be a nonnegative integer`.
If Index less than 0, Storage.Reject `index must be a nonnegative integer`.
Create "Fraction" as a New Storage's Number's Fraction of Index.
If Fraction isnt 0, Storage.Reject `index must be a nonnegative integer`.
Create "Limit" as a New Storage's Index's Limit.
If Index greater than Limit, Storage.Reject `index must be a nonnegative integer`.
If Index is Limit, Storage.Reject `index must be a nonnegative integer`.
Create "Integral" as a New Storage's Integral of Index and `usize`.
Return a New Storage's Success of Integral.

    Publicly create Storage's Bytes's "Read|-s||-ing|-er" as a Generic `l_byte_get` Function with Inputs of "bytes" as `&[u8]` and "index" as `f64` and Outputs of `runtime::Result<f64>`.
Create "Position" as a New Storage.Checked Binding of a New Storage's Index's Checker of Index.
Create "Found" as a New Storage's Getter of Bytes and Position.
Create "Reference" as a New Storage.Checked Binding of a New Storage's Requirer of Found and `byte index out of bounds`.
Create "Value" as a New Storage's Dereference of Reference.
Create "Number" as a New Storage's Integral of Value and `f64`.
Return a New Storage's Success of Number.

    Publicly create Storage's List's "Check|-s|-ed|-ing|-er" as a Generic Function with Inputs of "value" as `&LValue` and "expected" as `&[u64]` and Outputs of `runtime::Result<()>`.
Return a New Storage's Value's Checker of Value, Expected, and `List element type mismatch`.

    Publicly create Storage's List's "Create|-s|-d|--ing|--or" as a Generic `l_list_new` Function with Inputs of "element" as `Vec<u64>` and "values" as `Vec<LValue>` and Outputs of `runtime::Result<LList>`.
Create "Children" as a New Storage's Borrower of Values.
Create "Expected" as a New Storage's Borrower of Element.
Storage.Inspect Children and Storage's List's Checker and Expected.
Create "List" as a New Storage's List's Content of Element and Values.
Return a New Storage's Success of List.

Define Storage's List's "Editable" as a New Adjective describing any Storage's List's Values.
It describes any Storage's List's Values as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~&mut <@Value@>.values~|}<DEF>`.
Use Storage.Environment.Values as Storage.Map's "Values".
Use Storage.List.Editable as Storage.Map's "Editable".
Define Storage's "Cell" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~std::cell::RefCell::new~, Inputs: ~&[<@Value@>]&~|}<DEF>`.

    Publicly create Storage's Value's "Check|-s|-ed|-ing|-er" as a Generic Function with Inputs of "value" as `&LValue` and "expected" as `&[u64]` and "message" as `&str` and Outputs of `runtime::Result<()>`.
Create "Signature" as a New Storage's Value's Signature of Value.
Create "Accepted" as a New Storage's Signature's Matcher of Expected and Signature.
Return a New Storage's Validator of Accepted and Message.

    Publicly create Storage's List's "Read|-s||-ing|-er" as a Generic `l_list_get` Function with Inputs of "list" as `&LList` and "index" as `f64` and Outputs of `runtime::Result<LValue>`.
Create "Position" as a New Storage.Checked Binding of a New Storage's Index's Checker of Index.
Create "Values" as a New Storage's List's Values of List.
Create "Found" as a New Storage's Getter of Values and Position.
Create "Required" as a New Storage's Requirer of Found and `List index out of bounds`.
Return a New Storage's Resolver of Required and Storage's Value's Copier.

    Publicly create Storage's List's "Append|-s|-ed|-ing|-er" as a Generic `l_list_put` Function with Inputs of "list" as `&mut LList` and "value" as `LValue` and Outputs of `runtime::Result<()>`.
Create "Element" as a New Storage's List's Element of List.
Create "Borrowed" as a New Storage's Borrower of Value.
Create "Checked" as a New Storage's List's Checker of Borrowed and Element.
Storage.Verify Checked.
Create "Values" as a New Storage.List.Editable Storage's List's Values of List.
Create "Added" as a New Storage's Appender of Values and Value.
Return a New Storage's Success of Added.

    Publicly create Storage's List's "Update|-s|-d|--ing|--er" as a Generic `l_list_set` Function with Inputs of "list" as `&mut LList` and "index" as `f64` and "value" as `LValue` and Outputs of `runtime::Result<()>`.
Create "Element" as a New Storage's List's Element of List.
Create "Borrowed" as a New Storage's Borrower of Value.
Create "Checked" as a New Storage's List's Checker of Borrowed and Element.
Storage.Verify Checked.
Create "Position" as a New Storage.Checked Binding of a New Storage's Index's Checker of Index.
Create "Values" as a New Storage.List.Editable Storage's List's Values of List.
Create "Found" as a New Storage's Locator of Values and Position.
Create "Target" as a New Storage.Checked Binding of a New Storage's Requirer of Found and `List index out of bounds`.
Create "Previous" as a New Storage's Replacer of Target and Value.
Create "Unit" as a New Storage's Nothing.
Return a New Storage's Success of Unit.

    Publicly create Storage's Map's "Create|-s|-d|--ing|--or" as a Generic `l_map_new` Function with Inputs of "element" as `Vec<u64>` and Outputs of Storage's Map.
Create "Values" as a New Storage's Map's Buffer of `Vec<u8>` and `LValue`.
Return a New Storage's Map's Content of Element and Values.

    Publicly create Storage's Map's "Read|-s||-ing|-er" as a Generic `l_map_get` Function with Inputs of "map" as `&LMap` and "key" as `&[u8]` and Outputs of `runtime::Result<LValue>`.
Create "Values" as a New Storage's Map's Values of Map.
Create "Found" as a New Storage's Getter of Values and Key.
Create "Required" as a New Storage's Requirer of Found and `Map key is absent`.
Return a New Storage's Resolver of Required and Storage's Value's Copier.

    Publicly create Storage's Map's "Update|-s|-d|--ing|--er" as a Generic `l_map_set` Function with Inputs of "map" as `&mut LMap` and "key" as `Vec<u8>` and "value" as `LValue` and Outputs of `runtime::Result<()>`.
Create "Element" as a New Storage's Map's Element of Map.
Create "Borrowed" as a New Storage's Borrower of Value.
Create "Checked" as a New Storage's Value's Checker of Borrowed, Element, and `Map element type mismatch`.
Storage.Verify Checked.
Create "Values" as a New Storage.Map.Editable Storage's Map's Values of Map.
Create "Previous" as a New Storage's Inserter of Values, Key, and Value.
Create "Unit" as a New Storage's Nothing.
Return a New Storage's Success of Unit.

    Publicly create Storage's Pointer's "Create|-s|-d|--ing|--or" as a Generic `l_pointer_new` Function with Inputs of "element" as `Vec<u64>` and "value" as `LValue` and Outputs of `runtime::Result<LPointer>`.
Create "Expected" as a New Storage's Borrower of Element.
Create "Borrowed" as a New Storage's Borrower of Value.
Create "Checked" as a New Storage's Value's Checker of Borrowed, Expected, and `Pointer initialization type mismatch`.
Storage.Verify Checked.
Create "Cell" as a New Storage's Cell of Value.
Create "Target" as a New Storage's Shared of Cell.
Create "Pointer" as a New Storage's Pointer's Content of Element and Target.
Return a New Storage's Success of Pointer.

Define Storage's "Read|-s||-ing|-er" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~try_borrow~, Receiver: ~<@Value@>~|}<DEF>`.
Define Storage's "Reference" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~&*<@Value@>~|}<DEF>`.
Define Storage's "Address" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~std::rc::Rc::as_ptr~, Inputs: ~&[<@Value@>]&~|}<DEF>`.
Define Storage's "History" as `<DEF>MOB{|Key: ~mut <@Name|Label@>~|}^KY{|Key: ~std::collections::BTreeSet::new~|}<DEF>`.
Define Storage's "Remember|-s|-ed|-ing|-er" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~insert~, Receiver: ~<@Collection@>~, Inputs: ~&[<@Value@>]&~|}<DEF>`.

    Publicly create Storage's "Identify|--ies|--ied|-ing|--ier" as a Generic Storage.Polymorphic Storage.Routine with Inputs of "value" as `&std::rc::Rc<T>` and Outputs of `usize`.
Create "Address" as a New Storage's Address of Value.
Return a New Storage's Integral of Address and `usize`.

    Publicly create Storage's Pointer's "Read|-s||-ing|-er" as a Generic `l_pointer_get` Function with Inputs of "pointer" as `&LPointer` and Outputs of `runtime::Result<LValue>`.
Create "Target" as a New Storage's Pointer's Target of Pointer.
Create "Borrowed" as a New Storage's Reader of Target.
Create "Guard" as a New Storage.Checked Binding of a New Storage's Accepter of Borrowed and `Pointer is already borrowed`.
Create "Value" as a New Storage's Reference of Guard.
Create "Copied" as a New Storage's Value's Copier of Value.
Return a New Storage's Success of Copied.

    Publicly create Storage's Value's "Protect|-s|-ed|-ing|-or" as a Generic Function with Inputs of "value" as `&LValue` and "target" as `usize` and "message" as `&str` and Outputs of `runtime::Result<()>`.
Create "History" as a New Storage's History.
Create "Seen" as a New Storage.Editable Storage's Borrower of History.
Create "Cyclic" as a New Storage.Checked Binding of a New Storage's Value's Reacher of Value, Target, and Seen.
If Cyclic, Create "Failure" as a New Storage's Failure of Message and then Create "Rejected" as a New Storage's Rejection of Failure and then Return Rejected.
Create "Unit" as a New Storage's Nothing.
Return a New Storage's Success of Unit.

    Publicly create Storage's Pointer's "Reach|-es|-ed|-ing|-er" as a Generic `l_pointer_reaches` Function with Inputs of "pointer" as `&LPointer` and "target" as `usize` and "seen" as `&mut std::collections::BTreeSet<usize>` and Outputs of `runtime::Result<bool>`.
Create "Cell" as a New Storage's Pointer's Target of Pointer.
Create "Identity" as a New Storage's Identifier of Cell.
If Identity is Target, Create "Success" as a New Storage's Success of true and then Return Success.
Create "Fresh" as a New Storage's Rememberer of Seen and Identity.
If Fresh is false, Create "Missing" as a New Storage's Success of false and then Return Missing.
Create "Borrowed" as a New Storage's Reader of Cell.
Create "Guard" as a New Storage.Checked Binding of a New Storage's Accepter of Borrowed and `Pointer is already borrowed`.
Create "Value" as a New Storage's Reference of Guard.
Return a New Storage's Value's Reacher of Value, Target, and Seen.

    Publicly create Storage's Pointer's "Update|-s|-d|--ing|--er" as a Generic `l_pointer_set` Function with Inputs of "pointer" as `&LPointer` and "value" as `LValue` and Outputs of `runtime::Result<()>`.
Create "Element" as a New Storage's Pointer's Element of Pointer.
Create "Borrowed" as a New Storage's Borrower of Value.
Create "Checked" as a New Storage's Value's Checker of Borrowed, Element, and `Pointer target type mismatch`.
Storage.Verify Checked.
Create "Cell" as a New Storage's Pointer's Target of Pointer.
Create "Identity" as a New Storage's Identifier of Cell.
Create "Protected" as a New Storage's Value's Protector of Borrowed, Identity, and `Pointer write would create an ownership cycle`.
Storage.Verify Protected.
Create "Editable" as a New Storage.Editable Storage's Reader of Cell.
Create "Checkedguard" as a New Storage.Checked Binding of a New Storage's Accepter of Editable and `Pointer is already borrowed`.
Create "Guard" as a New Storage.Mutable Binding of Checkedguard.
Create "Target" as a New Storage.Editable Storage's Reference of Guard.
Create "Previous" as a New Storage's Replacer of Target and Value.
Create "Unit" as a New Storage's Nothing.
Return a New Storage's Success of Unit.
