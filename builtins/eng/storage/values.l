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
It describes any Storage's Binding as `<DEF>OB{|Key: ~<@Checked|Label@>~, Value: ~<@Attempt@>?~|}<DEF>`.

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
Create Storage's "Enumerate|-s|-d|--ing|--or" for Processing a New Storage.Paired Storage's Iteration of any Thing and a New Process with Inputs of "First" as `_` and "Second" as `_` for Creating "Outcome" as a New Storage.Paired Storage's Applier of any Thing, First, Second, and any Thing.

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

Define Storage's Tuple's "First" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~<@Value@>.0~|}<DEF>`.
Define Storage's Tuple's "Second" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~<@Value@>.1~|}<DEF>`.
