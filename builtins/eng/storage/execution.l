Define Storage's "Cloneable" as a New Adjective describing any Storage's Routine.
It describes a Generic Storage's Routine of `T: Clone`.
Define Storage's "Specialize|-s|-d|--ing|--er" to `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~<@Function@>::<<@Type|Type@>>~|}<DEF>`.

    Publicly create Storage's "Clone|-s|-d|--ing|--er" as a Generic Storage.Cloneable Storage.Routine with Inputs of "value" as `&T` and Outputs of `T`.
Return a New Storage's Copier of Value.

Define Storage's "Future" as `<DEF>COB{|Key: ~<@Future|Label@>~|}^CV{|Capture: ~Move~|}^(<@Body@>)<DEF>`.
Define Storage's "Pin|-s|-ned|-ning|-ner" to `<DEF>MOB{|Key: ~<@Pinned|Label@>~|}^KY{|Key: ~Box::pin~, Inputs: ~&[<@Value@>]&~|}<DEF>`.
Define "Tail" as `<DEF>BOD^E{|Statement: ~<@Pinned@>~|}<DEF>`.
Create Storage's Function's "Suspend|-s|-ed|-ing|-er" for Processing a New Storage's Future, a New Storage's Pinner, and a New Tail.
Define Storage's Function's "Continue|-s|-d|--ing|--er" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~std::ops::ControlFlow::Continue~, Inputs: ~&[<@Values@>]&~|}<DEF>`.
Define Storage's Function's "Return" as `<DEF><@Return@><DEF>`.
Create Storage's Function's "Yield|-s|-ed|-ing|-er" for Processing a New Storage's Function's Continuer, a New Storage's Success, and a New Storage's Function's Return.
Create Storage's Function's "Exit|-s|-ed|-ing|-er" for Processing a New Storage's Success and a New Storage's Function's Return.
Define "Capture" as `<DEF>MOB{|Key: ~<@Capture|Label@>~|}^(<@Environment@>)<DEF>`.
Create Storage's Function's "Bind|-s|----ound|-ing|-er" for Processing a New Capture, a New Storage.Checked Storage's Binding, a New Storage's Caster, and a New Storage's Tuple.

Create Storage's "Reject|-s|-ed|-ing|-or" for Function.Returning a New Storage's Rejection of a New Storage's Failure of any Text.

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
Define "Runtime" as `<DEF>MOB{|Key: ~<@Cloned|Label@>~|}^KY{|Key: ~clone~, Receiver: ~<@Runtime@>~|}<DEF>`.
Define "Program" as `<DEF>OB{|Key: ~<@Binding|Label@>~, Value: ~&<@Name@>.1~|}<DEF>`.
Create Storage's Context's "Create|-s|-d|--ing|--or" for Processing a New Runtime, a New Storage's Tuple, and a New Program.
Define "Host" as `<DEF>OB{|Key: ~<@RuntimeBinding|Label@>~, Value: ~&<@Context@>.0~|}<DEF>`.
Define "Data" as `<DEF>OB{|Key: ~<@ProgramBinding|Label@>~, Value: ~&<@Context@>.1~|}<DEF>`.
Create Storage's Context's "Read|-s||-ing|-er" for Processing a New Host and a New Data.
Define "Snapshot" as `<DEF>MOB{|Key: ~<@Captured|Label@>~|}^KY{|Key: ~<@Copy@>~, Inputs: ~&[&<@Environment@>]&~|}<DEF>`.
Create Storage's Function's "Retain|-s|-ed|-ing|-er" for Processing a New Snapshot, a New Storage's Caster, and a New Storage's Tuple.
Define Storage's Program's "Apps" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~&<@Program@>.apps~|}<DEF>`.

    Publicly create Storage's App's "Retain|-s|-ed|-ing|-er" as a Generic Function with Inputs of "value" as `&std::sync::Arc<runtime::App>` and Outputs of `std::sync::Arc<runtime::App>`.
Return a New Retainer of Value.

    Publicly create Storage's Program's "Select|-s|-ed|-ing|-or" as a Generic Function with Inputs of "program" as `&LProgram` and "index" as `usize` and Outputs of `runtime::Result<std::sync::Arc<runtime::App>>`.
Create "Apps" as a New Storage's Program's Apps of Program.
Create "Slot" as a New Storage's Getter of Apps and Index.
Create "Borrowed" as a New Storage.Checked Binding of a New Storage's Requirer of Slot and `App handle is not owned by this program`.
Create "Retained" as a New Storage's App's Retainer of Borrowed.
Return a New Storage's Success of Retained.

Define "Path" as `<DEF>MOB{|Key: ~<@PathBinding|Label@>~|}^KY{|Key: ~std::path::PathBuf::from~, Inputs: ~&[<@Path|Text@>]&~|}<DEF>`.
Define "Application" as `<DEF>MOB{|Key: ~<@AppResult|Label@>~|}^KY{|Key: ~runtime::runtime_app~, Imports: ~&[runtime::runtime_app as _]&~, Inputs: ~&[<@Runtime@>, <@PathBinding@>]&~|}<DEF>`.
Create Storage's App's "Create|-s|-d|--ing|--or" for Processing a New Path and a New Application.
Define "Source" as `<DEF>MOB{|Key: ~<@SourceBinding|Label@>~|}^KY{|Key: ~String::from~, Inputs: ~&[<@Source|Text@>]&~|}<DEF>`.
Define "Attribute|-s|-d|--ing|--or" to `<DEF>MOB{|Key: ~_~|}^KY{|Key: ~runtime::app_set_source~, Imports: ~&[runtime::app_set_source as _]&~, Inputs: ~&[&<@App@>, <@SourceBinding@>]&~|}<DEF>`.
Create Storage's App's "Attribute|-s|-d|--ing|--or" for Processing a New Source and a New Attributor.
Define Storage's App's "Output" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~runtime::app_output~, Imports: ~&[runtime::app_output as _]&~, Inputs: ~&[&<@App@>]&~|}<DEF>`.
Define Storage's App's "Error" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~runtime::app_error~, Imports: ~&[runtime::app_error as _]&~, Inputs: ~&[&<@App@>]&~|}<DEF>`.
Define Storage's Error's "Create|-s|-d|--ing|--or" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~runtime::error_new~, Imports: ~&[runtime::error_new as _]&~, Inputs: ~&[<@Value@>]&~|}<DEF>`.
Define Storage's "Show|-s|-ed|-ing|-er" to `<DEF>COB{|Key: ~<@Name|Label@>~|}^DV^KY{|Key: ~runtime::runtime_show~, Imports: ~&[runtime::runtime_show as _]&~, Inputs: ~&[runtime, &<@Value@>]&~|}<DEF>`.
Define "Fail|-s|-ed|-ing|-er" to `<DEF>MOB{|Key: ~<@Error|Label@>~|}^KY{|Key: ~runtime::runtime_fail~, Imports: ~&[runtime::runtime_fail as _]&~, Inputs: ~&[<@Runtime@>, <@FailureInput@>]&~|}<DEF>`.
Create Storage's Program's "Fail|-s|-ed|-ing|-er" for Processing a New Failer and a New Storage's Rejection.

Create Storage's "Entry" as a Dictionary.
Define "Entry" as `<DEF>MOB{|Key: ~<@MainFuture|Label@>~|}^KY{|Key: ~<@Main@>~, Inputs: ~&[<@Runtime@>]&~|}<DEF>`.
Define "Observation" as `<DEF>MOB{|Key: ~<@Future|Label@>~|}^KY{|Key: ~runtime::runtime_run~, Imports: ~&[runtime::runtime_run as _]&~, Inputs: ~&[<@Runtime@>, <@MainFuture@>]&~|}<DEF>`.
Create Storage's Entry's "Observe|-s|-d|--ing|--er" for Processing a New Entry, a New Observation, a New Storage's Awaiter, and a New Storage's Function's Return.
