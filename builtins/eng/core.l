Create "Start|-s|-ed|-ing|-er" for Building any App and then Running it.
Create "Execute|-s|-d|--ing|--or" for Starting any App and then Awaiting it.
Define "Process|-es|-ive|-ing" as `<DEF>W^<@Definitions|Syllables@><DEF>`.
Define "Function|-s|-al|-ing" as `<DEF>FW{|Func{Key: <@Name|Text@>, Inputs: <@Inputs|Props@>, Outputs: <@Outputs|Texts@>}Func|}^(<@Process@>)<DEF>`.
Define any Function's "Return|-s|-ed|-ing|-er" to `<DEF>BOD^E{|Statement: ~return <@Statement@>;~|}<DEF>`.
Define any Function's "Suspend" as `<DEF>W^(COB{|Key: ~lPending~|}^(CV{|Capture: ~Move~|}^(<@Body@>))*MOB{|Key: ~lPinned~|}^(KY{|Key: ~Box::pin~, Inputs: ~&[lPending]&~|})*BOD^E{|Statement: ~lPinned~|})<DEF>`.
Define any Function's "Invoke" as `<DEF>[["OB{|Key: ~<@Name@>~, Value: ~<@Target@>~|}", "COB{|Key: ~<@Name@>Outcome~|}^(DV^(KY{|Key: ~(<@Name@>.1)~, Inputs: ~&[<@Arguments@>]&~|}))", "GOB{|Key: ~<@Name@>Result~|}^GI{|Key: ~<@Name@>Outcome?~|}^(E{|Statement: ~std::ops::ControlFlow::Continue(lValue)~|}*W^(BOD^E{|Statement: ~lValue~|})*E{|Statement: ~std::ops::ControlFlow::Break(())~|}*<@Exit@>)"]]<DEF>`.
Define any Function's "Yield" as `<DEF>[["MOB{|Key: ~lReturned~|}^KY{|Key: ~std::ops::ControlFlow::Continue~, Inputs: ~&[<@Values@>]&~|}", "MOB{|Key: ~lSuccess~|}^KY{|Key: ~Ok~, Inputs: ~&[lReturned]&~|}", "<@Return@>"]]<DEF>`.
Define any Function's "Bind" as `<DEF>[["MOB{|Key: ~<@Name@>Environment~|}^KY{|Key: ~LEnvironment::typed~, Inputs: ~&[<@Types@>, <@Captures@>]&~|}", "OB{|Key: ~<@Name@>~, Value: ~(<@Name@>Environment?, <@Function@> as <@Type@>)~|}"]]<DEF>`.
Define any Function's "Exit" as `<DEF>[["MOB{|Key: ~lExit~|}^KY{|Key: ~Ok~, Inputs: ~&[<@Value@>]&~|}", "<@Return@>"]]<DEF>`.
"Objective" describes any Function as `<DEF>[["A", "<@Function@>"]]<DEF>`.
"Main" is a New Adjective describing any Service, Object, or Function.
It describes a Generic `main` Function of any Process with No Inputs and No Outputs.
It describes a Generic `main` Object with a New Objective Main Function using any Process.
It describes a Generic `main` Service with a New Main Object using any Process.
"Basic" describes a Generic App with a New Main Service using any Process.
