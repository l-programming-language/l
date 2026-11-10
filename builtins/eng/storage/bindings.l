Define Storage's "Routine" as `<DEF>FW{|Func{Key: <@Name|Text@>, Inputs: <@Inputs|Props@>, Outputs: <@Outputs|Texts@>}Func, GenericTypes: ~<@Types|Type@>~|}^(<@Process@>)<DEF>`.
Use Function's Return as Storage.Routine's "Return".
Use Function's Invoke as Storage.Routine's "Invoke".
Define Storage's "Lending" as a New Adjective describing any Storage's Routine.
It describes any Storage's Routine as `<DEF>FW{|Func{Key: <@Name|Text@>, Inputs: <@Inputs|Props@>, Outputs: <@Outputs|Texts@>}Func, GenericTypes: ~<@Types|Type@>~, Lifetimes: ~<@Lifetimes|Type@>~|}^(<@Process@>)<DEF>`.
Define Storage's "Iteration" as `<DEF>BOP^X{|Iterable: ~<@Values@>~, Key: ~<@Item|Label@>~, RangeArity: ~1~|}^(<@Body@>)<DEF>`.
Define Storage's "Apply|--ies|--ied|-ing|--ier" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~<@Action@>~, Inputs: ~&[<@Item@>, <@Context@>]&~|}<DEF>`.
Define Storage's "Paired" as a New Adjective describing any Storage's Iteration or Storage's Applier.
It describes any Storage's Iteration as `<DEF>BOP^X{|Iterable: ~<@Values@>~, Index: ~<@First|Label@>~, Key: ~<@Second|Label@>~, RangeArity: ~2~|}^(<@Body@>)<DEF>`.
It describes any Storage's Applier as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~<@Action@>~, Inputs: ~&[<@First@>, <@Second@>, <@Context@>]&~|}<DEF>`.
Define Storage's "Discard|-s|-ed|-ing|-er" to `<DEF>OB{|Key: ~_~, Value: ~<@Value@>~|}<DEF>`.
Create Storage's "Traverse|-s|-d|--ing|--er" for Processing a New Storage's Iteration of any Thing and a New Process with Inputs of "Item" as `_` for Creating "Outcome" as a New Storage's Applier of any Thing, Item, and any Thing.
Define Storage's "Binding|-s" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~<@Value@>~|}<DEF>`.
Define "Enclosure" as `<DEF>MOB{|Key: ~<@Captured|Label@>~|}^KY{|Key: ~<@Copy@>~, Inputs: ~&[&<@Name@>.0]&~|}<DEF>`.
Define "Reference" as `<DEF>OB{|Key: ~<@Context|Label@>~, Value: ~&<@SourceContext@>~|}<DEF>`.
Define "Invocation" as `<DEF>MOB{|Key: ~<@Future|Label@>~|}^(<@Invocation@>)<DEF>`.
Define Storage's "Await|-s|-ed|-ing|-er" to `<DEF>COB{|Key: ~<@Outcome|Label@>~|}^DV{|Key: ~<@Future@>~|}<DEF>`.
Define "Choice" as `<DEF>GOB{|Key: ~<@Selection|Label@>~|}^GI{|Key: ~<@Subject@>~|}^(<@Continued@>*<@Stopped@>)<DEF>`.
Define "Continuation" as `<DEF>E{|Statement: ~std::ops::ControlFlow::Continue(<@Payload|Label@>)~|}*<@Body@><DEF>`.
Define "Termination" as `<DEF>E{|Statement: ~std::ops::ControlFlow::Break(())~|}*<@Exit@><DEF>`.
Define "Payload" as `<DEF>W^(BOD^E{|Statement: ~<@Payload@>~|})<DEF>`.
Create Storage's "Invoke|-s|-d|--ing|--er" for Processing a New Storage's Binding, a New Enclosure, a New Reference, a New Invocation, a New Storage's Awaiter, a New Storage.Checked Storage's Binding, and a New Choice of a New Continuation of a New Payload and a New Termination.
Define Storage's "Mutable" as a New Adjective describing any Storage's Binding.
It describes any Storage's Binding as `<DEF>OB{|Key: ~mut <@Name|Label@>~, Value: ~<@Value@>~|}<DEF>`.
Define Storage's "Comparison|-s" as `<DEF>DOB{|Key: ~<@Name|Label@>~|}^DE{|Ops: ~&[Op{Type: Calculate, Operator: <@Operator@>, Operands: &[<@Left@>, <@Right@>]&}Op]&~|}<DEF>`.
Define Storage's "Identical" as a New Adjective describing any Storage's Comparison.
It describes any Storage's Comparison as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~std::sync::Arc::ptr_eq~, Inputs: ~&[&<@Left@>, &<@Right@>]&~|}<DEF>`.
Create Storage's "Compare|-s|-d|--ing|--er" for Processing a New Storage's Comparison.
Define Storage's "Produced" as a New Adjective describing any Storage's Binding.
It describes any Storage's Binding as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^(<@Value@>)<DEF>`.

