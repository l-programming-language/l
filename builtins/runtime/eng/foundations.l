*** Private numeric foundations. These are native value operations, not a second
    implementation of formatting policy. The named Functions compose them. ***
Use Storage from `storage`.
Use Storage.Presence as "Presence".
Use Storage.Equality as "Equality".
Use Storage.Description as "Description".
Use Storage.Encoding as "Encoding".
Use Storage.Length as "Length".
Use Storage.Binding as "Binding".
Use Storage.Mutable as "Mutable".
Define "Absolute" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~abs~, Receiver: ~<@Value@>~|}<DEF>`.
Define "Representation" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~to_bits~, Receiver: ~<@Value@>~|}<DEF>`.
Define "Trailing" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~trailing_zeros~, Receiver: ~<@Value@>~|}<DEF>`.
Define "Difference" as `<DEF>DOB{|Key: ~<@Name|Label@>~|}^DE{|Ops: ~&[Op{Type: Calculate, Operator: -, Operands: &[<@Left@>, <@Right@>]&}Op]&~|}<DEF>`.
Define "Sum" as `<DEF>DOB{|Key: ~<@Name|Label@>~|}^DE{|Ops: ~&[Op{Type: Calculate, Operator: +, Operands: &[<@Left@>, <@Right@>]&}Op]&~|}<DEF>`.
Define "Product" as `<DEF>DOB{|Key: ~<@Name|Label@>~|}^DE{|Ops: ~&[Op{Type: Calculate, Operator: *, Operands: &[<@Left@>, <@Right@>]&}Op]&~|}<DEF>`.
Define "Quotient" as `<DEF>DOB{|Key: ~<@Name|Label@>~|}^DE{|Ops: ~&[Op{Type: Calculate, Operator: /, Operands: &[<@Left@>, <@Right@>]&}Op]&~|}<DEF>`.
Define "Remainder" as `<DEF>DOB{|Key: ~<@Name|Label@>~|}^DE{|Ops: ~&[Op{Type: Calculate, Operator: %, Operands: &[<@Left@>, <@Right@>]&}Op]&~|}<DEF>`.

*** Arithmetic and storage leaves keep native atoms local. Control, ordering,
    recurrence and return behavior are authored by the Functions below. ***
Define "Integer" as `<DEF>DOB{|Key: ~<@Name|Label@>~|}^DE{|Ops: ~&[Op{Type: Calculate, Operator: as, Operands: &[<@Value@>, i32]&}Op]&~|}<DEF>`.
Define "Parse|-s|-d|--ing|--er" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~parse~, Receiver: ~<@Value@>~, GenericTypes: ~i32~|}<DEF>`.
Define "Wide" as a New Adjective describing any Integer or Parser.
It describes any Integer as `<DEF>DOB{|Key: ~<@Name|Label@>~|}^DE{|Ops: ~&[Op{Type: Calculate, Operator: as, Operands: &[<@Value@>, u128]&}Op]&~|}<DEF>`.
It describes any Parser as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~parse~, Receiver: ~<@Value@>~, GenericTypes: ~u128~|}<DEF>`.
Define "Indexed" as a New Adjective describing any Integer.
It describes any Integer as `<DEF>DOB{|Key: ~<@Name|Label@>~|}^DE{|Ops: ~&[Op{Type: Calculate, Operator: as, Operands: &[<@Value@>, usize]&}Op]&~|}<DEF>`.
Define "Unsigned" as a New Adjective describing any Integer.
It describes any Integer as `<DEF>DOB{|Key: ~<@Name|Label@>~|}^DE{|Ops: ~&[Op{Type: Calculate, Operator: as, Operands: &[<@Value@>, u64]&}Op]&~|}<DEF>`.
Define "Direction" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~is_positive~, Receiver: ~<@Value@>~|}<DEF>`.
Use Storage.Present as "Present".
Use Storage.Absent as "Absent".
Define "Multiple" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~checked_mul~, Receiver: ~<@Value@>~, Inputs: ~&[<@Factor@>]&~|}<DEF>`.
Define "Contents" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~expect~, Receiver: ~<@Value@>~, Inputs: ~&[<@Message|Text@>]&~|}<DEF>`.
Define "Pair" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~(<@First@>, <@Second@>)~|}<DEF>`.

Define "Choice" as `<DEF>GOB{|Key: ~<@Name|Label@>~|}^GI^(E{|Statement: ~<@Condition@>~|}*W^BOD^E{|Statement: ~<@Then@>~|}*E{|Statement: ~Void~|}*W^BOD^E{|Statement: ~<@Else@>~|})<DEF>`.
Define "Agreement" as `<DEF>DOB{|Key: ~<@Name|Label@>~|}^DE{|Ops: ~&[Op{Type: Calculate, Operator: &&, Operands: &[<@Left@>, <@Right@>]&}Op]&~|}<DEF>`.
Define "Alternative" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~unwrap_or~, Receiver: ~<@Value@>~, Inputs: ~&[<@Fallback@>]&~|}<DEF>`.

Define "Shift|-s|-ed|-ing|-er" to `<DEF>DOB{|Key: ~<@Name|Label@>~|}^DE{|Ops: ~&[Op{Type: Calculate, Operator: >>, Operands: &[<@Value@>, <@Count@>]&}Op]&~|}<DEF>`.
Define "Leftward" as a New Adjective describing any Shifter.
It describes any Shifter as `<DEF>DOB{|Key: ~<@Name|Label@>~|}^DE{|Ops: ~&[Op{Type: Calculate, Operator: <<, Operands: &[<@Value@>, <@Count@>]&}Op]&~|}<DEF>`.
Define "Intersection" as `<DEF>DOB{|Key: ~<@Name|Label@>~|}^DE{|Ops: ~&[Op{Type: Calculate, Operator: &, Operands: &[<@Left@>, <@Right@>]&}Op]&~|}<DEF>`.
Define "Union" as `<DEF>DOB{|Key: ~<@Name|Label@>~|}^DE{|Ops: ~&[Op{Type: Calculate, Operator: |, Operands: &[<@Left@>, <@Right@>]&}Op]&~|}<DEF>`.

Define "Notation" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~format!~, Inputs: ~&[\"{:e}\", <@Value@>]&~|}<DEF>`.
Define "Split|-s|--|-ting|-ter" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~split_once~, Receiver: ~<@Value@>~, Inputs: ~&[<@Separator|Text@>]&~|}<DEF>`.
Define "Trim|-s|-med|-ming|-mer" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~trim_start_matches~, Receiver: ~<@Value@>~, Inputs: ~&[<@Prefix|Text@>]&~|}<DEF>`.
Define "Replacement" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~replace~, Receiver: ~<@Value@>~, Inputs: ~&[<@From|Text@>, <@To|Text@>]&~|}<DEF>`.
Define "First" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~<@Value@>.0~|}<DEF>`.
Define "Second" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~<@Value@>.1~|}<DEF>`.
Define "Interval" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^Y{|Key: ~std::ops::Range~, Props: ~&[Prop{Key: start, Value: <@Minimum@>}Prop, Prop{Key: end, Value: <@Maximum@>}Prop]&~|}<DEF>`.
Define "Membership" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~contains~, Receiver: ~<@Range@>~, Inputs: ~&[&<@Value@>]&~|}<DEF>`.
Define "Insert|-s|-ed|-ing|-er" to `<DEF>MOB{|Key: ~_~|}^KY{|Key: ~insert_str~, Receiver: ~<@Value@>~, Inputs: ~&[<@Index@>, <@Text|Text@>]&~|}<DEF>`.
Define "Sign" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~is_sign_negative~, Receiver: ~<@Value@>~|}<DEF>`.
Define "Zeros" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~repeat~, Receiver: ~\"0\"~, Inputs: ~&[<@Count@>]&~|}<DEF>`.
Define "Prefix" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~format!~, Inputs: ~&[\"0.{}{}\", <@Zeros@>, <@Digits@>]&~|}<DEF>`.
Define "Suffix" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~format!~, Inputs: ~&[\"{}{}\", <@Digits@>, <@Zeros@>]&~|}<DEF>`.
Define "Scientific" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~format!~, Inputs: ~&[\"{}e{:+03}\", <@Digits@>, <@Exponent@>]&~|}<DEF>`.
