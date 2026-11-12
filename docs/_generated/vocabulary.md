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
| [it](#vocabulary-nouns-it) | <code>IT(inputs[0])</code> | <code>["*.noun"]</code> | <code>["value"]</code> | partial | [Pronouns](../reference/word-glossary.md#it) | [Note 1](#vocabulary-note-1) |
| [that](#vocabulary-nouns-that) | <code>THAT(inputs[0])</code> | <code>["*.noun"]</code> | <code>["value"]</code> | partial | [Pronouns](../reference/word-glossary.md#that) | [Note 1](#vocabulary-note-1) |
| [this](#vocabulary-nouns-this) | <code>THIS(inputs[0])</code> | <code>["*.noun"]</code> | <code>["value"]</code> | partial | [Pronouns](../reference/word-glossary.md#this) | [Note 1](#vocabulary-note-1) |
| [input](#vocabulary-nouns-input) | <code>INPUT(inputs[0])</code> | <code>["*.thing"]</code> | <code>["value"]</code> | partial | [Input metadata and runtime boundary](../reference/word-glossary.md#input) | [Note 1](#vocabulary-note-1) |
| [output](#vocabulary-nouns-output) | <code>OUTPUT(inputs[0])</code> | <code>["*.thing"]</code> | <code>["value"]</code> | rejected | [Process members](../reference/word-glossary.md#appoutput) | [Note 1](#vocabulary-note-1) |
| [error](#vocabulary-nouns-error) | <code>ERROR(inputs[0])</code> | <code>["*.thing"]</code> | <code>["value"]</code> | rejected | [Process members](../reference/word-glossary.md#apperror) | [Note 1](#vocabulary-note-1) |
| [list](#vocabulary-nouns-list) | <code>LIST(inputs[0])</code> | <code>["*.thing"]</code> | <code>["value"]</code> | partial | [list](../reference/word-glossary.md#list) | [Note 1](#vocabulary-note-1) |
| [dictionary](#vocabulary-nouns-dictionary) | <code>DICTIONARY(inputs[0], inputs[1])</code> | <code>["text","*.thing"]</code> | <code>["value"]</code> | partial | [dictionary](../reference/word-glossary.md#dictionary) | [Note 1](#vocabulary-note-1) |
| [null](#vocabulary-nouns-null) | <code>NULL(inputs[0])</code> | <code>["*.thing"]</code> | <code>["value"]</code> | partial | [null](../reference/word-glossary.md#null) | [Note 1](#vocabulary-note-1) |
| [thing](#vocabulary-nouns-thing) | <code>THING(inputs[0])</code> | <code>["*.thing"]</code> | <code>["value"]</code> | partial | [thing](../reference/word-glossary.md#thing) | [Note 1](#vocabulary-note-1) |
| [text](#vocabulary-nouns-text) | <code>TEXT(inputs[0])</code> | <code>["*.text"]</code> | <code>["value"]</code> | partial | [text](../reference/word-glossary.md#text) | [Note 1](#vocabulary-note-1) |
| [number](#vocabulary-nouns-number) | <code>NUMBER(inputs[0])</code> | <code>["*.number"]</code> | <code>["value"]</code> | partial | [number](../reference/word-glossary.md#number) | [Note 1](#vocabulary-note-1) |
| [byte](#vocabulary-nouns-byte) | <code>BYTES(inputs[0])</code> | <code>["*.byte"]</code> | <code>["value"]</code> | partial | [byte](../reference/word-glossary.md#byte) | [Note 1](#vocabulary-note-1) |
| [truth](#vocabulary-nouns-truth) | <code>TRUTH(inputs[0])</code> | <code>["*.truth"]</code> | <code>["value"]</code> | partial | [truth](../reference/word-glossary.md#truth) | [Note 1](#vocabulary-note-1) |
| [pointer](#vocabulary-nouns-pointer) | <code>POINTER(inputs[0])</code> | <code>["*.thing"]</code> | <code>["value"]</code> | partial | [pointer](../reference/word-glossary.md#pointer) | [Note 1](#vocabulary-note-1) |
| [op](#vocabulary-nouns-op) | <code>OP(inputs[0], inputs[1], inputs[2], inputs[3])</code> | <code>["text","numbers","text","text"]</code> | <code>["value"]</code> | internal | [op](../spec/applicational-types.md#functions-properties-and-operations) | [Note 2](#vocabulary-note-2) |
| [prop](#vocabulary-nouns-prop) | <code>PROP(inputs[0], inputs[1], inputs[2])</code> | <code>["text","text","thing"]</code> | <code>["value"]</code> | internal | [prop](../spec/applicational-types.md#functions-properties-and-operations) | [Note 2](#vocabulary-note-2) |
| [func](#vocabulary-nouns-func) | <code>FUNC(inputs[0], inputs[1], inputs[2], inputs[3])</code> | <code>["text","props","texts","prop"]</code> | <code>["value"]</code> | internal | [func](../spec/applicational-types.md#functions-properties-and-operations) | [Note 2](#vocabulary-note-2) |
| [definition](#vocabulary-nouns-definition) | <code>DEF(inputs[0])</code> | <code>["text"]</code> | <code>["value"]</code> | partial | [definition](../reference/word-glossary.md#definition) | [Note 1](#vocabulary-note-1) |
| [attribute](#vocabulary-nouns-attribute) | <code>ATTRIBUTE(inputs[0], inputs[1])</code> | <code>["definition","attributes"]</code> | <code>["value"]</code> | partial | [attribute](../reference/word-glossary.md#attribute) | [Note 1](#vocabulary-note-1) |
| [object](#vocabulary-nouns-object) | <code>OBJECT(inputs[0], inputs[1])</code> | <code>["definition","attributes"]</code> | <code>["value"]</code> | partial | [object](../reference/word-glossary.md#object) | [Note 1](#vocabulary-note-1) |
| [class](#vocabulary-nouns-class) | <code>CLASS(inputs[0], inputs[1])</code> | <code>["objects","classes"]</code> | <code>["value"]</code> | partial | [class](../reference/word-glossary.md#class) | [Note 1](#vocabulary-note-1) |
| [service](#vocabulary-nouns-service) | <code>SERVICE(inputs[0], inputs[1], inputs[2])</code> | <code>["objects","classes","services"]</code> | <code>["value"]</code> | partial | [service](../reference/word-glossary.md#service) | [Note 1](#vocabulary-note-1) |
| [app](#vocabulary-nouns-app) | <code>APP(inputs[0], inputs[1], inputs[2])</code> | <code>["objects","services","dictionary"]</code> | <code>["value"]</code> | partial | [app](../reference/word-glossary.md#app) | [Note 1](#vocabulary-note-1) |

## verbs

**Word forms**

| Word | root | pos | type | postfixes |
| --- | --- | --- | --- | --- |
| <a id="passive"></a><a id="vocabulary-verbs-passive"></a>passive | <code>is</code> | <code>verb</code> | <code>action.passive</code> | <code>{"-nt":"null"}</code> |
| <a id="associative"></a><a id="vocabulary-verbs-associative"></a>associative | <code>use</code> | <code>verb</code> | <code>action.associative</code> | <code>{"-s":"active","-d":"passive","--ing":"applicative","-r":"objective"}</code> |
| <a id="declarative"></a><a id="vocabulary-verbs-declarative"></a>declarative | <code>create</code> | <code>verb</code> | <code>action.declarative</code> | <code>{"-s":"active","-d":"passive","--ing":"applicative","--or":"objective"}</code> |
| <a id="definitive"></a><a id="vocabulary-verbs-definitive"></a>definitive | <code>define</code> | <code>verb</code> | <code>action.definitive</code> | <code>{"-s":"active","-d":"passive","--ing":"applicative","-r":"objective"}</code> |
| <a id="descriptive"></a><a id="vocabulary-verbs-descriptive"></a>descriptive | <code>describe</code> | <code>verb</code> | <code>action.descriptive</code> | <code>{"-s":"active","-d":"passive","--ing":"applicative","-r":"objective"}</code> |
| <a id="expansive"></a><a id="vocabulary-verbs-expansive"></a>expansive | <code>put</code> | <code>verb</code> | <code>action.expansive</code> | <code>{"-s":"active","-":"passive","-ting":"applicative","-ter":"objective"}</code> |
| <a id="locative"></a><a id="vocabulary-verbs-locative"></a>locative | <code>place</code> | <code>verb</code> | <code>action.locative</code> | <code>{"-s":"active","-d":"passive","--ing":"applicative","-r":"objective"}</code> |
| <a id="receptive"></a><a id="vocabulary-verbs-receptive"></a>receptive | <code>get</code> | <code>verb</code> | <code>action.receptive</code> | <code>{"-s":"active","---otten":"passive","-ting":"applicative","-ter":"objective"}</code> |
| <a id="reductive"></a><a id="vocabulary-verbs-reductive"></a>reductive | <code>take</code> | <code>verb</code> | <code>action.reductive</code> | <code>{"-s":"active","-n":"passive","--ing":"applicative","-r":"objective"}</code> |
| <a id="generative"></a><a id="vocabulary-verbs-generative"></a>generative | <code>build</code> | <code>verb</code> | <code>action.generative</code> | <code>{"-s":"active","--t":"passive","-ing":"applicative","-er":"objective"}</code> |
| <a id="executive"></a><a id="vocabulary-verbs-executive"></a>executive | <code>run</code> | <code>verb</code> | <code>action.executive</code> | <code>{"-s":"active","---an":"passive","-ning":"applicative","-ner":"objective"}</code> |
| <a id="suspensive"></a><a id="vocabulary-verbs-suspensive"></a>suspensive | <code>await</code> | <code>verb</code> | <code>action.suspensive</code> | <code>{"-s":"active","-ed":"passive","-ing":"applicative"}</code> |
| <a id="interruptive"></a><a id="vocabulary-verbs-interruptive"></a>interruptive | <code>stop</code> | <code>verb</code> | <code>action.interruptive</code> | <code>{"-s":"active","-ped":"passive","-ping":"applicative","-per":"objective"}</code> |
| <a id="inflective"></a><a id="vocabulary-verbs-inflective"></a>inflective | <code>import</code> | <code>verb</code> | <code>action.inflective</code> | <code>{"-s":"active","-ed":"passive","-ing":"applicative","-er":"objective"}</code> |
| <a id="deflective"></a><a id="vocabulary-verbs-deflective"></a>deflective | <code>export</code> | <code>verb</code> | <code>action.deflective</code> | <code>{"-s":"active","-ed":"passive","-ing":"applicative","-er":"objective"}</code> |
| <a id="observative"></a><a id="vocabulary-verbs-observative"></a>observative | <code>show</code> | <code>verb</code> | <code>action.observative</code> | <code>{"-s":"active","-ed":"passive","-ing":"applicative","-er":"objective"}</code> |
| <a id="terminative"></a><a id="vocabulary-verbs-terminative"></a>terminative | <code>exit</code> | <code>verb</code> | <code>action.terminative</code> | <code>{"-s":"active","-ed":"passive","-ing":"applicative","-er":"objective"}</code> |
| <a id="appropriative"></a><a id="vocabulary-verbs-appropriative"></a>appropriative | <code>have</code> | <code>verb</code> | <code>operation.appropriative</code> | <code>{"---s":"active","---d":"passive","--ing":"applicative","-r":"objective"}</code> |
| <a id="comparative"></a><a id="vocabulary-verbs-comparative"></a>comparative | <code>equal</code> | <code>verb</code> | <code>operation.comparative</code> | <code>{"-s":"active","-ed":"passive","-ing":"applicative","-er":"objective"}</code> |

**Handlers and support**

| Word | value | inputs | outputs | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| [passive](#vocabulary-verbs-passive) | <code>IS(inputs[0])</code> | <code>["*.thing"]</code> | <code>["value"]</code> | partial | [Assignment](../reference/word-glossary.md#is) | [Note 3](#vocabulary-note-3) |
| [associative](#vocabulary-verbs-associative) | <code>USE(inputs[0])</code> | <code>["*.noun"]</code> | <code>["value"]</code> | partial | [Aliases](../reference/word-glossary.md#use) | [Note 4](#vocabulary-note-4) |
| [declarative](#vocabulary-verbs-declarative) | <code>DECLARE(inputs[0])</code> | <code>["*.noun"]</code> | <code>["value"]</code> | partial | [Declarations](../reference/word-glossary.md#create) | [Note 5](#vocabulary-note-5) |
| [definitive](#vocabulary-verbs-definitive) | <code>DEFINE(inputs[0])</code> | <code>["*.noun"]</code> | <code>["value"]</code> | partial | [Definitions](../reference/word-glossary.md#define) | [Note 6](#vocabulary-note-6) |
| [descriptive](#vocabulary-verbs-descriptive) | <code>DESCRIBE(inputs[0])</code> | <code>["*.noun"]</code> | <code>["value"]</code> | partial | [Descriptors](../reference/word-glossary.md#describe) | [Note 1](#vocabulary-note-1) |
| [expansive](#vocabulary-verbs-expansive) | <code>PUT(inputs[0], inputs[1], inputs[2])</code> | <code>["*.thing","*.noun.iterative","text"]</code> | <code>["value"]</code> | rejected | [Collection operations](../reference/word-glossary.md#put) | [Note 1](#vocabulary-note-1) |
| [locative](#vocabulary-verbs-locative) | <code>PLACE(inputs[0], inputs[1], inputs[2])</code> | <code>["*.thing","*.noun.iterative","text"]</code> | <code>["value"]</code> | rejected | [Collection operations](../reference/word-glossary.md#place) | [Note 1](#vocabulary-note-1) |
| [receptive](#vocabulary-verbs-receptive) | <code>GET(inputs[0], inputs[1])</code> | <code>["text","*.noun.iterative"]</code> | <code>["value"]</code> | rejected | [Collection operations](../reference/word-glossary.md#get) | [Note 1](#vocabulary-note-1) |
| [reductive](#vocabulary-verbs-reductive) | <code>TAKE(inputs[0], inputs[1])</code> | <code>["text","*.noun.iterative"]</code> | <code>["value"]</code> | rejected | [Collection operations](../reference/word-glossary.md#take) | [Note 1](#vocabulary-note-1) |
| [generative](#vocabulary-verbs-generative) | <code>BUILD(inputs[0])</code> | <code>["*.noun.app"]</code> | <code>["value"]</code> | rejected | [Build](../reference/word-glossary.md#build) | [Note 1](#vocabulary-note-1) |
| [executive](#vocabulary-verbs-executive) | <code>RUN(inputs[0])</code> | <code>["*.noun.app"]</code> | <code>["value"]</code> | rejected | [Run](../reference/word-glossary.md#run) | [Note 1](#vocabulary-note-1) |
| [suspensive](#vocabulary-verbs-suspensive) | <code>AWAIT(inputs[0])</code> | <code>["*.noun.app"]</code> | <code>["value"]</code> | rejected | [Await](../reference/word-glossary.md#await) | [Note 1](#vocabulary-note-1) |
| [interruptive](#vocabulary-verbs-interruptive) | <code>STOP(inputs[0])</code> | <code>["*.noun.app"]</code> | <code>["value"]</code> | rejected | [Stop](../reference/word-glossary.md#stop) | [Note 1](#vocabulary-note-1) |
| [inflective](#vocabulary-verbs-inflective) | <code>IMPORT(inputs[0])</code> | <code>["*.thing"]</code> | <code>["value"]</code> | rejected | [Import/export distinction](../reference/word-glossary.md#import) | [Note 1](#vocabulary-note-1) |
| [deflective](#vocabulary-verbs-deflective) | <code>EXPORT(inputs[0])</code> | <code>["*.thing"]</code> | <code>["value"]</code> | rejected | [Import/export distinction](../reference/word-glossary.md#export) | [Note 1](#vocabulary-note-1) |
| [observative](#vocabulary-verbs-observative) | <code>SHOW(inputs[0])</code> | <code>["*.thing"]</code> | <code>["value"]</code> | rejected | [Show](../reference/word-glossary.md#show) | [Note 1](#vocabulary-note-1) |
| [terminative](#vocabulary-verbs-terminative) | <code>EXIT(inputs[0])</code> | <code>["*.thing"]</code> | <code>["value"]</code> | rejected | [Exit](../reference/word-glossary.md#exit) | [Note 1](#vocabulary-note-1) |
| [appropriative](#vocabulary-verbs-appropriative) | <code>HAVE(inputs[0])</code> | <code>["*.thing"]</code> | <code>["value"]</code> | partial | [Comparisons](../reference/word-glossary.md#have) | [Note 1](#vocabulary-note-1) |
| [comparative](#vocabulary-verbs-comparative) | <code>EQUAL(inputs[0])</code> | <code>["*.thing"]</code> | <code>["value"]</code> | partial | [Comparisons](../reference/word-glossary.md#equal) | [Note 1](#vocabulary-note-1) |

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
| [visibility.public](#vocabulary-adjectives-visibility.public) | Not declared | Not declared | Not declared | partial | [Visibility](../reference/word-glossary.md#public) | [Note 1](#vocabulary-note-1) |
| [initialization](#vocabulary-adjectives-initialization) | <code>NEW(inputs[0])</code> | <code>["*.noun"]</code> | <code>["value"]</code> | partial | [Descriptors](../reference/word-glossary.md#new) | [Note 1](#vocabulary-note-1) |
| [abstraction](#vocabulary-adjectives-abstraction) | <code>GENERIC(inputs[0])</code> | <code>["*.noun"]</code> | <code>["value"]</code> | partial | [Descriptors](../reference/word-glossary.md#generic) | [Note 1](#vocabulary-note-1) |

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
| [visibility.public](#vocabulary-adverbs-visibility.public) | Not declared | Not declared | Not declared | partial | [Visibility](../reference/word-glossary.md#publicly) | [Note 1](#vocabulary-note-1) |
| [nullification](#vocabulary-adverbs-nullification) | <code>NOT(inputs)</code> | <code>["*.thing"]</code> | <code>["value"]</code> | partial | [Conditions and modifiers](../reference/word-glossary.md#not) | [Note 1](#vocabulary-note-1) |
| [comparison.subtraction](#vocabulary-adverbs-comparison.subtraction) | <code>LESSER(inputs)</code> | <code>["*.thing"]</code> | <code>["value"]</code> | partial | [Conditions and modifiers](../reference/word-glossary.md#less) | [Note 1](#vocabulary-note-1) |
| [comparison.multiplication](#vocabulary-adverbs-comparison.multiplication) | <code>GREATER(inputs)</code> | <code>["*.thing"]</code> | <code>["value"]</code> | partial | [Conditions and modifiers](../reference/word-glossary.md#greater) | [Note 1](#vocabulary-note-1) |
| [progression](#vocabulary-adverbs-progression) | <code>THEN(inputs)</code> | <code>["*.thing"]</code> | <code>["value"]</code> | partial | [Conditions and modifiers](../reference/word-glossary.md#then) | [Note 1](#vocabulary-note-1) |

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
| [definement.declarative](#vocabulary-prepositions-definement.declarative) | <code>AS(inputs)</code> | <code>["*.thing"]</code> | <code>["value"]</code> | partial | [Prepositions](../reference/word-glossary.md#as) | [Note 7](#vocabulary-note-7) |
| [definement.definitive](#vocabulary-prepositions-definement.definitive) | <code>OF(inputs)</code> | <code>["*.thing"]</code> | <code>["value"]</code> | partial | [Prepositions](../reference/word-glossary.md#of) | [Note 1](#vocabulary-note-1) |
| [definement.locative](#vocabulary-prepositions-definement.locative) | <code>AT(inputs)</code> | <code>["*.thing"]</code> | <code>["value"]</code> | partial | [Prepositions](../reference/word-glossary.md#at) | [Note 1](#vocabulary-note-1) |
| [consignment.associative](#vocabulary-prepositions-consignment.associative) | <code>WITH(inputs)</code> | <code>["*.thing"]</code> | <code>["value"]</code> | partial | [Prepositions](../reference/word-glossary.md#with) | [Note 1](#vocabulary-note-1) |
| [consignment.inflective](#vocabulary-prepositions-consignment.inflective) | <code>IN(inputs)</code> | <code>["*.thing"]</code> | <code>["value"]</code> | partial | [Prepositions](../reference/word-glossary.md#in) | [Note 1](#vocabulary-note-1) |
| [consignment.deflective](#vocabulary-prepositions-consignment.deflective) | <code>OUT(inputs)</code> | <code>["*.thing"]</code> | <code>["value"]</code> | partial | [Prepositions](../reference/word-glossary.md#out) | [Note 1](#vocabulary-note-1) |
| [consignment.indicative](#vocabulary-prepositions-consignment.indicative) | <code>BY(inputs)</code> | <code>["*.thing"]</code> | <code>["value"]</code> | partial | [Prepositions](../reference/word-glossary.md#by) | [Note 1](#vocabulary-note-1) |
| [conjoinment.objective](#vocabulary-prepositions-conjoinment.objective) | <code>FROM(inputs)</code> | <code>["*.thing"]</code> | <code>["value"]</code> | partial | [Prepositions](../reference/word-glossary.md#from) | [Note 8](#vocabulary-note-8) |
| [conjoinment.subjective](#vocabulary-prepositions-conjoinment.subjective) | <code>TO(inputs)</code> | <code>["*.thing"]</code> | <code>["value"]</code> | partial | [Prepositions](../reference/word-glossary.md#to) | [Note 9](#vocabulary-note-9) |
| [conjoinment.comparative](#vocabulary-prepositions-conjoinment.comparative) | <code>THAN(inputs)</code> | <code>["*.thing"]</code> | <code>["value"]</code> | partial | [Prepositions](../reference/word-glossary.md#than) | [Note 1](#vocabulary-note-1) |
| [conjoinment.directive](#vocabulary-prepositions-conjoinment.directive) | <code>FOR(inputs)</code> | <code>["*.thing"]</code> | <code>["value"]</code> | partial | [Prepositions](../reference/word-glossary.md#for) | [Note 10](#vocabulary-note-10) |

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
| [variable](#vocabulary-articles-variable) | <code>ANY(inputs)</code> | <code>["*.thing"]</code> | <code>["value"]</code> | partial | [Articles and typed arguments](../reference/word-glossary.md#any) | [Note 1](#vocabulary-note-1) |
| [none](#vocabulary-articles-none) | <code>NO(inputs)</code> | <code>["*.thing"]</code> | <code>["value"]</code> | partial | [Articles and typed arguments](../reference/word-glossary.md#no) | [Note 1](#vocabulary-note-1) |
| [indefinite](#vocabulary-articles-indefinite) | <code>A(inputs)</code> | <code>["*.noun"]</code> | <code>["value"]</code> | partial | [Articles and typed arguments](../reference/word-glossary.md#a-1) | [Note 1](#vocabulary-note-1) |
| [definite](#vocabulary-articles-definite) | <code>THE(inputs)</code> | <code>["*.noun"]</code> | <code>["value"]</code> | partial | [Articles and typed arguments](../reference/word-glossary.md#the) | [Note 1](#vocabulary-note-1) |

## conjunctions

**Word forms**

| Word | root | pos | type | postfixes |
| --- | --- | --- | --- | --- |
| <a id="mutual"></a><a id="vocabulary-conjunctions-mutual"></a>mutual | <code>and</code> | <code>conjunction</code> | <code>mutual</code> | <code>{}</code> |
| <a id="exclusive"></a><a id="vocabulary-conjunctions-exclusive"></a>exclusive | <code>or</code> | <code>conjunction</code> | <code>exclusive</code> | <code>{}</code> |

**Handlers and support**

| Word | value | inputs | outputs | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| [mutual](#vocabulary-conjunctions-mutual) | <code>AND(inputs)</code> | <code>["*.thing"]</code> | <code>["value"]</code> | partial | [Conditions and modifiers](../reference/word-glossary.md#and) | [Note 1](#vocabulary-note-1) |
| [exclusive](#vocabulary-conjunctions-exclusive) | <code>OR(inputs)</code> | <code>["*.thing"]</code> | <code>["value"]</code> | partial | [Conditions and modifiers](../reference/word-glossary.md#or) | [Note 1](#vocabulary-note-1) |

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
| [conditional](#vocabulary-evaluations-conditional) | <code>IF(inputs)</code> | <code>["*.thing"]</code> | <code>["value"]</code> | partial | [Conditions and modifiers](../reference/word-glossary.md#if) | [Note 1](#vocabulary-note-1) |
| [exceptional](#vocabulary-evaluations-exceptional) | <code>ELSE(inputs)</code> | <code>["*.thing"]</code> | <code>["value"]</code> | partial | [Conditions and modifiers](../reference/word-glossary.md#else) | [Note 1](#vocabulary-note-1) |
| [alternative](#vocabulary-evaluations-alternative) | <code>OTHERWISE(inputs)</code> | <code>["*.thing"]</code> | <code>["value"]</code> | partial | [Conditions and modifiers](../reference/word-glossary.md#otherwise) | [Note 1](#vocabulary-note-1) |
| [iterative](#vocabulary-evaluations-iterative) | <code>WHILE(inputs)</code> | <code>["*.thing"]</code> | <code>["value"]</code> | rejected | [Iteration](../reference/word-glossary.md#while) | [Note 1](#vocabulary-note-1) |

## Support notes

| Note | Reviewed explanation |
| --- | --- |
| <a id="vocabulary-note-1"></a>1 | Configured spelling and grammatical handlers are retained. The linked entry describes the core-only reset boundary; this inventory does not assert runtime execution. |
| <a id="vocabulary-note-2"></a>2 | Configured native representation spelling, not a public application constructor or alias for the source Function, Property or Operation category. Its configuration handler is not executable API evidence. |
| <a id="vocabulary-note-3"></a>3 | Configured Is spellings and handlers are unchanged. Ordinary Function-value assignment does not declare a new invocation family; a framed native definition in a value position requires a complete noun family. The current CLI builds one standalone App with core Print and Text; this is compiler analysis, not runtime evidence. |
| <a id="vocabulary-note-4"></a>4 | Configured Use spellings and handlers are unchanged. A lexical alias inherits forms only when removing qualification under the unchanged canonical root of a complete validated family. Renames and selected actor aliases declare complete families; alias chains cannot bypass validation. Ordinary value aliases remain data. Dot-qualified authored references, declarations and import aliases reject with source attribution. Text, comments, numbers, paths, native/template contents and internal identities retain their separate meanings and bytes. The current CLI builds one standalone App with core Print and Text; this is compiler analysis, not runtime evidence. |
| <a id="vocabulary-note-5"></a>5 | Configured Create spellings and handlers are unchanged. Authored source actions and named Functions require complete verb families; native noun declarations require complete noun families. Ordinary value and non-callable application labels remain single names. The current CLI builds one standalone App with core Print and Text; this is compiler analysis, not runtime evidence. |
| <a id="vocabulary-note-6"></a>6 | Configured Define spellings and handlers are unchanged. Authored noun families require root/plural/descriptive/active and verb families require root/active/passive/applicative/objective, with at least one expanded derived form different from the root case-insensitively. Missing, empty, malformed, surplus or all-identical families reject. Ordinary values remain labels, and native/template bytes retain their separate syntax. The current CLI builds one standalone App with core Print and Text; this is compiler analysis, not runtime evidence. |
| <a id="vocabulary-note-7"></a>7 | Configured As retains its value, construction, signature and alias relations. The selected declaration role determines family requirements; as does not universally register a noun. The current CLI builds one standalone App with core Print and Text; this is compiler analysis, not runtime evidence. |
| <a id="vocabulary-note-8"></a>8 | Configured From retains its selected import-provider and value relations. Dotted literal paths remain paths; authored names and aliases reject dots. The current CLI builds one standalone App with core Print and Text; this is compiler analysis, not runtime evidence. |
| <a id="vocabulary-note-9"></a>9 | Configured To retains its utilization relation. Authored native verb definitions require complete five-form families while their framed native bodies retain their bytes. The current CLI builds one standalone App with core Print and Text; this is compiler analysis, not runtime evidence. |
| <a id="vocabulary-note-10"></a>10 | Configured For retains its source-action/body relation. A newly declared source action supplies all five verb forms; typed parameters retain their ordinary binding roles. The current CLI builds one standalone App with core Print and Text; this is compiler analysis, not runtime evidence. |
