# Grammar rule inventory

This mechanical reference preserves the configured English recognizers. It does not claim all recognized forms can be lowered or executed. Read [grammar notation](../spec/grammar.md) before interpreting these tables: ` | ` is sequencing, not an alternative operator. Support notes are separately reviewed. Rules without named cases use their structure directly.

## value

<a id="grammar-value"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>NOUN</code> | <code>VALUE</code> | <code>CASE</code> | partial | [Values and operands](../spec/values.md#declarations-assignment-and-reading-values) | [Note 1](#grammar-note-1) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>definitive</code> | <code>ARTICLE* &#124; [ADJECTIVE]* &#124; ID</code> | [Bound references](../spec/names-and-scope.md#possessives-and-member-lookup) · supported |
| <code>text</code> | <code>`~`</code> | [Text](../spec/values.md#text) · supported |
| <code>number</code> | <code>%~%</code> | [Number](../spec/values.md#number) · supported |
| <code>truth</code> | <code>?~?</code> | [Truth](../spec/values.md#truth) · supported |
| <code>null</code> | <code>!~!</code> | [Null](../spec/values.md#null-and-internal-states) · supported |
| <code>type</code> | <code>ARTICLE* &#124; #~#</code> | [Native type categories](../spec/lexical-types.md#attributes-definitions-and-lexical-categories) · partial |
| <code>list</code> | <code>&amp;~&amp;</code> | [List](../spec/collections-and-categories.md#lists) · partial |
| <code>dict</code> | <code>{~}</code> | [Dictionary](../spec/collections-and-categories.md#dictionaries) · partial |

## subject

<a id="grammar-subject"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>NOUN</code> | <code>SUBJECT</code> | <code>ARTICLE* &#124; [ADJECTIVE]* &#124; CASE</code> | partial | [Subjects and declarations](../spec/names-and-scope.md#declarations-and-definitions) | [Note 2](#grammar-note-2) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>type</code> | <code>"ID" &#124; PHRASE.DECLARATIVE.TYPE</code> | [Typed declarations](../spec/lexical-types.md#attributes-definitions-and-lexical-categories) · partial |
| <code>list</code> | <code>"ID" &#124; PHRASE.DECLARATIVE.LIST</code> | [List declarations](../spec/collections-and-categories.md#lists) · partial |
| <code>dict</code> | <code>"ID" &#124; PHRASE.DECLARATIVE.DICT</code> | [Dictionary declarations](../spec/collections-and-categories.md#dictionaries) · partial |
| <code>declarative</code> | <code>"ID" &#124; PHRASE.DECLARATIVE.DEFINITIVE</code> | [Declarations](../spec/definitions.md#create-define-and-describe) · partial |
| <code>definitional</code> | <code>"ID" &#124; PHRASE.DECLARATIVE.DEFINITIONAL</code> | [Native-backed definitions](../spec/definitions.md#framing-and-escaping) · partial |
| <code>value</code> | <code>"ID" &#124; PHRASE.DECLARATIVE.VALUE</code> | [Value declarations](../spec/values.md#declarations-assignment-and-reading-values) · partial |
| <code>declaration</code> | <code>"ID" &#124; PHRASE.GENERIC*</code> | [Declaration sites](../spec/names-and-scope.md#declarations-and-definitions) · partial |
| <code>definitive</code> | <code>ID</code> | [Bound subjects](../spec/names-and-scope.md#possessives-and-member-lookup) · partial |
| <code>reference</code> | <code>#~#</code> | [Native-category reference](../spec/lexical-types.md#attributes-definitions-and-lexical-categories) · partial |
| <code>definition</code> | <code>ID &#124; PHRASE.GENERIC</code> | [Definition use with relations](../spec/definitions.md#composition-and-source-ownership) · partial |
| <code>directive</code> | <code>"ID" &#124; PHRASE.DIRECTIVE</code> | [Source action declaration](../spec/definitions.md#source-actions) · partial |
| <code>utilization</code> | <code>"ID" &#124; PHRASE.UTILIZATION</code> | [Native/source definition body](../spec/definitions.md#framing-and-escaping) · partial |
| <code>callable</code> | <code>"ID" &#124; PHRASE.DECLARATIVE.CALLABLE</code> | [Typed Functions](../spec/applicational-types.md#functions-properties-and-operations) · supported |
| <code>signature</code> | <code>"ID" &#124; PHRASE.DECLARATIVE.SIGNATURE</code> | [Typed Functions](../spec/applicational-types.md#functions-properties-and-operations) · supported |

## object

<a id="grammar-object"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>NOUN</code> | <code>OBJECT</code> | <code>ARTICLE* &#124; [ADJECTIVE]* &#124; CASE</code> | partial | [Values and operands](../spec/values.md#declarations-assignment-and-reading-values) | [Note 1](#grammar-note-1) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>named</code> | <code>VALUE.TEXT &#124; ID &#124; PHRASE* &#124; CLAUSE*</code> | [Named construction](../spec/definitions.md#descriptors-and-constructors) · partial |
| <code>named_type</code> | <code>VALUE.TEXT &#124; #~# &#124; PHRASE* &#124; CLAUSE*</code> | [Named native-category construction](../spec/definitions.md#descriptors-and-constructors) · partial |
| <code>declarative</code> | <code>"ID" &#124; PHRASE*</code> | [Declared object name](../spec/names-and-scope.md#declarations-and-definitions) · partial |
| <code>definitive</code> | <code>ID &#124; PHRASE* &#124; CLAUSE*</code> | [Bound references](../spec/names-and-scope.md#possessives-and-member-lookup) · supported |
| <code>dependent</code> | <code>ID &#124; PHRASE.DEPENDENCY &#124; CLAUSE*</code> | [Dependency-qualified reference](../spec/imports.md#explicit-source-file-imports) · partial |
| <code>text</code> | <code>`~`</code> | [Text](../spec/values.md#text) · supported |
| <code>number</code> | <code>%~%</code> | [Number](../spec/values.md#number) · supported |
| <code>truth</code> | <code>?~?</code> | [Truth](../spec/values.md#truth) · supported |
| <code>null</code> | <code>!~!</code> | [Null](../spec/values.md#null-and-internal-states) · supported |
| <code>type</code> | <code>#~# &#124; PHRASE* &#124; CLAUSE*</code> | [Native type categories](../spec/lexical-types.md#attributes-definitions-and-lexical-categories) · partial |
| <code>list</code> | <code>&amp;~&amp; &#124; PHRASE* &#124; CLAUSE*</code> | [List](../spec/collections-and-categories.md#lists) · partial |
| <code>dict</code> | <code>{~} &#124; PHRASE* &#124; CLAUSE*</code> | [Dictionary](../spec/collections-and-categories.md#dictionaries) · partial |
| <code>argument</code> | <code>ID</code> | [Bound update operands](../spec/collections-and-categories.md#lists) · partial |
| <code>type_argument</code> | <code>#~#</code> | [Atomic type operands](../spec/collections-and-categories.md#lists) · partial |
| <code>callable</code> | <code>ID &#124; PHRASE.SIGNATURE</code> | [Typed Functions](../spec/applicational-types.md#functions-properties-and-operations) · supported |

## phrase.generic

<a id="grammar-phrase-generic"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>PHRASE</code> | <code>GENERIC</code> | <code>[PREPOSITION] &#124; [OBJECT] &#124; CLAUSE* &#124; PHRASE*</code> | partial | [Phrase association](../spec/grammar.md#phrases) | [Note 3](#grammar-note-3) |

## phrase.directive

<a id="grammar-phrase-directive"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>PHRASE</code> | <code>DIRECTIVE</code> | <code>PREPOSITION.CONJOINMENT.DIRECTIVE &#124; CASE</code> | partial | [Construction and typed arguments](../spec/definitions.md#typed-slots) | [Note 4](#grammar-note-4) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>definitive</code> | <code>[CLAUSE.APPLICATIVE]</code> | [Source action body](../spec/definitions.md#source-actions) · supported |

## phrase.declarative

<a id="grammar-phrase-declarative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>PHRASE</code> | <code>DECLARATIVE</code> | <code>PREPOSITION.DEFINEMENT.DECLARATIVE &#124; CASE</code> | partial | [Construction and typed arguments](../spec/definitions.md#typed-slots) | [Note 4](#grammar-note-4) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>type</code> | <code>OBJECT.TYPE</code> | [Native type categories](../spec/lexical-types.md#attributes-definitions-and-lexical-categories) · partial |
| <code>list</code> | <code>OBJECT.LIST</code> | [List](../spec/collections-and-categories.md#lists) · partial |
| <code>dict</code> | <code>OBJECT.DICT</code> | [Dictionary](../spec/collections-and-categories.md#dictionaries) · partial |
| <code>declarative</code> | <code>OBJECT.DECLARATIVE</code> | [Nested declaration role](../spec/names-and-scope.md#declarations-and-definitions) · partial |
| <code>definitive</code> | <code>OBJECT.DEFINITIVE</code> | [Assigned values](../spec/values.md#declarations-assignment-and-reading-values) · partial |
| <code>definitional</code> | <code>OBJECT.TEXT</code> | [Native definition Text](../spec/definitions.md#framing-and-escaping) · partial |
| <code>value</code> | <code>OBJECT</code> | [Assigned values](../spec/values.md#declarations-assignment-and-reading-values) · partial |
| <code>alias</code> | <code>SUBJECT.DECLARATION</code> | [Aliases](../spec/names-and-scope.md#aliases-and-canonical-identity) · partial |
| <code>callable</code> | <code>OBJECT.CALLABLE</code> | [Typed Functions](../spec/applicational-types.md#functions-properties-and-operations) · supported |
| <code>signature</code> | <code>OBJECT</code> | [Typed Functions](../spec/applicational-types.md#functions-properties-and-operations) · supported |

## phrase.locative

<a id="grammar-phrase-locative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>PHRASE</code> | <code>LOCATIVE</code> | <code>PREPOSITION.DEFINEMENT.LOCATIVE &#124; OBJECT</code> | partial | [Typed runtime access](../spec/collections-and-categories.md#lists) | [Note 5](#grammar-note-5) |

## phrase.objective

<a id="grammar-phrase-objective"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>PHRASE</code> | <code>OBJECTIVE</code> | <code>PREPOSITION.CONJOINMENT.OBJECTIVE &#124; SUBJECT.DEFINITIVE &#124; PHRASE.LOCATIVE* &#124; PHRASE.DECLARATIVE.DECLARATIVE*</code> | partial | [Typed runtime access](../spec/collections-and-categories.md#lists) | [Note 6](#grammar-note-6) |

## phrase.inflective

<a id="grammar-phrase-inflective"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>PHRASE</code> | <code>INFLECTIVE</code> | <code>PREPOSITION.CONSIGNMENT.INFLECTIVE &#124; CASE</code> | partial | [Typed runtime access](../spec/collections-and-categories.md#lists) | [Note 7](#grammar-note-7) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>list</code> | <code>SUBJECT.DEFINITIVE</code> | [Collection association](../spec/collections-and-categories.md#lists) · partial |
| <code>dictionary</code> | <code>SUBJECT.DEFINITIVE &#124; PHRASE.DECLARATIVE.DEFINITIONAL</code> | [Collection association](../spec/collections-and-categories.md#dictionaries) · partial |
| <code>locative</code> | <code>SUBJECT.DEFINITIVE &#124; PHRASE.LOCATIVE</code> | [Collection association](../spec/collections-and-categories.md#lists) · partial |

## phrase.associative

<a id="grammar-phrase-associative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>PHRASE</code> | <code>ASSOCIATIVE</code> | <code>PREPOSITION.CONSIGNMENT.ASSOCIATIVE &#124; CASE</code> | partial | [Phrase association](../spec/grammar.md#phrases) | [Note 3](#grammar-note-3) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>subjective</code> | <code>SUBJECT.DEFINITIVE</code> | [Phrase association](../spec/grammar.md#phrases) · partial |
| <code>objective</code> | <code>OBJECT.DEFINITIVE</code> | [Phrase association](../spec/grammar.md#phrases) · partial |

## phrase.indicative

<a id="grammar-phrase-indicative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>PHRASE</code> | <code>INDICATIVE</code> | <code>PREPOSITION.CONSIGNMENT.INDICATIVE &#124; OBJECT.DEFINITIVE</code> | partial | [Phrase association](../spec/grammar.md#phrases) | [Note 3](#grammar-note-3) |

## phrase.utilization

<a id="grammar-phrase-utilization"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>PHRASE</code> | <code>UTILIZATION</code> | <code>PREPOSITION.CONJOINMENT.SUBJECTIVE &#124; CASE</code> | partial | [Definition framing](../spec/definitions.md#framing-and-escaping) | [Note 8](#grammar-note-8) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>definitive</code> | <code>CLAUSE.IMPERATIVE</code> | [Source action relation](../spec/definitions.md#source-actions) · partial |
| <code>definitional</code> | <code>OBJECT.TEXT</code> | [Native-backed Text definition](../spec/definitions.md#framing-and-escaping) · supported |

## phrase.dependency

<a id="grammar-phrase-dependency"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>PHRASE</code> | <code>DEPENDENCY</code> | <code>PREPOSITION.CONJOINMENT.OBJECTIVE &#124; `~`</code> | partial | [Imports and dependencies](../spec/imports.md#explicit-source-file-imports) | [Note 9](#grammar-note-9) |

## clause.assign.passive

<a id="grammar-clause-assign-passive"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>ASSIGN.PASSIVE</code> | <code>SUBJECT.DECLARATION &#124; PREDICATE.PASSIVE.ASSIGN</code> | partial | [Assignment](../spec/control-flow.md#assignment) | [Note 10](#grammar-note-10) |

## clause.generic.imperative

<a id="grammar-clause-generic-imperative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>GENERIC.IMPERATIVE</code> | <code>[ADVERB]* &#124; VERB &#124; CASE</code> | partial | [Source actions](../spec/definitions.md#source-actions) | [Note 11](#grammar-note-11) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>subjective</code> | <code>SUBJECT</code> | [Bound source-action arguments](../spec/definitions.md#source-actions) · partial |
| <code>objective</code> | <code>[OBJECT]</code> | [Bound source-action arguments](../spec/definitions.md#source-actions) · partial |

## clause.generic.active

<a id="grammar-clause-generic-active"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>GENERIC.ACTIVE</code> | <code>[ADVERB]* &#124; VERB.ACTIVE &#124; OBJECT</code> | partial | [Source actions](../spec/definitions.md#source-actions) | [Note 12](#grammar-note-12) |

## clause.generic.passive

<a id="grammar-clause-generic-passive"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>GENERIC.PASSIVE</code> | <code>[ADVERB]* &#124; VERB.PASSIVE &#124; OBJECT</code> | partial | [Source actions](../spec/definitions.md#source-actions) | [Note 13](#grammar-note-13) |

## clause.generic.applicative

<a id="grammar-clause-generic-applicative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>GENERIC.APPLICATIVE</code> | <code>[ADVERB]* &#124; VERB.APPLICATIVE &#124; CASE</code> | partial | [Source actions](../spec/definitions.md#source-actions) | [Note 14](#grammar-note-14) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>subjective</code> | <code>SUBJECT</code> | [Bound source-action arguments](../spec/definitions.md#source-actions) · partial |
| <code>objective</code> | <code>[OBJECT]</code> | [Bound source-action arguments](../spec/definitions.md#source-actions) · partial |

## clause.declarative.imperative

<a id="grammar-clause-declarative-imperative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>DECLARATIVE.IMPERATIVE</code> | <code>[ADVERB]* &#124; VERB.DECLARATIVE &#124; CASE</code> | partial | [Declarations](../spec/definitions.md#create-define-and-describe) | [Note 15](#grammar-note-15) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>noun</code> | <code>SUBJECT.DECLARATIVE</code> | [Declarations](../spec/definitions.md#create-define-and-describe) · partial |
| <code>verb</code> | <code>SUBJECT.DIRECTIVE</code> | [Declarations](../spec/definitions.md#create-define-and-describe) · partial |
| <code>type</code> | <code>SUBJECT.TYPE</code> | [Declarations](../spec/definitions.md#create-define-and-describe) · partial |
| <code>list</code> | <code>SUBJECT.LIST</code> | [Declarations](../spec/definitions.md#create-define-and-describe) · partial |
| <code>dict</code> | <code>SUBJECT.DICT</code> | [Declarations](../spec/definitions.md#create-define-and-describe) · partial |
| <code>callable</code> | <code>SUBJECT.CALLABLE</code> | [Typed Functions](../spec/applicational-types.md#functions-properties-and-operations) · supported |

## clause.declarative.active

<a id="grammar-clause-declarative-active"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>DECLARATIVE.ACTIVE</code> | <code>[ADVERB]* &#124; VERB.DECLARATIVE.ACTIVE &#124; CASE</code> | partial | [Declarations](../spec/definitions.md#create-define-and-describe) | [Note 16](#grammar-note-16) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>noun</code> | <code>SUBJECT.DECLARATIVE</code> | [Declarations](../spec/definitions.md#create-define-and-describe) · partial |
| <code>verb</code> | <code>SUBJECT.DIRECTIVE</code> | [Declarations](../spec/definitions.md#create-define-and-describe) · partial |
| <code>type</code> | <code>SUBJECT.TYPE</code> | [Declarations](../spec/definitions.md#create-define-and-describe) · partial |
| <code>list</code> | <code>SUBJECT.LIST</code> | [Declarations](../spec/definitions.md#create-define-and-describe) · partial |
| <code>dict</code> | <code>SUBJECT.DICT</code> | [Declarations](../spec/definitions.md#create-define-and-describe) · partial |

## clause.declarative.passive

<a id="grammar-clause-declarative-passive"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>DECLARATIVE.PASSIVE</code> | <code>[ADVERB]* &#124; VERB.DECLARATIVE.PASSIVE &#124; CASE</code> | partial | [Declarations](../spec/definitions.md#create-define-and-describe) | [Note 17](#grammar-note-17) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>noun</code> | <code>PHRASE.DECLARATIVE.DEFINITIVE</code> | [Declarations](../spec/definitions.md#create-define-and-describe) · partial |
| <code>verb</code> | <code>PHRASE.DIRECTIVE</code> | [Declarations](../spec/definitions.md#create-define-and-describe) · partial |

## clause.declarative.applicative

<a id="grammar-clause-declarative-applicative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>DECLARATIVE.APPLICATIVE</code> | <code>[ADVERB]* &#124; VERB.DECLARATIVE.APPLICATIVE &#124; CASE</code> | partial | [Declarations](../spec/definitions.md#create-define-and-describe) | [Note 18](#grammar-note-18) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>noun</code> | <code>SUBJECT.DECLARATIVE</code> | [Declarations](../spec/definitions.md#create-define-and-describe) · partial |
| <code>verb</code> | <code>SUBJECT.DIRECTIVE</code> | [Declarations](../spec/definitions.md#create-define-and-describe) · partial |
| <code>type</code> | <code>SUBJECT.TYPE</code> | [Declarations](../spec/definitions.md#create-define-and-describe) · partial |
| <code>list</code> | <code>SUBJECT.LIST</code> | [Declarations](../spec/definitions.md#create-define-and-describe) · partial |
| <code>dict</code> | <code>SUBJECT.DICT</code> | [Declarations](../spec/definitions.md#create-define-and-describe) · partial |

## clause.definitive.imperative

<a id="grammar-clause-definitive-imperative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>DEFINITIVE.IMPERATIVE</code> | <code>[ADVERB]* &#124; VERB.DEFINITIVE &#124; CASE</code> | partial | [Definitions](../spec/definitions.md#create-define-and-describe) | [Note 19](#grammar-note-19) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>noun</code> | <code>SUBJECT.DEFINITIONAL</code> | [Definitions](../spec/definitions.md#create-define-and-describe) · partial |
| <code>value</code> | <code>SUBJECT.VALUE</code> | [Definitions](../spec/definitions.md#create-define-and-describe) · partial |
| <code>verb</code> | <code>SUBJECT.UTILIZATION</code> | [Definitions](../spec/definitions.md#create-define-and-describe) · partial |
| <code>interface</code> | <code>SUBJECT.DECLARATIVE</code> | [Definitions](../spec/definitions.md#create-define-and-describe) · partial |

## clause.definitive.active

<a id="grammar-clause-definitive-active"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>DEFINITIVE.ACTIVE</code> | <code>[ADVERB]* &#124; VERB.DEFINITIVE.ACTIVE &#124; CASE</code> | partial | [Definitions](../spec/definitions.md#create-define-and-describe) | [Note 20](#grammar-note-20) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>noun</code> | <code>SUBJECT.DEFINITIONAL</code> | [Definitions](../spec/definitions.md#create-define-and-describe) · partial |
| <code>verb</code> | <code>SUBJECT.UTILIZATION</code> | [Definitions](../spec/definitions.md#create-define-and-describe) · partial |

## clause.definitive.passive

<a id="grammar-clause-definitive-passive"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>DEFINITIVE.PASSIVE</code> | <code>[ADVERB]* &#124; VERB.DEFINITIVE.PASSIVE &#124; CASE</code> | partial | [Definitions](../spec/definitions.md#create-define-and-describe) | [Note 21](#grammar-note-21) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>noun</code> | <code>PHRASE.DECLARATIVE.DEFINITIONAL</code> | [Definitions](../spec/definitions.md#create-define-and-describe) · partial |
| <code>verb</code> | <code>PHRASE.UTILIZATION</code> | [Definitions](../spec/definitions.md#create-define-and-describe) · partial |

## clause.definitive.applicative

<a id="grammar-clause-definitive-applicative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>DEFINITIVE.APPLICATIVE</code> | <code>[ADVERB]* &#124; VERB.DEFINITIVE.APPLICATIVE &#124; CASE</code> | partial | [Definitions](../spec/definitions.md#create-define-and-describe) | [Note 22](#grammar-note-22) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>noun</code> | <code>SUBJECT.DEFINITIONAL</code> | [Definitions](../spec/definitions.md#create-define-and-describe) · partial |
| <code>verb</code> | <code>SUBJECT.UTILIZATION</code> | [Definitions](../spec/definitions.md#create-define-and-describe) · partial |

## clause.descriptive.imperative

<a id="grammar-clause-descriptive-imperative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>DESCRIPTIVE.IMPERATIVE</code> | <code>[ADVERB]* &#124; VERB.DESCRIPTIVE &#124; CASE</code> | partial | [Descriptors](../spec/definitions.md#descriptors-and-constructors) | [Note 23](#grammar-note-23) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>adjective</code> | <code>SUBJECT.DECLARATIVE</code> | [Descriptors](../spec/definitions.md#descriptors-and-constructors) · partial |
| <code>adverb</code> | <code>SUBJECT.UTILIZATION</code> | [Descriptors](../spec/definitions.md#descriptors-and-constructors) · partial |

## clause.descriptive.active

<a id="grammar-clause-descriptive-active"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>DESCRIPTIVE.ACTIVE</code> | <code>[ADVERB]* &#124; VERB.DESCRIPTIVE.ACTIVE &#124; CASE</code> | partial | [Descriptors](../spec/definitions.md#descriptors-and-constructors) | [Note 24](#grammar-note-24) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>adjective</code> | <code>OBJECT</code> | [Descriptors](../spec/definitions.md#descriptors-and-constructors) · partial |
| <code>adverb</code> | <code>CLAUSE.ASSOCIATIVE.APPLICATIVE</code> | [Descriptors](../spec/definitions.md#descriptors-and-constructors) · partial |

## clause.descriptive.passive

<a id="grammar-clause-descriptive-passive"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>DESCRIPTIVE.PASSIVE</code> | <code>[ADVERB]* &#124; VERB.DESCRIPTIVE.PASSIVE &#124; CASE</code> | partial | [Descriptors](../spec/definitions.md#descriptors-and-constructors) | [Note 25](#grammar-note-25) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>adjective</code> | <code>PHRASE.DECLARATIVE.DECLARATIVE</code> | [Descriptors](../spec/definitions.md#descriptors-and-constructors) · partial |
| <code>adverb</code> | <code>CLAUSE.ASSOCIATIVE.APPLICATIVE</code> | [Descriptors](../spec/definitions.md#descriptors-and-constructors) · partial |

## clause.descriptive.applicative

<a id="grammar-clause-descriptive-applicative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>DESCRIPTIVE.APPLICATIVE</code> | <code>[ADVERB]* &#124; VERB.DESCRIPTIVE.APPLICATIVE &#124; CASE</code> | partial | [Descriptors](../spec/definitions.md#descriptors-and-constructors) | [Note 26](#grammar-note-26) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>adjective</code> | <code>[OBJECT] &#124; PHRASE.DECLARATIVE.DECLARATIVE*</code> | [Descriptors](../spec/definitions.md#descriptors-and-constructors) · partial |
| <code>adverb</code> | <code>SUBJECT.DEFINITIVE &#124; CLAUSE.ASSOCIATIVE.APPLICATIVE</code> | [Descriptors](../spec/definitions.md#descriptors-and-constructors) · partial |

## clause.expansive.imperative

<a id="grammar-clause-expansive-imperative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>EXPANSIVE.IMPERATIVE</code> | <code>[ADVERB]* &#124; VERB.EXPANSIVE &#124; CASE</code> | partial | [Typed updates](../spec/collections-and-categories.md#lists) | [Note 27](#grammar-note-27) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>reference</code> | <code>OBJECT.ARGUMENT &#124; PHRASE.INFLECTIVE</code> | [Typed updates](../spec/collections-and-categories.md#lists) · partial |
| <code>type</code> | <code>OBJECT.TYPE_ARGUMENT &#124; PHRASE.INFLECTIVE</code> | [Typed updates](../spec/collections-and-categories.md#lists) · partial |
| <code>text</code> | <code>OBJECT.TEXT &#124; PHRASE.INFLECTIVE</code> | [Typed updates](../spec/collections-and-categories.md#lists) · partial |
| <code>number</code> | <code>OBJECT.NUMBER &#124; PHRASE.INFLECTIVE</code> | [Typed updates](../spec/collections-and-categories.md#lists) · partial |
| <code>truth</code> | <code>OBJECT.TRUTH &#124; PHRASE.INFLECTIVE</code> | [Typed updates](../spec/collections-and-categories.md#lists) · partial |
| <code>null</code> | <code>OBJECT.NULL &#124; PHRASE.INFLECTIVE</code> | [Typed updates](../spec/collections-and-categories.md#lists) · partial |

## clause.expansive.active

<a id="grammar-clause-expansive-active"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>EXPANSIVE.ACTIVE</code> | <code>[ADVERB]* &#124; VERB.EXPANSIVE.ACTIVE &#124; CASE</code> | partial | [Typed updates](../spec/collections-and-categories.md#lists) | [Note 27](#grammar-note-27) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>reference</code> | <code>OBJECT.ARGUMENT &#124; PHRASE.INFLECTIVE</code> | [Typed updates](../spec/collections-and-categories.md#lists) · partial |
| <code>type</code> | <code>OBJECT.TYPE_ARGUMENT &#124; PHRASE.INFLECTIVE</code> | [Typed updates](../spec/collections-and-categories.md#lists) · partial |
| <code>text</code> | <code>OBJECT.TEXT &#124; PHRASE.INFLECTIVE</code> | [Typed updates](../spec/collections-and-categories.md#lists) · partial |
| <code>number</code> | <code>OBJECT.NUMBER &#124; PHRASE.INFLECTIVE</code> | [Typed updates](../spec/collections-and-categories.md#lists) · partial |
| <code>truth</code> | <code>OBJECT.TRUTH &#124; PHRASE.INFLECTIVE</code> | [Typed updates](../spec/collections-and-categories.md#lists) · partial |
| <code>null</code> | <code>OBJECT.NULL &#124; PHRASE.INFLECTIVE</code> | [Typed updates](../spec/collections-and-categories.md#lists) · partial |

## clause.expansive.passive

<a id="grammar-clause-expansive-passive"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>EXPANSIVE.PASSIVE</code> | <code>[ADVERB]* &#124; VERB.EXPANSIVE.PASSIVE &#124; CASE</code> | partial | [Collections](../spec/collections-and-categories.md#lists) | [Note 28](#grammar-note-28) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>list</code> | <code>PHRASE.INFLECTIVE.LIST</code> | [Collections](../spec/collections-and-categories.md#lists) · partial |
| <code>dictionary</code> | <code>PHRASE.INFLECTIVE.DICTIONARY</code> | [Collections](../spec/collections-and-categories.md#lists) · partial |

## clause.expansive.applicative

<a id="grammar-clause-expansive-applicative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>EXPANSIVE.APPLICATIVE</code> | <code>[ADVERB]* &#124; VERB.EXPANSIVE.APPLICATIVE &#124; CASE</code> | partial | [Typed updates](../spec/collections-and-categories.md#lists) | [Note 27](#grammar-note-27) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>reference</code> | <code>OBJECT.ARGUMENT &#124; PHRASE.INFLECTIVE</code> | [Typed updates](../spec/collections-and-categories.md#lists) · partial |
| <code>type</code> | <code>OBJECT.TYPE_ARGUMENT &#124; PHRASE.INFLECTIVE</code> | [Typed updates](../spec/collections-and-categories.md#lists) · partial |
| <code>text</code> | <code>OBJECT.TEXT &#124; PHRASE.INFLECTIVE</code> | [Typed updates](../spec/collections-and-categories.md#lists) · partial |
| <code>number</code> | <code>OBJECT.NUMBER &#124; PHRASE.INFLECTIVE</code> | [Typed updates](../spec/collections-and-categories.md#lists) · partial |
| <code>truth</code> | <code>OBJECT.TRUTH &#124; PHRASE.INFLECTIVE</code> | [Typed updates](../spec/collections-and-categories.md#lists) · partial |
| <code>null</code> | <code>OBJECT.NULL &#124; PHRASE.INFLECTIVE</code> | [Typed updates](../spec/collections-and-categories.md#lists) · partial |

## clause.locative.imperative

<a id="grammar-clause-locative-imperative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>LOCATIVE.IMPERATIVE</code> | <code>[ADVERB]* &#124; VERB.LOCATIVE &#124; CASE</code> | partial | [Typed updates](../spec/collections-and-categories.md#lists) | [Note 27](#grammar-note-27) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>reference</code> | <code>OBJECT.ARGUMENT &#124; PHRASE.INFLECTIVE</code> | [Typed updates](../spec/collections-and-categories.md#lists) · partial |
| <code>type</code> | <code>OBJECT.TYPE_ARGUMENT &#124; PHRASE.INFLECTIVE</code> | [Typed updates](../spec/collections-and-categories.md#lists) · partial |
| <code>text</code> | <code>OBJECT.TEXT &#124; PHRASE.INFLECTIVE</code> | [Typed updates](../spec/collections-and-categories.md#lists) · partial |
| <code>number</code> | <code>OBJECT.NUMBER &#124; PHRASE.INFLECTIVE</code> | [Typed updates](../spec/collections-and-categories.md#lists) · partial |
| <code>truth</code> | <code>OBJECT.TRUTH &#124; PHRASE.INFLECTIVE</code> | [Typed updates](../spec/collections-and-categories.md#lists) · partial |
| <code>null</code> | <code>OBJECT.NULL &#124; PHRASE.INFLECTIVE</code> | [Typed updates](../spec/collections-and-categories.md#lists) · partial |

## clause.locative.active

<a id="grammar-clause-locative-active"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>LOCATIVE.ACTIVE</code> | <code>[ADVERB]* &#124; VERB.LOCATIVE.ACTIVE &#124; CASE</code> | partial | [Typed updates](../spec/collections-and-categories.md#lists) | [Note 27](#grammar-note-27) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>reference</code> | <code>OBJECT.ARGUMENT &#124; PHRASE.INFLECTIVE</code> | [Typed updates](../spec/collections-and-categories.md#lists) · partial |
| <code>type</code> | <code>OBJECT.TYPE_ARGUMENT &#124; PHRASE.INFLECTIVE</code> | [Typed updates](../spec/collections-and-categories.md#lists) · partial |
| <code>text</code> | <code>OBJECT.TEXT &#124; PHRASE.INFLECTIVE</code> | [Typed updates](../spec/collections-and-categories.md#lists) · partial |
| <code>number</code> | <code>OBJECT.NUMBER &#124; PHRASE.INFLECTIVE</code> | [Typed updates](../spec/collections-and-categories.md#lists) · partial |
| <code>truth</code> | <code>OBJECT.TRUTH &#124; PHRASE.INFLECTIVE</code> | [Typed updates](../spec/collections-and-categories.md#lists) · partial |
| <code>null</code> | <code>OBJECT.NULL &#124; PHRASE.INFLECTIVE</code> | [Typed updates](../spec/collections-and-categories.md#lists) · partial |

## clause.locative.passive

<a id="grammar-clause-locative-passive"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>LOCATIVE.PASSIVE</code> | <code>[ADVERB]* &#124; VERB.LOCATIVE.PASSIVE &#124; CASE</code> | partial | [Collections](../spec/collections-and-categories.md#lists) | [Note 28](#grammar-note-28) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>list</code> | <code>PHRASE.INFLECTIVE.LOCATIVE</code> | [Collections](../spec/collections-and-categories.md#lists) · partial |
| <code>dictionary</code> | <code>PHRASE.INFLECTIVE.DICTIONARY</code> | [Collections](../spec/collections-and-categories.md#lists) · partial |

## clause.locative.applicative

<a id="grammar-clause-locative-applicative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>LOCATIVE.APPLICATIVE</code> | <code>[ADVERB]* &#124; VERB.LOCATIVE.APPLICATIVE &#124; CASE</code> | partial | [Typed updates](../spec/collections-and-categories.md#lists) | [Note 27](#grammar-note-27) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>reference</code> | <code>OBJECT.ARGUMENT &#124; PHRASE.INFLECTIVE</code> | [Typed updates](../spec/collections-and-categories.md#lists) · partial |
| <code>type</code> | <code>OBJECT.TYPE_ARGUMENT &#124; PHRASE.INFLECTIVE</code> | [Typed updates](../spec/collections-and-categories.md#lists) · partial |
| <code>text</code> | <code>OBJECT.TEXT &#124; PHRASE.INFLECTIVE</code> | [Typed updates](../spec/collections-and-categories.md#lists) · partial |
| <code>number</code> | <code>OBJECT.NUMBER &#124; PHRASE.INFLECTIVE</code> | [Typed updates](../spec/collections-and-categories.md#lists) · partial |
| <code>truth</code> | <code>OBJECT.TRUTH &#124; PHRASE.INFLECTIVE</code> | [Typed updates](../spec/collections-and-categories.md#lists) · partial |
| <code>null</code> | <code>OBJECT.NULL &#124; PHRASE.INFLECTIVE</code> | [Typed updates](../spec/collections-and-categories.md#lists) · partial |

## clause.receptive.imperative

<a id="grammar-clause-receptive-imperative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>RECEPTIVE.IMPERATIVE</code> | <code>[ADVERB]* &#124; VERB.RECEPTIVE &#124; CASE &#124; PHRASE.OBJECTIVE</code> | partial | [Collections](../spec/collections-and-categories.md#lists) | [Note 29](#grammar-note-29) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>list</code> | <code>OBJECT.NUMBER</code> | [Collections](../spec/collections-and-categories.md#lists) · partial |
| <code>dictionary</code> | <code>OBJECT.TEXT</code> | [Collections](../spec/collections-and-categories.md#lists) · partial |

## clause.receptive.active

<a id="grammar-clause-receptive-active"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>RECEPTIVE.ACTIVE</code> | <code>[ADVERB]* &#124; VERB.RECEPTIVE.ACTIVE &#124; CASE &#124; PHRASE.OBJECTIVE</code> | partial | [Collections](../spec/collections-and-categories.md#lists) | [Note 30](#grammar-note-30) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>list</code> | <code>OBJECT.NUMBER</code> | [Collections](../spec/collections-and-categories.md#lists) · partial |
| <code>dictionary</code> | <code>OBJECT.TEXT</code> | [Collections](../spec/collections-and-categories.md#lists) · partial |

## clause.receptive.passive

<a id="grammar-clause-receptive-passive"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>RECEPTIVE.PASSIVE</code> | <code>[ADVERB]* &#124; VERB.RECEPTIVE.PASSIVE &#124; PHRASE.OBJECTIVE</code> | partial | [Collections](../spec/collections-and-categories.md#lists) | [Note 28](#grammar-note-28) |

## clause.receptive.applicative

<a id="grammar-clause-receptive-applicative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>RECEPTIVE.APPLICATIVE</code> | <code>[ADVERB]* &#124; VERB.RECEPTIVE.APPLICATIVE &#124; CASE &#124; PHRASE.OBJECTIVE</code> | partial | [Collections](../spec/collections-and-categories.md#lists) | [Note 31](#grammar-note-31) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>list</code> | <code>OBJECT.NUMBER</code> | [Collections](../spec/collections-and-categories.md#lists) · partial |
| <code>dictionary</code> | <code>OBJECT.TEXT</code> | [Collections](../spec/collections-and-categories.md#lists) · partial |

## clause.reductive.imperative

<a id="grammar-clause-reductive-imperative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>REDUCTIVE.IMPERATIVE</code> | <code>[ADVERB]* &#124; VERB.RECEPTIVE &#124; CASE &#124; PHRASE.OBJECTIVE</code> | partial | [Collections](../spec/collections-and-categories.md#lists) | [Note 29](#grammar-note-29) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>list</code> | <code>OBJECT.NUMBER</code> | [Collections](../spec/collections-and-categories.md#lists) · partial |
| <code>dictionary</code> | <code>OBJECT.TEXT</code> | [Collections](../spec/collections-and-categories.md#lists) · partial |

## clause.reductive.active

<a id="grammar-clause-reductive-active"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>REDUCTIVE.ACTIVE</code> | <code>[ADVERB]* &#124; VERB.RECEPTIVE.ACTIVE &#124; CASE &#124; PHRASE.OBJECTIVE</code> | partial | [Collections](../spec/collections-and-categories.md#lists) | [Note 30](#grammar-note-30) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>list</code> | <code>OBJECT.NUMBER</code> | [Collections](../spec/collections-and-categories.md#lists) · partial |
| <code>dictionary</code> | <code>OBJECT.TEXT</code> | [Collections](../spec/collections-and-categories.md#lists) · partial |

## clause.reductive.passive

<a id="grammar-clause-reductive-passive"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>REDUCTIVE.PASSIVE</code> | <code>[ADVERB]* &#124; VERB.RECEPTIVE.PASSIVE &#124; PHRASE.OBJECTIVE</code> | partial | [Collections](../spec/collections-and-categories.md#lists) | [Note 28](#grammar-note-28) |

## clause.reductive.applicative

<a id="grammar-clause-reductive-applicative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>REDUCTIVE.APPLICATIVE</code> | <code>[ADVERB]* &#124; VERB.RECEPTIVE.APPLICATIVE &#124; CASE &#124; PHRASE.OBJECTIVE</code> | partial | [Collections](../spec/collections-and-categories.md#lists) | [Note 31](#grammar-note-31) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>list</code> | <code>OBJECT.NUMBER</code> | [Collections](../spec/collections-and-categories.md#lists) · partial |
| <code>dictionary</code> | <code>OBJECT.TEXT</code> | [Collections](../spec/collections-and-categories.md#lists) · partial |

## clause.associative.imperative

<a id="grammar-clause-associative-imperative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>ASSOCIATIVE.IMPERATIVE</code> | <code>[ADVERB]* &#124; VERB.ASSOCIATIVE &#124; SUBJECT.DEFINITIVE &#124; CASE*</code> | partial | [Aliases](../spec/names-and-scope.md#aliases-and-canonical-identity) | [Note 32](#grammar-note-32) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>alias</code> | <code>PHRASE.DECLARATIVE.ALIAS</code> | [Aliases](../spec/names-and-scope.md#aliases-and-canonical-identity) · partial |

## clause.associative.active

<a id="grammar-clause-associative-active"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>ASSOCIATIVE.ACTIVE</code> | <code>[ADVERB]* &#124; VERB.ASSOCIATIVE.ACTIVE &#124; SUBJECT.DEFINITIVE &#124; CASE*</code> | partial | [Aliases](../spec/names-and-scope.md#aliases-and-canonical-identity) | [Note 33](#grammar-note-33) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>alias</code> | <code>PHRASE.DECLARATIVE.ALIAS</code> | [Aliases](../spec/names-and-scope.md#aliases-and-canonical-identity) · partial |

## clause.associative.passive

<a id="grammar-clause-associative-passive"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>ASSOCIATIVE.PASSIVE</code> | <code>[ADVERB]* &#124; VERB.ASSOCIATIVE.PASSIVE &#124; CASE</code> | partial | [Aliases](../spec/names-and-scope.md#aliases-and-canonical-identity) | [Note 34](#grammar-note-34) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>definitive</code> | <code>PHRASE.DECLARATIVE</code> | [Aliases](../spec/names-and-scope.md#aliases-and-canonical-identity) · partial |
| <code>associative</code> | <code>PHRASE.INDICATIVE</code> | [Aliases](../spec/names-and-scope.md#aliases-and-canonical-identity) · partial |

## clause.associative.applicative

<a id="grammar-clause-associative-applicative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>ASSOCIATIVE.APPLICATIVE</code> | <code>[ADVERB]* &#124; VERB.ASSOCIATIVE.APPLICATIVE &#124; OBJECT &#124; CASE*</code> | partial | [Aliases](../spec/names-and-scope.md#aliases-and-canonical-identity) | [Note 35](#grammar-note-35) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>alias</code> | <code>PHRASE.DECLARATIVE.ALIAS</code> | [Aliases](../spec/names-and-scope.md#aliases-and-canonical-identity) · partial |

## clause.generative.imperative

<a id="grammar-clause-generative-imperative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>GENERATIVE.IMPERATIVE</code> | <code>[ADVERB]* &#124; VERB.GENERATIVE &#124; OBJECT</code> | partial | [Build](../spec/execution.md#build-run-stop-and-await) | [Note 36](#grammar-note-36) |

## clause.generative.active

<a id="grammar-clause-generative-active"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>GENERATIVE.ACTIVE</code> | <code>[ADVERB]* &#124; VERB.GENERATIVE.ACTIVE &#124; OBJECT</code> | partial | [Build](../spec/execution.md#build-run-stop-and-await) | [Note 37](#grammar-note-37) |

## clause.generative.passive

<a id="grammar-clause-generative-passive"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>GENERATIVE.PASSIVE</code> | <code>[ADVERB]* &#124; VERB.GENERATIVE.PASSIVE</code> | partial | [Build](../spec/execution.md#build-run-stop-and-await) | [Note 38](#grammar-note-38) |

## clause.generative.applicative

<a id="grammar-clause-generative-applicative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>GENERATIVE.APPLICATIVE</code> | <code>[ADVERB]* &#124; VERB.GENERATIVE.APPLICATIVE &#124; OBJECT</code> | partial | [Build](../spec/execution.md#build-run-stop-and-await) | [Note 39](#grammar-note-39) |

## clause.executive.imperative

<a id="grammar-clause-executive-imperative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>EXECUTIVE.IMPERATIVE</code> | <code>[ADVERB]* &#124; VERB.EXECUTIVE &#124; OBJECT</code> | partial | [Run](../spec/execution.md#build-run-stop-and-await) | [Note 40](#grammar-note-40) |

## clause.executive.active

<a id="grammar-clause-executive-active"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>EXECUTIVE.ACTIVE</code> | <code>[ADVERB]* &#124; VERB.EXECUTIVE.ACTIVE &#124; OBJECT</code> | partial | [Run](../spec/execution.md#build-run-stop-and-await) | [Note 41](#grammar-note-41) |

## clause.executive.passive

<a id="grammar-clause-executive-passive"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>EXECUTIVE.PASSIVE</code> | <code>[ADVERB]* &#124; VERB.EXECUTIVE.PASSIVE</code> | partial | [Run](../spec/execution.md#build-run-stop-and-await) | [Note 42](#grammar-note-42) |

## clause.executive.applicative

<a id="grammar-clause-executive-applicative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>EXECUTIVE.APPLICATIVE</code> | <code>[ADVERB]* &#124; VERB.EXECUTIVE.APPLICATIVE &#124; OBJECT</code> | partial | [Run](../spec/execution.md#build-run-stop-and-await) | [Note 43](#grammar-note-43) |

## clause.suspensive.imperative

<a id="grammar-clause-suspensive-imperative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>SUSPENSIVE.IMPERATIVE</code> | <code>[ADVERB]* &#124; VERB.SUSPENSIVE &#124; OBJECT</code> | partial | [Await](../spec/execution.md#continuations-and-activation) | [Note 44](#grammar-note-44) |

## clause.suspensive.active

<a id="grammar-clause-suspensive-active"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>SUSPENSIVE.ACTIVE</code> | <code>[ADVERB]* &#124; VERB.SUSPENSIVE.ACTIVE &#124; OBJECT</code> | partial | [Await](../spec/execution.md#continuations-and-activation) | [Note 45](#grammar-note-45) |

## clause.suspensive.passive

<a id="grammar-clause-suspensive-passive"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>SUSPENSIVE.PASSIVE</code> | <code>[ADVERB]* &#124; VERB.SUSPENSIVE.PASSIVE</code> | partial | [Await](../spec/execution.md#continuations-and-activation) | [Note 46](#grammar-note-46) |

## clause.suspensive.applicative

<a id="grammar-clause-suspensive-applicative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>SUSPENSIVE.APPLICATIVE</code> | <code>[ADVERB]* &#124; VERB.SUSPENSIVE.APPLICATIVE &#124; OBJECT</code> | partial | [Await](../spec/execution.md#continuations-and-activation) | [Note 47](#grammar-note-47) |

## clause.interruptive.imperative

<a id="grammar-clause-interruptive-imperative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>INTERRUPTIVE.IMPERATIVE</code> | <code>[ADVERB]* &#124; VERB.INTERRUPTIVE &#124; OBJECT</code> | partial | [Stop](../spec/execution.md#build-run-stop-and-await) | [Note 48](#grammar-note-48) |

## clause.interruptive.active

<a id="grammar-clause-interruptive-active"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>INTERRUPTIVE.ACTIVE</code> | <code>[ADVERB]* &#124; VERB.INTERRUPTIVE.ACTIVE &#124; OBJECT</code> | partial | [Stop](../spec/execution.md#build-run-stop-and-await) | [Note 49](#grammar-note-49) |

## clause.interruptive.passive

<a id="grammar-clause-interruptive-passive"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>INTERRUPTIVE.PASSIVE</code> | <code>[ADVERB]* &#124; VERB.INTERRUPTIVE.PASSIVE</code> | partial | [Stop](../spec/execution.md#build-run-stop-and-await) | [Note 50](#grammar-note-50) |

## clause.interruptive.applicative

<a id="grammar-clause-interruptive-applicative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>INTERRUPTIVE.APPLICATIVE</code> | <code>[ADVERB]* &#124; VERB.INTERRUPTIVE.APPLICATIVE &#124; OBJECT</code> | partial | [Stop](../spec/execution.md#build-run-stop-and-await) | [Note 51](#grammar-note-51) |

## clause.inflective.imperative

<a id="grammar-clause-inflective-imperative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>INFLECTIVE.IMPERATIVE</code> | <code>[ADVERB]* &#124; VERB.INFLECTIVE &#124; OBJECT</code> | partial | [Import and export vocabulary](../spec/imports.md#indexed-dictionaries) | [Note 52](#grammar-note-52) |

## clause.inflective.active

<a id="grammar-clause-inflective-active"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>INFLECTIVE.ACTIVE</code> | <code>[ADVERB]* &#124; VERB.INFLECTIVE.ACTIVE &#124; OBJECT</code> | partial | [Import and export vocabulary](../spec/imports.md#indexed-dictionaries) | [Note 53](#grammar-note-53) |

## clause.inflective.passive

<a id="grammar-clause-inflective-passive"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>INFLECTIVE.PASSIVE</code> | <code>[ADVERB]* &#124; VERB.INFLECTIVE.PASSIVE</code> | partial | [Import and export vocabulary](../spec/imports.md#indexed-dictionaries) | [Note 54](#grammar-note-54) |

## clause.inflective.applicative

<a id="grammar-clause-inflective-applicative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>INFLECTIVE.APPLICATIVE</code> | <code>[ADVERB]* &#124; VERB.INFLECTIVE.APPLICATIVE &#124; OBJECT</code> | partial | [Import and export vocabulary](../spec/imports.md#indexed-dictionaries) | [Note 55](#grammar-note-55) |

## clause.deflective.imperative

<a id="grammar-clause-deflective-imperative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>DEFLECTIVE.IMPERATIVE</code> | <code>[ADVERB]* &#124; VERB.DEFLECTIVE &#124; OBJECT</code> | partial | [Import and export vocabulary](../spec/imports.md#indexed-dictionaries) | [Note 52](#grammar-note-52) |

## clause.deflective.active

<a id="grammar-clause-deflective-active"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>DEFLECTIVE.ACTIVE</code> | <code>[ADVERB]* &#124; VERB.DEFLECTIVE.ACTIVE &#124; OBJECT</code> | partial | [Import and export vocabulary](../spec/imports.md#indexed-dictionaries) | [Note 53](#grammar-note-53) |

## clause.deflective.passive

<a id="grammar-clause-deflective-passive"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>DEFLECTIVE.PASSIVE</code> | <code>[ADVERB]* &#124; VERB.DEFLECTIVE.PASSIVE</code> | partial | [Import and export vocabulary](../spec/imports.md#indexed-dictionaries) | [Note 54](#grammar-note-54) |

## clause.deflective.applicative

<a id="grammar-clause-deflective-applicative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>DEFLECTIVE.APPLICATIVE</code> | <code>[ADVERB]* &#124; VERB.DEFLECTIVE.APPLICATIVE &#124; OBJECT</code> | partial | [Import and export vocabulary](../spec/imports.md#indexed-dictionaries) | [Note 55](#grammar-note-55) |

## clause.observative.imperative

<a id="grammar-clause-observative-imperative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>OBSERVATIVE.IMPERATIVE</code> | <code>[ADVERB]* &#124; VERB.OBSERVATIVE &#124; OBJECT</code> | partial | [Show](../spec/control-flow.md#show-and-output) | [Note 56](#grammar-note-56) |

## clause.observative.active

<a id="grammar-clause-observative-active"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>OBSERVATIVE.ACTIVE</code> | <code>[ADVERB]* &#124; VERB.OBSERVATIVE.ACTIVE &#124; OBJECT</code> | partial | [Show](../spec/control-flow.md#show-and-output) | [Note 57](#grammar-note-57) |

## clause.observative.passive

<a id="grammar-clause-observative-passive"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>OBSERVATIVE.PASSIVE</code> | <code>[ADVERB]* &#124; VERB.OBSERVATIVE.PASSIVE</code> | partial | [Show](../spec/control-flow.md#show-and-output) | [Note 58](#grammar-note-58) |

## clause.observative.applicative

<a id="grammar-clause-observative-applicative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>OBSERVATIVE.APPLICATIVE</code> | <code>[ADVERB]* &#124; VERB.OBSERVATIVE.APPLICATIVE &#124; OBJECT</code> | partial | [Show](../spec/control-flow.md#show-and-output) | [Note 59](#grammar-note-59) |

## clause.terminative.imperative

<a id="grammar-clause-terminative-imperative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>TERMINATIVE.IMPERATIVE</code> | <code>[ADVERB]* &#124; VERB.TERMINATIVE &#124; CASE*</code> | partial | [Exit](../spec/control-flow.md#exit-and-return) | [Note 60](#grammar-note-60) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>objective</code> | <code>OBJECT</code> | [Exit](../spec/control-flow.md#exit-and-return) · partial |
| <code>associative</code> | <code>PHRASE.ASSOCIATIVE</code> | [Exit](../spec/control-flow.md#exit-and-return) · partial |

## clause.terminative.active

<a id="grammar-clause-terminative-active"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>TERMINATIVE.ACTIVE</code> | <code>[ADVERB]* &#124; VERB.TERMINATIVE.ACTIVE &#124; CASE</code> | partial | [Exit](../spec/control-flow.md#exit-and-return) | [Note 61](#grammar-note-61) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>objective</code> | <code>OBJECT</code> | [Exit](../spec/control-flow.md#exit-and-return) · partial |
| <code>associative</code> | <code>PHRASE.ASSOCIATIVE</code> | [Exit](../spec/control-flow.md#exit-and-return) · partial |

## clause.terminative.passive

<a id="grammar-clause-terminative-passive"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>TERMINATIVE.PASSIVE</code> | <code>[ADVERB]* &#124; VERB.TERMINATIVE.PASSIVE &#124; CASE</code> | partial | [Exit](../spec/control-flow.md#exit-and-return) | [Note 62](#grammar-note-62) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>objective</code> | <code>OBJECT</code> | [Exit](../spec/control-flow.md#exit-and-return) · partial |
| <code>associative</code> | <code>PHRASE.ASSOCIATIVE</code> | [Exit](../spec/control-flow.md#exit-and-return) · partial |

## clause.terminative.applicative

<a id="grammar-clause-terminative-applicative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>TERMINATIVE.APPLICATIVE</code> | <code>[ADVERB]* &#124; VERB.TERMINATIVE.APPLICATIVE &#124; CASE</code> | partial | [Exit](../spec/control-flow.md#exit-and-return) | [Note 63](#grammar-note-63) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>objective</code> | <code>OBJECT</code> | [Exit](../spec/control-flow.md#exit-and-return) · partial |
| <code>associative</code> | <code>PHRASE.ASSOCIATIVE</code> | [Exit](../spec/control-flow.md#exit-and-return) · partial |

## condition

<a id="grammar-condition"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CONDITION</code> | <code>CONDITION</code> | <code>CASE</code> | partial | [Conditions](../spec/control-flow.md#comparisons) | [Note 64](#grammar-note-64) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>truth</code> | <code>ADVERB.NULLIFICATION* &#124; VALUE</code> | [Truth condition](../spec/values.md#truth) · partial |
| <code>has</code> | <code>VALUE &#124; VERB.APPROPRIATIVE.ACTIVE &#124; ARTICLE.NONE* &#124; VALUE</code> | [Process Error condition](../spec/execution.md#output-and-error) · partial |
| <code>equals</code> | <code>VALUE &#124; VERB.COMPARATIVE.ACTIVE &#124; ARTICLE.NONE* &#124; VALUE</code> | [Scalar comparison](../spec/control-flow.md#comparisons) · partial |
| <code>is</code> | <code>VALUE &#124; VERB.PASSIVE &#124; ARTICLE.NONE* &#124; VALUE</code> | [Scalar comparison](../spec/control-flow.md#comparisons) · partial |
| <code>isnt</code> | <code>VALUE &#124; VERB.PASSIVE.NULL &#124; VALUE</code> | [Scalar comparison](../spec/control-flow.md#comparisons) · partial |
| <code>less</code> | <code>VALUE &#124; ARTICLE.NONE* &#124; ADVERB.COMPARISON.SUBTRACTION &#124; PREPOSITION.CONJOINMENT.COMPARATIVE &#124; VALUE</code> | [Scalar comparison](../spec/control-flow.md#comparisons) · partial |
| <code>greater</code> | <code>VALUE &#124; ARTICLE.NONE* &#124; ADVERB.COMPARISON.MULTIPLICATION &#124; PREPOSITION.CONJOINMENT.COMPARATIVE &#124; VALUE</code> | [Scalar comparison](../spec/control-flow.md#comparisons) · partial |
| <code>less_equal</code> | <code>VALUE &#124; ARTICLE.NONE* &#124; ADVERB.COMPARISON.SUBTRACTION &#124; PREPOSITION.CONJOINMENT.COMPARATIVE &#124; VERB.COMPARATIVE &#124; PREPOSITION.CONJOINMENT.SUBJECTIVE &#124; VALUE</code> | [Scalar comparison](../spec/control-flow.md#comparisons) · partial |
| <code>greater_equal</code> | <code>VALUE &#124; ARTICLE.NONE* &#124; ADVERB.COMPARISON.MULTIPLICATION &#124; PREPOSITION.CONJOINMENT.COMPARATIVE &#124; VERB.COMPARATIVE &#124; PREPOSITION.CONJOINMENT.SUBJECTIVE &#124; VALUE</code> | [Scalar comparison](../spec/control-flow.md#comparisons) · partial |

## conditional.if

<a id="grammar-conditional-if"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CONDITIONAL</code> | <code>IF</code> | <code>EVALUATION.CONDITIONAL &#124; [CONDITION] &#124; [CLAUSE.IMPERATIVE]</code> | partial | [Conditions](../spec/control-flow.md#comparisons) | [Note 64](#grammar-note-64) |

## conditional.else

<a id="grammar-conditional-else"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CONDITIONAL</code> | <code>ELSE</code> | <code>EVALUATION.EXCEPTIONAL &#124; CASE</code> | partial | [Conditions](../spec/control-flow.md#comparisons) | [Note 64](#grammar-note-64) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>if</code> | <code>CONDITIONAL.IF</code> | [Conditions](../spec/control-flow.md#comparisons) · partial |
| <code>else</code> | <code>[CLAUSE.IMPERATIVE]</code> | [Conditions](../spec/control-flow.md#comparisons) · partial |

## conditional.while

<a id="grammar-conditional-while"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CONDITIONAL</code> | <code>WHILE</code> | <code>EVALUATION.ITERATIVE &#124; [CONDITION] &#124; CLAUSE.IMPERATIVE</code> | rejected | [Iteration rejection](../spec/control-flow.md#iteration-and-collection-operations) | [Note 65](#grammar-note-65) |

## conditional.other

<a id="grammar-conditional-other"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CONDITIONAL</code> | <code>OTHER</code> | <code>EVALUATION.ALTERNATIVE &#124; [CLAUSE.IMPERATIVE]</code> | partial | [Conditions](../spec/control-flow.md#comparisons) | [Note 64](#grammar-note-64) |

## predicate.imperative

<a id="grammar-predicate-imperative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>PREDICATE</code> | <code>IMPERATIVE</code> | <code>CASE &#124; [CLAUSE]*</code> | partial | [Predicates and statements](../spec/grammar.md#predicates-and-statements) | [Note 66](#grammar-note-66) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>generic</code> | <code>CLAUSE.GENERIC.IMPERATIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>declarative</code> | <code>CLAUSE.DECLARATIVE.IMPERATIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>definitive</code> | <code>CLAUSE.DEFINITIVE.IMPERATIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>descriptive</code> | <code>CLAUSE.DESCRIPTIVE.IMPERATIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>expansive</code> | <code>CLAUSE.EXPANSIVE.IMPERATIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>locative</code> | <code>CLAUSE.LOCATIVE.IMPERATIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>receptive</code> | <code>CLAUSE.RECEPTIVE.IMPERATIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>reductive</code> | <code>CLAUSE.REDUCTIVE.IMPERATIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>associative</code> | <code>CLAUSE.ASSOCIATIVE.IMPERATIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>generative</code> | <code>CLAUSE.GENERATIVE.IMPERATIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>executive</code> | <code>CLAUSE.EXECUTIVE.IMPERATIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>interruptive</code> | <code>CLAUSE.INTERRUPTIVE.IMPERATIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>inflective</code> | <code>CLAUSE.INFLECTIVE.IMPERATIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>deflective</code> | <code>CLAUSE.DEFLECTIVE.IMPERATIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>observative</code> | <code>CLAUSE.OBSERVATIVE.IMPERATIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>terminative</code> | <code>CLAUSE.TERMINATIVE.IMPERATIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>suspensive</code> | <code>CLAUSE.SUSPENSIVE.IMPERATIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>returning</code> | <code>CLAUSE.RETURNING.IMPERATIVE</code> | [Typed Functions](../spec/applicational-types.md#functions-properties-and-operations) · supported |
| <code>invocative</code> | <code>CLAUSE.INVOCATIVE.IMPERATIVE</code> | [Typed Functions](../spec/applicational-types.md#functions-properties-and-operations) · supported |

## predicate.active

<a id="grammar-predicate-active"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>PREDICATE</code> | <code>ACTIVE</code> | <code>CASE &#124; [CLAUSE]*</code> | partial | [Predicates and statements](../spec/grammar.md#predicates-and-statements) | [Note 66](#grammar-note-66) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>generic</code> | <code>CLAUSE.GENERIC.ACTIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>declarative</code> | <code>CLAUSE.DECLARATIVE.ACTIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>definitive</code> | <code>CLAUSE.DEFINITIVE.ACTIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>descriptive</code> | <code>CLAUSE.DESCRIPTIVE.ACTIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>expansive</code> | <code>CLAUSE.EXPANSIVE.ACTIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>locative</code> | <code>CLAUSE.LOCATIVE.ACTIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>receptive</code> | <code>CLAUSE.RECEPTIVE.ACTIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>reductive</code> | <code>CLAUSE.REDUCTIVE.ACTIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>associative</code> | <code>CLAUSE.ASSOCIATIVE.ACTIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>generative</code> | <code>CLAUSE.GENERATIVE.ACTIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>executive</code> | <code>CLAUSE.EXECUTIVE.ACTIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>interruptive</code> | <code>CLAUSE.INTERRUPTIVE.ACTIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>inflective</code> | <code>CLAUSE.INFLECTIVE.ACTIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>deflective</code> | <code>CLAUSE.DEFLECTIVE.ACTIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>observative</code> | <code>CLAUSE.OBSERVATIVE.ACTIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>terminative</code> | <code>CLAUSE.TERMINATIVE.ACTIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>suspensive</code> | <code>CLAUSE.SUSPENSIVE.ACTIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |

## predicate.passive

<a id="grammar-predicate-passive"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>PREDICATE</code> | <code>PASSIVE</code> | <code>VERB.PASSIVE &#124; CASE &#124; [CLAUSE]*</code> | partial | [Predicates and statements](../spec/grammar.md#predicates-and-statements) | [Note 66](#grammar-note-66) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>assign</code> | <code>OBJECT</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>generic</code> | <code>CLAUSE.GENERIC.PASSIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>declarative</code> | <code>CLAUSE.DECLARATIVE.PASSIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>definitive</code> | <code>CLAUSE.DEFINITIVE.PASSIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>descriptive</code> | <code>CLAUSE.DESCRIPTIVE.PASSIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>expansive</code> | <code>CLAUSE.EXPANSIVE.PASSIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>locative</code> | <code>CLAUSE.LOCATIVE.PASSIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>receptive</code> | <code>CLAUSE.RECEPTIVE.PASSIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>reductive</code> | <code>CLAUSE.REDUCTIVE.PASSIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>associative</code> | <code>CLAUSE.ASSOCIATIVE.PASSIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>generative</code> | <code>CLAUSE.GENERATIVE.PASSIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>executive</code> | <code>CLAUSE.EXECUTIVE.PASSIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>interruptive</code> | <code>CLAUSE.INTERRUPTIVE.PASSIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>inflective</code> | <code>CLAUSE.INFLECTIVE.PASSIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>deflective</code> | <code>CLAUSE.DEFLECTIVE.PASSIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>observative</code> | <code>CLAUSE.OBSERVATIVE.PASSIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>terminative</code> | <code>CLAUSE.TERMINATIVE.PASSIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>suspensive</code> | <code>CLAUSE.SUSPENSIVE.PASSIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |

## predicate.applicative

<a id="grammar-predicate-applicative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>PREDICATE</code> | <code>APPLICATIVE</code> | <code>CASE &#124; [CLAUSE]*</code> | partial | [Predicates and statements](../spec/grammar.md#predicates-and-statements) | [Note 66](#grammar-note-66) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>generic</code> | <code>CLAUSE.GENERIC.APPLICATIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>declarative</code> | <code>CLAUSE.DECLARATIVE.APPLICATIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>definitive</code> | <code>CLAUSE.DEFINITIVE.APPLICATIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>descriptive</code> | <code>CLAUSE.DESCRIPTIVE.APPLICATIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>expansive</code> | <code>CLAUSE.EXPANSIVE.APPLICATIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>locative</code> | <code>CLAUSE.LOCATIVE.APPLICATIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>receptive</code> | <code>CLAUSE.RECEPTIVE.APPLICATIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>reductive</code> | <code>CLAUSE.REDUCTIVE.APPLICATIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>associative</code> | <code>CLAUSE.ASSOCIATIVE.APPLICATIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>generative</code> | <code>CLAUSE.GENERATIVE.APPLICATIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>executive</code> | <code>CLAUSE.EXECUTIVE.APPLICATIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>interruptive</code> | <code>CLAUSE.INTERRUPTIVE.APPLICATIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>inflective</code> | <code>CLAUSE.INFLECTIVE.APPLICATIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>deflective</code> | <code>CLAUSE.DEFLECTIVE.APPLICATIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>observative</code> | <code>CLAUSE.OBSERVATIVE.APPLICATIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>terminative</code> | <code>CLAUSE.TERMINATIVE.APPLICATIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>suspensive</code> | <code>CLAUSE.SUSPENSIVE.APPLICATIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |

## predicate.conditional

<a id="grammar-predicate-conditional"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>PREDICATE</code> | <code>CONDITIONAL</code> | <code>CASE</code> | partial | [Predicates and statements](../spec/grammar.md#predicates-and-statements) | [Note 66](#grammar-note-66) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>condition</code> | <code>CONDITIONAL.IF &#124; [CONDITIONAL.ELSE.IF]* &#124; CONDITIONAL.ELSE*</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>iteration</code> | <code>CONDITIONAL.WHILE &#124; ADVERB.PROGRESSION* &#124; CLAUSE.IMPERATIVE*</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>other</code> | <code>CONDITIONAL.OTHER</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |

## statement.objective.imperative

<a id="grammar-statement-objective-imperative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>STATEMENT</code> | <code>OBJECTIVE.IMPERATIVE</code> | <code>CASE</code> | partial | [Predicates and statements](../spec/grammar.md#predicates-and-statements) | [Note 66](#grammar-note-66) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>generic</code> | <code>PREDICATE.IMPERATIVE.GENERIC</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>declarative</code> | <code>PREDICATE.IMPERATIVE.DECLARATIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>definitive</code> | <code>PREDICATE.IMPERATIVE.DEFINITIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>descriptive</code> | <code>PREDICATE.IMPERATIVE.DESCRIPTIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>expansive</code> | <code>PREDICATE.IMPERATIVE.EXPANSIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>locative</code> | <code>PREDICATE.IMPERATIVE.LOCATIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>receptive</code> | <code>PREDICATE.IMPERATIVE.RECEPTIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>reductive</code> | <code>PREDICATE.IMPERATIVE.REDUCTIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>associative</code> | <code>PREDICATE.IMPERATIVE.ASSOCIATIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>generative</code> | <code>PREDICATE.IMPERATIVE.GENERATIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>executive</code> | <code>PREDICATE.IMPERATIVE.EXECUTIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>interruptive</code> | <code>PREDICATE.IMPERATIVE.INTERRUPTIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>inflective</code> | <code>PREDICATE.IMPERATIVE.INFLECTIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>deflective</code> | <code>PREDICATE.IMPERATIVE.DEFLECTIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>observative</code> | <code>PREDICATE.IMPERATIVE.OBSERVATIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>terminative</code> | <code>PREDICATE.IMPERATIVE.TERMINATIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>suspensive</code> | <code>PREDICATE.IMPERATIVE.SUSPENSIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>returning</code> | <code>PREDICATE.IMPERATIVE.RETURNING</code> | [Typed Functions](../spec/applicational-types.md#functions-properties-and-operations) · supported |
| <code>invocative</code> | <code>PREDICATE.IMPERATIVE.INVOCATIVE</code> | [Typed Functions](../spec/applicational-types.md#functions-properties-and-operations) · supported |

## statement.objective.conditional

<a id="grammar-statement-objective-conditional"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>STATEMENT</code> | <code>OBJECTIVE.CONDITIONAL</code> | <code>CASE</code> | partial | [Predicates and statements](../spec/grammar.md#predicates-and-statements) | [Note 66](#grammar-note-66) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>conditional</code> | <code>PREDICATE.CONDITIONAL.CONDITION</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>iterative</code> | <code>PREDICATE.CONDITIONAL.ITERATION</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>alternative</code> | <code>PREDICATE.CONDITIONAL.OTHER</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |

## statement.subjective.active

<a id="grammar-statement-subjective-active"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>STATEMENT</code> | <code>SUBJECTIVE.ACTIVE</code> | <code>SUBJECT &#124; CASE</code> | partial | [Predicates and statements](../spec/grammar.md#predicates-and-statements) | [Note 66](#grammar-note-66) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>generic</code> | <code>PREDICATE.ACTIVE.GENERIC</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>declarative</code> | <code>PREDICATE.ACTIVE.DECLARATIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>definitive</code> | <code>PREDICATE.ACTIVE.DEFINITIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>descriptive</code> | <code>PREDICATE.ACTIVE.DESCRIPTIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>expansive</code> | <code>PREDICATE.ACTIVE.EXPANSIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>locative</code> | <code>PREDICATE.ACTIVE.LOCATIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>receptive</code> | <code>PREDICATE.ACTIVE.RECEPTIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>reductive</code> | <code>PREDICATE.ACTIVE.REDUCTIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>associative</code> | <code>PREDICATE.ACTIVE.ASSOCIATIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>generative</code> | <code>PREDICATE.ACTIVE.GENERATIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>executive</code> | <code>PREDICATE.ACTIVE.EXECUTIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>interruptive</code> | <code>PREDICATE.ACTIVE.INTERRUPTIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>inflective</code> | <code>PREDICATE.ACTIVE.INFLECTIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>deflective</code> | <code>PREDICATE.ACTIVE.DEFLECTIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>observative</code> | <code>PREDICATE.ACTIVE.OBSERVATIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>terminative</code> | <code>PREDICATE.ACTIVE.TERMINATIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>suspensive</code> | <code>PREDICATE.ACTIVE.SUSPENSIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |

## statement.subjective.passive

<a id="grammar-statement-subjective-passive"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>STATEMENT</code> | <code>SUBJECTIVE.PASSIVE</code> | <code>SUBJECT &#124; CASE</code> | partial | [Predicates and statements](../spec/grammar.md#predicates-and-statements) | [Note 66](#grammar-note-66) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>assign</code> | <code>PREDICATE.PASSIVE.ASSIGN &#124; [CLAUSE.ASSIGN.PASSIVE]*</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>generic</code> | <code>PREDICATE.PASSIVE.GENERIC</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>declarative</code> | <code>PREDICATE.PASSIVE.DECLARATIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>definitive</code> | <code>PREDICATE.PASSIVE.DEFINITIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>descriptive</code> | <code>PREDICATE.PASSIVE.DESCRIPTIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>expansive</code> | <code>PREDICATE.PASSIVE.EXPANSIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>locative</code> | <code>PREDICATE.PASSIVE.LOCATIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>receptive</code> | <code>PREDICATE.PASSIVE.RECEPTIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>reductive</code> | <code>PREDICATE.PASSIVE.REDUCTIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>associative</code> | <code>PREDICATE.PASSIVE.ASSOCIATIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>generative</code> | <code>PREDICATE.PASSIVE.GENERATIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>executive</code> | <code>PREDICATE.PASSIVE.EXECUTIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>interruptive</code> | <code>PREDICATE.PASSIVE.INTERRUPTIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>inflective</code> | <code>PREDICATE.PASSIVE.INFLECTIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>deflective</code> | <code>PREDICATE.PASSIVE.DEFLECTIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>observative</code> | <code>PREDICATE.PASSIVE.OBSERVATIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>terminative</code> | <code>PREDICATE.PASSIVE.TERMINATIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>suspensive</code> | <code>PREDICATE.PASSIVE.SUSPENSIVE</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |

## phrase.signature

<a id="grammar-phrase-signature"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>PHRASE</code> | <code>SIGNATURE</code> | <code>PREPOSITION.CONSIGNMENT.ASSOCIATIVE &#124; CASE</code> | supported | [Typed Functions](../spec/applicational-types.md#functions-properties-and-operations) | [Note 67](#grammar-note-67) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>both</code> | <code>NOUN.FUNCTIONAL.INFLECTIVE &#124; PREPOSITION.DEFINEMENT.DEFINITIVE &#124; [SUBJECT.SIGNATURE] &#124; CONJUNCTION.MUTUAL &#124; NOUN.FUNCTIONAL.DEFLECTIVE &#124; PREPOSITION.DEFINEMENT.DEFINITIVE &#124; [OBJECT]</code> | [Typed Functions](../spec/applicational-types.md#functions-properties-and-operations) · supported |
| <code>inputs</code> | <code>NOUN.FUNCTIONAL.INFLECTIVE &#124; PREPOSITION.DEFINEMENT.DEFINITIVE &#124; [SUBJECT.SIGNATURE] &#124; CONJUNCTION.MUTUAL &#124; ARTICLE.NONE &#124; NOUN.FUNCTIONAL.DEFLECTIVE</code> | [Typed Functions](../spec/applicational-types.md#functions-properties-and-operations) · supported |
| <code>outputs</code> | <code>ARTICLE.NONE &#124; NOUN.FUNCTIONAL.INFLECTIVE &#124; CONJUNCTION.MUTUAL &#124; NOUN.FUNCTIONAL.DEFLECTIVE &#124; PREPOSITION.DEFINEMENT.DEFINITIVE &#124; [OBJECT]</code> | [Typed Functions](../spec/applicational-types.md#functions-properties-and-operations) · supported |
| <code>empty</code> | <code>ARTICLE.NONE &#124; NOUN.FUNCTIONAL.INFLECTIVE &#124; CONJUNCTION.MUTUAL &#124; ARTICLE.NONE &#124; NOUN.FUNCTIONAL.DEFLECTIVE</code> | [Typed Functions](../spec/applicational-types.md#functions-properties-and-operations) · supported |

## phrase.arguments

<a id="grammar-phrase-arguments"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>PHRASE</code> | <code>ARGUMENTS</code> | <code>PREPOSITION.CONSIGNMENT.ASSOCIATIVE &#124; [VALUE]</code> | supported | [Typed Functions](../spec/applicational-types.md#functions-properties-and-operations) | [Note 68](#grammar-note-68) |

## phrase.results

<a id="grammar-phrase-results"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>PHRASE</code> | <code>RESULTS</code> | <code>PREPOSITION.DEFINEMENT.DECLARATIVE &#124; [SUBJECT.DECLARATION]</code> | supported | [Typed Functions](../spec/applicational-types.md#functions-properties-and-operations) | [Note 69](#grammar-note-69) |

## clause.invocative.imperative

<a id="grammar-clause-invocative-imperative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>INVOCATIVE.IMPERATIVE</code> | <code>VERB.INVOCATIVE &#124; SUBJECT.DEFINITIVE &#124; PHRASE.ARGUMENTS* &#124; PHRASE.RESULTS*</code> | supported | [Typed Functions](../spec/applicational-types.md#functions-properties-and-operations) | [Note 70](#grammar-note-70) |

## clause.returning.imperative

<a id="grammar-clause-returning-imperative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>RETURNING.IMPERATIVE</code> | <code>VERB.RETURNING &#124; [VALUE]*</code> | supported | [Typed Functions](../spec/applicational-types.md#functions-properties-and-operations) | [Note 71](#grammar-note-71) |

## Support notes

| Note | Reviewed explanation |
| --- | --- |
| <a id="grammar-note-1"></a>1 | The cases enumerate recognizers, including categories whose general executable construction is not established. Consult each type’s support boundary before using a configured value pattern. |
| <a id="grammar-note-2"></a>2 | A subject participates in declaration/binding and paragraph antecedent selection. Its cases distinguish declared names, resolved references and definitions. |
| <a id="grammar-note-3"></a>3 | This phrase associates grammatical objects, prepositions or clauses as given by its sequence. Its location and optional/repeated terms are defined by the configured matcher, not conventional EBNF. |
| <a id="grammar-note-4"></a>4 | This phrase participates in construction or argument attachment. Native placeholders have explicit type and encoding contracts; do not infer arbitrary coercions. |
| <a id="grammar-note-5"></a>5 | Supplies a Number index for Lists/Bytes or a Text key for Map/Instance reads and writes. Runtime checks reject invalid keys and bounds. |
| <a id="grammar-note-6"></a>6 | Selects the source of a typed read, with an optional indexed key. Complete requested runtime type identities are checked before extraction. |
| <a id="grammar-note-7"></a>7 | Selects a typed update destination and optional index or field key; runtime mutation does not change vocabulary namespaces. |
| <a id="grammar-note-8"></a>8 | The to phrase supplies an imperative source clause or Text body to a definition. It is distinct from Use/import dependency resolution. |
| <a id="grammar-note-9"></a>9 | Source-provider ownership governs relative imports and expanded declarations. Grammatical dependency attachment does not override those ownership rules. |
| <a id="grammar-note-10"></a>10 | Assignment follows resolved typed-storage rules and definite assignment. Runtime ordering is preserved independently of parser retry order. The passive case is the passive grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries. |
| <a id="grammar-note-11"></a>11 | Generic clauses call source-defined vocabulary with resolved typed arguments. Runtime captures use retained typed environments; asynchronous methods are not introduced. The imperative case is the imperative grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries. |
| <a id="grammar-note-12"></a>12 | Generic clauses call source-defined vocabulary with resolved typed arguments. Runtime captures use retained typed environments; asynchronous methods are not introduced. The active case is the active grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries. |
| <a id="grammar-note-13"></a>13 | Generic clauses call source-defined vocabulary with resolved typed arguments. Runtime captures use retained typed environments; asynchronous methods are not introduced. The passive case is the passive grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries. |
| <a id="grammar-note-14"></a>14 | Generic clauses call source-defined vocabulary with resolved typed arguments. Runtime captures use retained typed environments; asynchronous methods are not introduced. The applicative case is the applicative grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries. |
| <a id="grammar-note-15"></a>15 | A declaration associates a source name with the indicated category or value. Voice-specific recognizers do not imply interchangeable native behavior. The imperative case is the imperative grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries. |
| <a id="grammar-note-16"></a>16 | A declaration associates a source name with the indicated category or value. Voice-specific recognizers do not imply interchangeable native behavior. The active case is the active grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries. |
| <a id="grammar-note-17"></a>17 | A declaration associates a source name with the indicated category or value. Voice-specific recognizers do not imply interchangeable native behavior. The passive case is the passive grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries. |
| <a id="grammar-note-18"></a>18 | A declaration associates a source name with the indicated category or value. Voice-specific recognizers do not imply interchangeable native behavior. The applicative case is the applicative grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries. |
| <a id="grammar-note-19"></a>19 | Definition clauses establish source vocabulary or typed native-backed forms. Accepted construction and framing are bounded by definition expansion. The imperative case is the imperative grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries. |
| <a id="grammar-note-20"></a>20 | Definition clauses establish source vocabulary or typed native-backed forms. Accepted construction and framing are bounded by definition expansion. The active case is the active grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries. |
| <a id="grammar-note-21"></a>21 | Definition clauses establish source vocabulary or typed native-backed forms. Accepted construction and framing are bounded by definition expansion. The passive case is the passive grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries. |
| <a id="grammar-note-22"></a>22 | Definition clauses establish source vocabulary or typed native-backed forms. Accepted construction and framing are bounded by definition expansion. The applicative case is the applicative grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries. |
| <a id="grammar-note-23"></a>23 | Descriptive clauses apply source-defined adjective/adverb meanings. Constructor descriptors preserve the target and source-unit ownership. The imperative case is the imperative grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries. |
| <a id="grammar-note-24"></a>24 | Descriptive clauses apply source-defined adjective/adverb meanings. Constructor descriptors preserve the target and source-unit ownership. The active case is the active grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries. |
| <a id="grammar-note-25"></a>25 | Descriptive clauses apply source-defined adjective/adverb meanings. Constructor descriptors preserve the target and source-unit ownership. The passive case is the passive grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries. |
| <a id="grammar-note-26"></a>26 | Descriptive clauses apply source-defined adjective/adverb meanings. Constructor descriptors preserve the target and source-unit ownership. The applicative case is the applicative grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries. |
| <a id="grammar-note-27"></a>27 | Put/Place resolve typed runtime updates. Keys, element/field types and ownership cycles are checked; this recognizer does not add iteration or removal. |
| <a id="grammar-note-28"></a>28 | This configured collection form is not an additional operation API. Use the documented Put/Place and typed-read forms; removal and iteration remain unsupported. The passive case is the passive grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries. |
| <a id="grammar-note-29"></a>29 | This configured collection form is not an additional operation API. Use the documented Put/Place and typed-read forms; removal and iteration remain unsupported. The imperative case is the imperative grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries. |
| <a id="grammar-note-30"></a>30 | This configured collection form is not an additional operation API. Use the documented Put/Place and typed-read forms; removal and iteration remain unsupported. The active case is the active grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries. |
| <a id="grammar-note-31"></a>31 | This configured collection form is not an additional operation API. Use the documented Put/Place and typed-read forms; removal and iteration remain unsupported. The applicative case is the applicative grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries. |
| <a id="grammar-note-32"></a>32 | Association names an existing canonical value or definition. It does not create an independent runtime copy. The imperative case is the imperative grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries. |
| <a id="grammar-note-33"></a>33 | Association names an existing canonical value or definition. It does not create an independent runtime copy. The active case is the active grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries. |
| <a id="grammar-note-34"></a>34 | Association names an existing canonical value or definition. It does not create an independent runtime copy. The passive case is the passive grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries. |
| <a id="grammar-note-35"></a>35 | Association names an existing canonical value or definition. It does not create an independent runtime copy. The applicative case is the applicative grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries. |
| <a id="grammar-note-36"></a>36 | Build is the runtime readiness operation. It is distinct from the CLI command that compiles native output. The imperative case is the imperative grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries. |
| <a id="grammar-note-37"></a>37 | Build is the runtime readiness operation. It is distinct from the CLI command that compiles native output. The active case is the active grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries. |
| <a id="grammar-note-38"></a>38 | Build is the runtime readiness operation. It is distinct from the CLI command that compiles native output. The passive case is the passive grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries. |
| <a id="grammar-note-39"></a>39 | Build is the runtime readiness operation. It is distinct from the CLI command that compiles native output. The applicative case is the applicative grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries. |
| <a id="grammar-note-40"></a>40 | Run launches the selected process. The grammar does not imply that every applicative category is runnable. The imperative case is the imperative grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries. |
| <a id="grammar-note-41"></a>41 | Run launches the selected process. The grammar does not imply that every applicative category is runnable. The active case is the active grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries. |
| <a id="grammar-note-42"></a>42 | Run launches the selected process. The grammar does not imply that every applicative category is runnable. The passive case is the passive grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries. |
| <a id="grammar-note-43"></a>43 | Run launches the selected process. The grammar does not imply that every applicative category is runnable. The applicative case is the applicative grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries. |
| <a id="grammar-note-44"></a>44 | Await joins a selected process invocation and activates its continuation once. A parsed clause alone does not establish valid activation placement. The imperative case is the imperative grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries. |
| <a id="grammar-note-45"></a>45 | Await joins a selected process invocation and activates its continuation once. A parsed clause alone does not establish valid activation placement. The active case is the active grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries. |
| <a id="grammar-note-46"></a>46 | Await joins a selected process invocation and activates its continuation once. A parsed clause alone does not establish valid activation placement. The passive case is the passive grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries. |
| <a id="grammar-note-47"></a>47 | Await joins a selected process invocation and activates its continuation once. A parsed clause alone does not establish valid activation placement. The applicative case is the applicative grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries. |
| <a id="grammar-note-48"></a>48 | Stop requests termination without joining. Final output still requires the Await contract. The imperative case is the imperative grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries. |
| <a id="grammar-note-49"></a>49 | Stop requests termination without joining. Final output still requires the Await contract. The active case is the active grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries. |
| <a id="grammar-note-50"></a>50 | Stop requests termination without joining. Final output still requires the Await contract. The passive case is the passive grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries. |
| <a id="grammar-note-51"></a>51 | Stop requests termination without joining. Final output still requires the Await contract. The applicative case is the applicative grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries. |
| <a id="grammar-note-52"></a>52 | Import/Export-style configured vocabulary is not a native outer instruction. The supported Use dictionary/source import path is documented separately. The imperative case is the imperative grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries. |
| <a id="grammar-note-53"></a>53 | Import/Export-style configured vocabulary is not a native outer instruction. The supported Use dictionary/source import path is documented separately. The active case is the active grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries. |
| <a id="grammar-note-54"></a>54 | Import/Export-style configured vocabulary is not a native outer instruction. The supported Use dictionary/source import path is documented separately. The passive case is the passive grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries. |
| <a id="grammar-note-55"></a>55 | Import/Export-style configured vocabulary is not a native outer instruction. The supported Use dictionary/source import path is documented separately. The applicative case is the applicative grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries. |
| <a id="grammar-note-56"></a>56 | Show emits the resolved operand without Print’s added newline; Output reads are subject to process finalization. The imperative case is the imperative grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries. |
| <a id="grammar-note-57"></a>57 | Show emits the resolved operand without Print’s added newline; Output reads are subject to process finalization. The active case is the active grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries. |
| <a id="grammar-note-58"></a>58 | Show emits the resolved operand without Print’s added newline; Output reads are subject to process finalization. The passive case is the passive grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries. |
| <a id="grammar-note-59"></a>59 | Show emits the resolved operand without Print’s added newline; Output reads are subject to process finalization. The applicative case is the applicative grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries. |
| <a id="grammar-note-60"></a>60 | Exit ends the compiled program; ordinary Return exits its Function, while native Function.Return remains a separate builtin expansion. The imperative case is the imperative grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries. |
| <a id="grammar-note-61"></a>61 | Exit ends the compiled program; ordinary Return exits its Function, while native Function.Return remains a separate builtin expansion. The active case is the active grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries. |
| <a id="grammar-note-62"></a>62 | Exit ends the compiled program; ordinary Return exits its Function, while native Function.Return remains a separate builtin expansion. The passive case is the passive grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries. |
| <a id="grammar-note-63"></a>63 | Exit ends the compiled program; ordinary Return exits its Function, while native Function.Return remains a separate builtin expansion. The applicative case is the applicative grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries. |
| <a id="grammar-note-64"></a>64 | Scalar conditions have specific type constraints. Joined conditions are rejected by outer lowering; no general short-circuit Boolean expression language is established. |
| <a id="grammar-note-65"></a>65 | The parser recognizes this shape, but outer instruction lowering explicitly rejects while. Rust backend iteration cases do not establish source-language loop support. |
| <a id="grammar-note-66"></a>66 | This wrapper selects a clause or conditional family. Complete statements must consume their associated tokens; a named alternative is not proof of native operational support. |
| <a id="grammar-note-67"></a>67 | An ordinary Function signature declares ordered named inputs and ordered result types, or explicit No Inputs/No Outputs. Nested runtime and callable identities remain complete. |
| <a id="grammar-note-68"></a>68 | Call supplies ordered existing values or literals. Construct complex argument values before the call; argument count and complete runtime types are checked. |
| <a id="grammar-note-69"></a>69 | Call declares one ordered result name per output. Names receive checked returned values only after the invocation completes. |
| <a id="grammar-note-70"></a>70 | Call invokes an ordinary Function with fresh parameter/local storage and its selected capture environment; it does not run an App process. |
| <a id="grammar-note-71"></a>71 | Return is valid inside an ordinary Function and must match its declared output count and types. Every continuing result-bearing path must return. |
