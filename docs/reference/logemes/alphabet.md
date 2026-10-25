# The alphabet and composition

## Nine logical pairs

The nine vowels name concepts and nine consonants name actions. These are all 18 selected Rust segments, confirmed by the [frozen asset manifest](../../../logiceme/testdata/lucid/templates.sha256.json).

| Action | Concept | Responsibility |
| --- | --- | --- |
| B — declare | A — declaration | Introduce names, storage and source units. |
| D — calculate | E — calculation | Produce expressions and conditions. |
| G — evaluate | I — evaluation | Select conditional alternatives. |
| F — define | O — definition | Supply values, assignments and bodies. |
| H — describe | U — description | Express structural properties and ordered constituents. |
| K — operate | W — operation | Group definitions and form callable operations. |
| P — iterate | X — iteration | Repeat or traverse supplied logic. |
| M — instantiate | Y — instantiation | Type, construct and invoke. |
| C — execute/asynchronize | V — execution/asynchrony | Scheduling, suspension and asynchronous compositions. |

A comparison such as native `x < y` belongs to calculation, even when an evaluation consumes it. Declaring a variable does not choose its initializer. A record's braces do not make it an operation. Template-time loops over fields do not establish runtime iteration. C/V's abstract ownership is broader than the implemented Rust scheduling forms.

## Syllables

A syllable has a vowel nucleus, an optional consonant onset and optional coda, plus metadata. A vowel can stand alone; a consonant cannot. The design discusses a possible two-vowel nucleus and four-letter form, but does not establish any such four-letter form as operational. Do not freely concatenate vowels or assign meanings to unused permutations.

The nucleus determines the logical category exposed to a parent. An onset selects an action or specialization; a coda adds one qualification while retaining its owner's meaning. For example FW defines an operation; KY invokes a callable. Metadata supplies data to a composition and does not silently redefine the letters.

## Sequence operators

`^` integrates a child composition into the preceding composition. `*` combines siblings; parentheses group a sequence. Child logical categories and singular/plural slots control which combinations are valid. The generator preserves explicit child occurrences and source order. These operators describe source composition, not a universal runtime evaluation-order promise for arbitrary native operands.

```text
FW^(W^(BOD^E * BOD^E))
```

This schematic shape supplies an ordered operation body to a named function. It omits required metadata and is not a complete executable AML example. Use a complete [statements witness](../../../logiceme/testdata/lucid/statements.aml.txt) to see the actual framing and explicit native statement punctuation.

## Declaration versus definition

O supplies an existing target's value. OB supplies the value and declares the binding. DO/DOB obtain a value from one Calculation child. BOD exposes one Calculation through the definition-category declaration adapter; it is not interchangeable with DOB. [Declarations](declarations.md) explain the corresponding shapes and cardinalities.

The maintained [logical model](../../../builtins/LOGEMES.md) describes design ownership. [Implementation contracts](../../../builtins/CONTRACTS.md) require explicit child roles, preservation of native bytes and separate evidence for generation, compilation and behavior. They do not authorize expanding the frozen Rust surface.
