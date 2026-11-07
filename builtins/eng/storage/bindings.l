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

