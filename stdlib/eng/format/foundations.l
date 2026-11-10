Use Storage from `storage`.
Use Storage.Presence as "Presence".
Use Storage.Equality as "Equality".
Use Storage.Description as "Description".
Use Storage.Encoding as "Encoding".
Use Storage.Length as "Length".
Use Storage.Binding as "Binding".
Use Storage.Mutable as "Mutable".
Use Storage.Present as "Present".
Use Storage.Absent as "Absent".
Use Storage.Sum as "Sum".
Use Storage.Choice as "Choice".
Use Storage.Agreement as "Agreement".
Use Storage.Unwrap as "Unwrap".
Use Storage.Tuple.First as "First".
Use Storage.Tuple.Second as "Second".

*** Private numeric foundations. These are native value operations, not a second
    implementation of formatting policy. The named Functions compose them. ***
Define "Magnitude" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~abs~, Receiver: ~<@Value@>~|}<DEF>`.
Define "Representation" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~to_bits~, Receiver: ~<@Value@>~|}<DEF>`.
Define "Trailing" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~trailing_zeros~, Receiver: ~<@Value@>~|}<DEF>`.
Define "Difference" as `<DEF>DOB{|Key: ~<@Name|Label@>~|}^DE{|Ops: ~&[Op{Type: Calculate, Operator: -, Operands: &[<@Left@>, <@Right@>]&}Op]&~|}<DEF>`.
Define "Product" as `<DEF>DOB{|Key: ~<@Name|Label@>~|}^DE{|Ops: ~&[Op{Type: Calculate, Operator: *, Operands: &[<@Left@>, <@Right@>]&}Op]&~|}<DEF>`.
Define "Quotient" as `<DEF>DOB{|Key: ~<@Name|Label@>~|}^DE{|Ops: ~&[Op{Type: Calculate, Operator: /, Operands: &[<@Left@>, <@Right@>]&}Op]&~|}<DEF>`.
Define "Remainder" as `<DEF>DOB{|Key: ~<@Name|Label@>~|}^DE{|Ops: ~&[Op{Type: Calculate, Operator: %, Operands: &[<@Left@>, <@Right@>]&}Op]&~|}<DEF>`.

*** Arithmetic and storage leaves keep native atoms local. Control, ordering,
    recurrence and return behavior are authored by the Functions below. ***
Define "Integer" as a New Storage's Caster of any Thing and `i32`.
Define "Parse|-s|-d|--ing|--er" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~parse~, Receiver: ~<@Value@>~, GenericTypes: ~i32~|}<DEF>`.
Define "Wide" as a New Adjective describing any Integer or Parser.
It describes any Integer as a New Storage's Caster of any Thing and `u128`.
It describes any Parser as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~parse~, Receiver: ~<@Value@>~, GenericTypes: ~u128~|}<DEF>`.
Define "Indexed" as a New Adjective describing any Integer.
It describes any Integer as a New Storage's Caster of any Thing and `usize`.
Define "Unsigned" as a New Adjective describing any Integer.
It describes any Integer as a New Storage's Caster of any Thing and `u64`.
Define "Direction" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~is_positive~, Receiver: ~<@Value@>~|}<DEF>`.
Define "Multiple" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~checked_mul~, Receiver: ~<@Value@>~, Inputs: ~&[<@Factor@>]&~|}<DEF>`.


Define "Shift|-s|-ed|-ing|-er" to `<DEF>DOB{|Key: ~<@Name|Label@>~|}^DE{|Ops: ~&[Op{Type: Calculate, Operator: >>, Operands: &[<@Value@>, <@Count@>]&}Op]&~|}<DEF>`.
Define "Leftward" as a New Adjective describing any Shifter.
It describes any Shifter as `<DEF>DOB{|Key: ~<@Name|Label@>~|}^DE{|Ops: ~&[Op{Type: Calculate, Operator: <<, Operands: &[<@Value@>, <@Count@>]&}Op]&~|}<DEF>`.
Define "Intersection" as `<DEF>DOB{|Key: ~<@Name|Label@>~|}^DE{|Ops: ~&[Op{Type: Calculate, Operator: &, Operands: &[<@Left@>, <@Right@>]&}Op]&~|}<DEF>`.
Define "Union" as `<DEF>DOB{|Key: ~<@Name|Label@>~|}^DE{|Ops: ~&[Op{Type: Calculate, Operator: |, Operands: &[<@Left@>, <@Right@>]&}Op]&~|}<DEF>`.

Define "Notation" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~format!~, Inputs: ~&[\"{:e}\", <@Value@>]&~|}<DEF>`.
Define "Split|-s|-|-ting|-ter" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~split_once~, Receiver: ~<@Value@>~, Inputs: ~&[<@Separator|Text@>]&~|}<DEF>`.
Define "Trim|-s|-med|-ming|-mer" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~trim_start_matches~, Receiver: ~<@Value@>~, Inputs: ~&[<@Prefix|Text@>]&~|}<DEF>`.
Define "Replacement" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~replace~, Receiver: ~<@Value@>~, Inputs: ~&[<@From|Text@>, <@To|Text@>]&~|}<DEF>`.
Define "Interval" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^Y{|Key: ~std::ops::Range~, Props: ~&[Prop{Key: start, Value: <@Minimum@>}Prop, Prop{Key: end, Value: <@Maximum@>}Prop]&~|}<DEF>`.
Define "Membership" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~contains~, Receiver: ~<@Range@>~, Inputs: ~&[&<@Value@>]&~|}<DEF>`.
Define "Insert|-s|-ed|-ing|-er" to `<DEF>MOB{|Key: ~_~|}^KY{|Key: ~insert_str~, Receiver: ~<@Value@>~, Inputs: ~&[<@Index@>, <@Text|Text@>]&~|}<DEF>`.
Define "Sign" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~is_sign_negative~, Receiver: ~<@Value@>~|}<DEF>`.
Define "Zeros" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~repeat~, Receiver: ~\"0\"~, Inputs: ~&[<@Count@>]&~|}<DEF>`.
Define "Prefix" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~format!~, Inputs: ~&[\"0.{}{}\", <@Zeros@>, <@Digits@>]&~|}<DEF>`.
Define "Suffix" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~format!~, Inputs: ~&[\"{}{}\", <@Digits@>, <@Zeros@>]&~|}<DEF>`.
Define "Scientific" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~format!~, Inputs: ~&[\"{}e{:+03}\", <@Digits@>, <@Exponent@>]&~|}<DEF>`.
