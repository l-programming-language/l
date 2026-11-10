Create "IO" as a Dictionary.
Define any Thing's "Render|-s|-ed|-ing|-er" as a New Verb.
Use Formatting's Format from `format` as "Format".
Use Storage from `storage`.

Define "Raw" as a New Adjective describing any Storage's Routine.
It describes a Generic Storage's Routine of `T: AsRef<[u8]>`.
Define "Bytes" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~std::convert::AsRef::<[u8]>::as_ref~, Inputs: ~&[&<@Value@>]&~|}<DEF>`.
Define "Absence" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~"<nil>"~|}<DEF>`.
Publicly define IO's "Output" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~std::io::stdout~|}<DEF>`.
Publicly define IO's "Diagnostic" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~std::io::stderr~|}<DEF>`.
Define "Lock|-s|-ed|-ing|-er" to `<DEF>MOB{|Key: ~mut <@Name|Label@>~|}^KY{|Key: ~lock~, Receiver: ~<@Value@>~|}<DEF>`.
Define "Deliver|-s|-ed|-ing|-er" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~std::io::Write::write_all~, Inputs: ~&[<@Writer@>, <@Value@>]&~|}<DEF>`.
Publicly define IO's "Flush|-es|-ed|-ing|-er" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~std::io::Write::flush~, Inputs: ~&[<@Writer@>]&~|}<DEF>`.
Define "Line" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~b"\n"~|}<DEF>`.

    Publicly create IO's "Send|-s|--t|-ing|-er" as a Generic Storage.Routine of `W: std::io::Write + ?Sized, P: AsRef<[u8]> + ?Sized` with Inputs of "writer" as `&mut W` and "payload" as `&P` and Outputs of `std::io::Result<()>`.
Create "Borrowed" as a New Bytes of Payload.
Return a New Deliverer of Writer and Borrowed.

    Create "Terminate|-s|-d|--ing|--or" as a Generic Storage.Routine of `W: std::io::Write + ?Sized` with Inputs of "writer" as `&mut W` and "receipt" as `std::io::Result<()>` and Outputs of `std::io::Result<()>`.
Storage.Result.Select Receipt, a New Process with Inputs of "Payload" as `()` for Creating "Ending" as a New Line and then Returning a New IO's Sender of Writer and Ending, and a New Process with Inputs of "Failure" as `std::io::Error` for Returning a New Storage's Rejection of Failure.
Publicly define IO's "Lined" as a New Adjective describing any IO's Sender.
It describes any IO's Sender as a New Terminator of any Thing and any IO's Sender.

    Publicly create "Write|-s|----ote|--ing|-r" as a Generic Function with Inputs of "value" as `Vec<u8>` and Outputs of `()`.
Create "Device" as a New IO's Output.
Create "Held" as a New Locker of Device.
Create "Destination" as a New Storage.Editable Storage's Borrower of Held.
Create "Payload" as a New Bytes of Value.
Create "Receipt" as a New IO.Lined IO's Sender of Destination and Payload.
Create "Accepted" as a New Storage's Expecter of Receipt and `failed printing to stdout`.
Return a New Storage's Nothing.

    Publicly create any Number's "Render|-s|-ed|-ing|-er" as a Generic Function with Inputs of "value" as `f64` and Outputs of `Vec<u8>`.
Return a New Formatter of Value.

    Publicly create any Text's "Render|-s|-ed|-ing|-er" as a Generic Raw Storage.Routine with Inputs of "value" as `T` and Outputs of `Vec<u8>`.
Create "Borrowed" as a New Bytes of Value.
Return a New Storage's Owner of Borrowed.

    Publicly create any Truth's "Render|-s|-ed|-ing|-er" as a Generic Function with Inputs of "value" as `bool` and Outputs of `Vec<u8>`.
Create "String" as a New Storage's Description of Value.
Return a New Storage's Encoding of String.

    Publicly create any Null's "Render|-s|-ed|-ing|-er" as a Generic Function with Inputs of "value" as `()` and Outputs of `Vec<u8>`.
Create "Message" as a New Absence.
Return a New Text's Renderer of Message.

Create IO's "Print|-s|-ed|-ing|-er" for Processing a New Storage's Discarder of a New Writer of any Thing's Renderer.
