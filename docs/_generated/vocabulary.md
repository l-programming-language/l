# Vocabulary inventory

The configured lexicon supplies roots, inflection transformations, grammatical categories and handler descriptions. These tables are exact configuration data, not a claim that every handler has native instruction support. See [grammar and vocabulary](../spec/grammar.md) and the linked semantic subjects. Source-defined words, including builtins and dictionary exports, are documented separately.

## nouns

**Word forms**

| Word | root | pos | type | postfixes |
| --- | --- | --- | --- | --- |
| <a id="it"></a><a id="vocabulary-nouns-it"></a>it | <code>it</code> | <code>noun</code> | <code>pronoun.definitive</code> | <code>{"---they":"plural"}</code> |
| <a id="that"></a><a id="vocabulary-nouns-that"></a>that | <code>that</code> | <code>noun</code> | <code>pronoun.objective</code> | <code>{"---ose":"plural"}</code> |
| <a id="this"></a><a id="vocabulary-nouns-this"></a>this | <code>this</code> | <code>noun</code> | <code>pronoun.subjective</code> | <code>{"---ese":"plural"}</code> |
| <a id="input"></a><a id="vocabulary-nouns-input"></a>input | <code>input</code> | <code>noun</code> | <code>functional.inflective</code> | <code>{"-s":"plural"}</code> |
| <a id="output"></a><a id="vocabulary-nouns-output"></a>output | <code>output</code> | <code>noun</code> | <code>functional.deflective</code> | <code>{"-s":"plural"}</code> |
| <a id="error"></a><a id="vocabulary-nouns-error"></a>error | <code>error</code> | <code>noun</code> | <code>functional.error</code> | <code>{"-s":"plural"}</code> |
| <a id="list"></a><a id="vocabulary-nouns-list"></a>list | <code>list</code> | <code>noun</code> | <code>iterative.list</code> | <code>{"-s":"plural"}</code> |
| <a id="dictionary"></a><a id="vocabulary-nouns-dictionary"></a>dictionary | <code>dictionary</code> | <code>noun</code> | <code>iterative.dictionary</code> | <code>{"--ies":"plural"}</code> |
| <a id="null"></a><a id="vocabulary-nouns-null"></a>null | <code>null</code> | <code>noun</code> | <code>primitive.null</code> | <code>{"-s":"plural"}</code> |
| <a id="thing"></a><a id="vocabulary-nouns-thing"></a>thing | <code>thing</code> | <code>noun</code> | <code>primitive.thing</code> | <code>{"-s":"plural"}</code> |
| <a id="text"></a><a id="vocabulary-nouns-text"></a>text | <code>text</code> | <code>noun</code> | <code>primitive.text</code> | <code>{"-s":"plural"}</code> |
| <a id="number"></a><a id="vocabulary-nouns-number"></a>number | <code>number</code> | <code>noun</code> | <code>primitive.number</code> | <code>{"-s":"plural"}</code> |
| <a id="byte"></a><a id="vocabulary-nouns-byte"></a>byte | <code>byte</code> | <code>noun</code> | <code>primitive.byte</code> | <code>{"-s":"plural"}</code> |
| <a id="truth"></a><a id="vocabulary-nouns-truth"></a>truth | <code>truth</code> | <code>noun</code> | <code>primitive.truth</code> | <code>{"-s":"plural"}</code> |
| <a id="pointer"></a><a id="vocabulary-nouns-pointer"></a>pointer | <code>pointer</code> | <code>noun</code> | <code>primitive.pointer</code> | <code>{"-s":"plural"}</code> |
| <a id="op"></a><a id="vocabulary-nouns-op"></a>op | <code>op</code> | <code>noun</code> | <code>applicative.operation</code> | <code>{"-s":"plural"}</code> |
| <a id="prop"></a><a id="vocabulary-nouns-prop"></a>prop | <code>prop</code> | <code>noun</code> | <code>applicative.property</code> | <code>{"-s":"plural"}</code> |
| <a id="func"></a><a id="vocabulary-nouns-func"></a>func | <code>func</code> | <code>noun</code> | <code>applicative.function</code> | <code>{"-s":"plural"}</code> |
| <a id="definition"></a><a id="vocabulary-nouns-definition"></a>definition | <code>definition</code> | <code>noun</code> | <code>applicative.definition</code> | <code>{"-s":"plural"}</code> |
| <a id="attribute"></a><a id="vocabulary-nouns-attribute"></a>attribute | <code>attribute</code> | <code>noun</code> | <code>applicative.attribute</code> | <code>{"-s":"plural"}</code> |
| <a id="object"></a><a id="vocabulary-nouns-object"></a>object | <code>object</code> | <code>noun</code> | <code>applicative.object</code> | <code>{"-s":"plural"}</code> |
| <a id="class"></a><a id="vocabulary-nouns-class"></a>class | <code>class</code> | <code>noun</code> | <code>applicative.class</code> | <code>{"-es":"plural"}</code> |
| <a id="service"></a><a id="vocabulary-nouns-service"></a>service | <code>service</code> | <code>noun</code> | <code>applicative.service</code> | <code>{"-s":"plural"}</code> |
| <a id="app"></a><a id="vocabulary-nouns-app"></a>app | <code>app</code> | <code>noun</code> | <code>applicative.app</code> | <code>{"-s":"plural"}</code> |

**Handlers and support**

| Word | value | inputs | outputs | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| [it](#vocabulary-nouns-it) | <code>IT(inputs[0])</code> | <code>["*.noun"]</code> | <code>["value"]</code> | partial | [Pronouns](../spec/names-and-scope.md#pronouns) | [Note 1](#vocabulary-note-1) |
| [that](#vocabulary-nouns-that) | <code>THAT(inputs[0])</code> | <code>["*.noun"]</code> | <code>["value"]</code> | partial | [Pronouns](../spec/names-and-scope.md#pronouns) | [Note 1](#vocabulary-note-1) |
| [this](#vocabulary-nouns-this) | <code>THIS(inputs[0])</code> | <code>["*.noun"]</code> | <code>["value"]</code> | partial | [Pronouns](../spec/names-and-scope.md#pronouns) | [Note 1](#vocabulary-note-1) |
| [input](#vocabulary-nouns-input) | <code>INPUT(inputs[0])</code> | <code>["*.thing"]</code> | <code>["value"]</code> | partial | [Input metadata and runtime boundary](../reference/builtins/index.md#runtime-members) | [Note 2](#vocabulary-note-2) |
| [output](#vocabulary-nouns-output) | <code>OUTPUT(inputs[0])</code> | <code>["*.thing"]</code> | <code>["value"]</code> | partial | [Process members](../spec/execution.md#output-and-error) | [Note 3](#vocabulary-note-3) |
| [error](#vocabulary-nouns-error) | <code>ERROR(inputs[0])</code> | <code>["*.thing"]</code> | <code>["value"]</code> | partial | [Process members](../spec/execution.md#output-and-error) | [Note 3](#vocabulary-note-3) |
| [list](#vocabulary-nouns-list) | <code>LIST(inputs[0])</code> | <code>["*.thing"]</code> | <code>["value"]</code> | partial | [list](../spec/collections-and-categories.md#lists) | [Note 4](#vocabulary-note-4) |
| [dictionary](#vocabulary-nouns-dictionary) | <code>DICTIONARY(inputs[0], inputs[1])</code> | <code>["text","*.thing"]</code> | <code>["value"]</code> | partial | [dictionary](../spec/collections-and-categories.md#dictionaries) | [Note 5](#vocabulary-note-5) |
| [null](#vocabulary-nouns-null) | <code>NULL(inputs[0])</code> | <code>["*.thing"]</code> | <code>["value"]</code> | supported | [null](../spec/values.md#null-and-internal-states) | [Note 6](#vocabulary-note-6) |
| [thing](#vocabulary-nouns-thing) | <code>THING(inputs[0])</code> | <code>["*.thing"]</code> | <code>["value"]</code> | partial | [thing](../spec/values.md#thing) | [Note 7](#vocabulary-note-7) |
| [text](#vocabulary-nouns-text) | <code>TEXT(inputs[0])</code> | <code>["*.text"]</code> | <code>["value"]</code> | partial | [text](../spec/values.md#text) | [Note 8](#vocabulary-note-8) |
| [number](#vocabulary-nouns-number) | <code>NUMBER(inputs[0])</code> | <code>["*.number"]</code> | <code>["value"]</code> | partial | [number](../spec/values.md#number) | [Note 9](#vocabulary-note-9) |
| [byte](#vocabulary-nouns-byte) | <code>BYTES(inputs[0])</code> | <code>["*.byte"]</code> | <code>["value"]</code> | partial | [byte](../spec/values.md#byte-and-bytes) | [Note 10](#vocabulary-note-10) |
| [truth](#vocabulary-nouns-truth) | <code>TRUTH(inputs[0])</code> | <code>["*.truth"]</code> | <code>["value"]</code> | partial | [truth](../spec/values.md#truth) | [Note 11](#vocabulary-note-11) |
| [pointer](#vocabulary-nouns-pointer) | <code>POINTER(inputs[0])</code> | <code>["*.thing"]</code> | <code>["value"]</code> | supported | [pointer](../spec/values.md#pointer) | [Note 12](#vocabulary-note-12) |
| [op](#vocabulary-nouns-op) | <code>OP(inputs[0], inputs[1], inputs[2], inputs[3])</code> | <code>["text","numbers","text","text"]</code> | <code>["value"]</code> | partial | [op](../spec/applicational-types.md#functions-properties-and-operations) | [Note 5](#vocabulary-note-5) |
| [prop](#vocabulary-nouns-prop) | <code>PROP(inputs[0], inputs[1], inputs[2])</code> | <code>["text","text","thing"]</code> | <code>["value"]</code> | partial | [prop](../spec/applicational-types.md#functions-properties-and-operations) | [Note 5](#vocabulary-note-5) |
| [func](#vocabulary-nouns-func) | <code>FUNC(inputs[0], inputs[1], inputs[2], inputs[3])</code> | <code>["text","props","texts","prop"]</code> | <code>["value"]</code> | partial | [func](../spec/applicational-types.md#functions-properties-and-operations) | [Note 13](#vocabulary-note-13) |
| [definition](#vocabulary-nouns-definition) | <code>DEF(inputs[0])</code> | <code>["text"]</code> | <code>["value"]</code> | partial | [definition](../spec/applicational-types.md#attributes-and-definitions) | [Note 5](#vocabulary-note-5) |
| [attribute](#vocabulary-nouns-attribute) | <code>ATTRIBUTE(inputs[0], inputs[1])</code> | <code>["definition","attributes"]</code> | <code>["value"]</code> | partial | [attribute](../spec/applicational-types.md#attributes-and-definitions) | [Note 5](#vocabulary-note-5) |
| [object](#vocabulary-nouns-object) | <code>OBJECT(inputs[0], inputs[1])</code> | <code>["definition","attributes"]</code> | <code>["value"]</code> | partial | [object](../spec/applicational-types.md#classes-and-objects) | [Note 5](#vocabulary-note-5) |
| [class](#vocabulary-nouns-class) | <code>CLASS(inputs[0], inputs[1])</code> | <code>["objects","classes"]</code> | <code>["value"]</code> | partial | [class](../spec/applicational-types.md#classes-and-objects) | [Note 5](#vocabulary-note-5) |
| [service](#vocabulary-nouns-service) | <code>SERVICE(inputs[0], inputs[1], inputs[2])</code> | <code>["objects","classes","services"]</code> | <code>["value"]</code> | partial | [service](../spec/applicational-types.md#services) | [Note 5](#vocabulary-note-5) |
| [app](#vocabulary-nouns-app) | <code>APP(inputs[0], inputs[1], inputs[2])</code> | <code>["objects","services","dictionary"]</code> | <code>["value"]</code> | partial | [app](../spec/applicational-types.md#apps) | [Note 5](#vocabulary-note-5) |

## verbs

**Word forms**

| Word | root | pos | type | postfixes |
| --- | --- | --- | --- | --- |
| <a id="passive"></a><a id="vocabulary-verbs-passive"></a>passive | <code>is</code> | <code>verb</code> | <code>action.passive</code> | <code>{"-nt":"null"}</code> |
| <a id="associative"></a><a id="vocabulary-verbs-associative"></a>associative | <code>use</code> | <code>verb</code> | <code>action.associative</code> | <code>{"-s":"active","-d":"passive","--ing":"applicative","-r":"objective"}</code> |
| <a id="declarative"></a><a id="vocabulary-verbs-declarative"></a>declarative | <code>create</code> | <code>verb</code> | <code>action.declarative</code> | <code>{"-s":"active","-d":"passive","--ing":"applicative","--or":"objective"}</code> |
| <a id="definitive"></a><a id="vocabulary-verbs-definitive"></a>definitive | <code>define</code> | <code>verb</code> | <code>action.definitive</code> | <code>{"-s":"active","-d":"passive","--ing":"applicative","-r":"objective"}</code> |
| <a id="descriptive"></a><a id="vocabulary-verbs-descriptive"></a>descriptive | <code>describe</code> | <code>verb</code> | <code>action.descriptive</code> | <code>{"-s":"active","-d":"passive","--ing":"applicative","-r":"objective"}</code> |
| <a id="expansive"></a><a id="vocabulary-verbs-expansive"></a>expansive | <code>put</code> | <code>verb</code> | <code>action.expansive</code> | <code>{"-s":"active","-":"passive","-ting":"applicative","-er":"objective"}</code> |
| <a id="locative"></a><a id="vocabulary-verbs-locative"></a>locative | <code>place</code> | <code>verb</code> | <code>action.locative</code> | <code>{"-s":"active","-d":"passive","--ing":"applicative","-r":"objective"}</code> |
| <a id="receptive"></a><a id="vocabulary-verbs-receptive"></a>receptive | <code>get</code> | <code>verb</code> | <code>action.receptive</code> | <code>{"-s":"active","---otten":"passive","-ting":"applicative","-ter":"objective"}</code> |
| <a id="reductive"></a><a id="vocabulary-verbs-reductive"></a>reductive | <code>take</code> | <code>verb</code> | <code>action.reductive</code> | <code>{"-s":"active","-n":"passive","--ing":"applicative","-r":"objective"}</code> |
| <a id="generative"></a><a id="vocabulary-verbs-generative"></a>generative | <code>build</code> | <code>verb</code> | <code>action.generative</code> | <code>{"-s":"active","--t":"passive","-ing":"applicative","-er":"objective"}</code> |
| <a id="executive"></a><a id="vocabulary-verbs-executive"></a>executive | <code>run</code> | <code>verb</code> | <code>action.executive</code> | <code>{"-s":"active","--an":"passive","-ning":"applicative","-ner":"objective"}</code> |
| <a id="suspensive"></a><a id="vocabulary-verbs-suspensive"></a>suspensive | <code>await</code> | <code>verb</code> | <code>action.suspensive</code> | <code>{"-s":"active","-ed":"passive","-ing":"applicative"}</code> |
| <a id="interruptive"></a><a id="vocabulary-verbs-interruptive"></a>interruptive | <code>stop</code> | <code>verb</code> | <code>action.interruptive</code> | <code>{"-s":"active","-ped":"passive","-ping":"applicative","-pper":"objective"}</code> |
| <a id="inflective"></a><a id="vocabulary-verbs-inflective"></a>inflective | <code>import</code> | <code>verb</code> | <code>action.inflective</code> | <code>{"-s":"active","-ed":"passive","-ing":"applicative","-er":"objective"}</code> |
| <a id="deflective"></a><a id="vocabulary-verbs-deflective"></a>deflective | <code>export</code> | <code>verb</code> | <code>action.deflective</code> | <code>{"-s":"active","-ed":"passive","-ing":"applicative","-er":"objective"}</code> |
| <a id="observative"></a><a id="vocabulary-verbs-observative"></a>observative | <code>show</code> | <code>verb</code> | <code>action.observative</code> | <code>{"-s":"active","-ed":"passive","-ing":"applicative","-er":"objective"}</code> |
| <a id="terminative"></a><a id="vocabulary-verbs-terminative"></a>terminative | <code>exit</code> | <code>verb</code> | <code>action.terminative</code> | <code>{"-s":"active","-ed":"passive","-ing":"applicative","-er":"objective"}</code> |
| <a id="appropriative"></a><a id="vocabulary-verbs-appropriative"></a>appropriative | <code>have</code> | <code>verb</code> | <code>operation.appropriative</code> | <code>{"---s":"active","---d":"passive","--ing":"applicative","-r":"objective"}</code> |
| <a id="comparative"></a><a id="vocabulary-verbs-comparative"></a>comparative | <code>equal</code> | <code>verb</code> | <code>operation.comparative</code> | <code>{"-s":"active","-ed":"passive","-ing":"applicative","-er":"objective"}</code> |

**Handlers and support**

| Word | value | inputs | outputs | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| [passive](#vocabulary-verbs-passive) | <code>IS(inputs[0])</code> | <code>["*.thing"]</code> | <code>["value"]</code> | partial | [Assignment](../spec/control-flow.md#assignment) | [Note 14](#vocabulary-note-14) |
| [associative](#vocabulary-verbs-associative) | <code>USE(inputs[0])</code> | <code>["*.noun"]</code> | <code>["value"]</code> | partial | [Aliases](../spec/names-and-scope.md#aliases-and-canonical-identity) | [Note 15](#vocabulary-note-15) |
| [declarative](#vocabulary-verbs-declarative) | <code>DECLARE(inputs[0])</code> | <code>["*.noun"]</code> | <code>["value"]</code> | partial | [Declarations](../spec/definitions.md#create-define-and-describe) | [Note 16](#vocabulary-note-16) |
| [definitive](#vocabulary-verbs-definitive) | <code>DEFINE(inputs[0])</code> | <code>["*.noun"]</code> | <code>["value"]</code> | partial | [Definitions](../spec/definitions.md#create-define-and-describe) | [Note 17](#vocabulary-note-17) |
| [descriptive](#vocabulary-verbs-descriptive) | <code>DESCRIBE(inputs[0])</code> | <code>["*.noun"]</code> | <code>["value"]</code> | partial | [Descriptors](../spec/definitions.md#descriptors-and-constructors) | [Note 18](#vocabulary-note-18) |
| [expansive](#vocabulary-verbs-expansive) | <code>PUT(inputs[0], inputs[1], inputs[2])</code> | <code>["*.thing","*.noun.iterative","text"]</code> | <code>["value"]</code> | partial | [Collection operations](../spec/control-flow.md#iteration-and-collection-operations) | [Note 19](#vocabulary-note-19) |
| [locative](#vocabulary-verbs-locative) | <code>PLACE(inputs[0], inputs[1], inputs[2])</code> | <code>["*.thing","*.noun.iterative","text"]</code> | <code>["value"]</code> | partial | [Collection operations](../spec/control-flow.md#iteration-and-collection-operations) | [Note 20](#vocabulary-note-20) |
| [receptive](#vocabulary-verbs-receptive) | <code>GET(inputs[0], inputs[1])</code> | <code>["text","*.noun.iterative"]</code> | <code>["value"]</code> | partial | [Collection operations](../spec/control-flow.md#iteration-and-collection-operations) | [Note 21](#vocabulary-note-21) |
| [reductive](#vocabulary-verbs-reductive) | <code>TAKE(inputs[0], inputs[1])</code> | <code>["text","*.noun.iterative"]</code> | <code>["value"]</code> | partial | [Collection operations](../spec/control-flow.md#iteration-and-collection-operations) | [Note 21](#vocabulary-note-21) |
| [generative](#vocabulary-verbs-generative) | <code>BUILD(inputs[0])</code> | <code>["*.noun.app"]</code> | <code>["value"]</code> | partial | [Build](../spec/execution.md#build-run-stop-and-await) | [Note 22](#vocabulary-note-22) |
| [executive](#vocabulary-verbs-executive) | <code>RUN(inputs[0])</code> | <code>["*.noun.app"]</code> | <code>["value"]</code> | partial | [Run](../spec/execution.md#build-run-stop-and-await) | [Note 23](#vocabulary-note-23) |
| [suspensive](#vocabulary-verbs-suspensive) | <code>AWAIT(inputs[0])</code> | <code>["*.noun.app"]</code> | <code>["value"]</code> | partial | [Await](../spec/execution.md#continuations-and-activation) | [Note 24](#vocabulary-note-24) |
| [interruptive](#vocabulary-verbs-interruptive) | <code>STOP(inputs[0])</code> | <code>["*.noun.app"]</code> | <code>["value"]</code> | partial | [Stop](../spec/execution.md#build-run-stop-and-await) | [Note 25](#vocabulary-note-25) |
| [inflective](#vocabulary-verbs-inflective) | <code>IMPORT(inputs[0])</code> | <code>["*.thing"]</code> | <code>["value"]</code> | partial | [Import/export distinction](../spec/imports.md#indexed-dictionaries) | [Note 26](#vocabulary-note-26) |
| [deflective](#vocabulary-verbs-deflective) | <code>EXPORT(inputs[0])</code> | <code>["*.thing"]</code> | <code>["value"]</code> | partial | [Import/export distinction](../spec/imports.md#indexed-dictionaries) | [Note 26](#vocabulary-note-26) |
| [observative](#vocabulary-verbs-observative) | <code>SHOW(inputs[0])</code> | <code>["*.thing"]</code> | <code>["value"]</code> | partial | [Show](../spec/control-flow.md#show-and-output) | [Note 27](#vocabulary-note-27) |
| [terminative](#vocabulary-verbs-terminative) | <code>EXIT(inputs[0])</code> | <code>["*.thing"]</code> | <code>["value"]</code> | partial | [Exit](../spec/control-flow.md#exit-and-return) | [Note 28](#vocabulary-note-28) |
| [appropriative](#vocabulary-verbs-appropriative) | <code>HAVE(inputs[0])</code> | <code>["*.thing"]</code> | <code>["value"]</code> | partial | [Comparisons](../spec/control-flow.md#comparisons) | [Note 29](#vocabulary-note-29) |
| [comparative](#vocabulary-verbs-comparative) | <code>EQUAL(inputs[0])</code> | <code>["*.thing"]</code> | <code>["value"]</code> | partial | [Comparisons](../spec/control-flow.md#comparisons) | [Note 29](#vocabulary-note-29) |

## adjectives

**Word forms**

| Word | root | pos | type | postfixes |
| --- | --- | --- | --- | --- |
| <a id="visibilitypublic"></a><a id="vocabulary-adjectives-visibility.public"></a>visibility.public | <code>public</code> | <code>adjective</code> | <code>visibility.public</code> | <code>{}</code> |
| <a id="initialization"></a><a id="vocabulary-adjectives-initialization"></a>initialization | <code>new</code> | <code>adjective</code> | <code>initialization</code> | <code>{}</code> |
| <a id="abstraction"></a><a id="vocabulary-adjectives-abstraction"></a>abstraction | <code>generic</code> | <code>adjective</code> | <code>abstraction</code> | <code>{}</code> |

**Handlers and support**

| Word | value | inputs | outputs | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| [visibility.public](#vocabulary-adjectives-visibility.public) | — | — | — | partial | [Visibility](../spec/names-and-scope.md#paragraph-scope) | [Note 30](#vocabulary-note-30) |
| [initialization](#vocabulary-adjectives-initialization) | <code>NEW(inputs[0])</code> | <code>["*.noun"]</code> | <code>["value"]</code> | partial | [Descriptors](../spec/definitions.md#descriptors-and-constructors) | [Note 31](#vocabulary-note-31) |
| [abstraction](#vocabulary-adjectives-abstraction) | <code>GENERIC(inputs[0])</code> | <code>["*.noun"]</code> | <code>["value"]</code> | partial | [Descriptors](../spec/definitions.md#descriptors-and-constructors) | [Note 31](#vocabulary-note-31) |

## adverbs

**Word forms**

| Word | root | pos | type | postfixes |
| --- | --- | --- | --- | --- |
| <a id="visibilitypublic-1"></a><a id="vocabulary-adverbs-visibility.public"></a>visibility.public | <code>publicly</code> | <code>adverb</code> | <code>visibility.public</code> | <code>{}</code> |
| <a id="nullification"></a><a id="vocabulary-adverbs-nullification"></a>nullification | <code>not</code> | <code>adverb</code> | <code>nullification</code> | <code>{}</code> |
| <a id="comparisonsubtraction"></a><a id="vocabulary-adverbs-comparison.subtraction"></a>comparison.subtraction | <code>less</code> | <code>adverb</code> | <code>comparison.subtraction</code> | <code>{}</code> |
| <a id="comparisonmultiplication"></a><a id="vocabulary-adverbs-comparison.multiplication"></a>comparison.multiplication | <code>greater</code> | <code>adverb</code> | <code>comparison.multiplication</code> | <code>{}</code> |
| <a id="progression"></a><a id="vocabulary-adverbs-progression"></a>progression | <code>then</code> | <code>adverb</code> | <code>abstraction</code> | <code>{}</code> |

**Handlers and support**

| Word | value | inputs | outputs | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| [visibility.public](#vocabulary-adverbs-visibility.public) | — | — | — | partial | [Visibility](../spec/names-and-scope.md#paragraph-scope) | [Note 30](#vocabulary-note-30) |
| [nullification](#vocabulary-adverbs-nullification) | <code>NOT(inputs)</code> | <code>["*.thing"]</code> | <code>["value"]</code> | partial | [Conditions and modifiers](../spec/control-flow.md#comparisons) | [Note 32](#vocabulary-note-32) |
| [comparison.subtraction](#vocabulary-adverbs-comparison.subtraction) | <code>LESSER(inputs)</code> | <code>["*.thing"]</code> | <code>["value"]</code> | partial | [Conditions and modifiers](../spec/control-flow.md#comparisons) | [Note 32](#vocabulary-note-32) |
| [comparison.multiplication](#vocabulary-adverbs-comparison.multiplication) | <code>GREATER(inputs)</code> | <code>["*.thing"]</code> | <code>["value"]</code> | partial | [Conditions and modifiers](../spec/control-flow.md#comparisons) | [Note 32](#vocabulary-note-32) |
| [progression](#vocabulary-adverbs-progression) | <code>THEN(inputs)</code> | <code>["*.thing"]</code> | <code>["value"]</code> | partial | [Conditions and modifiers](../spec/control-flow.md#comparisons) | [Note 32](#vocabulary-note-32) |

## prepositions

**Word forms**

| Word | root | pos | type | postfixes |
| --- | --- | --- | --- | --- |
| <a id="definementdeclarative"></a><a id="vocabulary-prepositions-definement.declarative"></a>definement.declarative | <code>as</code> | <code>preposition</code> | <code>definement.declarative</code> | <code>{}</code> |
| <a id="definementdefinitive"></a><a id="vocabulary-prepositions-definement.definitive"></a>definement.definitive | <code>of</code> | <code>preposition</code> | <code>definement.definitive</code> | <code>{}</code> |
| <a id="definementlocative"></a><a id="vocabulary-prepositions-definement.locative"></a>definement.locative | <code>at</code> | <code>preposition</code> | <code>definement.locative</code> | <code>{}</code> |
| <a id="consignmentassociative"></a><a id="vocabulary-prepositions-consignment.associative"></a>consignment.associative | <code>with</code> | <code>preposition</code> | <code>consignment.associative</code> | <code>{}</code> |
| <a id="consignmentinflective"></a><a id="vocabulary-prepositions-consignment.inflective"></a>consignment.inflective | <code>in</code> | <code>preposition</code> | <code>consignment.inflective</code> | <code>{}</code> |
| <a id="consignmentdeflective"></a><a id="vocabulary-prepositions-consignment.deflective"></a>consignment.deflective | <code>out</code> | <code>preposition</code> | <code>consignment.deflective</code> | <code>{}</code> |
| <a id="consignmentindicative"></a><a id="vocabulary-prepositions-consignment.indicative"></a>consignment.indicative | <code>by</code> | <code>preposition</code> | <code>consignment.indicative</code> | <code>{}</code> |
| <a id="conjoinmentobjective"></a><a id="vocabulary-prepositions-conjoinment.objective"></a>conjoinment.objective | <code>from</code> | <code>preposition</code> | <code>conjoinment.objective</code> | <code>{}</code> |
| <a id="conjoinmentsubjective"></a><a id="vocabulary-prepositions-conjoinment.subjective"></a>conjoinment.subjective | <code>to</code> | <code>preposition</code> | <code>conjoinment.subjective</code> | <code>{}</code> |
| <a id="conjoinmentcomparative"></a><a id="vocabulary-prepositions-conjoinment.comparative"></a>conjoinment.comparative | <code>than</code> | <code>preposition</code> | <code>conjoinment.comparative</code> | <code>{}</code> |
| <a id="conjoinmentdirective"></a><a id="vocabulary-prepositions-conjoinment.directive"></a>conjoinment.directive | <code>for</code> | <code>preposition</code> | <code>conjoinment.directive</code> | <code>{}</code> |

**Handlers and support**

| Word | value | inputs | outputs | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| [definement.declarative](#vocabulary-prepositions-definement.declarative) | <code>AS(inputs)</code> | <code>["*.thing"]</code> | <code>["value"]</code> | partial | [Prepositions](../spec/grammar.md#phrases) | [Note 33](#vocabulary-note-33) |
| [definement.definitive](#vocabulary-prepositions-definement.definitive) | <code>OF(inputs)</code> | <code>["*.thing"]</code> | <code>["value"]</code> | partial | [Prepositions](../spec/grammar.md#phrases) | [Note 33](#vocabulary-note-33) |
| [definement.locative](#vocabulary-prepositions-definement.locative) | <code>AT(inputs)</code> | <code>["*.thing"]</code> | <code>["value"]</code> | partial | [Prepositions](../spec/grammar.md#phrases) | [Note 34](#vocabulary-note-34) |
| [consignment.associative](#vocabulary-prepositions-consignment.associative) | <code>WITH(inputs)</code> | <code>["*.thing"]</code> | <code>["value"]</code> | partial | [Prepositions](../spec/grammar.md#phrases) | [Note 33](#vocabulary-note-33) |
| [consignment.inflective](#vocabulary-prepositions-consignment.inflective) | <code>IN(inputs)</code> | <code>["*.thing"]</code> | <code>["value"]</code> | partial | [Prepositions](../spec/grammar.md#phrases) | [Note 35](#vocabulary-note-35) |
| [consignment.deflective](#vocabulary-prepositions-consignment.deflective) | <code>OUT(inputs)</code> | <code>["*.thing"]</code> | <code>["value"]</code> | partial | [Prepositions](../spec/grammar.md#phrases) | [Note 33](#vocabulary-note-33) |
| [consignment.indicative](#vocabulary-prepositions-consignment.indicative) | <code>BY(inputs)</code> | <code>["*.thing"]</code> | <code>["value"]</code> | partial | [Prepositions](../spec/grammar.md#phrases) | [Note 33](#vocabulary-note-33) |
| [conjoinment.objective](#vocabulary-prepositions-conjoinment.objective) | <code>FROM(inputs)</code> | <code>["*.thing"]</code> | <code>["value"]</code> | partial | [Prepositions](../spec/grammar.md#phrases) | [Note 33](#vocabulary-note-33) |
| [conjoinment.subjective](#vocabulary-prepositions-conjoinment.subjective) | <code>TO(inputs)</code> | <code>["*.thing"]</code> | <code>["value"]</code> | partial | [Prepositions](../spec/grammar.md#phrases) | [Note 33](#vocabulary-note-33) |
| [conjoinment.comparative](#vocabulary-prepositions-conjoinment.comparative) | <code>THAN(inputs)</code> | <code>["*.thing"]</code> | <code>["value"]</code> | partial | [Prepositions](../spec/grammar.md#phrases) | [Note 33](#vocabulary-note-33) |
| [conjoinment.directive](#vocabulary-prepositions-conjoinment.directive) | <code>FOR(inputs)</code> | <code>["*.thing"]</code> | <code>["value"]</code> | partial | [Prepositions](../spec/grammar.md#phrases) | [Note 33](#vocabulary-note-33) |

## articles

**Word forms**

| Word | root | pos | type | postfixes |
| --- | --- | --- | --- | --- |
| <a id="variable"></a><a id="vocabulary-articles-variable"></a>variable | <code>any</code> | <code>article</code> | <code>variable</code> | <code>{}</code> |
| <a id="none"></a><a id="vocabulary-articles-none"></a>none | <code>no</code> | <code>article</code> | <code>none</code> | <code>{}</code> |
| <a id="indefinite"></a><a id="vocabulary-articles-indefinite"></a>indefinite | <code>a</code> | <code>article</code> | <code>indefinite</code> | <code>{"-n":"alt"}</code> |
| <a id="definite"></a><a id="vocabulary-articles-definite"></a>definite | <code>the</code> | <code>article</code> | <code>definite</code> | <code>{}</code> |

**Handlers and support**

| Word | value | inputs | outputs | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| [variable](#vocabulary-articles-variable) | <code>ANY(inputs)</code> | <code>["*.thing"]</code> | <code>["value"]</code> | partial | [Articles and typed arguments](../spec/definitions.md#typed-slots) | [Note 36](#vocabulary-note-36) |
| [none](#vocabulary-articles-none) | <code>NO(inputs)</code> | <code>["*.thing"]</code> | <code>["value"]</code> | partial | [Articles and typed arguments](../spec/definitions.md#typed-slots) | [Note 36](#vocabulary-note-36) |
| [indefinite](#vocabulary-articles-indefinite) | <code>A(inputs)</code> | <code>["*.noun"]</code> | <code>["value"]</code> | partial | [Articles and typed arguments](../spec/definitions.md#typed-slots) | [Note 36](#vocabulary-note-36) |
| [definite](#vocabulary-articles-definite) | <code>THE(inputs)</code> | <code>["*.noun"]</code> | <code>["value"]</code> | partial | [Articles and typed arguments](../spec/definitions.md#typed-slots) | [Note 36](#vocabulary-note-36) |

## conjunctions

**Word forms**

| Word | root | pos | type | postfixes |
| --- | --- | --- | --- | --- |
| <a id="mutual"></a><a id="vocabulary-conjunctions-mutual"></a>mutual | <code>and</code> | <code>conjunction</code> | <code>mutual</code> | <code>{}</code> |
| <a id="exclusive"></a><a id="vocabulary-conjunctions-exclusive"></a>exclusive | <code>or</code> | <code>conjunction</code> | <code>exclusive</code> | <code>{}</code> |

**Handlers and support**

| Word | value | inputs | outputs | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| [mutual](#vocabulary-conjunctions-mutual) | <code>AND(inputs)</code> | <code>["*.thing"]</code> | <code>["value"]</code> | partial | [Conditions and modifiers](../spec/control-flow.md#comparisons) | [Note 32](#vocabulary-note-32) |
| [exclusive](#vocabulary-conjunctions-exclusive) | <code>OR(inputs)</code> | <code>["*.thing"]</code> | <code>["value"]</code> | partial | [Conditions and modifiers](../spec/control-flow.md#comparisons) | [Note 32](#vocabulary-note-32) |

## evaluations

**Word forms**

| Word | root | pos | type | postfixes |
| --- | --- | --- | --- | --- |
| <a id="conditional"></a><a id="vocabulary-evaluations-conditional"></a>conditional | <code>if</code> | <code>evaluation</code> | <code>conditional</code> | <code>{}</code> |
| <a id="exceptional"></a><a id="vocabulary-evaluations-exceptional"></a>exceptional | <code>else</code> | <code>evaluation</code> | <code>exceptional</code> | <code>{}</code> |
| <a id="alternative"></a><a id="vocabulary-evaluations-alternative"></a>alternative | <code>otherwise</code> | <code>evaluation</code> | <code>alternative</code> | <code>{}</code> |
| <a id="iterative"></a><a id="vocabulary-evaluations-iterative"></a>iterative | <code>while</code> | <code>evaluation</code> | <code>iterative</code> | <code>{}</code> |

**Handlers and support**

| Word | value | inputs | outputs | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| [conditional](#vocabulary-evaluations-conditional) | <code>IF(inputs)</code> | <code>["*.thing"]</code> | <code>["value"]</code> | partial | [Conditions and modifiers](../spec/control-flow.md#comparisons) | [Note 32](#vocabulary-note-32) |
| [exceptional](#vocabulary-evaluations-exceptional) | <code>ELSE(inputs)</code> | <code>["*.thing"]</code> | <code>["value"]</code> | partial | [Conditions and modifiers](../spec/control-flow.md#comparisons) | [Note 32](#vocabulary-note-32) |
| [alternative](#vocabulary-evaluations-alternative) | <code>OTHERWISE(inputs)</code> | <code>["*.thing"]</code> | <code>["value"]</code> | partial | [Conditions and modifiers](../spec/control-flow.md#comparisons) | [Note 32](#vocabulary-note-32) |
| [iterative](#vocabulary-evaluations-iterative) | <code>WHILE(inputs)</code> | <code>["*.thing"]</code> | <code>["value"]</code> | rejected | [Iteration](../spec/control-flow.md#iteration-and-collection-operations) | [Note 37](#vocabulary-note-37) |

## Support notes

| Note | Reviewed explanation |
| --- | --- |
| <a id="vocabulary-note-1"></a>1 | Pronouns bind through source paragraph antecedents, not the last dynamically evaluated value. Separate generic-action argument rules also apply. |
| <a id="vocabulary-note-2"></a>2 | Inputs name ordinary Function parameters or native signature metadata. They do not introduce process stdin/argument transport. |
| <a id="vocabulary-note-3"></a>3 | Runtime member access requires a compiler-resolved process target. Output is finalized raw stdout; Error has its own lifecycle and diagnostic boundaries. |
| <a id="vocabulary-note-4"></a>4 | Typed owned Lists construct, append, replace and read indexed values; heterogeneous construction infers Thing without coercion. Indices and complete element types are checked. Iteration/removal remain outside this API. |
| <a id="vocabulary-note-5"></a>5 | This lexical category must be distinguished from similarly named compiler representations. Read the construction and native-lowering boundaries at the semantic target. |
| <a id="vocabulary-note-6"></a>6 | Explicit Null initializes an absence value and prints null. It is distinct from an unassigned slot and from internal None/Unknown states. |
| <a id="vocabulary-note-7"></a>7 | Explicit Thing boundaries retain concrete runtime types; typed reads validate full element, shape and callable identities. No implicit coercion or dynamic arithmetic. |
| <a id="vocabulary-note-8"></a>8 | The configured text root is a source spelling. Its storage and executable support are documented independently; the handler description below is not a callable API. |
| <a id="vocabulary-note-9"></a>9 | The configured number root is a source spelling. Its storage and executable support are documented independently; the handler description below is not a callable API. |
| <a id="vocabulary-note-10"></a>10 | The configured byte root is a source spelling. Its storage and executable support are documented independently; the handler description below is not a callable API. |
| <a id="vocabulary-note-11"></a>11 | The configured truth root is a source spelling. Its storage and executable support are documented independently; the handler description below is not a callable API. |
| <a id="vocabulary-note-12"></a>12 | Managed typed handles share targets; reads snapshot ordinary contents. Checked mutation rejects ownership cycles, including container and callable-environment edges. No pointer arithmetic or raw-address API. |
| <a id="vocabulary-note-13"></a>13 | Function declarations retain typed signatures and use the canonical builtin definition. Their invocation verbs participate in ordinary sentences. |
| <a id="vocabulary-note-14"></a>14 | Assignment follows resolved typed-storage rules and definite assignment. Runtime ordering is preserved independently of parser retry order. |
| <a id="vocabulary-note-15"></a>15 | Association names an existing canonical value or definition. It does not create an independent runtime copy. |
| <a id="vocabulary-note-16"></a>16 | A declaration associates a source name with the indicated category or value. Voice-specific recognizers do not imply interchangeable native behavior. |
| <a id="vocabulary-note-17"></a>17 | Definition clauses establish source vocabulary or typed native-backed forms. Accepted construction and framing are bounded by definition expansion. |
| <a id="vocabulary-note-18"></a>18 | Descriptive clauses apply source-defined adjective/adverb meanings. Constructor descriptors preserve the target and source-unit ownership. |
| <a id="vocabulary-note-19"></a>19 | Put appends to a List, stores Map values or replaces a Pointer target through checked runtime updates. It does not modify vocabulary Dictionaries. |
| <a id="vocabulary-note-20"></a>20 | Place replaces a List element at a checked Number index or a keyed runtime field. Required target/key/value types are checked. |
| <a id="vocabulary-note-21"></a>21 | The configured word is recognized vocabulary, but no native outer collection instruction is supplied for this operation. Do not treat it as a supported general collection API. |
| <a id="vocabulary-note-22"></a>22 | Build is the runtime readiness operation. It is distinct from the CLI command that compiles native output. |
| <a id="vocabulary-note-23"></a>23 | Run launches the selected process. The grammar does not imply that every applicative category is runnable. |
| <a id="vocabulary-note-24"></a>24 | Await joins a selected process invocation and activates its continuation once. A parsed clause alone does not establish valid activation placement. |
| <a id="vocabulary-note-25"></a>25 | Stop requests termination without joining. Final output still requires the Await contract. |
| <a id="vocabulary-note-26"></a>26 | Configured Import/Export handlers do not replace the supported Use import syntax and are absent from native outer instruction dispatch. |
| <a id="vocabulary-note-27"></a>27 | Show emits the resolved operand without Print’s added newline; Output reads are subject to process finalization. |
| <a id="vocabulary-note-28"></a>28 | Exit ends the compiled program. Function-local Return uses the selected Function.Return definition and checked outputs. |
| <a id="vocabulary-note-29"></a>29 | The configured comparison/appropriation root is used in bounded condition forms. Operand types and Error-member restrictions are enforced by native lowering. |
| <a id="vocabulary-note-30"></a>30 | Public/Publicly marks exactly the associated declaration. It does not publish unrelated owners or sibling declarations. |
| <a id="vocabulary-note-31"></a>31 | Constructor/descriptor vocabulary determines source construction roles. Its presence does not imply arbitrary type construction or inherited native metadata. |
| <a id="vocabulary-note-32"></a>32 | These tokens participate in configured grammatical conditions/modifiers. In particular, joined conditions do not yet have a native outer instruction operator; configured presence is not a general Boolean API. |
| <a id="vocabulary-note-33"></a>33 | This preposition establishes the specified grammatical relationship between source elements. Interpret its category through the configured phrase rules. |
| <a id="vocabulary-note-34"></a>34 | At selects a Number index or Text field key; its enclosing typed read/update determines the permitted key type. |
| <a id="vocabulary-note-35"></a>35 | In selects a runtime update destination; the following index or field key remains part of that update. |
| <a id="vocabulary-note-36"></a>36 | Articles qualify source bindings and typed argument roles. The any form in source actions is not an unrestricted Rust generic or implicit conversion. |
| <a id="vocabulary-note-37"></a>37 | While is recognized but explicitly rejected by outer lowering. |
