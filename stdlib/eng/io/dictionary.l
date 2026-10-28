Create "IO" as a Dictionary.
Define any Thing's "Render|-s|-ed|-ing|-er" as a New Verb.
Use Formatting's Format from `format` as "Format".

Define "Raw" as a New Adjective describing any Function.
It describes any Function as `<DEF>FW{|Func{Key: <@Name|Text@>, Inputs: <@Inputs|Props@>, Outputs: <@Outputs|Texts@>}Func, GenericTypes: ~T: AsRef<[u8]>~|}^(<@Process@>)<DEF>`.
Define "Bytes" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~std::convert::AsRef::<[u8]>::as_ref~, Inputs: ~&[&<@Value@>]&~|}<DEF>`.
Define "Copy|--ies|--ied|-ing|--ier" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~to_vec~, Receiver: ~<@Value@>~|}<DEF>`.
Define "Absence" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~"<nil>"~|}<DEF>`.
Define "Description" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~to_string~, Receiver: ~<@Value@>~|}<DEF>`.
Define "Encoding" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~into_bytes~, Receiver: ~<@Value@>~|}<DEF>`.
Define "Stream" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~std::io::stdout~|}<DEF>`.
Define "Lock|-s|-ed|-ing|-er" to `<DEF>MOB{|Key: ~mut <@Name|Label@>~|}^KY{|Key: ~lock~, Receiver: ~<@Value@>~|}<DEF>`.
Define "Slice" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~as_slice~, Receiver: ~<@Value@>~|}<DEF>`.
Define "Deliver|-s|-ed|-ing|-er" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~std::io::Write::write_all~, Inputs: ~&[&mut <@Stream@>, <@Value@>]&~|}<DEF>`.
Define "Check|-s|-ed|-ing|-er" to `<DEF>MOB{|Key: ~_~|}^KY{|Key: ~expect~, Receiver: ~<@Value@>~, Inputs: ~&[\"failed printing to stdout\"]&~|}<DEF>`.
Define "Line" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~b"\n"~|}<DEF>`.
Define "Nothing" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~()~|}<DEF>`.
Define "Discard|-s|-ed|-ing|-er" to `<DEF>MOB{|Key: ~_~|}^(<@Value@>)<DEF>`.

    Publicly create "Write|-s|----ote|--ing|-r" as a Generic Function with Inputs of "value" as `Vec<u8>` and Outputs of `()`.
Create "Device" as a New Stream.
Create "Held" as a New Locker of Device.
Create "Payload" as a New Slice of Value.
Create "Receipt" as a New Deliverer of Held and Payload.
Check Receipt.
Create "Ending" as a New Line.
Create "Terminated" as a New Deliverer of Held and Ending.
Check Terminated.
Return a New Nothing.

    Publicly create any Number's "Render|-s|-ed|-ing|-er" as a Generic Function with Inputs of "value" as `f64` and Outputs of `Vec<u8>`.
Return a New Formatter of Value.

    Publicly create any Text's "Render|-s|-ed|-ing|-er" as a Generic Raw Function with Inputs of "value" as `T` and Outputs of `Vec<u8>`.
Create "Borrowed" as a New Bytes of Value.
Return a New Copier of Borrowed.

    Publicly create any Truth's "Render|-s|-ed|-ing|-er" as a Generic Function with Inputs of "value" as `bool` and Outputs of `Vec<u8>`.
Create "String" as a New Description of Value.
Return a New Encoding of String.

    Publicly create any Null's "Render|-s|-ed|-ing|-er" as a Generic Function with Inputs of "value" as `()` and Outputs of `Vec<u8>`.
Create "Message" as a New Absence.
Return a New Text's Renderer of Message.

Create IO's "Print|-s|-ed|-ing|-er" for Processing a New Discarder of a New Writer of any Thing's Renderer.
