Use IO from `io`.
Create "Writing" as a Dictionary.

    Publicly create Writing's "Output" as a Generic `output` Function with No Inputs and Outputs of `std::io::Stdout`.
Return a New IO's Output.

    Publicly create Writing's "Diagnostic" as a Generic `diagnostic` Function with No Inputs and Outputs of `std::io::Stderr`.
Return a New IO's Diagnostic.

    Publicly create Writing's "Send" as a Generic `send` Function with Inputs of "writer" as `&mut Vec<u8>` and "payload" as `&[u8]` and Outputs of `std::io::Result<()>`.
Return a New IO's Sender of Writer and Payload.

    Publicly create Writing's "Line" as a Generic `line` Function with Inputs of "writer" as `&mut Vec<u8>` and "payload" as `&[u8]` and Outputs of `std::io::Result<()>`.
Return a New IO.Lined IO's Sender of Writer and Payload.

    Publicly create Writing's "Flush" as a Generic `flush` Function with Inputs of "writer" as `&mut Vec<u8>` and Outputs of `std::io::Result<()>`.
Return a New IO's Flusher of Writer.
