Create "Formatting" as a Dictionary.
Use `foundations.l`.

    Create Formatting's "Scale|-s|-d|--ing|--er" as a Generic Function with Inputs of "value" as `u128` and "exponent" as `i32` and Outputs of `Option<u128>`.
Create "Zero" as a New Integer of 0.
Create "Original" as a New Present of Value.
Create "Failure" as a New Absent.
If Exponent is Zero, Return Original.
Create "Increasing" as a New Direction of Exponent.
Create "Five" as a New Wide Integer of 5.
Create "Optional" as a New Multiple of Value and Five.
Create "Available" as a New Presence of Optional.
Create "Missing" as a New Equality of Available and false.
Create "Overflowed" as a New Agreement of Increasing and Missing.
If Overflowed, Return Failure.
Create "Rest" as a New Remainder of Value and Five.
Create "Empty" as a New Wide Integer of 0.
Create "Divisible" as a New Equality of Rest and Empty.
Create "Uneven" as a New Equality of Divisible and false.
Create "Decreasing" as a New Equality of Increasing and false.
Create "Invalid" as a New Agreement of Decreasing and Uneven.
If Invalid, Return Failure.
Create "Expanded" as a New Alternative of Optional and Value.
Create "Shortened" as a New Quotient of Value and Five.
Create "Selected" as a New Choice of Increasing and Expanded and Shortened.
Create "One" as a New Integer of 1.
Create "Lower" as a New Difference of Exponent and One.
Create "Higher" as a New Sum of Exponent and One.
Create "Next" as a New Choice of Increasing and Lower and Higher.
Return a New Formatting's Scaler of Selected and Next.

    Create Formatting's "Balance|-s|-d|--ing|--er" as a Generic Function with Inputs of "value" as `f64` and "coefficient" as `u128` and "exponent" as `i32` and Outputs of `bool`.
Create "Bits" as a New Representation of a New Absolute of Value.
Create "Precision" as a New Unsigned Integer of 52.
Create "Shifted" as a New Shifter of Bits and Precision.
Create "Mask" as a New Unsigned Integer of 2047.
Create "Binary" as a New Integer of a New Intersection of Shifted and Mask.
Create "One" as a New Unsigned Integer of 1.
Create "High" as a New Leftward Shifter of One and Precision.
Create "Fraction" as a New Difference of High and One.
Create "Significand" as a New Intersection of Bits and Fraction.
Create "Expanded" as a New Union of Significand and High.
Create "Zero" as a New Integer of 0.
Create "Denormal" as a New Equality of Binary and Zero.
Create "Selected" as a New Choice of Denormal and Significand and Expanded.
Create "Bias" as a New Integer of 1023.
Create "Unbiased" as a New Difference of Binary and Bias.
Create "Width" as a New Integer of Precision.
Create "Ordinary" as a New Difference of Unbiased and Width.
Create "Minimum" as a New Integer of -1074.
Create "Initial" as a New Choice of Denormal and Minimum and Ordinary.
Create "Zeros" as a New Trailing of Selected.
Create "Significant" as a New Shifter of Selected and Zeros.
Create "Adjustment" as a New Integer of Zeros.
Create "Power" as a New Sum of Initial and Adjustment.
Create "Two" as a New Wide Integer of 2.
Create "Doubled" as a New Product of Two and Coefficient.
Create "Unit" as a New Wide Integer of 1.
Create "Numerator" as a New Difference of Doubled and Unit.
Create "Outcome" as a New Formatting's Scaler of Numerator and Exponent.
Create "Expected" as a New Present of a New Wide Integer of Significant.
Create "Equal" as a New Equality of Outcome and Expected.
Create "Step" as a New Integer of 1.
Create "Decimal" as a New Difference of Exponent and Step.
Create "Aligned" as a New Equality of Decimal and Power.
Return a New Agreement of Equal and Aligned.

    Create Formatting's "Reduce|-s|-d|--ing|--er" as a Generic Function with Inputs of "coefficient" as `u128` and "exponent" as `i32` and Outputs of `(u128, i32)`.
Create "Ten" as a New Wide Integer of 10.
Create "Rest" as a New Remainder of Coefficient and Ten.
Create "Zero" as a New Wide Integer of 0.
Create "Original" as a New Pair of Coefficient and Exponent.
If Rest isnt Zero, Return Original.
Create "Shortened" as a New Quotient of Coefficient and Ten.
Create "One" as a New Integer of 1.
Create "Higher" as a New Sum of Exponent and One.
Return a New Formatting's Reducer of Shortened and Higher.

    Create Formatting's "Prepend|-s|-ed|-ing|-er" as a Generic Function with Inputs of "digits" as `String` and "point" as `i32` and Outputs of `String`.
Create "Zero" as a New Integer of 0.
Create "Padding" as a New Zeros of a New Indexed Integer of a New Difference of Zero and Point.
Return a New Prefix of Padding and Digits.

    Create Formatting's "Append|-s|-ed|-ing|-er" as a Generic Function with Inputs of "digits" as `String` and "point" as `i32` and Outputs of `String`.
Create "Width" as a New Integer of a New Length of Digits.
Create "Padding" as a New Zeros of a New Indexed Integer of a New Difference of Point and Width.
Return a New Suffix of Digits and Padding.

    Create Formatting's "Expand|-s|-ed|-ing|-er" as a Generic Function with Inputs of "digits" as `String` and "exponent" as `i32` and Outputs of `String`.
Create "Width" as a New Integer of a New Length of Digits.
Create "Point" as a New Sum of Width and Exponent.
Create "Zero" as a New Integer of 0.
If Point is Zero, Create "Before" as a New Formatting's Prepender of Digits and Point and then Return Before.
If Point less than Zero, Create "Leading" as a New Formatting's Prepender of Digits and Point and then Return Leading.
If Point is Width, Create "After" as a New Formatting's Appender of Digits and Point and then Return After.
If Point greater than Width, Create "Extended" as a New Formatting's Appender of Digits and Point and then Return Extended.
Create "Result" as a New Mutable Binding of Digits.
Create "Position" as a New Indexed Integer of Point.
Insert Result and Position and `.`.
Return Result.

    Create Formatting's "Contract|-s|-ed|-ing|-or" as a Generic Function with Inputs of "digits" as `String` and "exponent" as `i32` and Outputs of `String`.
Create "Width" as a New Integer of a New Length of Digits.
Create "One" as a New Integer of 1.
Create "Last" as a New Difference of Width and One.
Create "Power" as a New Sum of Exponent and Last.
Create "Position" as a New Indexed Integer of 1.
Create "Symbols" as a New Mutable Binding of Digits.
If Width greater than One, Insert Symbols and Position and `.`.
Return a New Scientific of Symbols and Power.

    Create Formatting's "Finish|-es|-ed|-ing|-er" as a Generic Function with Inputs of "digits" as `String` and "value" as `f64` and Outputs of `Vec<u8>`.
Create "Negative" as a New Sign of Value.
Create "Position" as a New Indexed Integer of 0.
Create "Buffer" as a New Mutable Binding of Digits.
If Negative, Insert Buffer and Position and `-`.
Return a New Encoding of Buffer.

    Create Formatting's "Render|-s|-ed|-ing|-er" as a Generic Function with Inputs of "digits" as `&str` and "exponent" as `&str` and "value" as `f64` and Outputs of `Vec<u8>`.
Create "Reading" as a New Parser of Exponent.
Create "Power" as a New Contents of Reading and `native finite float exponent`.
Create "Magnitude" as a New Trimmer of Digits and `-`.
Create "Separated" as a New Splitter of Magnitude and `.`.
Create "Default" as a New Pair of Magnitude and ``.
Create "Size" as a New Length of a New Second of a New Alternative of Separated and Default.
Create "Places" as a New Integer of Size.
Create "Decimal" as a New Difference of Power and Places.
Create "Amount" as a New Wide Parser of a New Replacement of Magnitude and `.` and ``.
Create "Whole" as a New Contents of Amount and `native finite float digits`.
Create "Two" as a New Wide Integer of 2.
Create "Rest" as a New Remainder of Whole and Two.
Create "One" as a New Wide Integer of 1.
Create "Odd" as a New Equality of Rest and One.
Create "Balanced" as a New Formatting's Balancer of Value and Whole and Decimal.
Create "Adjust" as a New Agreement of Odd and Balanced.
Create "Previous" as a New Difference of Whole and One.
Create "Chosen" as a New Choice of Adjust and Previous and Whole.
Create "Normal" as a New Formatting's Reducer of Chosen and Decimal.
Create "Significant" as a New First of Normal.
Create "Distance" as a New Second of Normal.
Create "Symbols" as a New Description of Significant.
Create "Measure" as a New Absolute of Value.
Create "Bounds" as a New Interval of 0.0001 and 1000000.
Create "Ordinary" as a New Membership of Bounds and Measure.
If Ordinary, Create "Decimaltext" as a New Formatting's Expander of Symbols and Distance and then Create "Ready" as a New Formatting's Finisher of Decimaltext and Value and then Return Ready.
Create "Scientifictext" as a New Formatting's Contractor of Symbols and Distance.
Return a New Formatting's Finisher of Scientifictext and Value.

    Publicly create Formatting's "Format|-s|-ted|-ting|-ter" as a Generic `format_number` Function with Inputs of "value" as `f64` and Outputs of `Vec<u8>`.
If Value is 0, Create "Plain" as a New Description of Value and then Create "Raw" as a New Encoding of Plain and then Return Raw.
Create "Spelling" as a New Notation of Value.
Create "Separated" as a New Splitter of Spelling and `e`.
Create "Exists" as a New Presence of Separated.
If Exists is false, Create "Encoded" as a New Encoding of Spelling and then Return Encoded.
Create "Components" as a New Contents of Separated and `native scientific notation`.
Create "Characters" as a New First of Components.
Create "Order" as a New Second of Components.
Return a New Formatting's Renderer of Characters and Order and Value.
