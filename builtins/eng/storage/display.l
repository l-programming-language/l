Define Storage's "Annotation" as `<DEF>OB{|Key: ~<@Name|Label@>: <@Type|Type@>~, Value: ~<@Value@>~|}<DEF>`.
Define Storage's "Description" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~to_string~, Receiver: ~<@Value@>~|}<DEF>`.
Define Storage's "Encoding" as `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~into_bytes~, Receiver: ~<@Value@>~|}<DEF>`.
Create Storage's "Bytes" as a Dictionary.
Define Storage's Bytes's "Extend|-s|-ed|-ing|-er" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~extend_from_slice~, Receiver: ~<@Buffer@>~, Inputs: ~&[<@Bytes@>]&~|}<DEF>`.
Create Storage's "Number" as a Dictionary.
Create Storage's "Truth" as a Dictionary.
Create Storage's "Error" as a Dictionary.
Create Storage's "Null" as a Dictionary.
Create Storage's "App" as a Dictionary.
Create Storage's "Function" as a Dictionary.
Define Storage's Number's "Format|-s|-ted|-ting|-ter" to `<DEF>MOB{|Key: ~<@Name|Label@>~|}^KY{|Key: ~runtime::format_number~, Imports: ~&[runtime::format_number as _]&~, Inputs: ~&[*<@Value@>]&~|}<DEF>`.

    Publicly create Storage's Text's "Restore|-s|-d|--ing|--er" as a Generic Function with Inputs of "data" as `&dyn std::any::Any` and "message" as `&str` and Outputs of `runtime::Result<Vec<u8>>`.
Create "Value" as a New Storage.Checked Binding of a New Storage's Annotation of `runtime::Result<&Vec<u8>>` and a New Storage's Projector of Data and Message.
Create "Bytes" as a New Storage's Copier of Value.
Return a New Storage's Success of Bytes.

    Publicly create Storage's Text's "Render|-s|-ed|-ing|-er" as a Generic `l_show_text` Function with Inputs of "data" as `&dyn std::any::Any` and Outputs of `runtime::Result<Vec<u8>>`.
Return a New Storage's Text's Restorer of Data and `expected Text`.

    Publicly create Storage's Bytes's "Render|-s|-ed|-ing|-er" as a Generic `l_show_bytes` Function with Inputs of "data" as `&dyn std::any::Any` and Outputs of `runtime::Result<Vec<u8>>`.
Return a New Storage's Text's Restorer of Data and `expected Bytes`.

    Publicly create Storage's Number's "Render|-s|-ed|-ing|-er" as a Generic `l_show_number` Function with Inputs of "data" as `&dyn std::any::Any` and Outputs of `runtime::Result<Vec<u8>>`.
Create "Value" as a New Storage.Checked Binding of a New Storage's Annotation of `runtime::Result<&f64>` and a New Storage's Projector of Data and `expected Number`.
Create "Bytes" as a New Storage's Number's Formatter of Value.
Return a New Storage's Success of Bytes.

    Publicly create Storage's Truth's "Render|-s|-ed|-ing|-er" as a Generic `l_show_truth` Function with Inputs of "data" as `&dyn std::any::Any` and Outputs of `runtime::Result<Vec<u8>>`.
Create "Value" as a New Storage.Checked Binding of a New Storage's Annotation of `runtime::Result<&bool>` and a New Storage's Projector of Data and `expected Truth`.
Create "Text" as a New Storage's Description of Value.
Create "Bytes" as a New Storage's Encoding of Text.
Return a New Storage's Success of Bytes.

    Publicly create Storage's Error's "Render|-s|-ed|-ing|-er" as a Generic `l_show_error` Function with Inputs of "data" as `&dyn std::any::Any` and Outputs of `runtime::Result<Vec<u8>>`.
Create "Value" as a New Storage.Checked Binding of a New Storage's Annotation of `runtime::Result<&runtime::Error>` and a New Storage's Projector of Data and `expected Error`.
Create "Text" as a New Storage's Description of Value.
Create "Bytes" as a New Storage's Encoding of Text.
Return a New Storage's Success of Bytes.

    Publicly create Storage's Null's "Render|-s|-ed|-ing|-er" as a Generic `l_show_null` Function with Inputs of "data" as `&dyn std::any::Any` and Outputs of `runtime::Result<Vec<u8>>`.
Create "Bytes" as a New Storage's Text's Owner of `null`.
Return a New Storage's Success of Bytes.

    Publicly create Storage's App's "Render|-s|-ed|-ing|-er" as a Generic `l_show_app` Function with Inputs of "data" as `&dyn std::any::Any` and Outputs of `runtime::Result<Vec<u8>>`.
Create "Bytes" as a New Storage's Text's Owner of `App`.
Return a New Storage's Success of Bytes.

    Publicly create Storage's Function's "Render|-s|-ed|-ing|-er" as a Generic `l_show_function` Function with Inputs of "data" as `&dyn std::any::Any` and Outputs of `runtime::Result<Vec<u8>>`.
Create "Bytes" as a New Storage's Text's Owner of `Function`.
Return a New Storage's Success of Bytes.

    Publicly create Storage's Environment's "Render|-s|-ed|-ing|-er" as a Generic `l_show_environment` Function with Inputs of "data" as `&dyn std::any::Any` and Outputs of `runtime::Result<Vec<u8>>`.
Create "Failure" as a New Storage's Failure of `capture environment is not an L value`.
Return a New Storage's Rejection of Failure.

Define Storage's Function's "Environment" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~&<@Value@>.0~|}<DEF>`.
Use Storage.Tuple.Second as Storage.Function's "Target".
Define Storage's "Copyable" as a New Adjective describing any Storage's Routine.
It describes a Generic Storage's Routine of `T: Copy`.

    Publicly create Storage's Function's "Copy|--ies|--ied|-ing|--ier" as a Generic Storage.Copyable `l_function_copy` Storage.Routine with Inputs of "value" as `&(LEnvironment, T)` and Outputs of `(LEnvironment, T)`.
Create "Environment" as a New Storage's Function's Environment of Value.
Create "Copied" as a New Storage's Environment's Copier of Environment.
Create "Target" as a New Storage's Function's Target of Value.
Return a New Storage's Tuple of Copied and Target.

Define Storage's "Nothing" as `<DEF>OB{|Key: ~<@Name|Label@>~, Value: ~()~|}<DEF>`.
