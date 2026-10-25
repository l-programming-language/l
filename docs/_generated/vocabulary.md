# Vocabulary inventory

The configured lexicon supplies roots, inflection transformations, grammatical categories and handler descriptions. These tables are exact configuration data, not a claim that every handler has native instruction support. See [grammar and vocabulary](../spec/grammar.md) and the linked semantic subjects. Source-defined words, including builtins and dictionary exports, are documented separately.

## nouns

### it

<a id="vocabulary-nouns-it"></a>

Pronouns bind through source paragraph antecedents, not the last dynamically evaluated value. Separate generic-action argument rules also apply.

Semantic context: [Pronouns](../spec/names-and-scope.md#pronouns).

| Field | Value |
| --- | --- |
| root | <code>it</code> |
| pos | <code>noun</code> |
| type | <code>pronoun.definitive</code> |
| postfixes | <code>{"---they":"plural"}</code> |
| value | <code>IT(inputs[0])</code> |
| inputs | <code>["*.noun"]</code> |
| outputs | <code>["value"]</code> |

### that

<a id="vocabulary-nouns-that"></a>

Pronouns bind through source paragraph antecedents, not the last dynamically evaluated value. Separate generic-action argument rules also apply.

Semantic context: [Pronouns](../spec/names-and-scope.md#pronouns).

| Field | Value |
| --- | --- |
| root | <code>that</code> |
| pos | <code>noun</code> |
| type | <code>pronoun.objective</code> |
| postfixes | <code>{"---ose":"plural"}</code> |
| value | <code>THAT(inputs[0])</code> |
| inputs | <code>["*.noun"]</code> |
| outputs | <code>["value"]</code> |

### this

<a id="vocabulary-nouns-this"></a>

Pronouns bind through source paragraph antecedents, not the last dynamically evaluated value. Separate generic-action argument rules also apply.

Semantic context: [Pronouns](../spec/names-and-scope.md#pronouns).

| Field | Value |
| --- | --- |
| root | <code>this</code> |
| pos | <code>noun</code> |
| type | <code>pronoun.subjective</code> |
| postfixes | <code>{"---ese":"plural"}</code> |
| value | <code>THIS(inputs[0])</code> |
| inputs | <code>["*.noun"]</code> |
| outputs | <code>["value"]</code> |

### input

<a id="vocabulary-nouns-input"></a>

Input participates in callable signature metadata; the current process member implementation supports Output/Error. Runtime argument transport is outside the supported contract.

Semantic context: [Input metadata and runtime boundary](../reference/builtins/index.md#runtime-members).

| Field | Value |
| --- | --- |
| root | <code>input</code> |
| pos | <code>noun</code> |
| type | <code>functional.inflective</code> |
| postfixes | <code>{"-s":"plural"}</code> |
| value | <code>INPUT(inputs[0])</code> |
| inputs | <code>["*.thing"]</code> |
| outputs | <code>["value"]</code> |

### output

<a id="vocabulary-nouns-output"></a>

Runtime member access requires a compiler-resolved process target. Output is finalized raw stdout; Error has its own lifecycle and diagnostic boundaries.

Semantic context: [Process members](../spec/execution.md#output-and-error).

| Field | Value |
| --- | --- |
| root | <code>output</code> |
| pos | <code>noun</code> |
| type | <code>functional.deflective</code> |
| postfixes | <code>{"-s":"plural"}</code> |
| value | <code>OUTPUT(inputs[0])</code> |
| inputs | <code>["*.thing"]</code> |
| outputs | <code>["value"]</code> |

### error

<a id="vocabulary-nouns-error"></a>

Runtime member access requires a compiler-resolved process target. Output is finalized raw stdout; Error has its own lifecycle and diagnostic boundaries.

Semantic context: [Process members](../spec/execution.md#output-and-error).

| Field | Value |
| --- | --- |
| root | <code>error</code> |
| pos | <code>noun</code> |
| type | <code>functional.error</code> |
| postfixes | <code>{"-s":"plural"}</code> |
| value | <code>ERROR(inputs[0])</code> |
| inputs | <code>["*.thing"]</code> |
| outputs | <code>["value"]</code> |

### list

<a id="vocabulary-nouns-list"></a>

This lexical category must be distinguished from similarly named compiler representations. Read the construction and native-lowering boundaries at the semantic target.

Semantic context: [list](../spec/collections-and-categories.md#lists).

| Field | Value |
| --- | --- |
| root | <code>list</code> |
| pos | <code>noun</code> |
| type | <code>iterative.list</code> |
| postfixes | <code>{"-s":"plural"}</code> |
| value | <code>LIST(inputs[0])</code> |
| inputs | <code>["*.thing"]</code> |
| outputs | <code>["value"]</code> |

### dictionary

<a id="vocabulary-nouns-dictionary"></a>

This lexical category must be distinguished from similarly named compiler representations. Read the construction and native-lowering boundaries at the semantic target.

Semantic context: [dictionary](../spec/collections-and-categories.md#dictionaries).

| Field | Value |
| --- | --- |
| root | <code>dictionary</code> |
| pos | <code>noun</code> |
| type | <code>iterative.dictionary</code> |
| postfixes | <code>{"--ies":"plural"}</code> |
| value | <code>DICTIONARY(inputs[0], inputs[1])</code> |
| inputs | <code>["text","*.thing"]</code> |
| outputs | <code>["value"]</code> |

### null

<a id="vocabulary-nouns-null"></a>

The configured null root is a source spelling. Its storage and executable support are documented independently; the handler description below is not a callable API.

Semantic context: [null](../spec/values.md#null-and-internal-states).

| Field | Value |
| --- | --- |
| root | <code>null</code> |
| pos | <code>noun</code> |
| type | <code>primitive.null</code> |
| postfixes | <code>{"-s":"plural"}</code> |
| value | <code>NULL(inputs[0])</code> |
| inputs | <code>["*.thing"]</code> |
| outputs | <code>["value"]</code> |

### thing

<a id="vocabulary-nouns-thing"></a>

The configured thing root is a source spelling. Its storage and executable support are documented independently; the handler description below is not a callable API.

Semantic context: [thing](../spec/values.md#thing).

| Field | Value |
| --- | --- |
| root | <code>thing</code> |
| pos | <code>noun</code> |
| type | <code>primitive.thing</code> |
| postfixes | <code>{"-s":"plural"}</code> |
| value | <code>THING(inputs[0])</code> |
| inputs | <code>["*.thing"]</code> |
| outputs | <code>["value"]</code> |

### text

<a id="vocabulary-nouns-text"></a>

The configured text root is a source spelling. Its storage and executable support are documented independently; the handler description below is not a callable API.

Semantic context: [text](../spec/values.md#text).

| Field | Value |
| --- | --- |
| root | <code>text</code> |
| pos | <code>noun</code> |
| type | <code>primitive.text</code> |
| postfixes | <code>{"-s":"plural"}</code> |
| value | <code>TEXT(inputs[0])</code> |
| inputs | <code>["*.text"]</code> |
| outputs | <code>["value"]</code> |

### number

<a id="vocabulary-nouns-number"></a>

The configured number root is a source spelling. Its storage and executable support are documented independently; the handler description below is not a callable API.

Semantic context: [number](../spec/values.md#number).

| Field | Value |
| --- | --- |
| root | <code>number</code> |
| pos | <code>noun</code> |
| type | <code>primitive.number</code> |
| postfixes | <code>{"-s":"plural"}</code> |
| value | <code>NUMBER(inputs[0])</code> |
| inputs | <code>["*.number"]</code> |
| outputs | <code>["value"]</code> |

### byte

<a id="vocabulary-nouns-byte"></a>

The configured byte root is a source spelling. Its storage and executable support are documented independently; the handler description below is not a callable API.

Semantic context: [byte](../spec/values.md#byte-and-bytes).

| Field | Value |
| --- | --- |
| root | <code>byte</code> |
| pos | <code>noun</code> |
| type | <code>primitive.byte</code> |
| postfixes | <code>{"-s":"plural"}</code> |
| value | <code>BYTES(inputs[0])</code> |
| inputs | <code>["*.byte"]</code> |
| outputs | <code>["value"]</code> |

### truth

<a id="vocabulary-nouns-truth"></a>

The configured truth root is a source spelling. Its storage and executable support are documented independently; the handler description below is not a callable API.

Semantic context: [truth](../spec/values.md#truth).

| Field | Value |
| --- | --- |
| root | <code>truth</code> |
| pos | <code>noun</code> |
| type | <code>primitive.truth</code> |
| postfixes | <code>{"-s":"plural"}</code> |
| value | <code>TRUTH(inputs[0])</code> |
| inputs | <code>["*.truth"]</code> |
| outputs | <code>["value"]</code> |

### pointer

<a id="vocabulary-nouns-pointer"></a>

The configured pointer root is a source spelling. Its storage and executable support are documented independently; the handler description below is not a callable API.

Semantic context: [pointer](../spec/values.md#pointer).

| Field | Value |
| --- | --- |
| root | <code>pointer</code> |
| pos | <code>noun</code> |
| type | <code>primitive.pointer</code> |
| postfixes | <code>{"-s":"plural"}</code> |
| value | <code>POINTER(inputs[0])</code> |
| inputs | <code>["*.thing"]</code> |
| outputs | <code>["value"]</code> |

### op

<a id="vocabulary-nouns-op"></a>

This lexical category must be distinguished from similarly named compiler representations. Read the construction and native-lowering boundaries at the semantic target.

Semantic context: [op](../spec/collections-and-categories.md#functions-properties-and-operations).

| Field | Value |
| --- | --- |
| root | <code>op</code> |
| pos | <code>noun</code> |
| type | <code>applicative.operation</code> |
| postfixes | <code>{"-s":"plural"}</code> |
| value | <code>OP(inputs[0], inputs[1], inputs[2], inputs[3])</code> |
| inputs | <code>["text","numbers","text","text"]</code> |
| outputs | <code>["value"]</code> |

### prop

<a id="vocabulary-nouns-prop"></a>

This lexical category must be distinguished from similarly named compiler representations. Read the construction and native-lowering boundaries at the semantic target.

Semantic context: [prop](../spec/collections-and-categories.md#functions-properties-and-operations).

| Field | Value |
| --- | --- |
| root | <code>prop</code> |
| pos | <code>noun</code> |
| type | <code>applicative.property</code> |
| postfixes | <code>{"-s":"plural"}</code> |
| value | <code>PROP(inputs[0], inputs[1], inputs[2])</code> |
| inputs | <code>["text","text","thing"]</code> |
| outputs | <code>["value"]</code> |

### func

<a id="vocabulary-nouns-func"></a>

This lexical category must be distinguished from similarly named compiler representations. Read the construction and native-lowering boundaries at the semantic target.

Semantic context: [func](../spec/collections-and-categories.md#functions-properties-and-operations).

| Field | Value |
| --- | --- |
| root | <code>func</code> |
| pos | <code>noun</code> |
| type | <code>applicative.function</code> |
| postfixes | <code>{"-s":"plural"}</code> |
| value | <code>FUNC(inputs[0], inputs[1], inputs[2], inputs[3])</code> |
| inputs | <code>["text","props","texts","prop"]</code> |
| outputs | <code>["value"]</code> |

### definition

<a id="vocabulary-nouns-definition"></a>

This lexical category must be distinguished from similarly named compiler representations. Read the construction and native-lowering boundaries at the semantic target.

Semantic context: [definition](../spec/collections-and-categories.md#attributes-definitions-and-lexical-categories).

| Field | Value |
| --- | --- |
| root | <code>definition</code> |
| pos | <code>noun</code> |
| type | <code>applicative.definition</code> |
| postfixes | <code>{"-s":"plural"}</code> |
| value | <code>DEF(inputs[0])</code> |
| inputs | <code>["text"]</code> |
| outputs | <code>["value"]</code> |

### attribute

<a id="vocabulary-nouns-attribute"></a>

This lexical category must be distinguished from similarly named compiler representations. Read the construction and native-lowering boundaries at the semantic target.

Semantic context: [attribute](../spec/collections-and-categories.md#attributes-definitions-and-lexical-categories).

| Field | Value |
| --- | --- |
| root | <code>attribute</code> |
| pos | <code>noun</code> |
| type | <code>applicative.attribute</code> |
| postfixes | <code>{"-s":"plural"}</code> |
| value | <code>ATTRIBUTE(inputs[0], inputs[1])</code> |
| inputs | <code>["definition","attributes"]</code> |
| outputs | <code>["value"]</code> |

### object

<a id="vocabulary-nouns-object"></a>

This lexical category must be distinguished from similarly named compiler representations. Read the construction and native-lowering boundaries at the semantic target.

Semantic context: [object](../spec/collections-and-categories.md#classes-and-objects).

| Field | Value |
| --- | --- |
| root | <code>object</code> |
| pos | <code>noun</code> |
| type | <code>applicative.object</code> |
| postfixes | <code>{"-s":"plural"}</code> |
| value | <code>OBJECT(inputs[0], inputs[1])</code> |
| inputs | <code>["definition","attributes"]</code> |
| outputs | <code>["value"]</code> |

### class

<a id="vocabulary-nouns-class"></a>

This lexical category must be distinguished from similarly named compiler representations. Read the construction and native-lowering boundaries at the semantic target.

Semantic context: [class](../spec/collections-and-categories.md#classes-and-objects).

| Field | Value |
| --- | --- |
| root | <code>class</code> |
| pos | <code>noun</code> |
| type | <code>applicative.class</code> |
| postfixes | <code>{"-es":"plural"}</code> |
| value | <code>CLASS(inputs[0], inputs[1])</code> |
| inputs | <code>["objects","classes"]</code> |
| outputs | <code>["value"]</code> |

### service

<a id="vocabulary-nouns-service"></a>

This lexical category must be distinguished from similarly named compiler representations. Read the construction and native-lowering boundaries at the semantic target.

Semantic context: [service](../spec/collections-and-categories.md#services).

| Field | Value |
| --- | --- |
| root | <code>service</code> |
| pos | <code>noun</code> |
| type | <code>applicative.service</code> |
| postfixes | <code>{"-s":"plural"}</code> |
| value | <code>SERVICE(inputs[0], inputs[1], inputs[2])</code> |
| inputs | <code>["objects","classes","services"]</code> |
| outputs | <code>["value"]</code> |

### app

<a id="vocabulary-nouns-app"></a>

This lexical category must be distinguished from similarly named compiler representations. Read the construction and native-lowering boundaries at the semantic target.

Semantic context: [app](../spec/collections-and-categories.md#apps).

| Field | Value |
| --- | --- |
| root | <code>app</code> |
| pos | <code>noun</code> |
| type | <code>applicative.app</code> |
| postfixes | <code>{"-s":"plural"}</code> |
| value | <code>APP(inputs[0], inputs[1], inputs[2])</code> |
| inputs | <code>["objects","services","dictionary"]</code> |
| outputs | <code>["value"]</code> |

## verbs

### passive

<a id="vocabulary-verbs-passive"></a>

Assignment follows resolved scalar/storage rules and definite assignment. Runtime ordering is preserved independently of parser retry order.

Semantic context: [Assignment](../spec/control-flow.md#assignment).

| Field | Value |
| --- | --- |
| root | <code>is</code> |
| pos | <code>verb</code> |
| type | <code>action.passive</code> |
| postfixes | <code>{"-nt":"null"}</code> |
| value | <code>IS(inputs[0])</code> |
| inputs | <code>["*.thing"]</code> |
| outputs | <code>["value"]</code> |

### associative

<a id="vocabulary-verbs-associative"></a>

Association names an existing canonical value or definition. It does not create an independent runtime copy.

Semantic context: [Aliases](../spec/names-and-scope.md#aliases-and-canonical-identity).

| Field | Value |
| --- | --- |
| root | <code>use</code> |
| pos | <code>verb</code> |
| type | <code>action.associative</code> |
| postfixes | <code>{"-s":"active","-d":"passive","--ing":"applicative","-r":"objective"}</code> |
| value | <code>USE(inputs[0])</code> |
| inputs | <code>["*.noun"]</code> |
| outputs | <code>["value"]</code> |

### declarative

<a id="vocabulary-verbs-declarative"></a>

A declaration associates a source name with the indicated category or value. Voice-specific recognizers do not imply interchangeable native behavior.

Semantic context: [Declarations](../spec/definitions.md#create-define-and-describe).

| Field | Value |
| --- | --- |
| root | <code>create</code> |
| pos | <code>verb</code> |
| type | <code>action.declarative</code> |
| postfixes | <code>{"-s":"active","-d":"passive","--ing":"applicative","--or":"objective"}</code> |
| value | <code>DECLARE(inputs[0])</code> |
| inputs | <code>["*.noun"]</code> |
| outputs | <code>["value"]</code> |

### definitive

<a id="vocabulary-verbs-definitive"></a>

Definition clauses establish source vocabulary or typed native-backed forms. Accepted construction and framing are bounded by definition expansion.

Semantic context: [Definitions](../spec/definitions.md#create-define-and-describe).

| Field | Value |
| --- | --- |
| root | <code>define</code> |
| pos | <code>verb</code> |
| type | <code>action.definitive</code> |
| postfixes | <code>{"-s":"active","-d":"passive","--ing":"applicative","-r":"objective"}</code> |
| value | <code>DEFINE(inputs[0])</code> |
| inputs | <code>["*.noun"]</code> |
| outputs | <code>["value"]</code> |

### descriptive

<a id="vocabulary-verbs-descriptive"></a>

Descriptive clauses apply source-defined adjective/adverb meanings. Constructor descriptors preserve the target and source-unit ownership.

Semantic context: [Descriptors](../spec/definitions.md#descriptors-and-constructors).

| Field | Value |
| --- | --- |
| root | <code>describe</code> |
| pos | <code>verb</code> |
| type | <code>action.descriptive</code> |
| postfixes | <code>{"-s":"active","-d":"passive","--ing":"applicative","-r":"objective"}</code> |
| value | <code>DESCRIBE(inputs[0])</code> |
| inputs | <code>["*.noun"]</code> |
| outputs | <code>["value"]</code> |

### expansive

<a id="vocabulary-verbs-expansive"></a>

The configured word is recognized vocabulary, but no native outer collection instruction is supplied for this operation. Do not treat it as a supported general collection API.

Semantic context: [Collection operations](../spec/control-flow.md#iteration-and-collection-operations).

| Field | Value |
| --- | --- |
| root | <code>put</code> |
| pos | <code>verb</code> |
| type | <code>action.expansive</code> |
| postfixes | <code>{"-s":"active","-":"passive","-ting":"applicative","-er":"objective"}</code> |
| value | <code>PUT(inputs[0], inputs[1], inputs[2])</code> |
| inputs | <code>["*.thing","*.noun.iterative","text"]</code> |
| outputs | <code>["value"]</code> |

### locative

<a id="vocabulary-verbs-locative"></a>

The configured word is recognized vocabulary, but no native outer collection instruction is supplied for this operation. Do not treat it as a supported general collection API.

Semantic context: [Collection operations](../spec/control-flow.md#iteration-and-collection-operations).

| Field | Value |
| --- | --- |
| root | <code>place</code> |
| pos | <code>verb</code> |
| type | <code>action.locative</code> |
| postfixes | <code>{"-s":"active","-d":"passive","--ing":"applicative","-r":"objective"}</code> |
| value | <code>PLACE(inputs[0], inputs[1], inputs[2])</code> |
| inputs | <code>["*.thing","*.noun.iterative","text"]</code> |
| outputs | <code>["value"]</code> |

### receptive

<a id="vocabulary-verbs-receptive"></a>

The configured word is recognized vocabulary, but no native outer collection instruction is supplied for this operation. Do not treat it as a supported general collection API.

Semantic context: [Collection operations](../spec/control-flow.md#iteration-and-collection-operations).

| Field | Value |
| --- | --- |
| root | <code>get</code> |
| pos | <code>verb</code> |
| type | <code>action.receptive</code> |
| postfixes | <code>{"-s":"active","---otten":"passive","-ting":"applicative","-ter":"objective"}</code> |
| value | <code>GET(inputs[0], inputs[1])</code> |
| inputs | <code>["text","*.noun.iterative"]</code> |
| outputs | <code>["value"]</code> |

### reductive

<a id="vocabulary-verbs-reductive"></a>

The configured word is recognized vocabulary, but no native outer collection instruction is supplied for this operation. Do not treat it as a supported general collection API.

Semantic context: [Collection operations](../spec/control-flow.md#iteration-and-collection-operations).

| Field | Value |
| --- | --- |
| root | <code>take</code> |
| pos | <code>verb</code> |
| type | <code>action.reductive</code> |
| postfixes | <code>{"-s":"active","-n":"passive","--ing":"applicative","-r":"objective"}</code> |
| value | <code>TAKE(inputs[0], inputs[1])</code> |
| inputs | <code>["text","*.noun.iterative"]</code> |
| outputs | <code>["value"]</code> |

### generative

<a id="vocabulary-verbs-generative"></a>

Build is the runtime readiness operation. It is distinct from the CLI command that compiles native output.

Semantic context: [Build](../spec/execution.md#build-run-stop-and-await).

| Field | Value |
| --- | --- |
| root | <code>build</code> |
| pos | <code>verb</code> |
| type | <code>action.generative</code> |
| postfixes | <code>{"-s":"active","--t":"passive","-ing":"applicative","-er":"objective"}</code> |
| value | <code>BUILD(inputs[0])</code> |
| inputs | <code>["*.noun.app"]</code> |
| outputs | <code>["value"]</code> |

### executive

<a id="vocabulary-verbs-executive"></a>

Run launches the selected process. The grammar does not imply that every applicative category is runnable.

Semantic context: [Run](../spec/execution.md#build-run-stop-and-await).

| Field | Value |
| --- | --- |
| root | <code>run</code> |
| pos | <code>verb</code> |
| type | <code>action.executive</code> |
| postfixes | <code>{"-s":"active","--an":"passive","-ning":"applicative","-ner":"objective"}</code> |
| value | <code>RUN(inputs[0])</code> |
| inputs | <code>["*.noun.app"]</code> |
| outputs | <code>["value"]</code> |

### suspensive

<a id="vocabulary-verbs-suspensive"></a>

Await joins a selected process invocation and activates its continuation once. A parsed clause alone does not establish valid activation placement.

Semantic context: [Await](../spec/execution.md#continuations-and-activation).

| Field | Value |
| --- | --- |
| root | <code>await</code> |
| pos | <code>verb</code> |
| type | <code>action.suspensive</code> |
| postfixes | <code>{"-s":"active","-ed":"passive","-ing":"applicative"}</code> |
| value | <code>AWAIT(inputs[0])</code> |
| inputs | <code>["*.noun.app"]</code> |
| outputs | <code>["value"]</code> |

### interruptive

<a id="vocabulary-verbs-interruptive"></a>

Stop requests termination without joining. Final output still requires the Await contract.

Semantic context: [Stop](../spec/execution.md#build-run-stop-and-await).

| Field | Value |
| --- | --- |
| root | <code>stop</code> |
| pos | <code>verb</code> |
| type | <code>action.interruptive</code> |
| postfixes | <code>{"-s":"active","-ped":"passive","-ping":"applicative","-pper":"objective"}</code> |
| value | <code>STOP(inputs[0])</code> |
| inputs | <code>["*.noun.app"]</code> |
| outputs | <code>["value"]</code> |

### inflective

<a id="vocabulary-verbs-inflective"></a>

Configured Import/Export handlers do not replace the supported Use import syntax and are absent from native outer instruction dispatch.

Semantic context: [Import/export distinction](../spec/imports.md#indexed-dictionaries).

| Field | Value |
| --- | --- |
| root | <code>import</code> |
| pos | <code>verb</code> |
| type | <code>action.inflective</code> |
| postfixes | <code>{"-s":"active","-ed":"passive","-ing":"applicative","-er":"objective"}</code> |
| value | <code>IMPORT(inputs[0])</code> |
| inputs | <code>["*.thing"]</code> |
| outputs | <code>["value"]</code> |

### deflective

<a id="vocabulary-verbs-deflective"></a>

Configured Import/Export handlers do not replace the supported Use import syntax and are absent from native outer instruction dispatch.

Semantic context: [Import/export distinction](../spec/imports.md#indexed-dictionaries).

| Field | Value |
| --- | --- |
| root | <code>export</code> |
| pos | <code>verb</code> |
| type | <code>action.deflective</code> |
| postfixes | <code>{"-s":"active","-ed":"passive","-ing":"applicative","-er":"objective"}</code> |
| value | <code>EXPORT(inputs[0])</code> |
| inputs | <code>["*.thing"]</code> |
| outputs | <code>["value"]</code> |

### observative

<a id="vocabulary-verbs-observative"></a>

Show emits the resolved operand without Print’s added newline; Output reads are subject to process finalization.

Semantic context: [Show](../spec/control-flow.md#show-and-output).

| Field | Value |
| --- | --- |
| root | <code>show</code> |
| pos | <code>verb</code> |
| type | <code>action.observative</code> |
| postfixes | <code>{"-s":"active","-ed":"passive","-ing":"applicative","-er":"objective"}</code> |
| value | <code>SHOW(inputs[0])</code> |
| inputs | <code>["*.thing"]</code> |
| outputs | <code>["value"]</code> |

### terminative

<a id="vocabulary-verbs-terminative"></a>

Exit ends the compiled program; source Function.Return is a separate builtin expansion.

Semantic context: [Exit](../spec/control-flow.md#exit-and-return).

| Field | Value |
| --- | --- |
| root | <code>exit</code> |
| pos | <code>verb</code> |
| type | <code>action.terminative</code> |
| postfixes | <code>{"-s":"active","-ed":"passive","-ing":"applicative","-er":"objective"}</code> |
| value | <code>EXIT(inputs[0])</code> |
| inputs | <code>["*.thing"]</code> |
| outputs | <code>["value"]</code> |

### appropriative

<a id="vocabulary-verbs-appropriative"></a>

The configured comparison/appropriation root is used in bounded condition forms. Operand types and Error-member restrictions are enforced by native lowering.

Semantic context: [Comparisons](../spec/control-flow.md#comparisons).

| Field | Value |
| --- | --- |
| root | <code>have</code> |
| pos | <code>verb</code> |
| type | <code>operation.appropriative</code> |
| postfixes | <code>{"---s":"active","---d":"passive","--ing":"applicative","-r":"objective"}</code> |
| value | <code>HAVE(inputs[0])</code> |
| inputs | <code>["*.thing"]</code> |
| outputs | <code>["value"]</code> |

### comparative

<a id="vocabulary-verbs-comparative"></a>

The configured comparison/appropriation root is used in bounded condition forms. Operand types and Error-member restrictions are enforced by native lowering.

Semantic context: [Comparisons](../spec/control-flow.md#comparisons).

| Field | Value |
| --- | --- |
| root | <code>equal</code> |
| pos | <code>verb</code> |
| type | <code>operation.comparative</code> |
| postfixes | <code>{"-s":"active","-ed":"passive","-ing":"applicative","-er":"objective"}</code> |
| value | <code>EQUAL(inputs[0])</code> |
| inputs | <code>["*.thing"]</code> |
| outputs | <code>["value"]</code> |

## adjectives

### visibility.public

<a id="vocabulary-adjectives-visibility.public"></a>

Public/Publicly marks exactly the associated declaration. It does not publish unrelated owners or sibling declarations.

Semantic context: [Visibility](../spec/names-and-scope.md#paragraph-scope).

| Field | Value |
| --- | --- |
| root | <code>public</code> |
| pos | <code>adjective</code> |
| type | <code>visibility.public</code> |
| postfixes | <code>{}</code> |

### initialization

<a id="vocabulary-adjectives-initialization"></a>

Constructor/descriptor vocabulary determines source construction roles. Its presence does not imply arbitrary type construction or inherited native metadata.

Semantic context: [Descriptors](../spec/definitions.md#descriptors-and-constructors).

| Field | Value |
| --- | --- |
| root | <code>new</code> |
| pos | <code>adjective</code> |
| type | <code>initialization</code> |
| postfixes | <code>{}</code> |
| value | <code>NEW(inputs[0])</code> |
| inputs | <code>["*.noun"]</code> |
| outputs | <code>["value"]</code> |

### abstraction

<a id="vocabulary-adjectives-abstraction"></a>

Constructor/descriptor vocabulary determines source construction roles. Its presence does not imply arbitrary type construction or inherited native metadata.

Semantic context: [Descriptors](../spec/definitions.md#descriptors-and-constructors).

| Field | Value |
| --- | --- |
| root | <code>generic</code> |
| pos | <code>adjective</code> |
| type | <code>abstraction</code> |
| postfixes | <code>{}</code> |
| value | <code>GENERIC(inputs[0])</code> |
| inputs | <code>["*.noun"]</code> |
| outputs | <code>["value"]</code> |

## adverbs

### visibility.public

<a id="vocabulary-adverbs-visibility.public"></a>

Public/Publicly marks exactly the associated declaration. It does not publish unrelated owners or sibling declarations.

Semantic context: [Visibility](../spec/names-and-scope.md#paragraph-scope).

| Field | Value |
| --- | --- |
| root | <code>publicly</code> |
| pos | <code>adverb</code> |
| type | <code>visibility.public</code> |
| postfixes | <code>{}</code> |

### nullification

<a id="vocabulary-adverbs-nullification"></a>

These tokens participate in configured grammatical conditions/modifiers. In particular, joined conditions do not yet have a native outer instruction operator; configured presence is not a general Boolean API.

Semantic context: [Conditions and modifiers](../spec/control-flow.md#comparisons).

| Field | Value |
| --- | --- |
| root | <code>not</code> |
| pos | <code>adverb</code> |
| type | <code>nullification</code> |
| postfixes | <code>{}</code> |
| value | <code>NOT(inputs)</code> |
| inputs | <code>["*.thing"]</code> |
| outputs | <code>["value"]</code> |

### comparison.subtraction

<a id="vocabulary-adverbs-comparison.subtraction"></a>

These tokens participate in configured grammatical conditions/modifiers. In particular, joined conditions do not yet have a native outer instruction operator; configured presence is not a general Boolean API.

Semantic context: [Conditions and modifiers](../spec/control-flow.md#comparisons).

| Field | Value |
| --- | --- |
| root | <code>less</code> |
| pos | <code>adverb</code> |
| type | <code>comparison.subtraction</code> |
| postfixes | <code>{}</code> |
| value | <code>LESSER(inputs)</code> |
| inputs | <code>["*.thing"]</code> |
| outputs | <code>["value"]</code> |

### comparison.multiplication

<a id="vocabulary-adverbs-comparison.multiplication"></a>

These tokens participate in configured grammatical conditions/modifiers. In particular, joined conditions do not yet have a native outer instruction operator; configured presence is not a general Boolean API.

Semantic context: [Conditions and modifiers](../spec/control-flow.md#comparisons).

| Field | Value |
| --- | --- |
| root | <code>greater</code> |
| pos | <code>adverb</code> |
| type | <code>comparison.multiplication</code> |
| postfixes | <code>{}</code> |
| value | <code>GREATER(inputs)</code> |
| inputs | <code>["*.thing"]</code> |
| outputs | <code>["value"]</code> |

### progression

<a id="vocabulary-adverbs-progression"></a>

These tokens participate in configured grammatical conditions/modifiers. In particular, joined conditions do not yet have a native outer instruction operator; configured presence is not a general Boolean API.

Semantic context: [Conditions and modifiers](../spec/control-flow.md#comparisons).

| Field | Value |
| --- | --- |
| root | <code>then</code> |
| pos | <code>adverb</code> |
| type | <code>abstraction</code> |
| postfixes | <code>{}</code> |
| value | <code>THEN(inputs)</code> |
| inputs | <code>["*.thing"]</code> |
| outputs | <code>["value"]</code> |

## prepositions

### definement.declarative

<a id="vocabulary-prepositions-definement.declarative"></a>

This preposition establishes the specified grammatical relationship between source elements. Interpret its category through the configured phrase rules.

Semantic context: [Prepositions](../spec/grammar.md#phrases).

| Field | Value |
| --- | --- |
| root | <code>as</code> |
| pos | <code>preposition</code> |
| type | <code>definement.declarative</code> |
| postfixes | <code>{}</code> |
| value | <code>AS(inputs)</code> |
| inputs | <code>["*.thing"]</code> |
| outputs | <code>["value"]</code> |

### definement.definitive

<a id="vocabulary-prepositions-definement.definitive"></a>

This preposition establishes the specified grammatical relationship between source elements. Interpret its category through the configured phrase rules.

Semantic context: [Prepositions](../spec/grammar.md#phrases).

| Field | Value |
| --- | --- |
| root | <code>of</code> |
| pos | <code>preposition</code> |
| type | <code>definement.definitive</code> |
| postfixes | <code>{}</code> |
| value | <code>OF(inputs)</code> |
| inputs | <code>["*.thing"]</code> |
| outputs | <code>["value"]</code> |

### definement.locative

<a id="vocabulary-prepositions-definement.locative"></a>

This preposition establishes the specified grammatical relationship between source elements. Interpret its category through the configured phrase rules.

Semantic context: [Prepositions](../spec/grammar.md#phrases).

| Field | Value |
| --- | --- |
| root | <code>at</code> |
| pos | <code>preposition</code> |
| type | <code>definement.locative</code> |
| postfixes | <code>{}</code> |
| value | <code>AT(inputs)</code> |
| inputs | <code>["*.thing"]</code> |
| outputs | <code>["value"]</code> |

### consignment.associative

<a id="vocabulary-prepositions-consignment.associative"></a>

This preposition establishes the specified grammatical relationship between source elements. Interpret its category through the configured phrase rules.

Semantic context: [Prepositions](../spec/grammar.md#phrases).

| Field | Value |
| --- | --- |
| root | <code>with</code> |
| pos | <code>preposition</code> |
| type | <code>consignment.associative</code> |
| postfixes | <code>{}</code> |
| value | <code>WITH(inputs)</code> |
| inputs | <code>["*.thing"]</code> |
| outputs | <code>["value"]</code> |

### consignment.inflective

<a id="vocabulary-prepositions-consignment.inflective"></a>

This preposition establishes the specified grammatical relationship between source elements. Interpret its category through the configured phrase rules.

Semantic context: [Prepositions](../spec/grammar.md#phrases).

| Field | Value |
| --- | --- |
| root | <code>in</code> |
| pos | <code>preposition</code> |
| type | <code>consignment.inflective</code> |
| postfixes | <code>{}</code> |
| value | <code>IN(inputs)</code> |
| inputs | <code>["*.thing"]</code> |
| outputs | <code>["value"]</code> |

### consignment.deflective

<a id="vocabulary-prepositions-consignment.deflective"></a>

This preposition establishes the specified grammatical relationship between source elements. Interpret its category through the configured phrase rules.

Semantic context: [Prepositions](../spec/grammar.md#phrases).

| Field | Value |
| --- | --- |
| root | <code>out</code> |
| pos | <code>preposition</code> |
| type | <code>consignment.deflective</code> |
| postfixes | <code>{}</code> |
| value | <code>OUT(inputs)</code> |
| inputs | <code>["*.thing"]</code> |
| outputs | <code>["value"]</code> |

### consignment.indicative

<a id="vocabulary-prepositions-consignment.indicative"></a>

This preposition establishes the specified grammatical relationship between source elements. Interpret its category through the configured phrase rules.

Semantic context: [Prepositions](../spec/grammar.md#phrases).

| Field | Value |
| --- | --- |
| root | <code>by</code> |
| pos | <code>preposition</code> |
| type | <code>consignment.indicative</code> |
| postfixes | <code>{}</code> |
| value | <code>BY(inputs)</code> |
| inputs | <code>["*.thing"]</code> |
| outputs | <code>["value"]</code> |

### conjoinment.objective

<a id="vocabulary-prepositions-conjoinment.objective"></a>

This preposition establishes the specified grammatical relationship between source elements. Interpret its category through the configured phrase rules.

Semantic context: [Prepositions](../spec/grammar.md#phrases).

| Field | Value |
| --- | --- |
| root | <code>from</code> |
| pos | <code>preposition</code> |
| type | <code>conjoinment.objective</code> |
| postfixes | <code>{}</code> |
| value | <code>FROM(inputs)</code> |
| inputs | <code>["*.thing"]</code> |
| outputs | <code>["value"]</code> |

### conjoinment.subjective

<a id="vocabulary-prepositions-conjoinment.subjective"></a>

This preposition establishes the specified grammatical relationship between source elements. Interpret its category through the configured phrase rules.

Semantic context: [Prepositions](../spec/grammar.md#phrases).

| Field | Value |
| --- | --- |
| root | <code>to</code> |
| pos | <code>preposition</code> |
| type | <code>conjoinment.subjective</code> |
| postfixes | <code>{}</code> |
| value | <code>TO(inputs)</code> |
| inputs | <code>["*.thing"]</code> |
| outputs | <code>["value"]</code> |

### conjoinment.comparative

<a id="vocabulary-prepositions-conjoinment.comparative"></a>

This preposition establishes the specified grammatical relationship between source elements. Interpret its category through the configured phrase rules.

Semantic context: [Prepositions](../spec/grammar.md#phrases).

| Field | Value |
| --- | --- |
| root | <code>than</code> |
| pos | <code>preposition</code> |
| type | <code>conjoinment.comparative</code> |
| postfixes | <code>{}</code> |
| value | <code>THAN(inputs)</code> |
| inputs | <code>["*.thing"]</code> |
| outputs | <code>["value"]</code> |

### conjoinment.directive

<a id="vocabulary-prepositions-conjoinment.directive"></a>

This preposition establishes the specified grammatical relationship between source elements. Interpret its category through the configured phrase rules.

Semantic context: [Prepositions](../spec/grammar.md#phrases).

| Field | Value |
| --- | --- |
| root | <code>for</code> |
| pos | <code>preposition</code> |
| type | <code>conjoinment.directive</code> |
| postfixes | <code>{}</code> |
| value | <code>FOR(inputs)</code> |
| inputs | <code>["*.thing"]</code> |
| outputs | <code>["value"]</code> |

## articles

### variable

<a id="vocabulary-articles-variable"></a>

Articles qualify source bindings and typed argument roles. The any form in source actions is not an unrestricted Rust generic or implicit conversion.

Semantic context: [Articles and typed arguments](../spec/definitions.md#typed-slots).

| Field | Value |
| --- | --- |
| root | <code>any</code> |
| pos | <code>article</code> |
| type | <code>variable</code> |
| postfixes | <code>{}</code> |
| value | <code>ANY(inputs)</code> |
| inputs | <code>["*.thing"]</code> |
| outputs | <code>["value"]</code> |

### none

<a id="vocabulary-articles-none"></a>

Articles qualify source bindings and typed argument roles. The any form in source actions is not an unrestricted Rust generic or implicit conversion.

Semantic context: [Articles and typed arguments](../spec/definitions.md#typed-slots).

| Field | Value |
| --- | --- |
| root | <code>no</code> |
| pos | <code>article</code> |
| type | <code>none</code> |
| postfixes | <code>{}</code> |
| value | <code>NO(inputs)</code> |
| inputs | <code>["*.thing"]</code> |
| outputs | <code>["value"]</code> |

### indefinite

<a id="vocabulary-articles-indefinite"></a>

Articles qualify source bindings and typed argument roles. The any form in source actions is not an unrestricted Rust generic or implicit conversion.

Semantic context: [Articles and typed arguments](../spec/definitions.md#typed-slots).

| Field | Value |
| --- | --- |
| root | <code>a</code> |
| pos | <code>article</code> |
| type | <code>indefinite</code> |
| postfixes | <code>{"-n":"alt"}</code> |
| value | <code>A(inputs)</code> |
| inputs | <code>["*.noun"]</code> |
| outputs | <code>["value"]</code> |

### definite

<a id="vocabulary-articles-definite"></a>

Articles qualify source bindings and typed argument roles. The any form in source actions is not an unrestricted Rust generic or implicit conversion.

Semantic context: [Articles and typed arguments](../spec/definitions.md#typed-slots).

| Field | Value |
| --- | --- |
| root | <code>the</code> |
| pos | <code>article</code> |
| type | <code>definite</code> |
| postfixes | <code>{}</code> |
| value | <code>THE(inputs)</code> |
| inputs | <code>["*.noun"]</code> |
| outputs | <code>["value"]</code> |

## conjunctions

### mutual

<a id="vocabulary-conjunctions-mutual"></a>

These tokens participate in configured grammatical conditions/modifiers. In particular, joined conditions do not yet have a native outer instruction operator; configured presence is not a general Boolean API.

Semantic context: [Conditions and modifiers](../spec/control-flow.md#comparisons).

| Field | Value |
| --- | --- |
| root | <code>and</code> |
| pos | <code>conjunction</code> |
| type | <code>mutual</code> |
| postfixes | <code>{}</code> |
| value | <code>AND(inputs)</code> |
| inputs | <code>["*.thing"]</code> |
| outputs | <code>["value"]</code> |

### exclusive

<a id="vocabulary-conjunctions-exclusive"></a>

These tokens participate in configured grammatical conditions/modifiers. In particular, joined conditions do not yet have a native outer instruction operator; configured presence is not a general Boolean API.

Semantic context: [Conditions and modifiers](../spec/control-flow.md#comparisons).

| Field | Value |
| --- | --- |
| root | <code>or</code> |
| pos | <code>conjunction</code> |
| type | <code>exclusive</code> |
| postfixes | <code>{}</code> |
| value | <code>OR(inputs)</code> |
| inputs | <code>["*.thing"]</code> |
| outputs | <code>["value"]</code> |

## evaluations

### conditional

<a id="vocabulary-evaluations-conditional"></a>

These tokens participate in configured grammatical conditions/modifiers. In particular, joined conditions do not yet have a native outer instruction operator; configured presence is not a general Boolean API.

Semantic context: [Conditions and modifiers](../spec/control-flow.md#comparisons).

| Field | Value |
| --- | --- |
| root | <code>if</code> |
| pos | <code>evaluation</code> |
| type | <code>conditional</code> |
| postfixes | <code>{}</code> |
| value | <code>IF(inputs)</code> |
| inputs | <code>["*.thing"]</code> |
| outputs | <code>["value"]</code> |

### exceptional

<a id="vocabulary-evaluations-exceptional"></a>

These tokens participate in configured grammatical conditions/modifiers. In particular, joined conditions do not yet have a native outer instruction operator; configured presence is not a general Boolean API.

Semantic context: [Conditions and modifiers](../spec/control-flow.md#comparisons).

| Field | Value |
| --- | --- |
| root | <code>else</code> |
| pos | <code>evaluation</code> |
| type | <code>exceptional</code> |
| postfixes | <code>{}</code> |
| value | <code>ELSE(inputs)</code> |
| inputs | <code>["*.thing"]</code> |
| outputs | <code>["value"]</code> |

### alternative

<a id="vocabulary-evaluations-alternative"></a>

These tokens participate in configured grammatical conditions/modifiers. In particular, joined conditions do not yet have a native outer instruction operator; configured presence is not a general Boolean API.

Semantic context: [Conditions and modifiers](../spec/control-flow.md#comparisons).

| Field | Value |
| --- | --- |
| root | <code>otherwise</code> |
| pos | <code>evaluation</code> |
| type | <code>alternative</code> |
| postfixes | <code>{}</code> |
| value | <code>OTHERWISE(inputs)</code> |
| inputs | <code>["*.thing"]</code> |
| outputs | <code>["value"]</code> |

### iterative

<a id="vocabulary-evaluations-iterative"></a>

While is recognized but explicitly rejected by outer lowering.

Semantic context: [Iteration](../spec/control-flow.md#iteration-and-collection-operations).

| Field | Value |
| --- | --- |
| root | <code>while</code> |
| pos | <code>evaluation</code> |
| type | <code>iterative</code> |
| postfixes | <code>{}</code> |
| value | <code>WHILE(inputs)</code> |
| inputs | <code>["*.thing"]</code> |
| outputs | <code>["value"]</code> |
