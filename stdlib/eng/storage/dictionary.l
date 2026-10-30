Create "Storage" as a Dictionary.
Define Storage's "Routine" as `<DEF>FW{|Func{Key: <@Name|Text@>, Inputs: <@Inputs|Props@>, Outputs: <@Outputs|Texts@>}Func, GenericTypes: ~<@Types|Type@>~|}^(<@Process@>)<DEF>`.
Use Function's Return as Storage.Routine's "Return".
Use Function's Invoke as Storage.Routine's "Invoke".
Define Storage's "Lending" as a New Adjective describing any Storage's Routine.
It describes any Storage's Routine as `<DEF>FW{|Func{Key: <@Name|Text@>, Inputs: <@Inputs|Props@>, Outputs: <@Outputs|Texts@>}Func, GenericTypes: ~<@Types|Type@>~, Lifetimes: ~<@Lifetimes|Type@>~|}^(<@Process@>)<DEF>`.
Define Storage's "Iteration" as `<DEF>BOP^X{|Iterable: ~<@Values@>~, Key: ~item~, RangeArity: ~1~|}^(<@Body@>)<DEF>`.
Define Storage's "Apply|--ies|--ied|-ying|--ier" to `<DEF>MOB{|Key: ~outcome~|}^KY{|Key: ~<@Action@>~, Inputs: ~&[item, <@Context@>]&~|}<DEF>`.
Define Storage's "Indexed" as a New Adjective describing any Storage's Iteration or Storage's Applier.
It describes any Storage's Iteration as `<DEF>BOP^X{|Iterable: ~<@Values@>~, Key: ~item~, Index: ~key~, RangeArity: ~2~|}^(<@Body@>)<DEF>`.
It describes any Storage's Applier as `<DEF>MOB{|Key: ~outcome~|}^KY{|Key: ~<@Action@>~, Inputs: ~&[key, item, <@Context@>]&~|}<DEF>`.
Define Storage's "Discard|-s|-ed|-ing|-er" to `<DEF>OB{|Key: ~_~, Value: ~<@Value@>~|}<DEF>`.
Create Storage's "Traverse|-s|-d|--ing|--er" for Processing a New Storage's Iteration of any Thing and a New Process of a New Storage's Applier of any Thing and any Thing.
Define Storage's "Binding|-s" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~<@Value@>~|}<DEF>`.
Use Storage.Binding as Storage's "Closure".
Define Storage's Closure's "Environment" as `<DEF>MOB{|Key: ~<@Name@>Environment~|}^KY{|Key: ~<@Copy@>~, Inputs: ~&[&<@Name@>.0]&~|}<DEF>`.
Define Storage's Closure's "Context" as `<DEF>OB{|Key: ~<@Name@>Context~, Value: ~&<@Context@>~|}<DEF>`.
Define Storage's "Await|-s|-ed|-ing|-er" to `<DEF>COB{|Key: ~<@Name@>Outcome~|}^DV^(<@Invocation@>)<DEF>`.
Define Storage's "Outcome" as `<DEF>GOB{|Key: ~<@Name@>Result~|}^GI{|Key: ~<@Name@>Outcome?~|}^(E{|Statement: ~std::ops::ControlFlow::Continue(value)~|}*W^(BOD^E{|Statement: ~value~|})*E{|Statement: ~std::ops::ControlFlow::Break(())~|}*<@Exit@>)<DEF>`.
Create Storage's "Invoke|-s|-d|--ing|--er" for Processing a New Storage's Closure, a New Storage's Closure's Environment, a New Storage's Closure's Context, a New Storage's Awaiter, and a New Storage's Outcome.
Define Storage's "Mutable" as a New Adjective describing any Storage's Binding.
It describes any Storage's Binding as `<DEF>OB{|Key: ~mut <@Name|Label@>~, Value: ~<@Value@>~|}<DEF>`.
Define Storage's "Comparison|-s" as `<DEF>DOB{|Key: ~<@Name|Label@>~|}^DE{|Ops: ~&[Op{Type: Calculate, Operator: <@Operator@>, Operands: &[<@Left@>, <@Right@>]&}Op]&~|}<DEF>`.
Define Storage's "Identical" as a New Adjective describing any Storage's Comparison.
It describes any Storage's Comparison as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~std::sync::Arc::ptr_eq~, Inputs: ~&[&<@Left@>, &<@Right@>]&~|}<DEF>`.
Create Storage's "Compare|-s|-d|--ing|--er" for Processing a New Storage's Comparison.
Define Storage's "Produced" as a New Adjective describing any Storage's Binding.
It describes any Storage's Binding as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^(<@Value@>)<DEF>`.

Define Storage's Pointer's "Content" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^Y{|Key: ~LPointer~, Props: ~&[Prop{Key: element, Value: <@Element@>}Prop, Prop{Key: target, Value: <@Target@>}Prop]&~|}<DEF>`.
Define Storage's Layout's "Pointer" as `<DEF>HY{|Key: ~LPointer~, Props: ~&[Prop{Key: element, Type: Vec<u64>}Prop, Prop{Key: target, Type: std::rc::Rc<std::cell::RefCell<<@Value|Type@>>>}Prop]&~|}<DEF>`.
Define Storage's "Pointer" as a New Storage's Layout's Pointer of Storage's Value.
Define Storage's Pointer's "Element" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~&<@Value@>.element~|}<DEF>`.
Define Storage's Pointer's "Target" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~&<@Value@>.target~|}<DEF>`.
Define Storage's "Borrow|-s|-ed|-ing|-er" to `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~&<@Value@>~|}<DEF>`.
Define Storage's "Copy|--ies|--ied|-ing|--ier" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~std::clone::Clone::clone~, Inputs: ~&[<@Value@>]&~|}<DEF>`.
Define Storage's "Retain|-s|-ed|-ing|-er" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~std::rc::Rc::clone~, Inputs: ~&[<@Value@>]&~|}<DEF>`.

Define Storage's "Assign|-s|-ed|-ing|-er" to `<DEF>O{|Key: ~<@Name@>~, Value: ~<@Value@>~|}<DEF>`.
Define Storage's "Declaration|-s" as `<DEF>BO{|Key: ~<@Name@>~, Type: ~<@Type@>~|}<DEF>`.
Define Storage's "Publish|-es|-ed|-ing|-er" to `<DEF>MO{|Key: ~<@Name@>~|}^KY{|Key: ~Some~, Inputs: ~&[<@Value@>]&~|}<DEF>`.
Define Storage's "Presence" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~is_some~, Receiver: ~<@Value@>~|}<DEF>`.
Define Storage's "Vector|-s" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~Vec::from~, Inputs: ~&[[<@Values|Atoms@>]]&~|}<DEF>`.
Define Storage's "Tuple|-s" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~(<@Values|Atoms@>,)~|}<DEF>`.
Define Storage's "Empty" as a New Adjective describing any Process, Storage's Binding, Storage's Vector, or Storage's Tuple.
It describes any Process as `<DEF>W<DEF>`.
It describes any Storage's Binding as `<DEF>OB{|Key: ~mut <@Name|Label@>: Option<<@Type@>>~, Value: ~None~|}<DEF>`.
It describes any Storage's Vector as `<DEF>MOB{|Key: ~<@Name|Label@>: Vec<<@Type|Type@>>~|}^KY{|Key: ~Vec::new~|}<DEF>`.
It describes any Storage's Tuple as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~()~|}<DEF>`.
Define Storage's "Checked" as a New Adjective describing any Storage's Binding.
It describes any Storage's Binding as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~<@Value@>?~|}<DEF>`.

    Publicly create Storage's Pointer's "Copy|--ies|--ied|-ing|--ier" as a Generic Function with Inputs of "value" as `&LPointer` and Outputs of Storage's Pointer.
Create "Element" as a New Storage's Pointer's Element of Value.
Create "Signature" as a New Storage's Copier of Element.
Create "Target" as a New Storage's Pointer's Target of Value.
Create "Retained" as a New Storage's Retainer of Target.
Return a New Storage's Pointer's Content of Signature and Retained.

Create Storage's "Text" as a Dictionary.
Define Storage's Text's "Bytes" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~as_bytes~, Receiver: ~<@Value@>~|}<DEF>`.
Define Storage's "Own|-s|-ed|-ing|-er" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~to_vec~, Receiver: ~<@Value@>~|}<DEF>`.

    Publicly create Storage's Text's "Own|-s|-ed|-ing|-er" as a Generic Function with Inputs of "value" as `&str` and Outputs of `Vec<u8>`.
Create "Bytes" as a New Storage's Text's Bytes of Value.
Return a New Storage's Owner of Bytes.

Define Storage's Value's "Content" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^Y{|Key: ~LValue~, Props: ~&[Prop{Key: ty, Value: <@Signature@>}Prop, Prop{Key: data, Value: <@Data@>}Prop, Prop{Key: show, Value: <@Renderer@>}Prop, Prop{Key: reaches, Value: <@Visitor@>}Prop]&~|}<DEF>`.
Define Storage's "Value" as `<DEF>HY{|Key: ~LValue~, Props: ~&[Prop{Key: ty, Type: Vec<u64>}Prop, Prop{Key: data, Type: std::rc::Rc<dyn std::any::Any>}Prop, Prop{Key: show, Type: fn(&dyn std::any::Any) -> runtime::Result<Vec<u8>>}Prop, Prop{Key: reaches, Type: fn(&dyn std::any::Any, usize, &mut std::collections::BTreeSet<usize>) -> runtime::Result<bool>}Prop]&~|}<DEF>`.
Define Storage's Value's "Signature" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~&<@Value@>.ty~|}<DEF>`.
Define Storage's Value's "Data" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~&<@Value@>.data~|}<DEF>`.
Define Storage's Value's "Renderer" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~<@Value@>.show~|}<DEF>`.
Define Storage's Value's "Visitor" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~<@Value@>.reaches~|}<DEF>`.

    Publicly create Storage's Value's "Copy|--ies|--ied|-ing|--ier" as a Generic `l_value_copy` Function with Inputs of "value" as `&LValue` and Outputs of Storage's Value.
Create "Signature" as a New Storage's Value's Signature of Value.
Create "Copied" as a New Storage's Copier of Signature.
Create "Data" as a New Storage's Value's Data of Value.
Create "Retained" as a New Storage's Retainer of Data.
Create "Renderer" as a New Storage's Value's Renderer of Value.
Create "Visitor" as a New Storage's Value's Visitor of Value.
Return a New Storage's Value's Content of Copied, Retained, Renderer, and Visitor.

Define Storage's Environment's "Content" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^Y{|Key: ~LEnvironment~, Props: ~&[Prop{Key: types, Value: <@Types@>}Prop, Prop{Key: values, Value: <@Values@>}Prop]&~|}<DEF>`.
Define Storage's Layout's "Environment" as `<DEF>HY{|Key: ~LEnvironment~, Props: ~&[Prop{Key: values, Type: std::rc::Rc<std::cell::RefCell<Vec<<@Value|Type@>>>>}Prop, Prop{Key: types, Type: std::rc::Rc<Vec<Vec<u64>>>}Prop]&~|}<DEF>`.
Define Storage's "Environment" as a New Storage's Layout's Environment of Storage's Value.
Define Storage's Environment's "Types" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~&<@Value@>.types~|}<DEF>`.
Define Storage's Environment's "Values" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~&<@Value@>.values~|}<DEF>`.

    Publicly create Storage's Environment's "Copy|--ies|--ied|-ing|--ier" as a Generic `l_environment_copy` Function with Inputs of "value" as `&LEnvironment` and Outputs of Storage's Environment.
Create "Types" as a New Storage's Environment's Types of Value.
Create "Signatures" as a New Storage's Retainer of Types.
Create "Values" as a New Storage's Environment's Values of Value.
Create "Retained" as a New Storage's Retainer of Values.
Return a New Storage's Environment's Content of Signatures and Retained.

Use Storage's Binding as "Binding".
Define Storage's "Refer|-s|-red|-ring|-rer" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~as_ref~, Receiver: ~<@Value@>~|}<DEF>`.

    Publicly create Storage's Value's "Format|-s|-ted|-ting|-ter" as a Generic `l_display` Function with Inputs of "value" as `&LValue` and Outputs of `runtime::Result<Vec<u8>>`.
Create "Data" as a New Storage's Value's Data of Value.
Create "Borrowed" as a New Storage's Referrer of Data.
Create "Renderer" as a New Storage's Value's Renderer of Value.
Return a New Storage.Produced Binding of a New Function's Invoker of Renderer and Borrowed.

Define Storage's List's "Content" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^Y{|Key: ~LList~, Props: ~&[Prop{Key: element, Value: <@Element@>}Prop, Prop{Key: values, Value: <@Values@>}Prop]&~|}<DEF>`.
Define Storage's Layout's "List" as `<DEF>HY{|Key: ~LList~, Props: ~&[Prop{Key: element, Type: Vec<u64>}Prop, Prop{Key: values, Type: Vec<<@Value|Type@>>}Prop]&~|}<DEF>`.
Define Storage's "List" as a New Storage's Layout's List of Storage's Value.
Use Storage.Pointer.Element as Storage.List's "Element".
Use Storage.Environment.Values as Storage.List's "Values".
Define Storage's Vector's "Buffer" as `<DEF>MOB{|Key: ~mut <@Name|Label@>: Vec<<@Type|Type@>>~|}^KY{|Key: ~Vec::new~|}<DEF>`.
Define Storage's "Editable" as a New Adjective describing any Storage's Borrower, Storage's Reader, or Storage's Reference.
It describes any Storage's Borrower as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~&mut <@Value@>~|}<DEF>`.
It describes any Storage's Reader as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~try_borrow_mut~, Receiver: ~<@Value@>~|}<DEF>`.
It describes any Storage's Reference as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~&mut *<@Value@>~|}<DEF>`.
Define Storage's "Append|-s|-ed|-ing|-er" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~push~, Receiver: ~<@Collection@>~, Inputs: ~&[<@Value@>]&~|}<DEF>`.

    Publicly create Storage's List's "Collect|-s|-ed|-ing|-or" as a Generic Function with Inputs of "value" as `&LValue` and "collection" as `&mut Vec<LValue>` and Outputs of `()`.
Create "Copied" as a New Storage's Value's Copier of Value.
Return a New Storage's Appender of Collection and Copied.

    Publicly create Storage's List's "Copy|--ies|--ied|-ing|--ier" as a Generic `l_list_copy` Function with Inputs of "value" as `&LList` and Outputs of Storage's List.
Create "Element" as a New Storage's List's Element of Value.
Create "Signature" as a New Storage's Copier of Element.
Create "Children" as a New Storage's List's Values of Value.
Create "Values" as a New Storage's Vector's Buffer of `LValue`.
Create "Context" as a New Storage.Editable Storage's Borrower of Values.
Storage.Traverse Children and Storage's List's Collector and Context.
Return a New Storage's List's Content of Signature and Values.

Define Storage's Map's "Content" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^Y{|Key: ~LMap~, Props: ~&[Prop{Key: element, Value: <@Element@>}Prop, Prop{Key: values, Value: <@Values@>}Prop]&~|}<DEF>`.
Define Storage's Layout's "Map" as `<DEF>HY{|Key: ~LMap~, Props: ~&[Prop{Key: element, Type: Vec<u64>}Prop, Prop{Key: values, Type: std::collections::BTreeMap<Vec<u8>, <@Value|Type@>>}Prop]&~|}<DEF>`.
Define Storage's "Map" as a New Storage's Layout's Map of Storage's Value.
Use Storage.Pointer.Element as Storage.Map's "Element".
Define Storage's Map's "Entries" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~iter~, Receiver: ~<@Value@>.values~|}<DEF>`.
Define Storage's Map's "Buffer" as `<DEF>MOB{|Key: ~mut <@Name|Label@>: std::collections::BTreeMap<<@Key|Type@>, <@Value|Type@>>~|}^KY{|Key: ~std::collections::BTreeMap::new~|}<DEF>`.
Define Storage's "Insert|-s|-ed|-ing|-er" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~insert~, Receiver: ~<@Collection@>~, Inputs: ~&[<@Key@>, <@Value@>]&~|}<DEF>`.
Create Storage's "Enumerate|-s|-d|--ing|--or" for Processing a New Storage.Indexed Storage's Iteration of any Thing and a New Process of a New Storage.Indexed Storage's Applier of any Thing and any Thing.

    Publicly create Storage's Map's "Collect|-s|-ed|-ing|-or" as a Generic Function with Inputs of "key" as `&Vec<u8>` and "value" as `&LValue` and "collection" as `&mut std::collections::BTreeMap<Vec<u8>, LValue>` and Outputs of `Option<LValue>`.
Create "Copied" as a New Storage's Value's Copier of Value.
Create "Index" as a New Storage's Copier of Key.
Return a New Storage's Inserter of Collection, Index, and Copied.

    Publicly create Storage's Map's "Copy|--ies|--ied|-ing|--ier" as a Generic `l_map_copy` Function with Inputs of "value" as `&LMap` and Outputs of Storage's Map.
Create "Element" as a New Storage's Map's Element of Value.
Create "Signature" as a New Storage's Copier of Element.
Create "Entries" as a New Storage's Map's Entries of Value.
Create "Values" as a New Storage's Map's Buffer of `Vec<u8>` and `LValue`.
Create "Context" as a New Storage.Editable Storage's Borrower of Values.
Storage.Enumerate Entries and Storage's Map's Collector and Context.
Return a New Storage's Map's Content of Signature and Values.

Define Storage's "Polymorphic" as a New Adjective describing any Storage's Routine.
It describes a Generic Storage's Routine of `T`.
Define Storage's "Permanent" as a New Adjective describing any Storage's Routine.
It describes a Generic Storage's Routine of `T: 'static`.
Define Storage's "Fallible" as a New Adjective describing any Storage's Routine.
It describes a Generic Storage's Routine of `T, E`.
Define Storage's "Referenced" as a New Adjective describing any Storage's Routine.
It describes a Generic Storage.Lending Storage's Routine of `T: 'static` and `'a`.
Define Storage's "Failure" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~runtime::error_new_text~, Imports: ~&[runtime::error_new_text as _]&~, Inputs: ~&[<@Message@>]&~|}<DEF>`.
Define Storage's "Demand|-s|-ed|-ing|-er" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~ok_or~, Receiver: ~<@Value@>~, Inputs: ~&[<@Failure@>]&~|}<DEF>`.
Define Storage's "Optional" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~ok~, Receiver: ~<@Value@>~|}<DEF>`.
Define Storage's "Projection" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~downcast_ref~, Receiver: ~<@Value@>~|}<DEF>`.
Define Storage's "Shared" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~std::rc::Rc::new~, Inputs: ~&[<@Value@>]&~|}<DEF>`.
Define Storage's "Success" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~Ok~, Inputs: ~&[<@Value@>]&~|}<DEF>`.

    Publicly create Storage's "Require|-s|-d|--ing|--er" as a Generic Storage.Polymorphic `l_required` Storage.Routine with Inputs of "value" as `Option<T>` and "message" as `&str` and Outputs of `runtime::Result<T>`.
Create "Failure" as a New Storage's Failure of Message.
Return a New Storage's Demander of Value and Failure.

    Publicly create Storage's "Accept|-s|-ed|-ing|-er" as a Generic Storage.Fallible `l_borrowed` Storage.Routine with Inputs of "value" as `Result<T, E>` and "message" as `&str` and Outputs of `runtime::Result<T>`.
Create "Optional" as a New Storage's Optional of Value.
Return a New Storage's Requirer of Optional and Message.

    Publicly create Storage's "Project|-s|-ed|-ing|-or" as a Generic Storage.Referenced `l_project` Storage.Routine with Inputs of "value" as `&'a dyn std::any::Any` and "message" as `&str` and Outputs of `runtime::Result<&'a T>`.
Create "Projected" as a New Storage's Projection of Value.
Return a New Storage's Requirer of Projected and Message.

    Publicly create Storage's "Resolve|-s|-d|--ing|--er" as a Generic Storage.Polymorphic Storage.Routine with Inputs of "value" as `runtime::Result<&T>` and "copy" as `fn(&T) -> T` and Outputs of `runtime::Result<T>`.
Create "Checked" as a New Storage.Checked Binding of Value.
Create "Copied" as a New Storage.Produced Binding of a New Function's Invoker of Copy and Checked.
Return a New Storage's Success of Copied.

    Publicly create Storage's Value's "Wrap|-s|-ped|-ping|-per" as a Generic Storage.Permanent `l_value_wrap` Storage.Routine with Inputs of "value" as `T` and "signature" as `Vec<u64>` and "renderer" as `fn(&dyn std::any::Any) -> runtime::Result<Vec<u8>>` and "visitor" as `fn(&dyn std::any::Any, usize, &mut std::collections::BTreeSet<usize>) -> runtime::Result<bool>` and Outputs of Storage's Value.
Create "Data" as a New Storage's Shared of Value.
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

Define Storage's "Annotation" as `<DEF>OB{|Key: ~<@Name|Label@>: <@Type|Type@>~, Value: ~<@Value@>~|}<DEF>`.
Define Storage's "Description" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~to_string~, Receiver: ~<@Value@>~|}<DEF>`.
Define Storage's "Encoding" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~into_bytes~, Receiver: ~<@Value@>~|}<DEF>`.
Create Storage's "Bytes" as a Dictionary.
Create Storage's "Number" as a Dictionary.
Create Storage's "Truth" as a Dictionary.
Create Storage's "Error" as a Dictionary.
Create Storage's "Null" as a Dictionary.
Create Storage's "App" as a Dictionary.
Create Storage's "Function" as a Dictionary.
Define Storage's Number's "Format" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~runtime::format_number~, Imports: ~&[runtime::format_number as _]&~, Inputs: ~&[*<@Value@>]&~|}<DEF>`.

    Publicly create Storage's Text's "Restore|-s|-d|--ing|--er" as a Generic Function with Inputs of "data" as `&dyn std::any::Any` and "message" as `&str` and Outputs of `runtime::Result<Vec<u8>>`.
Create "Value" as a New Storage.Checked Binding of a New Storage's Annotation of `runtime::Result<&Vec<u8>>` and a New Storage's Projector of Data and Message.
Create "Bytes" as a New Storage's Copier of Value.
Return a New Storage's Success of Bytes.

    Publicly create Storage's Text's "Render|-s|-ed|-ing|-er" as a Generic `l_show_text` Function with Inputs of "data" as `&dyn std::any::Any` and Outputs of `runtime::Result<Vec<u8>>`.
Return a New Storage's Text's Restorer of Data and `expected Text`.

    Publicly create Storage's Bytes's "Render|-s|-ed|-ing|-er" as a Generic `l_show_bytes` Function with Inputs of "data" as `&dyn std::any::Any` and Outputs of `runtime::Result<Vec<u8>>`.
Return a New Storage's Text's Restorer of Data and `expected Bytes`.

    Publicly create Storage's Number's "Render|-s|-ed|-ing|-er" as a Generic `l_show_number` Function with Inputs of "data" as `&dyn std::any::Any` and Outputs of `runtime::Result<Vec<u8>>`.
Create "Value" as a New Storage.Checked Binding of a New Storage's Annotation of `runtime::Result<&f64>` and a New Storage's Projector of Data and `expected Number`.
Create "Bytes" as a New Storage's Number's Format of Value.
Return a New Storage's Success of Bytes.

    Publicly create Storage's Truth's "Render|-s|-ed|-ing|-er" as a Generic `l_show_truth` Function with Inputs of "data" as `&dyn std::any::Any` and Outputs of `runtime::Result<Vec<u8>>`.
Create "Value" as a New Storage.Checked Binding of a New Storage's Annotation of `runtime::Result<&bool>` and a New Storage's Projector of Data and `expected Truth`.
Create "Text" as a New Storage's Description of Value.
Create "Bytes" as a New Storage's Encoding of Text.
Return a New Storage's Success of Bytes.

    Publicly create Storage's Error's "Render|-s|-ed|-ing|-er" as a Generic `l_show_error` Function with Inputs of "data" as `&dyn std::any::Any` and Outputs of `runtime::Result<Vec<u8>>`.
Create "Value" as a New Storage.Checked Binding of a New Storage's Annotation of `runtime::Result<&runtime::Error>` and a New Storage's Projector of Data and `expected Error`.
Create "Text" as a New Storage's Description of Value.
Create "Bytes" as a New Storage's Encoding of Text.
Return a New Storage's Success of Bytes.

    Publicly create Storage's Null's "Render|-s|-ed|-ing|-er" as a Generic `l_show_null` Function with Inputs of "data" as `&dyn std::any::Any` and Outputs of `runtime::Result<Vec<u8>>`.
Create "Bytes" as a New Storage's Text's Owner of `null`.
Return a New Storage's Success of Bytes.

    Publicly create Storage's App's "Render|-s|-ed|-ing|-er" as a Generic `l_show_app` Function with Inputs of "data" as `&dyn std::any::Any` and Outputs of `runtime::Result<Vec<u8>>`.
Create "Bytes" as a New Storage's Text's Owner of `App`.
Return a New Storage's Success of Bytes.

    Publicly create Storage's Function's "Render|-s|-ed|-ing|-er" as a Generic `l_show_function` Function with Inputs of "data" as `&dyn std::any::Any` and Outputs of `runtime::Result<Vec<u8>>`.
Create "Bytes" as a New Storage's Text's Owner of `Function`.
Return a New Storage's Success of Bytes.

    Publicly create Storage's Environment's "Render|-s|-ed|-ing|-er" as a Generic `l_show_environment` Function with Inputs of "data" as `&dyn std::any::Any` and Outputs of `runtime::Result<Vec<u8>>`.
Create "Failure" as a New Storage's Failure of `capture environment is not an L value`.
Return a New Storage's Rejection of Failure.

Define Storage's Function's "Environment" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~&<@Value@>.0~|}<DEF>`.
Define Storage's Function's "Target" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~<@Value@>.1~|}<DEF>`.
Define Storage's Function's "Copyable" as a New Adjective describing any Storage's Routine.
It describes a Generic Storage's Routine of `T: Copy`.

    Publicly create Storage's Function's "Copy|--ies|--ied|-ing|--ier" as a Generic Storage.Function.Copyable `l_function_copy` Storage.Routine with Inputs of "value" as `&(LEnvironment, T)` and Outputs of `(LEnvironment, T)`.
Create "Environment" as a New Storage's Function's Environment of Value.
Create "Copied" as a New Storage's Environment's Copier of Environment.
Create "Target" as a New Storage's Function's Target of Value.
Return a New Storage's Tuple of Copied and Target.

Define Storage's "Nothing" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~()~|}<DEF>`.
Define Storage's "Verify|--ies|--ied|-ing|--ier" to `<DEF>OB{|Key: ~_~, Value: ~<@Value@>?~|}<DEF>`.
Define Storage's "Get|-s|--ot|-ting|-ter" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~get~, Receiver: ~<@Collection@>~, Inputs: ~&[<@Index@>]&~|}<DEF>`.
Define Storage's "Locate|-s|-d|--ing|--or" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~get_mut~, Receiver: ~<@Collection@>~, Inputs: ~&[<@Index@>]&~|}<DEF>`.
Define Storage's "Replace|-s|-d|--ing|--er" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~std::mem::replace~, Inputs: ~&[<@Target@>, <@Value@>]&~|}<DEF>`.
Define Storage's "Propagate|-s|-d|--ing|--or" to `<DEF>OB{|Key: ~_~, Value: ~outcome?~|}<DEF>`.
Create Storage's "Examine|-s|-d|--ing|--er" for Processing a New Storage's Applier of any Thing and any Thing and then Processing a New Storage's Propagator.
Create Storage's "Inspect|-s|-ed|-ing|-or" for Processing a New Storage's Iteration of any Thing and a New Process of a New Storage's Examiner of any Thing and any Thing.

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
Create "Failure" as a New Storage's Failure of `index must be a nonnegative integer`.
Create "Rejected" as a New Storage's Rejection of Failure.
If Finite is false, Return Rejected.
If Index less than 0, Return Rejected.
Create "Fraction" as a New Storage's Number's Fraction of Index.
If Fraction isnt 0, Return Rejected.
Create "Limit" as a New Storage's Index's Limit.
If Index greater than Limit, Return Rejected.
If Index is Limit, Return Rejected.
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

Create Storage's "Layout" as a Dictionary.
Define Storage's Layout's "Program" as `<DEF>HY{|Key: ~LProgram~, Props: ~&[Prop{Key: apps, Type: Vec<std::sync::Arc<runtime::App>>}Prop, Prop{Key: types, Type: Vec<Vec<u64>>}Prop, Prop{Key: values, Type: std::cell::RefCell<Vec<Option<<@Value|Type@>>>>}Prop, Prop{Key: actions, Type: std::cell::RefCell<Vec<Option<<@Environment|Type@>>>>}Prop]&~|}<DEF>`.
Define Storage's "Program" as a New Storage's Layout's Program of Storage's Value and Storage's Environment.
Define Storage's Program's "Content" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^Y{|Key: ~LProgram~, Props: ~&[Prop{Key: apps, Value: <@Apps@>}Prop, Prop{Key: types, Value: <@Types@>}Prop, Prop{Key: values, Value: <@Values@>}Prop, Prop{Key: actions, Value: <@Actions@>}Prop]&~|}<DEF>`.
Define Storage's Program's "Values" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~&<@Program@>.values~|}<DEF>`.
Define Storage's Program's "Actions" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~&<@Program@>.actions~|}<DEF>`.
Define Storage's Program's "Types" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~&<@Program@>.types~|}<DEF>`.
Define Storage's "Length" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~len~, Receiver: ~<@Value@>~|}<DEF>`.
Define Storage's "Range" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^Y{|Key: ~std::ops::Range~, Props: ~&[Prop{Key: start, Value: 0_usize}Prop, Prop{Key: end, Value: <@Limit@>}Prop]&~|}<DEF>`.
Define Storage's "Present" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~Some~, Inputs: ~&[<@Value@>]&~|}<DEF>`.
Define Storage's "Absent" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~None~|}<DEF>`.
Create Storage's "Slot" as a Dictionary.

    Publicly create Storage's Slot's "Populate|-s|-d|--ing|--or" as a Generic Storage.Polymorphic Storage.Routine with Inputs of "index" as `usize` and "values" as `&mut Vec<Option<T>>` and Outputs of `()`.
Create "Missing" as a New Storage's Absent.
Return a New Storage's Appender of Values and Missing.

    Publicly create Storage's Slot's "Allocate|-s|-d|--ing|--or" as a Generic Storage.Polymorphic Storage.Routine with Inputs of "count" as `usize` and Outputs of `std::cell::RefCell<Vec<Option<T>>>`.
Create "Values" as a New Storage's Vector's Buffer of `Option<T>`.
Create "Indices" as a New Storage's Range of Count.
Create "Context" as a New Storage.Editable Storage's Borrower of Values.
Storage.Traverse Indices and Storage's Slot's Populator and Context.
Return a New Storage's Cell of Values.

    Publicly create Storage's Slot's "Read|-s||-ing|-er" as a Generic Storage.Polymorphic Storage.Routine with Inputs of "slots" as `&std::cell::RefCell<Vec<Option<T>>>` and "index" as `usize` and "copy" as `fn(&T) -> T` and "borrowed" as `&str` and "missing" as `&str` and Outputs of `runtime::Result<T>`.
Create "Attempt" as a New Storage's Reader of Slots.
Create "Values" as a New Storage.Checked Binding of a New Storage's Accepter of Attempt and Borrowed.
Create "Found" as a New Storage's Getter of Values and Index.
Create "Optional" as a New Storage.Checked Binding of a New Storage's Requirer of Found and Missing.
Create "Referenced" as a New Storage's Referrer of Optional.
Create "Initialized" as a New Storage's Requirer of Referenced and Missing.
Return a New Storage's Resolver of Initialized and Copy.

    Publicly create Storage's Slot's "Write|-s|----ote|--ing|-r" as a Generic Storage.Polymorphic Storage.Routine with Inputs of "slots" as `&std::cell::RefCell<Vec<Option<T>>>` and "index" as `usize` and "value" as `T` and "borrowed" as `&str` and "missing" as `&str` and Outputs of `runtime::Result<()>`.
Create "Attempt" as a New Storage.Editable Storage's Reader of Slots.
Create "Checked" as a New Storage.Checked Binding of a New Storage's Accepter of Attempt and Borrowed.
Create "Values" as a New Storage.Mutable Binding of Checked.
Create "Found" as a New Storage's Locator of Values and Index.
Create "Target" as a New Storage.Checked Binding of a New Storage's Requirer of Found and Missing.
Create "Present" as a New Storage's Present of Value.
Create "Previous" as a New Storage's Replacer of Target and Present.
Create "Unit" as a New Storage's Nothing.
Return a New Storage's Success of Unit.

    Publicly create Storage's Program's "Create|-s|-d|--ing|--or" as a Generic `l_program_new` Function with Inputs of "apps" as `Vec<std::sync::Arc<runtime::App>>` and "types" as `Vec<Vec<u64>>` and "action_count" as `usize` and Outputs of Storage's Program.
Create "Count" as a New Storage's Length of Types.
Create "Values" as a New Storage's Slot's Allocator of Count.
Create "Actions" as a New Storage's Slot's Allocator of Action_count.
Return a New Storage's Program's Content of Apps, Types, Values, and Actions.

    Publicly create Storage's Program's Values's "Read|-s||-ing|-er" as a Generic `l_program_get` Function with Inputs of "program" as `&LProgram` and "index" as `usize` and Outputs of `runtime::Result<LValue>`.
Create "Slots" as a New Storage's Program's Values of Program.
Return a New Storage's Slot's Reader of Slots, Index, Storage's Value's Copier, `program storage is already borrowed`, and `program value is not initialized`.

    Publicly create Storage's Program's Actions's "Read|-s||-ing|-er" as a Generic `l_program_action` Function with Inputs of "program" as `&LProgram` and "index" as `usize` and Outputs of `runtime::Result<LEnvironment>`.
Create "Slots" as a New Storage's Program's Actions of Program.
Return a New Storage's Slot's Reader of Slots, Index, Storage's Environment's Copier, `program actions are already borrowed`, and `program action is not initialized`.

    Publicly create Storage's Program's Values's "Write|-s|----ote|--ing|-r" as a Generic `l_program_set` Function with Inputs of "program" as `&LProgram` and "index" as `usize` and "value" as `LValue` and Outputs of `runtime::Result<()>`.
Create "Types" as a New Storage's Program's Types of Program.
Create "Found" as a New Storage's Getter of Types and Index.
Create "Expected" as a New Storage.Checked Binding of a New Storage's Requirer of Found and `program storage index out of bounds`.
Create "Reference" as a New Storage's Borrower of Value.
Create "Accepted" as a New Storage's Value's Checker of Reference, Expected, and `program value type mismatch`.
Storage.Verify Accepted.
Create "Slots" as a New Storage's Program's Values of Program.
Return a New Storage's Slot's Writer of Slots, Index, Value, `program storage is already borrowed`, and `program storage index out of bounds`.

    Publicly create Storage's Program's Actions's "Write|-s|----ote|--ing|-r" as a Generic `l_program_bind_action` Function with Inputs of "program" as `&LProgram` and "index" as `usize` and "value" as `LEnvironment` and Outputs of `runtime::Result<()>`.
Create "Slots" as a New Storage's Program's Actions of Program.
Return a New Storage's Slot's Writer of Slots, Index, Value, `program actions are already borrowed`, and `program action index out of bounds`.

Define Storage's "Decide|-s|-d|--ing|--er" to `<DEF>BOG^GI^(E{|Statement: ~outcome?~|}*W^(<@Return@>))<DEF>`.
Create Storage's "Probe|-s|-d|--ing|--er" for Processing a New Storage's Applier of any Thing and any Thing and then Processing a New Storage's Decider of any Process.
Create Storage's "Search|-es|-ed|-ing|-er" for Processing a New Storage's Iteration of any Thing and a New Process of a New Storage's Prober of any Thing, any Thing, and any Process.
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
Create Storage's "Review|-s|-ed|-ing|-er" for Processing a New Storage.Indexed Storage's Applier of any Thing and any Thing and then Processing a New Storage's Propagator.
Create Storage's "Audit|-s|-ed|-ing|-or" for Processing a New Storage.Indexed Storage's Iteration of any Thing and a New Process of a New Storage's Reviewer of any Thing and any Thing.

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
Create "Shared" as a New Storage's Shared of Cell.
Create "Signature" as a New Storage's Shared of Types.
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
Create "Prefix" as a New Storage's Text's Owner of Opening.
Create "Suffix" as a New Storage's Text's Owner of Closing.
Create "Parts" as a New Storage's Vector of Prefix, Value, and Suffix.
Return a New Storage's Text's Joiner of Parts and ``.

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

Define Storage's "Cloneable" as a New Adjective describing any Storage's Routine.
It describes a Generic Storage's Routine of `T: Clone`.
Define Storage's "Specialize|-s|-d|--ing|--er" to `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~<@Function@>::<<@Type|Type@>>~|}<DEF>`.

    Publicly create Storage's "Clone|-s|-d|--ing|--er" as a Generic Storage.Cloneable Storage.Routine with Inputs of "value" as `&T` and Outputs of `T`.
Return a New Storage's Copier of Value.

Define Storage's "Future" as `<DEF>COB{|Key: ~<@Name|Label@>~|}^CV{|Capture: ~Move~|}^(<@Body@>)<DEF>`.
Define Storage's Function's "Pending" as `<DEF>COB{|Key: ~lPending~|}^CV{|Capture: ~Move~|}^(<@Body@>)<DEF>`.
Define Storage's Function's "Pinned" as `<DEF>MOB{|Key: ~lPinned~|}^KY{|Key: ~Box::pin~, Inputs: ~&[lPending]&~|}<DEF>`.
Define Storage's Function's "Ready" as `<DEF>BOD^E{|Statement: ~lPinned~|}<DEF>`.
Create Storage's Function's "Suspend|-s|-ed|-ing|-er" for Processing a New Storage's Function's Pending, a New Storage's Function's Pinned, and a New Storage's Function's Ready.
Define Storage's Function's "Continue|-s|-d|--ing|--er" to `<DEF>MOB{|Key: ~lReturned~|}^KY{|Key: ~std::ops::ControlFlow::Continue~, Inputs: ~&[<@Values@>]&~|}<DEF>`.
Define Storage's Function's "Success" as `<DEF>MOB{|Key: ~lSuccess~|}^KY{|Key: ~Ok~, Inputs: ~&[lReturned]&~|}<DEF>`.
Define Storage's Function's "Return" as `<DEF><@Return@><DEF>`.
Create Storage's Function's "Yield|-s|-ed|-ing|-er" for Processing a New Storage's Function's Continuer, a New Storage's Function's Success, and a New Storage's Function's Return.
Define Storage's Function's "Completion" as `<DEF>MOB{|Key: ~lExit~|}^KY{|Key: ~Ok~, Inputs: ~&[<@Value@>]&~|}<DEF>`.
Create Storage's Function's "Exit|-s|-ed|-ing|-er" for Processing a New Storage's Function's Completion and a New Storage's Function's Return.
Define Storage's Function's "Capture" as `<DEF>MOB{|Key: ~<@Name@>Environment~|}^(<@Environment@>)<DEF>`.
Define Storage's Function's "Pair" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~(<@Name|Label@>Environment?, <@Function@> as <@Type@>)~|}<DEF>`.
Create Storage's Function's "Bind|-s|---ound|-ing|-er" for Processing a New Storage's Function's Capture and a New Storage's Function's Pair.
Define Storage's Rejection's "Error" as `<DEF>MOB{|Key: ~<@Name@>Error~|}^KY{|Key: ~runtime::error_new_text~, Imports: ~&[runtime::error_new_text as _]&~, Inputs: ~&[<@Message|Text@>]&~|}<DEF>`.
Define Storage's Rejection's "Result" as `<DEF>MOB{|Key: ~<@Name@>Result~|}^KY{|Key: ~Err~, Inputs: ~&[<@Name@>Error]&~|}<DEF>`.
Create Storage's "Reject|-s|-ed|-ing|-or" for Processing a New Storage's Rejection's Error, a New Storage's Rejection's Result, and a New Storage's Function's Return.

Define Storage's "Select|-s|-ed|-ing|-or" to `<DEF>BOG^GI^(E{|Statement: ~<@Condition@>~|}*<@Then@>*E{|Statement: ~Void~|}*<@Else@>)<DEF>`.
Define Storage's "Negation" as `<DEF>DOB{|Key: ~<@Name|Label@>~|}^DE{|Ops: ~&[Op{Type: Calculate, Operator: ==, Operands: &[<@Value@>, false]&}Op]&~|}<DEF>`.
Define Storage's "Completion" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~Ok~, Inputs: ~&[()]&~|}<DEF>`.
Define Storage's "Test" as `<DEF>[["E{|Statement: ~#[cfg(test)] pub(crate)~|}", "<@Function@>"]]<DEF>`.
Define Storage's "Scope" as `<DEF>BOP^X^W^(BOK^<@Body@>*BOD^E{|Statement: ~break;~|})<DEF>`.
Define Storage's "Loop" as `<DEF>BOP^X^(<@Body@>)<DEF>`.
Define Storage's Function's "Stop" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~std::ops::ControlFlow::Break~, Inputs: ~&[()]&~|}<DEF>`.
Define Storage's "Literal" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~<@Value|Thing@>~|}<DEF>`.

Define Storage's "Lent" as a New Adjective describing any Storage's Routine.
It describes a Generic Storage.Lending Storage's Routine of `T` and `'a`.

    Publicly create Storage's "Access|-es|-ed|-ing|-or" as a Generic Storage.Lent Storage.Routine with Inputs of "value" as `&'a Option<T>` and "message" as `&str` and Outputs of `runtime::Result<&'a T>`.
Create "Optional" as a New Storage's Referrer of Value.
Return a New Storage's Requirer of Optional and Message.

    Publicly create Storage's Environment's "Require|-s|-d|--ing|--er" as a Generic Function with Inputs of "value" as `&Option<LEnvironment>` and "message" as `&str` and Outputs of `runtime::Result<LEnvironment>`.
Create "Borrowed" as a New Storage.Checked Binding of a New Storage's Accessor of Value and Message.
Create "Copied" as a New Storage's Environment's Copier of Borrowed.
Return a New Storage's Success of Copied.

Create Storage's "Context" as a Dictionary.
Define Storage's Context's "Runtime" as `<DEF>MOB{|Key: ~<@Name@>Runtime~|}^KY{|Key: ~clone~, Receiver: ~<@Runtime@>~|}<DEF>`.
Define Storage's Context's "Pair" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~(<@Name|Label@>Runtime, <@Program@>)~|}<DEF>`.
Define Storage's Context's "Binding" as `<DEF>OB{|Key: ~<@Binding@>~, Value: ~&<@Name@>.1~|}<DEF>`.
Create Storage's Context's "Create|-s|-d|--ing|--or" for Processing a New Storage's Context's Runtime, a New Storage's Context's Pair, and a New Storage's Context's Binding.
Define Storage's Context's "First" as `<DEF>OB{|Key: ~runtime~, Value: ~&<@Context@>.0~|}<DEF>`.
Define Storage's Context's "Second" as `<DEF>OB{|Key: ~lProgram~, Value: ~&<@Context@>.1~|}<DEF>`.
Create Storage's Context's "Read|-s||-ing|-er" for Processing a New Storage's Context's First and a New Storage's Context's Second.
Define Storage's Function's "Snapshot" as `<DEF>MOB{|Key: ~<@Name@>Environment~|}^KY{|Key: ~<@Copy@>~, Inputs: ~&[&<@Environment@>]&~|}<DEF>`.
Define Storage's Function's "Identity" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~(<@Name|Label@>Environment, <@Function@> as <@Type@>)~|}<DEF>`.
Create Storage's Function's "Retain|-s|-ed|-ing|-er" for Processing a New Storage's Function's Snapshot and a New Storage's Function's Identity.
Define Storage's Program's "Apps" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~&<@Program@>.apps~|}<DEF>`.
Define Storage's App's "Retain|-s|-ed|-ing|-er" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~std::sync::Arc::clone~, Inputs: ~&[<@Value@>]&~|}<DEF>`.

    Publicly create Storage's Program's "Select|-s|-ed|-ing|-or" as a Generic Function with Inputs of "program" as `&LProgram` and "index" as `usize` and Outputs of `runtime::Result<std::sync::Arc<runtime::App>>`.
Create "Apps" as a New Storage's Program's Apps of Program.
Create "Slot" as a New Storage's Getter of Apps and Index.
Create "Borrowed" as a New Storage.Checked Binding of a New Storage's Requirer of Slot and `App handle is not owned by this program`.
Create "Retained" as a New Storage's App's Retainer of Borrowed.
Return a New Storage's Success of Retained.

Define Storage's App's "Path" as `<DEF>MOB{|Key: ~<@Name@>Path~|}^KY{|Key: ~std::path::PathBuf::from~, Inputs: ~&[<@Path|Text@>]&~|}<DEF>`.
Define Storage's App's "Handle" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~runtime::runtime_app~, Imports: ~&[runtime::runtime_app as _]&~, Inputs: ~&[runtime, <@Name|Label@>Path]&~|}<DEF>`.
Create Storage's App's "Create|-s|-d|--ing|--or" for Processing a New Storage's App's Path and a New Storage's App's Handle.
Define Storage's App's "Source" as `<DEF>MOB{|Key: ~lAppSource~|}^KY{|Key: ~String::from~, Inputs: ~&[<@Source|Text@>]&~|}<DEF>`.
Define Storage's App's "Attribution" as `<DEF>MOB{|Key: ~_~|}^KY{|Key: ~runtime::app_set_source~, Imports: ~&[runtime::app_set_source as _]&~, Inputs: ~&[&<@App@>, lAppSource]&~|}<DEF>`.
Create Storage's App's "Attribute|-s|-d|--ing|--or" for Processing a New Storage's App's Source and a New Storage's App's Attribution.
Define Storage's App's "Output" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~runtime::app_output~, Imports: ~&[runtime::app_output as _]&~, Inputs: ~&[&<@App@>]&~|}<DEF>`.
Define Storage's App's "Error" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~runtime::app_error~, Imports: ~&[runtime::app_error as _]&~, Inputs: ~&[&<@App@>]&~|}<DEF>`.
Define Storage's Error's "Create|-s|-d|--ing|--or" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~runtime::error_new~, Imports: ~&[runtime::error_new as _]&~, Inputs: ~&[<@Value@>]&~|}<DEF>`.
Define Storage's "Show|-s|-ed|-ing|-er" to `<DEF>COB{|Key: ~<@Name|Label@>~|}^DV^KY{|Key: ~runtime::runtime_show~, Imports: ~&[runtime::runtime_show as _]&~, Inputs: ~&[runtime, &<@Value@>]&~|}<DEF>`.
Define Storage's Program's "Failure" as `<DEF>MOB{|Key: ~<@Name@>Error~|}^KY{|Key: ~runtime::runtime_fail~, Imports: ~&[runtime::runtime_fail as _]&~, Inputs: ~&[runtime, <@Value@>]&~|}<DEF>`.
Define Storage's Program's "Rejection" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~Err~, Inputs: ~&[<@Name|Label@>Error]&~|}<DEF>`.
Create Storage's Program's "Fail|-s|-ed|-ing|-er" for Processing a New Storage's Program's Failure and a New Storage's Program's Rejection.

Create Storage's "Entry" as a Dictionary.
Define Storage's Entry's "Future" as `<DEF>MOB{|Key: ~lObservedFuture~|}^KY{|Key: ~l_main~, Inputs: ~&[runtime]&~|}<DEF>`.
Define Storage's Entry's "Outcome" as `<DEF>COB{|Key: ~lObservedStatus~|}^DV^KY{|Key: ~runtime::runtime_run~, Imports: ~&[runtime::runtime_run as _]&~, Inputs: ~&[runtime, lObservedFuture]&~|}<DEF>`.
Create Storage's Entry's "Observe|-s|-d|--ing|--er" for Processing a New Storage's Entry's Future, a New Storage's Entry's Outcome, and a New Storage's Function's Return.
