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
| [it](#vocabulary-nouns-it) | <code>IT(inputs[0])</code> | <code>["*.noun"]</code> | <code>["value"]</code> | supported | [Pronouns](../spec/names-and-scope.md#it) | [Note 1](#vocabulary-note-1) |
| [that](#vocabulary-nouns-that) | <code>THAT(inputs[0])</code> | <code>["*.noun"]</code> | <code>["value"]</code> | supported | [Pronouns](../spec/names-and-scope.md#that) | [Note 1](#vocabulary-note-1) |
| [this](#vocabulary-nouns-this) | <code>THIS(inputs[0])</code> | <code>["*.noun"]</code> | <code>["value"]</code> | supported | [Pronouns](../spec/names-and-scope.md#this) | [Note 1](#vocabulary-note-1) |
| [input](#vocabulary-nouns-input) | <code>INPUT(inputs[0])</code> | <code>["*.thing"]</code> | <code>["value"]</code> | supported | [Input metadata and runtime boundary](../reference/builtins/index.md#input) | [Note 2](#vocabulary-note-2) |
| [output](#vocabulary-nouns-output) | <code>OUTPUT(inputs[0])</code> | <code>["*.thing"]</code> | <code>["value"]</code> | supported | [Process members](../reference/builtins/lifecycle.md#output) | [Note 3](#vocabulary-note-3) |
| [error](#vocabulary-nouns-error) | <code>ERROR(inputs[0])</code> | <code>["*.thing"]</code> | <code>["value"]</code> | supported | [Process members](../reference/builtins/lifecycle.md#error) | [Note 3](#vocabulary-note-3) |
| [list](#vocabulary-nouns-list) | <code>LIST(inputs[0])</code> | <code>["*.thing"]</code> | <code>["value"]</code> | supported | [list](../spec/collections-and-categories.md#lists) | [Note 4](#vocabulary-note-4) |
| [dictionary](#vocabulary-nouns-dictionary) | <code>DICTIONARY(inputs[0], inputs[1])</code> | <code>["text","*.thing"]</code> | <code>["value"]</code> | supported | [dictionary](../spec/collections-and-categories.md#dictionaries) | [Note 5](#vocabulary-note-5) |
| [null](#vocabulary-nouns-null) | <code>NULL(inputs[0])</code> | <code>["*.thing"]</code> | <code>["value"]</code> | supported | [null](../spec/values.md#null) | [Note 6](#vocabulary-note-6) |
| [thing](#vocabulary-nouns-thing) | <code>THING(inputs[0])</code> | <code>["*.thing"]</code> | <code>["value"]</code> | supported | [thing](../spec/values.md#thing) | [Note 7](#vocabulary-note-7) |
| [text](#vocabulary-nouns-text) | <code>TEXT(inputs[0])</code> | <code>["*.text"]</code> | <code>["value"]</code> | supported | [text](../spec/values.md#text) | [Note 8](#vocabulary-note-8) |
| [number](#vocabulary-nouns-number) | <code>NUMBER(inputs[0])</code> | <code>["*.number"]</code> | <code>["value"]</code> | supported | [number](../spec/values.md#number) | [Note 9](#vocabulary-note-9) |
| [byte](#vocabulary-nouns-byte) | <code>BYTES(inputs[0])</code> | <code>["*.byte"]</code> | <code>["value"]</code> | partial | [byte](../spec/values.md#byte) | [Note 10](#vocabulary-note-10) |
| [truth](#vocabulary-nouns-truth) | <code>TRUTH(inputs[0])</code> | <code>["*.truth"]</code> | <code>["value"]</code> | supported | [truth](../spec/values.md#truth) | [Note 11](#vocabulary-note-11) |
| [pointer](#vocabulary-nouns-pointer) | <code>POINTER(inputs[0])</code> | <code>["*.thing"]</code> | <code>["value"]</code> | supported | [pointer](../spec/values.md#pointer) | [Note 12](#vocabulary-note-12) |
| [op](#vocabulary-nouns-op) | <code>OP(inputs[0], inputs[1], inputs[2], inputs[3])</code> | <code>["text","numbers","text","text"]</code> | <code>["value"]</code> | internal | [op](../spec/applicational-types.md#functions-properties-and-operations) | [Note 13](#vocabulary-note-13) |
| [prop](#vocabulary-nouns-prop) | <code>PROP(inputs[0], inputs[1], inputs[2])</code> | <code>["text","text","thing"]</code> | <code>["value"]</code> | internal | [prop](../spec/applicational-types.md#functions-properties-and-operations) | [Note 13](#vocabulary-note-13) |
| [func](#vocabulary-nouns-func) | <code>FUNC(inputs[0], inputs[1], inputs[2], inputs[3])</code> | <code>["text","props","texts","prop"]</code> | <code>["value"]</code> | internal | [func](../spec/applicational-types.md#functions-properties-and-operations) | [Note 13](#vocabulary-note-13) |
| [definition](#vocabulary-nouns-definition) | <code>DEF(inputs[0])</code> | <code>["text"]</code> | <code>["value"]</code> | partial | [definition](../spec/applicational-types.md#definition) | [Note 14](#vocabulary-note-14) |
| [attribute](#vocabulary-nouns-attribute) | <code>ATTRIBUTE(inputs[0], inputs[1])</code> | <code>["definition","attributes"]</code> | <code>["value"]</code> | partial | [attribute](../spec/applicational-types.md#attribute) | [Note 15](#vocabulary-note-15) |
| [object](#vocabulary-nouns-object) | <code>OBJECT(inputs[0], inputs[1])</code> | <code>["definition","attributes"]</code> | <code>["value"]</code> | supported | [object](../spec/applicational-types.md#object) | [Note 16](#vocabulary-note-16) |
| [class](#vocabulary-nouns-class) | <code>CLASS(inputs[0], inputs[1])</code> | <code>["objects","classes"]</code> | <code>["value"]</code> | supported | [class](../spec/applicational-types.md#class) | [Note 17](#vocabulary-note-17) |
| [service](#vocabulary-nouns-service) | <code>SERVICE(inputs[0], inputs[1], inputs[2])</code> | <code>["objects","classes","services"]</code> | <code>["value"]</code> | supported | [service](../spec/applicational-types.md#services) | [Note 18](#vocabulary-note-18) |
| [app](#vocabulary-nouns-app) | <code>APP(inputs[0], inputs[1], inputs[2])</code> | <code>["objects","services","dictionary"]</code> | <code>["value"]</code> | supported | [app](../spec/applicational-types.md#apps) | [Note 19](#vocabulary-note-19) |

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
| [passive](#vocabulary-verbs-passive) | <code>IS(inputs[0])</code> | <code>["*.thing"]</code> | <code>["value"]</code> | supported | [Assignment](../spec/control-flow.md#is) | [Note 20](#vocabulary-note-20) |
| [associative](#vocabulary-verbs-associative) | <code>USE(inputs[0])</code> | <code>["*.noun"]</code> | <code>["value"]</code> | supported | [Aliases](../spec/imports.md#indexed-dictionaries) | [Note 21](#vocabulary-note-21) |
| [declarative](#vocabulary-verbs-declarative) | <code>DECLARE(inputs[0])</code> | <code>["*.noun"]</code> | <code>["value"]</code> | supported | [Declarations](../spec/definitions.md#create) | [Note 22](#vocabulary-note-22) |
| [definitive](#vocabulary-verbs-definitive) | <code>DEFINE(inputs[0])</code> | <code>["*.noun"]</code> | <code>["value"]</code> | supported | [Definitions](../spec/definitions.md#define) | [Note 23](#vocabulary-note-23) |
| [descriptive](#vocabulary-verbs-descriptive) | <code>DESCRIBE(inputs[0])</code> | <code>["*.noun"]</code> | <code>["value"]</code> | supported | [Descriptors](../spec/definitions.md#describe) | [Note 24](#vocabulary-note-24) |
| [expansive](#vocabulary-verbs-expansive) | <code>PUT(inputs[0], inputs[1], inputs[2])</code> | <code>["*.thing","*.noun.iterative","text"]</code> | <code>["value"]</code> | supported | [Collection operations](../spec/collections-and-categories.md#put) | [Note 25](#vocabulary-note-25) |
| [locative](#vocabulary-verbs-locative) | <code>PLACE(inputs[0], inputs[1], inputs[2])</code> | <code>["*.thing","*.noun.iterative","text"]</code> | <code>["value"]</code> | supported | [Collection operations](../spec/collections-and-categories.md#place) | [Note 25](#vocabulary-note-25) |
| [receptive](#vocabulary-verbs-receptive) | <code>GET(inputs[0], inputs[1])</code> | <code>["text","*.noun.iterative"]</code> | <code>["value"]</code> | rejected | [Collection operations](../spec/grammar.md#get) | [Note 26](#vocabulary-note-26) |
| [reductive](#vocabulary-verbs-reductive) | <code>TAKE(inputs[0], inputs[1])</code> | <code>["text","*.noun.iterative"]</code> | <code>["value"]</code> | rejected | [Collection operations](../spec/grammar.md#take) | [Note 27](#vocabulary-note-27) |
| [generative](#vocabulary-verbs-generative) | <code>BUILD(inputs[0])</code> | <code>["*.noun.app"]</code> | <code>["value"]</code> | supported | [Build](../reference/builtins/lifecycle.md#build) | [Note 3](#vocabulary-note-3) |
| [executive](#vocabulary-verbs-executive) | <code>RUN(inputs[0])</code> | <code>["*.noun.app"]</code> | <code>["value"]</code> | supported | [Run](../reference/builtins/lifecycle.md#run) | [Note 3](#vocabulary-note-3) |
| [suspensive](#vocabulary-verbs-suspensive) | <code>AWAIT(inputs[0])</code> | <code>["*.noun.app"]</code> | <code>["value"]</code> | supported | [Await](../reference/builtins/lifecycle.md#await) | [Note 3](#vocabulary-note-3) |
| [interruptive](#vocabulary-verbs-interruptive) | <code>STOP(inputs[0])</code> | <code>["*.noun.app"]</code> | <code>["value"]</code> | supported | [Stop](../reference/builtins/lifecycle.md#stop) | [Note 3](#vocabulary-note-3) |
| [inflective](#vocabulary-verbs-inflective) | <code>IMPORT(inputs[0])</code> | <code>["*.thing"]</code> | <code>["value"]</code> | rejected | [Import/export distinction](../spec/grammar.md#import) | [Note 28](#vocabulary-note-28) |
| [deflective](#vocabulary-verbs-deflective) | <code>EXPORT(inputs[0])</code> | <code>["*.thing"]</code> | <code>["value"]</code> | rejected | [Import/export distinction](../spec/grammar.md#export) | [Note 29](#vocabulary-note-29) |
| [observative](#vocabulary-verbs-observative) | <code>SHOW(inputs[0])</code> | <code>["*.thing"]</code> | <code>["value"]</code> | supported | [Show](../reference/builtins/lifecycle.md#show) | [Note 30](#vocabulary-note-30) |
| [terminative](#vocabulary-verbs-terminative) | <code>EXIT(inputs[0])</code> | <code>["*.thing"]</code> | <code>["value"]</code> | supported | [Exit](../reference/builtins/lifecycle.md#exit) | [Note 31](#vocabulary-note-31) |
| [appropriative](#vocabulary-verbs-appropriative) | <code>HAVE(inputs[0])</code> | <code>["*.thing"]</code> | <code>["value"]</code> | supported | [Comparisons](../spec/control-flow.md#have) | [Note 32](#vocabulary-note-32) |
| [comparative](#vocabulary-verbs-comparative) | <code>EQUAL(inputs[0])</code> | <code>["*.thing"]</code> | <code>["value"]</code> | supported | [Comparisons](../spec/control-flow.md#equal) | [Note 33](#vocabulary-note-33) |

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
| [visibility.public](#vocabulary-adjectives-visibility.public) | — | — | — | supported | [Visibility](../spec/names-and-scope.md#public) | [Note 34](#vocabulary-note-34) |
| [initialization](#vocabulary-adjectives-initialization) | <code>NEW(inputs[0])</code> | <code>["*.noun"]</code> | <code>["value"]</code> | supported | [Descriptors](../spec/definitions.md#new) | [Note 35](#vocabulary-note-35) |
| [abstraction](#vocabulary-adjectives-abstraction) | <code>GENERIC(inputs[0])</code> | <code>["*.noun"]</code> | <code>["value"]</code> | supported | [Descriptors](../spec/definitions.md#generic) | [Note 36](#vocabulary-note-36) |

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
| [visibility.public](#vocabulary-adverbs-visibility.public) | — | — | — | supported | [Visibility](../spec/names-and-scope.md#publicly) | [Note 37](#vocabulary-note-37) |
| [nullification](#vocabulary-adverbs-nullification) | <code>NOT(inputs)</code> | <code>["*.thing"]</code> | <code>["value"]</code> | supported | [Conditions and modifiers](../spec/control-flow.md#not) | [Note 38](#vocabulary-note-38) |
| [comparison.subtraction](#vocabulary-adverbs-comparison.subtraction) | <code>LESSER(inputs)</code> | <code>["*.thing"]</code> | <code>["value"]</code> | supported | [Conditions and modifiers](../spec/control-flow.md#less) | [Note 39](#vocabulary-note-39) |
| [comparison.multiplication](#vocabulary-adverbs-comparison.multiplication) | <code>GREATER(inputs)</code> | <code>["*.thing"]</code> | <code>["value"]</code> | supported | [Conditions and modifiers](../spec/control-flow.md#greater) | [Note 39](#vocabulary-note-39) |
| [progression](#vocabulary-adverbs-progression) | <code>THEN(inputs)</code> | <code>["*.thing"]</code> | <code>["value"]</code> | supported | [Conditions and modifiers](../spec/control-flow.md#then) | [Note 40](#vocabulary-note-40) |

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
| [definement.declarative](#vocabulary-prepositions-definement.declarative) | <code>AS(inputs)</code> | <code>["*.thing"]</code> | <code>["value"]</code> | supported | [Prepositions](../spec/grammar.md#as) | [Note 41](#vocabulary-note-41) |
| [definement.definitive](#vocabulary-prepositions-definement.definitive) | <code>OF(inputs)</code> | <code>["*.thing"]</code> | <code>["value"]</code> | supported | [Prepositions](../spec/grammar.md#of) | [Note 42](#vocabulary-note-42) |
| [definement.locative](#vocabulary-prepositions-definement.locative) | <code>AT(inputs)</code> | <code>["*.thing"]</code> | <code>["value"]</code> | supported | [Prepositions](../spec/grammar.md#at) | [Note 43](#vocabulary-note-43) |
| [consignment.associative](#vocabulary-prepositions-consignment.associative) | <code>WITH(inputs)</code> | <code>["*.thing"]</code> | <code>["value"]</code> | supported | [Prepositions](../spec/grammar.md#with) | [Note 44](#vocabulary-note-44) |
| [consignment.inflective](#vocabulary-prepositions-consignment.inflective) | <code>IN(inputs)</code> | <code>["*.thing"]</code> | <code>["value"]</code> | supported | [Prepositions](../spec/grammar.md#in) | [Note 45](#vocabulary-note-45) |
| [consignment.deflective](#vocabulary-prepositions-consignment.deflective) | <code>OUT(inputs)</code> | <code>["*.thing"]</code> | <code>["value"]</code> | partial | [Prepositions](../spec/grammar.md#out) | [Note 46](#vocabulary-note-46) |
| [consignment.indicative](#vocabulary-prepositions-consignment.indicative) | <code>BY(inputs)</code> | <code>["*.thing"]</code> | <code>["value"]</code> | partial | [Prepositions](../spec/grammar.md#by) | [Note 47](#vocabulary-note-47) |
| [conjoinment.objective](#vocabulary-prepositions-conjoinment.objective) | <code>FROM(inputs)</code> | <code>["*.thing"]</code> | <code>["value"]</code> | supported | [Prepositions](../spec/grammar.md#from) | [Note 48](#vocabulary-note-48) |
| [conjoinment.subjective](#vocabulary-prepositions-conjoinment.subjective) | <code>TO(inputs)</code> | <code>["*.thing"]</code> | <code>["value"]</code> | supported | [Prepositions](../spec/grammar.md#to) | [Note 49](#vocabulary-note-49) |
| [conjoinment.comparative](#vocabulary-prepositions-conjoinment.comparative) | <code>THAN(inputs)</code> | <code>["*.thing"]</code> | <code>["value"]</code> | supported | [Prepositions](../spec/grammar.md#than) | [Note 50](#vocabulary-note-50) |
| [conjoinment.directive](#vocabulary-prepositions-conjoinment.directive) | <code>FOR(inputs)</code> | <code>["*.thing"]</code> | <code>["value"]</code> | supported | [Prepositions](../spec/grammar.md#for) | [Note 51](#vocabulary-note-51) |

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
| [variable](#vocabulary-articles-variable) | <code>ANY(inputs)</code> | <code>["*.thing"]</code> | <code>["value"]</code> | supported | [Articles and typed arguments](../spec/grammar.md#any) | [Note 52](#vocabulary-note-52) |
| [none](#vocabulary-articles-none) | <code>NO(inputs)</code> | <code>["*.thing"]</code> | <code>["value"]</code> | supported | [Articles and typed arguments](../spec/grammar.md#no) | [Note 53](#vocabulary-note-53) |
| [indefinite](#vocabulary-articles-indefinite) | <code>A(inputs)</code> | <code>["*.noun"]</code> | <code>["value"]</code> | supported | [Articles and typed arguments](../spec/grammar.md#a) | [Note 54](#vocabulary-note-54) |
| [definite](#vocabulary-articles-definite) | <code>THE(inputs)</code> | <code>["*.noun"]</code> | <code>["value"]</code> | supported | [Articles and typed arguments](../spec/grammar.md#the) | [Note 55](#vocabulary-note-55) |

## conjunctions

**Word forms**

| Word | root | pos | type | postfixes |
| --- | --- | --- | --- | --- |
| <a id="mutual"></a><a id="vocabulary-conjunctions-mutual"></a>mutual | <code>and</code> | <code>conjunction</code> | <code>mutual</code> | <code>{}</code> |
| <a id="exclusive"></a><a id="vocabulary-conjunctions-exclusive"></a>exclusive | <code>or</code> | <code>conjunction</code> | <code>exclusive</code> | <code>{}</code> |

**Handlers and support**

| Word | value | inputs | outputs | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| [mutual](#vocabulary-conjunctions-mutual) | <code>AND(inputs)</code> | <code>["*.thing"]</code> | <code>["value"]</code> | partial | [Conditions and modifiers](../spec/control-flow.md#and) | [Note 56](#vocabulary-note-56) |
| [exclusive](#vocabulary-conjunctions-exclusive) | <code>OR(inputs)</code> | <code>["*.thing"]</code> | <code>["value"]</code> | partial | [Conditions and modifiers](../spec/control-flow.md#or) | [Note 57](#vocabulary-note-57) |

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
| [conditional](#vocabulary-evaluations-conditional) | <code>IF(inputs)</code> | <code>["*.thing"]</code> | <code>["value"]</code> | partial | [Conditions and modifiers](../spec/control-flow.md#if) | [Note 58](#vocabulary-note-58) |
| [exceptional](#vocabulary-evaluations-exceptional) | <code>ELSE(inputs)</code> | <code>["*.thing"]</code> | <code>["value"]</code> | partial | [Conditions and modifiers](../spec/control-flow.md#else) | [Note 59](#vocabulary-note-59) |
| [alternative](#vocabulary-evaluations-alternative) | <code>OTHERWISE(inputs)</code> | <code>["*.thing"]</code> | <code>["value"]</code> | supported | [Conditions and modifiers](../spec/control-flow.md#otherwise) | [Note 60](#vocabulary-note-60) |
| [iterative](#vocabulary-evaluations-iterative) | <code>WHILE(inputs)</code> | <code>["*.thing"]</code> | <code>["value"]</code> | rejected | [Iteration](../spec/control-flow.md#while) | [Note 61](#vocabulary-note-61) |

## Support notes

| Note | Reviewed explanation |
| --- | --- |
| <a id="vocabulary-note-1"></a>1 | Paragraph antecedents and declared source-action arguments resolve through source identity. A new paragraph lacking an antecedent is an enforced scope rule, not an unimplemented pronoun capability. Plural inflections retain the same configured lexical identity. |
| <a id="vocabulary-note-2"></a>2 | Input metadata covers Function parameters and scoped structured Process bindings. No App process input transport is supplied. |
| <a id="vocabulary-note-3"></a>3 | Native process operations and implicit Output/Error require an App; source-defined actions and explicit members retain lexical precedence. |
| <a id="vocabulary-note-4"></a>4 | Typed owned Lists construct, append, replace and read indexed values; heterogeneous construction infers Thing without coercion. Indices and complete element types are checked. Iteration/removal remain outside this API. |
| <a id="vocabulary-note-5"></a>5 | Dictionary owns lexical members, exports and canonical alias identity. It is a source namespace, not runtime key/value storage; use Map for that distinct meaning. |
| <a id="vocabulary-note-6"></a>6 | Null initializes an explicit absence value; typed reads preserve evaluation and type/index/key checks. Unassigned storage and the native-definition Null atom remain separate. |
| <a id="vocabulary-note-7"></a>7 | Explicit Thing boundaries retain concrete runtime types; typed reads validate full element types and callable signatures. No implicit coercion or dynamic arithmetic. |
| <a id="vocabulary-note-8"></a>8 | Raw byte Text construction, literal binding, transfer and output implement the documented meaning. Lack of Unicode normalization or a general string-processing library is an API boundary, not an incomplete Text type. |
| <a id="vocabulary-note-9"></a>9 | Number is explicitly finite binary64; parsing, native encoding and typed runtime use implement that meaning. Arbitrary precision and integer-width arithmetic are not promised. |
| <a id="vocabulary-note-10"></a>10 | The configured Byte/Bytes family is recognized, but only Bytes has owned storage. Singular Byte has no scalar source type or one-byte constructor; octet reads produce Number. |
| <a id="vocabulary-note-11"></a>11 | Truth maps to bool and retains exact typed values. Absence of truthiness and limits on joined condition grammar do not make the value type partial. |
| <a id="vocabulary-note-12"></a>12 | Managed typed handles share targets; reads snapshot ordinary contents. Checked mutation rejects ownership cycles, including container and callable-environment edges. No pointer arithmetic or raw-address API. |
| <a id="vocabulary-note-13"></a>13 | Configured native representation spelling, not a public application constructor or alias for the source Function, Property or Operation category. Its configuration handler is not executable API evidence. |
| <a id="vocabulary-note-14"></a>14 | Framed definitions attach to words and composed Functions. The recognized Definition category has no standalone New Definition source constructor for a runtime value to store or inspect. |
| <a id="vocabulary-note-15"></a>15 | Recognized applicational category and internal Function/Object attributes are implemented. No standalone New Attribute source data-field or runtime-value constructor is supplied. |
| <a id="vocabulary-note-16"></a>16 | Object constructs source-file content containing Functions and their definitions. Runtime initialization/lifetime is not part of the documented meaning. |
| <a id="vocabulary-note-17"></a>17 | Class constructs ordered module/source groups including nested Classes and Objects. It does not promise a runtime scope or instance lifetime. |
| <a id="vocabulary-note-18"></a>18 | Service declarations remain structural; standalone bodies, packages and native process operations are unsupported. App-constructed Service libraries retain their workspace topology. Static inspection; witnesses not run. |
| <a id="vocabulary-note-19"></a>19 | App construction preserves source graph identity and yields the executable/lifecycle boundary documented by this word. Apps are distinct from structural Services. |
| <a id="vocabulary-note-20"></a>20 | Assignment and same-type scalar equality/inequality are implemented. Definite assignment and rejection of mixed or non-scalar comparisons enforce the documented meaning rather than mark it unfinished. |
| <a id="vocabulary-note-21"></a>21 | One selected registry and source provider implement indexed imports, confined includes and canonical aliases. Unsupported provider names/remote package resolution are outside the explicit-source contract. |
| <a id="vocabulary-note-22"></a>22 | Registration constructs the named supported value/action with atomic declaration ownership and source identity. Arbitrary English or unspecified native constructors are not promised. |
| <a id="vocabulary-note-23"></a>23 | Registration binds supported values or explicitly framed typed native definitions, checking slots and lexical ownership. It does not bypass the declared type or visibility rules. |
| <a id="vocabulary-note-24"></a>24 | Descriptor construction binds canonical targets and ordered refinements, applies exact target rules and rejects mismatches/cycles. Aliases preserve identity while foreign homonyms remain distinct. |
| <a id="vocabulary-note-25"></a>25 | Typed updates append/replace List elements, write Map values or replace a Pointer target. Destination kind, keys, element types and managed ownership cycles are checked. |
| <a id="vocabulary-note-26"></a>26 | The configured receptive root parses but has no native instruction role or bound implementation. A public Compile probe of Get 0 from items rejects with unsupported outer instruction action action.receptive; the matched typed read control compiles. |
| <a id="vocabulary-note-27"></a>27 | The configured reductive root has no native instruction role or bound implementation. A public Compile probe of Take 0 from items reaches unsupported outer instruction action action.reductive; this is known unavailability, not unresolved semantics. |
| <a id="vocabulary-note-28"></a>28 | The configured inflective action is not the source loader. A public Compile probe of Import `io` rejects with unsupported outer instruction action action.inflective while Use IO from `io` compiles. |
| <a id="vocabulary-note-29"></a>29 | The configured deflective action has no lowering implementation. A public Compile probe of Export `io` rejects with unsupported outer instruction action action.deflective. Public/Publicly publish declarations; they are not runtime Export aliases. |
| <a id="vocabulary-note-30"></a>30 | The outer instruction uses the selected runtime-value display path, preserving raw Text and adding no LF. Direct App/Error display is outside the declared supported-value domain. |
| <a id="vocabulary-note-31"></a>31 | The outer instruction terminates the program and owns shutdown/cleanup even from a continuation or source action. Function.Return is deliberately separate; no unawaited continuation is promised. |
| <a id="vocabulary-note-32"></a>32 | Has tests the selected App Error, including same-App ownership. General collection membership is a different, unclaimed operation. |
| <a id="vocabulary-note-33"></a>33 | Equality of matching Text, Number and Truth operands is implemented; no coercion or arbitrary object equality is promised. |
| <a id="vocabulary-note-34"></a>34 | The declaration modifier publishes one canonical declaration without publishing siblings or cloning dictionary membership. Private captures and definite assignment retain their own checks. |
| <a id="vocabulary-note-35"></a>35 | Construction selects the declared noun and exact descriptor refinements with checked operands. There is intentionally no universal constructor, cloning fallback or invented initialization. |
| <a id="vocabulary-note-36"></a>36 | Native source constructors accept explicit ordered parameter atoms through their declared slots. GenericTypes and Lifetimes retain their approved local native-input contract; this does not infer parameters or introduce general L generics. |
| <a id="vocabulary-note-37"></a>37 | The publication adverb applies to the selected declaration, retaining its identity and defining paragraph. It is a distinct lexical role from Public, with the same checked publication mechanism. |
| <a id="vocabulary-note-38"></a>38 | Condition-owned negation is implemented by instructionCondition.Negated and native condition emission. Lack of joined conditions is an And/Or/If limitation, not a missing unary negation implementation. |
| <a id="vocabulary-note-39"></a>39 | Numeric strict and inclusive comparison are implemented. Rejecting Text ordering is the declared Number-operand contract. |
| <a id="vocabulary-note-40"></a>40 | And then preserves composed action order; source-action expansion evaluates each complete body before the next coordinated operand. It is a sequencing token, not an asynchronous scheduling feature. |
| <a id="vocabulary-note-41"></a>41 | Declaration construction, aliases, typed parameters and invocation result names use the selected owning grammar and canonical binding. No standalone conversion operation is claimed. |
| <a id="vocabulary-note-42"></a>42 | Element types/contents and Function signature components are implemented through owner-specific construction and signature extraction. |
| <a id="vocabulary-note-43"></a>43 | Checked reads and updates bind Number indices or Text keys and retain the selected container and result types. Key and bounds checks are intended behavior. |
| <a id="vocabulary-note-44"></a>44 | Function signatures and descriptor-associated child constructions use With under their selected owners. It does not erase argument or descriptor contracts. |
| <a id="vocabulary-note-45"></a>45 | List/Map/Pointer update destinations are lowered and type checked. In has no unclaimed lexical-namespace behavior. |
| <a id="vocabulary-note-46"></a>46 | Generic Function argument phrases accept Out and bind an ordinary argument. The configured consignment.deflective role has no dedicated grammar phrase or executable deflective/output/export operation; accepting a PREPOSITION in a generic call does not implement that missing role. |
| <a id="vocabulary-note-47"></a>47 | Generic Function argument phrases accept By and bind an ordinary argument. The configured phrase.indicative describes by Actor but no enclosing clause references that phrase to implement actor/passive execution. Generic separator acceptance does not implement that missing role. |
| <a id="vocabulary-note-48"></a>48 | Indexed/file import sources and checked typed-read sources are implemented by their distinct owners. Provider isolation and read typing are enforced contracts. |
| <a id="vocabulary-note-49"></a>49 | Native definition utilization, Function prepositional arguments and inclusive comparison phrases have their documented owner-specific implementations. To has no independent conversion/callback meaning. |
| <a id="vocabulary-note-50"></a>50 | Than supplies the comparison relationship in implemented numeric strict and inclusive conditions. Matching Number typing is intentional. |
| <a id="vocabulary-note-51"></a>51 | For introduces the source-action body; source instructions expand the selected action with scoped arguments in declared order. It does not introduce a general loop construct. |
| <a id="vocabulary-note-52"></a>52 | Any introduces typed source-action parameters or exact descriptor targets. Canonical target checking and rejection of ambiguous antecedents implement its contract, without unrestricted native generic promises. |
| <a id="vocabulary-note-53"></a>53 | Explicit empty Function Inputs/Outputs and grammar-owned condition negation are implemented. No does not supply implicit initialization. |
| <a id="vocabulary-note-54"></a>54 | A/An indefinite article family participates in implemented declarations, constructions and Error conditions. It need not select a type independently of its noun. |
| <a id="vocabulary-note-55"></a>55 | The selects a definite subject/value in its existing lexical context; App body effects and member Function invocations have maintained witnesses. Visibility rejection is the contract. |
| <a id="vocabulary-note-56"></a>56 | Operand/declaration/action coordination works, but recognized And-joined conditions are explicitly rejected by instructionBuilder.conditional because no Boolean join operator is implemented. |
| <a id="vocabulary-note-57"></a>57 | Canonical descriptor target alternatives work, but recognized Or-joined conditions are explicitly rejected by instructionBuilder.conditional because no Boolean join operator is implemented. |
| <a id="vocabulary-note-58"></a>58 | Single typed conditions produce branches with flow checks. Recognized joined conditions reject during instruction lowering and configured else-if continuation rejects during grammar processing. |
| <a id="vocabulary-note-59"></a>59 | Plain Else works, but the configured conditional.else.if continuation does not parse as an executable chain. Actual valid single-if/else control passes while else-if leaves unassociated else tokens. |
| <a id="vocabulary-note-60"></a>60 | Otherwise attaches to the immediately preceding unmatched conditional within the same paragraph. Intervening instructions or paragraph boundaries correctly reject; they are ownership rules rather than missing support. |
| <a id="vocabulary-note-61"></a>61 | The configured While loop has no outer instruction implementation: instructionBuilder.statement accepts only if/else/other. The documented While fragment additionally fails grammar processing; do not claim this fixture reaches lowering or that every possible configured While spelling was probed. |
