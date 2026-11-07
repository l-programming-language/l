
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

