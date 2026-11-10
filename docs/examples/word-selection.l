Use Storage from `storage`.
Create "Selection" as a Dictionary.

    Publicly create Selection's "Choose" as a Generic `choose` Function with Inputs of "value" as `Option<u64>` and Outputs of `u64`.
Storage.Option.Select Value, a New Process with Inputs of "Item" as `u64` for Returning Item, and a New Process for Returning a New Storage's Caster of 0 and `u64`.

    Publicly create Selection's "Preserve" as a Generic `preserve` Function with Inputs of "value" as `Result<u64, u8>` and Outputs of `Result<u64, u8>`.
Storage.Result.Select Value, a New Process with Inputs of "Item" as `u64` for Returning a New Storage's Success of Item, and a New Process with Inputs of "Failure" as `u8` for Returning a New Storage's Rejection of Failure.

    Publicly create Selection's "Require" as a Generic `require` Function with Inputs of "value" as `Option<Box<u8>>` and "message" as `&str` and Outputs of `runtime::Result<Box<u8>>`.
Return a New Storage's Requirer of Value and Message.

    Publicly create Selection's "Expect" as a Generic `expect` Function with Inputs of "value" as `Result<Box<u8>, u8>` and "message" as `&str` and Outputs of `Box<u8>`.
Return a New Storage's Expecter of Value and Message.
