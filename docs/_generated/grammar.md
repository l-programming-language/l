# Grammar rule inventory

This mechanical reference preserves the configured English recognizers. It does not claim all recognized forms can be lowered or executed. Read [grammar notation](../spec/grammar.md) before interpreting these tables: ` | ` is sequencing, not an alternative operator. Support notes are separately reviewed. Rules without named cases use their structure directly.

## value

<a id="grammar-value"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>NOUN</code> | <code>VALUE</code> | <code>CASE</code> | partial | [Values and operands](../spec/values.md#declarations-assignment-and-reading-values) | [Note 1](#grammar-note-1) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>definitive</code> | <code>ARTICLE* &#124; [ADJECTIVE]* &#124; ID</code> | [Bound references](../spec/names-and-scope.md#possessives-and-member-lookup) · partial |
| <code>text</code> | <code>`~`</code> | [Text](../spec/values.md#text) · partial |
| <code>number</code> | <code>%~%</code> | [Number](../spec/values.md#number) · partial |
| <code>truth</code> | <code>?~?</code> | [Truth](../spec/values.md#truth) · partial |
| <code>null</code> | <code>!~!</code> | [Null](../spec/values.md#null-and-internal-states) · partial |
| <code>type</code> | <code>ARTICLE* &#124; #~#</code> | [Native type categories](../spec/lexical-types.md#attributes-definitions-and-lexical-categories) · partial |
| <code>list</code> | <code>&amp;~&amp;</code> | [List](../spec/collections-and-categories.md#lists) · partial |
| <code>dict</code> | <code>{~}</code> | [Dictionary](../spec/collections-and-categories.md#dictionaries) · partial |

## subject

<a id="grammar-subject"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>NOUN</code> | <code>SUBJECT</code> | <code>ARTICLE* &#124; [ADJECTIVE]* &#124; CASE</code> | partial | [Subjects and declarations](../spec/names-and-scope.md#declarations-and-definitions) | [Note 1](#grammar-note-1) |

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
| <code>callable</code> | <code>"ID" &#124; PHRASE.DECLARATIVE.CALLABLE</code> | [Typed Functions](../spec/applicational-types.md#functions-properties-and-operations) · partial |
| <code>signature</code> | <code>"ID" &#124; PHRASE.DECLARATIVE.SIGNATURE</code> | [Typed Functions](../spec/applicational-types.md#functions-properties-and-operations) · partial |

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
| <code>definitive</code> | <code>ID &#124; PHRASE* &#124; CLAUSE*</code> | [Bound references](../spec/names-and-scope.md#possessives-and-member-lookup) · partial |
| <code>dependent</code> | <code>ID &#124; PHRASE.DEPENDENCY &#124; CLAUSE*</code> | [Dependency-qualified reference](../spec/imports.md#explicit-source-file-imports) · partial |
| <code>text</code> | <code>`~`</code> | [Text](../spec/values.md#text) · partial |
| <code>number</code> | <code>%~%</code> | [Number](../spec/values.md#number) · partial |
| <code>truth</code> | <code>?~?</code> | [Truth](../spec/values.md#truth) · partial |
| <code>null</code> | <code>!~!</code> | [Null](../spec/values.md#null-and-internal-states) · partial |
| <code>type</code> | <code>#~# &#124; PHRASE* &#124; CLAUSE*</code> | [Native type categories](../spec/lexical-types.md#attributes-definitions-and-lexical-categories) · partial |
| <code>list</code> | <code>&amp;~&amp; &#124; PHRASE* &#124; CLAUSE*</code> | [List](../spec/collections-and-categories.md#lists) · partial |
| <code>dict</code> | <code>{~} &#124; PHRASE* &#124; CLAUSE*</code> | [Dictionary](../spec/collections-and-categories.md#dictionaries) · partial |
| <code>argument</code> | <code>ID</code> | [Bound update operands](../spec/collections-and-categories.md#lists) · partial |
| <code>type_argument</code> | <code>#~#</code> | [Atomic type operands](../spec/collections-and-categories.md#lists) · partial |
| <code>callable</code> | <code>ID &#124; PHRASE</code> | [Typed Functions](../spec/applicational-types.md#functions-properties-and-operations) · partial |

## phrase.generic

<a id="grammar-phrase-generic"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>PHRASE</code> | <code>GENERIC</code> | <code>[PREPOSITION] &#124; [OBJECT] &#124; CLAUSE* &#124; PHRASE*</code> | partial | [Phrase association](../spec/grammar.md#phrases) | [Note 1](#grammar-note-1) |

## phrase.directive

<a id="grammar-phrase-directive"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>PHRASE</code> | <code>DIRECTIVE</code> | <code>PREPOSITION.CONJOINMENT.DIRECTIVE &#124; CASE</code> | partial | [Construction and typed arguments](../spec/definitions.md#typed-slots) | [Note 1](#grammar-note-1) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>definitive</code> | <code>[CLAUSE.APPLICATIVE]</code> | [Source action body](../spec/definitions.md#source-actions) · partial |

## phrase.declarative

<a id="grammar-phrase-declarative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>PHRASE</code> | <code>DECLARATIVE</code> | <code>PREPOSITION.DEFINEMENT.DECLARATIVE &#124; CASE</code> | partial | [Construction and typed arguments](../spec/definitions.md#typed-slots) | [Note 1](#grammar-note-1) |

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
| <code>callable</code> | <code>OBJECT.CALLABLE</code> | [Typed Functions](../spec/applicational-types.md#functions-properties-and-operations) · partial |
| <code>signature</code> | <code>OBJECT</code> | [Typed Functions](../spec/applicational-types.md#functions-properties-and-operations) · partial |

## phrase.locative

<a id="grammar-phrase-locative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>PHRASE</code> | <code>LOCATIVE</code> | <code>PREPOSITION.DEFINEMENT.LOCATIVE &#124; OBJECT</code> | partial | [Typed runtime access](../spec/collections-and-categories.md#lists) | [Note 1](#grammar-note-1) |

## phrase.objective

<a id="grammar-phrase-objective"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>PHRASE</code> | <code>OBJECTIVE</code> | <code>PREPOSITION.CONJOINMENT.OBJECTIVE &#124; SUBJECT.DEFINITIVE &#124; PHRASE.LOCATIVE* &#124; PHRASE.DECLARATIVE.DECLARATIVE*</code> | partial | [Typed runtime access](../spec/collections-and-categories.md#lists) | [Note 1](#grammar-note-1) |

## phrase.inflective

<a id="grammar-phrase-inflective"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>PHRASE</code> | <code>INFLECTIVE</code> | <code>PREPOSITION.CONSIGNMENT.INFLECTIVE &#124; CASE</code> | partial | [Typed runtime access](../spec/collections-and-categories.md#lists) | [Note 1](#grammar-note-1) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>list</code> | <code>SUBJECT.DEFINITIVE</code> | [Collection association](../spec/collections-and-categories.md#lists) · partial |
| <code>dictionary</code> | <code>SUBJECT.DEFINITIVE &#124; PHRASE.DECLARATIVE.DEFINITIONAL</code> | [Collection association](../spec/collections-and-categories.md#dictionaries) · partial |
| <code>locative</code> | <code>SUBJECT.DEFINITIVE &#124; PHRASE.LOCATIVE</code> | [Collection association](../spec/collections-and-categories.md#lists) · partial |

## phrase.associative

<a id="grammar-phrase-associative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>PHRASE</code> | <code>ASSOCIATIVE</code> | <code>PREPOSITION.CONSIGNMENT.ASSOCIATIVE &#124; CASE</code> | partial | [Phrase association](../spec/grammar.md#phrases) | [Note 1](#grammar-note-1) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>subjective</code> | <code>SUBJECT.DEFINITIVE</code> | [Phrase association](../spec/grammar.md#phrases) · partial |
| <code>objective</code> | <code>OBJECT.DEFINITIVE</code> | [Phrase association](../spec/grammar.md#phrases) · partial |

## phrase.indicative

<a id="grammar-phrase-indicative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>PHRASE</code> | <code>INDICATIVE</code> | <code>PREPOSITION.CONSIGNMENT.INDICATIVE &#124; OBJECT.DEFINITIVE</code> | partial | [Phrase association](../spec/grammar.md#phrases) | [Note 1](#grammar-note-1) |

## phrase.utilization

<a id="grammar-phrase-utilization"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>PHRASE</code> | <code>UTILIZATION</code> | <code>PREPOSITION.CONJOINMENT.SUBJECTIVE &#124; CASE</code> | partial | [Definition framing](../spec/definitions.md#framing-and-escaping) | [Note 1](#grammar-note-1) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>definitive</code> | <code>CLAUSE.IMPERATIVE</code> | [Source action relation](../spec/definitions.md#source-actions) · partial |
| <code>definitional</code> | <code>OBJECT.TEXT</code> | [Native-backed Text definition](../spec/definitions.md#framing-and-escaping) · partial |

## phrase.dependency

<a id="grammar-phrase-dependency"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>PHRASE</code> | <code>DEPENDENCY</code> | <code>PREPOSITION.CONJOINMENT.OBJECTIVE &#124; `~`</code> | partial | [Imports and dependencies](../spec/imports.md#explicit-source-file-imports) | [Note 1](#grammar-note-1) |

## clause.assign.passive

<a id="grammar-clause-assign-passive"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>ASSIGN.PASSIVE</code> | <code>SUBJECT.DECLARATION &#124; PREDICATE.PASSIVE.ASSIGN</code> | partial | [Assignment](../spec/control-flow.md#assignment) | [Note 1](#grammar-note-1) |

## clause.generic.imperative

<a id="grammar-clause-generic-imperative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>GENERIC.IMPERATIVE</code> | <code>[ADVERB]* &#124; VERB &#124; CASE</code> | partial | [Source actions](../spec/definitions.md#source-actions) | [Note 1](#grammar-note-1) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>callable</code> | <code>[VALUE]* &#124; [PHRASE.ARGUMENTS]* &#124; PHRASE.RESULTS*</code> | [Typed Function words](../spec/applicational-types.md#functions-properties-and-operations) · partial |
| <code>subjective</code> | <code>SUBJECT</code> | [Bound source-action arguments](../spec/definitions.md#source-actions) · partial |
| <code>objective</code> | <code>[OBJECT]</code> | [Bound source-action arguments](../spec/definitions.md#source-actions) · partial |

## clause.generic.active

<a id="grammar-clause-generic-active"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>GENERIC.ACTIVE</code> | <code>[ADVERB]* &#124; VERB.ACTIVE &#124; CASE</code> | partial | [Source actions](../spec/definitions.md#source-actions) | [Note 1](#grammar-note-1) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>callable</code> | <code>[VALUE]* &#124; [PHRASE.ARGUMENTS]* &#124; PHRASE.RESULTS*</code> | [Typed Function words](../spec/applicational-types.md#functions-properties-and-operations) · partial |
| <code>objective</code> | <code>OBJECT</code> | [Source actions](../spec/definitions.md#source-actions) · partial |

## clause.generic.passive

<a id="grammar-clause-generic-passive"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>GENERIC.PASSIVE</code> | <code>[ADVERB]* &#124; VERB.PASSIVE &#124; OBJECT</code> | partial | [Source actions](../spec/definitions.md#source-actions) | [Note 1](#grammar-note-1) |

## clause.generic.applicative

<a id="grammar-clause-generic-applicative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>GENERIC.APPLICATIVE</code> | <code>[ADVERB]* &#124; VERB.APPLICATIVE &#124; CASE</code> | partial | [Source actions](../spec/definitions.md#source-actions) | [Note 1](#grammar-note-1) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>callable</code> | <code>[VALUE]* &#124; [PHRASE.ARGUMENTS]* &#124; PHRASE.RESULTS*</code> | [Typed Function words](../spec/applicational-types.md#functions-properties-and-operations) · partial |
| <code>subjective</code> | <code>SUBJECT</code> | [Bound source-action arguments](../spec/definitions.md#source-actions) · partial |
| <code>objective</code> | <code>[OBJECT]</code> | [Bound source-action arguments](../spec/definitions.md#source-actions) · partial |

## clause.declarative.imperative

<a id="grammar-clause-declarative-imperative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>DECLARATIVE.IMPERATIVE</code> | <code>[ADVERB]* &#124; VERB.DECLARATIVE &#124; CASE</code> | partial | [Declarations](../spec/definitions.md#create-define-and-describe) | [Note 1](#grammar-note-1) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>noun</code> | <code>SUBJECT.DECLARATIVE</code> | [Declarations](../spec/definitions.md#create-define-and-describe) · partial |
| <code>verb</code> | <code>SUBJECT.DIRECTIVE</code> | [Declarations](../spec/definitions.md#create-define-and-describe) · partial |
| <code>type</code> | <code>SUBJECT.TYPE</code> | [Declarations](../spec/definitions.md#create-define-and-describe) · partial |
| <code>list</code> | <code>SUBJECT.LIST</code> | [Declarations](../spec/definitions.md#create-define-and-describe) · partial |
| <code>dict</code> | <code>SUBJECT.DICT</code> | [Declarations](../spec/definitions.md#create-define-and-describe) · partial |
| <code>callable</code> | <code>SUBJECT.CALLABLE</code> | [Typed Functions](../spec/applicational-types.md#functions-properties-and-operations) · partial |

## clause.declarative.active

<a id="grammar-clause-declarative-active"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>DECLARATIVE.ACTIVE</code> | <code>[ADVERB]* &#124; VERB.DECLARATIVE.ACTIVE &#124; CASE</code> | partial | [Declarations](../spec/definitions.md#create-define-and-describe) | [Note 1](#grammar-note-1) |

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
| <code>CLAUSE</code> | <code>DECLARATIVE.PASSIVE</code> | <code>[ADVERB]* &#124; VERB.DECLARATIVE.PASSIVE &#124; CASE</code> | partial | [Declarations](../spec/definitions.md#create-define-and-describe) | [Note 1](#grammar-note-1) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>noun</code> | <code>PHRASE.DECLARATIVE.DEFINITIVE</code> | [Declarations](../spec/definitions.md#create-define-and-describe) · partial |
| <code>verb</code> | <code>PHRASE.DIRECTIVE</code> | [Declarations](../spec/definitions.md#create-define-and-describe) · partial |

## clause.declarative.applicative

<a id="grammar-clause-declarative-applicative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>DECLARATIVE.APPLICATIVE</code> | <code>[ADVERB]* &#124; VERB.DECLARATIVE.APPLICATIVE &#124; CASE</code> | partial | [Declarations](../spec/definitions.md#create-define-and-describe) | [Note 1](#grammar-note-1) |

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
| <code>CLAUSE</code> | <code>DEFINITIVE.IMPERATIVE</code> | <code>[ADVERB]* &#124; VERB.DEFINITIVE &#124; CASE</code> | partial | [Definitions](../spec/definitions.md#create-define-and-describe) | [Note 1](#grammar-note-1) |

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
| <code>CLAUSE</code> | <code>DEFINITIVE.ACTIVE</code> | <code>[ADVERB]* &#124; VERB.DEFINITIVE.ACTIVE &#124; CASE</code> | partial | [Definitions](../spec/definitions.md#create-define-and-describe) | [Note 1](#grammar-note-1) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>noun</code> | <code>SUBJECT.DEFINITIONAL</code> | [Definitions](../spec/definitions.md#create-define-and-describe) · partial |
| <code>verb</code> | <code>SUBJECT.UTILIZATION</code> | [Definitions](../spec/definitions.md#create-define-and-describe) · partial |

## clause.definitive.passive

<a id="grammar-clause-definitive-passive"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>DEFINITIVE.PASSIVE</code> | <code>[ADVERB]* &#124; VERB.DEFINITIVE.PASSIVE &#124; CASE</code> | partial | [Definitions](../spec/definitions.md#create-define-and-describe) | [Note 1](#grammar-note-1) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>noun</code> | <code>PHRASE.DECLARATIVE.DEFINITIONAL</code> | [Definitions](../spec/definitions.md#create-define-and-describe) · partial |
| <code>verb</code> | <code>PHRASE.UTILIZATION</code> | [Definitions](../spec/definitions.md#create-define-and-describe) · partial |

## clause.definitive.applicative

<a id="grammar-clause-definitive-applicative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>DEFINITIVE.APPLICATIVE</code> | <code>[ADVERB]* &#124; VERB.DEFINITIVE.APPLICATIVE &#124; CASE</code> | partial | [Definitions](../spec/definitions.md#create-define-and-describe) | [Note 1](#grammar-note-1) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>noun</code> | <code>SUBJECT.DEFINITIONAL</code> | [Definitions](../spec/definitions.md#create-define-and-describe) · partial |
| <code>verb</code> | <code>SUBJECT.UTILIZATION</code> | [Definitions](../spec/definitions.md#create-define-and-describe) · partial |

## clause.descriptive.imperative

<a id="grammar-clause-descriptive-imperative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>DESCRIPTIVE.IMPERATIVE</code> | <code>[ADVERB]* &#124; VERB.DESCRIPTIVE &#124; CASE</code> | partial | [Descriptors](../spec/definitions.md#descriptors-and-constructors) | [Note 1](#grammar-note-1) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>adjective</code> | <code>SUBJECT.DECLARATIVE</code> | [Descriptors](../spec/definitions.md#descriptors-and-constructors) · partial |
| <code>adverb</code> | <code>SUBJECT.UTILIZATION</code> | [Descriptors](../spec/definitions.md#descriptors-and-constructors) · partial |

## clause.descriptive.active

<a id="grammar-clause-descriptive-active"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>DESCRIPTIVE.ACTIVE</code> | <code>[ADVERB]* &#124; VERB.DESCRIPTIVE.ACTIVE &#124; CASE</code> | partial | [Descriptors](../spec/definitions.md#descriptors-and-constructors) | [Note 1](#grammar-note-1) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>adjective</code> | <code>OBJECT</code> | [Descriptors](../spec/definitions.md#descriptors-and-constructors) · partial |
| <code>adverb</code> | <code>CLAUSE.ASSOCIATIVE.APPLICATIVE</code> | [Descriptors](../spec/definitions.md#descriptors-and-constructors) · partial |

## clause.descriptive.passive

<a id="grammar-clause-descriptive-passive"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>DESCRIPTIVE.PASSIVE</code> | <code>[ADVERB]* &#124; VERB.DESCRIPTIVE.PASSIVE &#124; CASE</code> | partial | [Descriptors](../spec/definitions.md#descriptors-and-constructors) | [Note 1](#grammar-note-1) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>adjective</code> | <code>PHRASE.DECLARATIVE.DECLARATIVE</code> | [Descriptors](../spec/definitions.md#descriptors-and-constructors) · partial |
| <code>adverb</code> | <code>CLAUSE.ASSOCIATIVE.APPLICATIVE</code> | [Descriptors](../spec/definitions.md#descriptors-and-constructors) · partial |

## clause.descriptive.applicative

<a id="grammar-clause-descriptive-applicative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>DESCRIPTIVE.APPLICATIVE</code> | <code>[ADVERB]* &#124; VERB.DESCRIPTIVE.APPLICATIVE &#124; CASE</code> | partial | [Descriptors](../spec/definitions.md#descriptors-and-constructors) | [Note 1](#grammar-note-1) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>adjective</code> | <code>[OBJECT] &#124; PHRASE.DECLARATIVE.DECLARATIVE*</code> | [Descriptors](../spec/definitions.md#descriptors-and-constructors) · partial |
| <code>adverb</code> | <code>SUBJECT.DEFINITIVE &#124; CLAUSE.ASSOCIATIVE.APPLICATIVE</code> | [Descriptors](../spec/definitions.md#descriptors-and-constructors) · partial |

## clause.expansive.imperative

<a id="grammar-clause-expansive-imperative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>EXPANSIVE.IMPERATIVE</code> | <code>[ADVERB]* &#124; VERB.EXPANSIVE &#124; CASE</code> | partial | [Typed updates](../spec/collections-and-categories.md#lists) | [Note 1](#grammar-note-1) |

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
| <code>CLAUSE</code> | <code>EXPANSIVE.ACTIVE</code> | <code>[ADVERB]* &#124; VERB.EXPANSIVE.ACTIVE &#124; CASE</code> | partial | [Typed updates](../spec/collections-and-categories.md#lists) | [Note 1](#grammar-note-1) |

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
| <code>CLAUSE</code> | <code>EXPANSIVE.PASSIVE</code> | <code>[ADVERB]* &#124; VERB.EXPANSIVE.PASSIVE &#124; CASE</code> | partial | [Collections](../spec/collections-and-categories.md#lists) | [Note 1](#grammar-note-1) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>list</code> | <code>PHRASE.INFLECTIVE.LIST</code> | [Collections](../spec/collections-and-categories.md#lists) · partial |
| <code>dictionary</code> | <code>PHRASE.INFLECTIVE.DICTIONARY</code> | [Collections](../spec/collections-and-categories.md#lists) · partial |

## clause.expansive.applicative

<a id="grammar-clause-expansive-applicative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>EXPANSIVE.APPLICATIVE</code> | <code>[ADVERB]* &#124; VERB.EXPANSIVE.APPLICATIVE &#124; CASE</code> | partial | [Typed updates](../spec/collections-and-categories.md#lists) | [Note 1](#grammar-note-1) |

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
| <code>CLAUSE</code> | <code>LOCATIVE.IMPERATIVE</code> | <code>[ADVERB]* &#124; VERB.LOCATIVE &#124; CASE</code> | partial | [Typed updates](../spec/collections-and-categories.md#lists) | [Note 1](#grammar-note-1) |

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
| <code>CLAUSE</code> | <code>LOCATIVE.ACTIVE</code> | <code>[ADVERB]* &#124; VERB.LOCATIVE.ACTIVE &#124; CASE</code> | partial | [Typed updates](../spec/collections-and-categories.md#lists) | [Note 1](#grammar-note-1) |

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
| <code>CLAUSE</code> | <code>LOCATIVE.PASSIVE</code> | <code>[ADVERB]* &#124; VERB.LOCATIVE.PASSIVE &#124; CASE</code> | partial | [Collections](../spec/collections-and-categories.md#lists) | [Note 1](#grammar-note-1) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>list</code> | <code>PHRASE.INFLECTIVE.LOCATIVE</code> | [Collections](../spec/collections-and-categories.md#lists) · partial |
| <code>dictionary</code> | <code>PHRASE.INFLECTIVE.DICTIONARY</code> | [Collections](../spec/collections-and-categories.md#lists) · partial |

## clause.locative.applicative

<a id="grammar-clause-locative-applicative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>LOCATIVE.APPLICATIVE</code> | <code>[ADVERB]* &#124; VERB.LOCATIVE.APPLICATIVE &#124; CASE</code> | partial | [Typed updates](../spec/collections-and-categories.md#lists) | [Note 1](#grammar-note-1) |

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
| <code>CLAUSE</code> | <code>RECEPTIVE.IMPERATIVE</code> | <code>[ADVERB]* &#124; VERB.RECEPTIVE &#124; CASE &#124; PHRASE.OBJECTIVE</code> | partial | [Collections](../spec/collections-and-categories.md#lists) | [Note 1](#grammar-note-1) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>list</code> | <code>OBJECT.NUMBER</code> | [Collections](../spec/collections-and-categories.md#lists) · partial |
| <code>dictionary</code> | <code>OBJECT.TEXT</code> | [Collections](../spec/collections-and-categories.md#lists) · partial |

## clause.receptive.active

<a id="grammar-clause-receptive-active"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>RECEPTIVE.ACTIVE</code> | <code>[ADVERB]* &#124; VERB.RECEPTIVE.ACTIVE &#124; CASE &#124; PHRASE.OBJECTIVE</code> | partial | [Collections](../spec/collections-and-categories.md#lists) | [Note 1](#grammar-note-1) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>list</code> | <code>OBJECT.NUMBER</code> | [Collections](../spec/collections-and-categories.md#lists) · partial |
| <code>dictionary</code> | <code>OBJECT.TEXT</code> | [Collections](../spec/collections-and-categories.md#lists) · partial |

## clause.receptive.passive

<a id="grammar-clause-receptive-passive"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>RECEPTIVE.PASSIVE</code> | <code>[ADVERB]* &#124; VERB.RECEPTIVE.PASSIVE &#124; PHRASE.OBJECTIVE</code> | partial | [Collections](../spec/collections-and-categories.md#lists) | [Note 1](#grammar-note-1) |

## clause.receptive.applicative

<a id="grammar-clause-receptive-applicative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>RECEPTIVE.APPLICATIVE</code> | <code>[ADVERB]* &#124; VERB.RECEPTIVE.APPLICATIVE &#124; CASE &#124; PHRASE.OBJECTIVE</code> | partial | [Collections](../spec/collections-and-categories.md#lists) | [Note 1](#grammar-note-1) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>list</code> | <code>OBJECT.NUMBER</code> | [Collections](../spec/collections-and-categories.md#lists) · partial |
| <code>dictionary</code> | <code>OBJECT.TEXT</code> | [Collections](../spec/collections-and-categories.md#lists) · partial |

## clause.reductive.imperative

<a id="grammar-clause-reductive-imperative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>REDUCTIVE.IMPERATIVE</code> | <code>[ADVERB]* &#124; VERB.RECEPTIVE &#124; CASE &#124; PHRASE.OBJECTIVE</code> | partial | [Collections](../spec/collections-and-categories.md#lists) | [Note 1](#grammar-note-1) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>list</code> | <code>OBJECT.NUMBER</code> | [Collections](../spec/collections-and-categories.md#lists) · partial |
| <code>dictionary</code> | <code>OBJECT.TEXT</code> | [Collections](../spec/collections-and-categories.md#lists) · partial |

## clause.reductive.active

<a id="grammar-clause-reductive-active"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>REDUCTIVE.ACTIVE</code> | <code>[ADVERB]* &#124; VERB.RECEPTIVE.ACTIVE &#124; CASE &#124; PHRASE.OBJECTIVE</code> | partial | [Collections](../spec/collections-and-categories.md#lists) | [Note 1](#grammar-note-1) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>list</code> | <code>OBJECT.NUMBER</code> | [Collections](../spec/collections-and-categories.md#lists) · partial |
| <code>dictionary</code> | <code>OBJECT.TEXT</code> | [Collections](../spec/collections-and-categories.md#lists) · partial |

## clause.reductive.passive

<a id="grammar-clause-reductive-passive"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>REDUCTIVE.PASSIVE</code> | <code>[ADVERB]* &#124; VERB.RECEPTIVE.PASSIVE &#124; PHRASE.OBJECTIVE</code> | partial | [Collections](../spec/collections-and-categories.md#lists) | [Note 1](#grammar-note-1) |

## clause.reductive.applicative

<a id="grammar-clause-reductive-applicative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>REDUCTIVE.APPLICATIVE</code> | <code>[ADVERB]* &#124; VERB.RECEPTIVE.APPLICATIVE &#124; CASE &#124; PHRASE.OBJECTIVE</code> | partial | [Collections](../spec/collections-and-categories.md#lists) | [Note 1](#grammar-note-1) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>list</code> | <code>OBJECT.NUMBER</code> | [Collections](../spec/collections-and-categories.md#lists) · partial |
| <code>dictionary</code> | <code>OBJECT.TEXT</code> | [Collections](../spec/collections-and-categories.md#lists) · partial |

## clause.associative.imperative

<a id="grammar-clause-associative-imperative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>ASSOCIATIVE.IMPERATIVE</code> | <code>[ADVERB]* &#124; VERB.ASSOCIATIVE &#124; SUBJECT.DEFINITIVE &#124; CASE*</code> | partial | [Aliases](../spec/names-and-scope.md#aliases-and-canonical-identity) | [Note 1](#grammar-note-1) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>alias</code> | <code>PHRASE.DECLARATIVE.ALIAS</code> | [Aliases](../spec/names-and-scope.md#aliases-and-canonical-identity) · partial |

## clause.associative.active

<a id="grammar-clause-associative-active"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>ASSOCIATIVE.ACTIVE</code> | <code>[ADVERB]* &#124; VERB.ASSOCIATIVE.ACTIVE &#124; SUBJECT.DEFINITIVE &#124; CASE*</code> | partial | [Aliases](../spec/names-and-scope.md#aliases-and-canonical-identity) | [Note 1](#grammar-note-1) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>alias</code> | <code>PHRASE.DECLARATIVE.ALIAS</code> | [Aliases](../spec/names-and-scope.md#aliases-and-canonical-identity) · partial |

## clause.associative.passive

<a id="grammar-clause-associative-passive"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>ASSOCIATIVE.PASSIVE</code> | <code>[ADVERB]* &#124; VERB.ASSOCIATIVE.PASSIVE &#124; CASE</code> | partial | [Aliases](../spec/names-and-scope.md#aliases-and-canonical-identity) | [Note 1](#grammar-note-1) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>definitive</code> | <code>PHRASE.DECLARATIVE</code> | [Aliases](../spec/names-and-scope.md#aliases-and-canonical-identity) · partial |
| <code>associative</code> | <code>PHRASE.INDICATIVE</code> | [Aliases](../spec/names-and-scope.md#aliases-and-canonical-identity) · partial |

## clause.associative.applicative

<a id="grammar-clause-associative-applicative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>ASSOCIATIVE.APPLICATIVE</code> | <code>[ADVERB]* &#124; VERB.ASSOCIATIVE.APPLICATIVE &#124; OBJECT &#124; CASE*</code> | partial | [Aliases](../spec/names-and-scope.md#aliases-and-canonical-identity) | [Note 1](#grammar-note-1) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>alias</code> | <code>PHRASE.DECLARATIVE.ALIAS</code> | [Aliases](../spec/names-and-scope.md#aliases-and-canonical-identity) · partial |

## clause.generative.imperative

<a id="grammar-clause-generative-imperative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>GENERATIVE.IMPERATIVE</code> | <code>[ADVERB]* &#124; VERB.GENERATIVE &#124; OBJECT</code> | partial | [Build](../spec/execution.md#build-run-stop-and-await) | [Note 1](#grammar-note-1) |

## clause.generative.active

<a id="grammar-clause-generative-active"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>GENERATIVE.ACTIVE</code> | <code>[ADVERB]* &#124; VERB.GENERATIVE.ACTIVE &#124; OBJECT</code> | partial | [Build](../spec/execution.md#build-run-stop-and-await) | [Note 1](#grammar-note-1) |

## clause.generative.passive

<a id="grammar-clause-generative-passive"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>GENERATIVE.PASSIVE</code> | <code>[ADVERB]* &#124; VERB.GENERATIVE.PASSIVE</code> | partial | [Build](../spec/execution.md#build-run-stop-and-await) | [Note 1](#grammar-note-1) |

## clause.generative.applicative

<a id="grammar-clause-generative-applicative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>GENERATIVE.APPLICATIVE</code> | <code>[ADVERB]* &#124; VERB.GENERATIVE.APPLICATIVE &#124; OBJECT</code> | partial | [Build](../spec/execution.md#build-run-stop-and-await) | [Note 1](#grammar-note-1) |

## clause.executive.imperative

<a id="grammar-clause-executive-imperative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>EXECUTIVE.IMPERATIVE</code> | <code>[ADVERB]* &#124; VERB.EXECUTIVE &#124; OBJECT</code> | partial | [Run](../spec/execution.md#build-run-stop-and-await) | [Note 1](#grammar-note-1) |

## clause.executive.active

<a id="grammar-clause-executive-active"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>EXECUTIVE.ACTIVE</code> | <code>[ADVERB]* &#124; VERB.EXECUTIVE.ACTIVE &#124; OBJECT</code> | partial | [Run](../spec/execution.md#build-run-stop-and-await) | [Note 1](#grammar-note-1) |

## clause.executive.passive

<a id="grammar-clause-executive-passive"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>EXECUTIVE.PASSIVE</code> | <code>[ADVERB]* &#124; VERB.EXECUTIVE.PASSIVE</code> | partial | [Run](../spec/execution.md#build-run-stop-and-await) | [Note 1](#grammar-note-1) |

## clause.executive.applicative

<a id="grammar-clause-executive-applicative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>EXECUTIVE.APPLICATIVE</code> | <code>[ADVERB]* &#124; VERB.EXECUTIVE.APPLICATIVE &#124; OBJECT</code> | partial | [Run](../spec/execution.md#build-run-stop-and-await) | [Note 1](#grammar-note-1) |

## clause.suspensive.imperative

<a id="grammar-clause-suspensive-imperative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>SUSPENSIVE.IMPERATIVE</code> | <code>[ADVERB]* &#124; VERB.SUSPENSIVE &#124; OBJECT</code> | partial | [Await](../spec/execution.md#continuations-and-activation) | [Note 1](#grammar-note-1) |

## clause.suspensive.active

<a id="grammar-clause-suspensive-active"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>SUSPENSIVE.ACTIVE</code> | <code>[ADVERB]* &#124; VERB.SUSPENSIVE.ACTIVE &#124; OBJECT</code> | partial | [Await](../spec/execution.md#continuations-and-activation) | [Note 1](#grammar-note-1) |

## clause.suspensive.passive

<a id="grammar-clause-suspensive-passive"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>SUSPENSIVE.PASSIVE</code> | <code>[ADVERB]* &#124; VERB.SUSPENSIVE.PASSIVE</code> | partial | [Await](../spec/execution.md#continuations-and-activation) | [Note 1](#grammar-note-1) |

## clause.suspensive.applicative

<a id="grammar-clause-suspensive-applicative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>SUSPENSIVE.APPLICATIVE</code> | <code>[ADVERB]* &#124; VERB.SUSPENSIVE.APPLICATIVE &#124; OBJECT</code> | partial | [Await](../spec/execution.md#continuations-and-activation) | [Note 1](#grammar-note-1) |

## clause.interruptive.imperative

<a id="grammar-clause-interruptive-imperative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>INTERRUPTIVE.IMPERATIVE</code> | <code>[ADVERB]* &#124; VERB.INTERRUPTIVE &#124; OBJECT</code> | partial | [Stop](../spec/execution.md#build-run-stop-and-await) | [Note 1](#grammar-note-1) |

## clause.interruptive.active

<a id="grammar-clause-interruptive-active"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>INTERRUPTIVE.ACTIVE</code> | <code>[ADVERB]* &#124; VERB.INTERRUPTIVE.ACTIVE &#124; OBJECT</code> | partial | [Stop](../spec/execution.md#build-run-stop-and-await) | [Note 1](#grammar-note-1) |

## clause.interruptive.passive

<a id="grammar-clause-interruptive-passive"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>INTERRUPTIVE.PASSIVE</code> | <code>[ADVERB]* &#124; VERB.INTERRUPTIVE.PASSIVE</code> | partial | [Stop](../spec/execution.md#build-run-stop-and-await) | [Note 1](#grammar-note-1) |

## clause.interruptive.applicative

<a id="grammar-clause-interruptive-applicative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>INTERRUPTIVE.APPLICATIVE</code> | <code>[ADVERB]* &#124; VERB.INTERRUPTIVE.APPLICATIVE &#124; OBJECT</code> | partial | [Stop](../spec/execution.md#build-run-stop-and-await) | [Note 1](#grammar-note-1) |

## clause.inflective.imperative

<a id="grammar-clause-inflective-imperative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>INFLECTIVE.IMPERATIVE</code> | <code>[ADVERB]* &#124; VERB.INFLECTIVE &#124; OBJECT</code> | partial | [Import and export vocabulary](../spec/imports.md#indexed-dictionaries) | [Note 1](#grammar-note-1) |

## clause.inflective.active

<a id="grammar-clause-inflective-active"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>INFLECTIVE.ACTIVE</code> | <code>[ADVERB]* &#124; VERB.INFLECTIVE.ACTIVE &#124; OBJECT</code> | partial | [Import and export vocabulary](../spec/imports.md#indexed-dictionaries) | [Note 1](#grammar-note-1) |

## clause.inflective.passive

<a id="grammar-clause-inflective-passive"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>INFLECTIVE.PASSIVE</code> | <code>[ADVERB]* &#124; VERB.INFLECTIVE.PASSIVE</code> | partial | [Import and export vocabulary](../spec/imports.md#indexed-dictionaries) | [Note 1](#grammar-note-1) |

## clause.inflective.applicative

<a id="grammar-clause-inflective-applicative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>INFLECTIVE.APPLICATIVE</code> | <code>[ADVERB]* &#124; VERB.INFLECTIVE.APPLICATIVE &#124; OBJECT</code> | partial | [Import and export vocabulary](../spec/imports.md#indexed-dictionaries) | [Note 1](#grammar-note-1) |

## clause.deflective.imperative

<a id="grammar-clause-deflective-imperative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>DEFLECTIVE.IMPERATIVE</code> | <code>[ADVERB]* &#124; VERB.DEFLECTIVE &#124; OBJECT</code> | partial | [Import and export vocabulary](../spec/imports.md#indexed-dictionaries) | [Note 1](#grammar-note-1) |

## clause.deflective.active

<a id="grammar-clause-deflective-active"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>DEFLECTIVE.ACTIVE</code> | <code>[ADVERB]* &#124; VERB.DEFLECTIVE.ACTIVE &#124; OBJECT</code> | partial | [Import and export vocabulary](../spec/imports.md#indexed-dictionaries) | [Note 1](#grammar-note-1) |

## clause.deflective.passive

<a id="grammar-clause-deflective-passive"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>DEFLECTIVE.PASSIVE</code> | <code>[ADVERB]* &#124; VERB.DEFLECTIVE.PASSIVE</code> | partial | [Import and export vocabulary](../spec/imports.md#indexed-dictionaries) | [Note 1](#grammar-note-1) |

## clause.deflective.applicative

<a id="grammar-clause-deflective-applicative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>DEFLECTIVE.APPLICATIVE</code> | <code>[ADVERB]* &#124; VERB.DEFLECTIVE.APPLICATIVE &#124; OBJECT</code> | partial | [Import and export vocabulary](../spec/imports.md#indexed-dictionaries) | [Note 1](#grammar-note-1) |

## clause.observative.imperative

<a id="grammar-clause-observative-imperative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>OBSERVATIVE.IMPERATIVE</code> | <code>[ADVERB]* &#124; VERB.OBSERVATIVE &#124; OBJECT</code> | partial | [Show](../spec/control-flow.md#show-and-output) | [Note 1](#grammar-note-1) |

## clause.observative.active

<a id="grammar-clause-observative-active"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>OBSERVATIVE.ACTIVE</code> | <code>[ADVERB]* &#124; VERB.OBSERVATIVE.ACTIVE &#124; OBJECT</code> | partial | [Show](../spec/control-flow.md#show-and-output) | [Note 1](#grammar-note-1) |

## clause.observative.passive

<a id="grammar-clause-observative-passive"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>OBSERVATIVE.PASSIVE</code> | <code>[ADVERB]* &#124; VERB.OBSERVATIVE.PASSIVE</code> | partial | [Show](../spec/control-flow.md#show-and-output) | [Note 1](#grammar-note-1) |

## clause.observative.applicative

<a id="grammar-clause-observative-applicative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>OBSERVATIVE.APPLICATIVE</code> | <code>[ADVERB]* &#124; VERB.OBSERVATIVE.APPLICATIVE &#124; OBJECT</code> | partial | [Show](../spec/control-flow.md#show-and-output) | [Note 1](#grammar-note-1) |

## clause.terminative.imperative

<a id="grammar-clause-terminative-imperative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>TERMINATIVE.IMPERATIVE</code> | <code>[ADVERB]* &#124; VERB.TERMINATIVE &#124; CASE*</code> | partial | [Exit](../spec/control-flow.md#exit-and-return) | [Note 1](#grammar-note-1) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>objective</code> | <code>OBJECT</code> | [Exit](../spec/control-flow.md#exit-and-return) · partial |
| <code>associative</code> | <code>PHRASE.ASSOCIATIVE</code> | [Exit](../spec/control-flow.md#exit-and-return) · partial |

## clause.terminative.active

<a id="grammar-clause-terminative-active"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>TERMINATIVE.ACTIVE</code> | <code>[ADVERB]* &#124; VERB.TERMINATIVE.ACTIVE &#124; CASE</code> | partial | [Exit](../spec/control-flow.md#exit-and-return) | [Note 1](#grammar-note-1) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>objective</code> | <code>OBJECT</code> | [Exit](../spec/control-flow.md#exit-and-return) · partial |
| <code>associative</code> | <code>PHRASE.ASSOCIATIVE</code> | [Exit](../spec/control-flow.md#exit-and-return) · partial |

## clause.terminative.passive

<a id="grammar-clause-terminative-passive"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>TERMINATIVE.PASSIVE</code> | <code>[ADVERB]* &#124; VERB.TERMINATIVE.PASSIVE &#124; CASE</code> | partial | [Exit](../spec/control-flow.md#exit-and-return) | [Note 1](#grammar-note-1) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>objective</code> | <code>OBJECT</code> | [Exit](../spec/control-flow.md#exit-and-return) · partial |
| <code>associative</code> | <code>PHRASE.ASSOCIATIVE</code> | [Exit](../spec/control-flow.md#exit-and-return) · partial |

## clause.terminative.applicative

<a id="grammar-clause-terminative-applicative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CLAUSE</code> | <code>TERMINATIVE.APPLICATIVE</code> | <code>[ADVERB]* &#124; VERB.TERMINATIVE.APPLICATIVE &#124; CASE</code> | partial | [Exit](../spec/control-flow.md#exit-and-return) | [Note 1](#grammar-note-1) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>objective</code> | <code>OBJECT</code> | [Exit](../spec/control-flow.md#exit-and-return) · partial |
| <code>associative</code> | <code>PHRASE.ASSOCIATIVE</code> | [Exit](../spec/control-flow.md#exit-and-return) · partial |

## condition

<a id="grammar-condition"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CONDITION</code> | <code>CONDITION</code> | <code>CASE</code> | partial | [Conditions](../spec/control-flow.md#comparisons) | [Note 1](#grammar-note-1) |

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
| <code>CONDITIONAL</code> | <code>IF</code> | <code>EVALUATION.CONDITIONAL &#124; [CONDITION] &#124; [CLAUSE.IMPERATIVE]</code> | partial | [Conditions](../spec/control-flow.md#comparisons) | [Note 1](#grammar-note-1) |

## conditional.else

<a id="grammar-conditional-else"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CONDITIONAL</code> | <code>ELSE</code> | <code>EVALUATION.EXCEPTIONAL &#124; CASE</code> | partial | [Conditions](../spec/control-flow.md#comparisons) | [Note 1](#grammar-note-1) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>if</code> | <code>CONDITIONAL.IF</code> | [Conditions](../spec/control-flow.md#comparisons) · partial |
| <code>else</code> | <code>[CLAUSE.IMPERATIVE]</code> | [Conditions](../spec/control-flow.md#comparisons) · partial |

## conditional.while

<a id="grammar-conditional-while"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CONDITIONAL</code> | <code>WHILE</code> | <code>EVALUATION.ITERATIVE &#124; [CONDITION] &#124; CLAUSE.IMPERATIVE</code> | rejected | [Iteration rejection](../spec/control-flow.md#iteration-and-collection-operations) | [Note 1](#grammar-note-1) |

## conditional.other

<a id="grammar-conditional-other"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>CONDITIONAL</code> | <code>OTHER</code> | <code>EVALUATION.ALTERNATIVE &#124; [CLAUSE.IMPERATIVE]</code> | partial | [Conditions](../spec/control-flow.md#comparisons) | [Note 1](#grammar-note-1) |

## predicate.imperative

<a id="grammar-predicate-imperative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>PREDICATE</code> | <code>IMPERATIVE</code> | <code>CASE &#124; [CLAUSE]*</code> | partial | [Predicates and statements](../spec/grammar.md#predicates-and-statements) | [Note 1](#grammar-note-1) |

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

## predicate.active

<a id="grammar-predicate-active"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>PREDICATE</code> | <code>ACTIVE</code> | <code>CASE &#124; [CLAUSE]*</code> | partial | [Predicates and statements](../spec/grammar.md#predicates-and-statements) | [Note 1](#grammar-note-1) |

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
| <code>PREDICATE</code> | <code>PASSIVE</code> | <code>VERB.PASSIVE &#124; CASE &#124; [CLAUSE]*</code> | partial | [Predicates and statements](../spec/grammar.md#predicates-and-statements) | [Note 1](#grammar-note-1) |

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
| <code>PREDICATE</code> | <code>APPLICATIVE</code> | <code>CASE &#124; [CLAUSE]*</code> | partial | [Predicates and statements](../spec/grammar.md#predicates-and-statements) | [Note 1](#grammar-note-1) |

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
| <code>PREDICATE</code> | <code>CONDITIONAL</code> | <code>CASE</code> | partial | [Predicates and statements](../spec/grammar.md#predicates-and-statements) | [Note 1](#grammar-note-1) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>condition</code> | <code>CONDITIONAL.IF &#124; [CONDITIONAL.ELSE.IF]* &#124; CONDITIONAL.ELSE*</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>iteration</code> | <code>CONDITIONAL.WHILE &#124; ADVERB.PROGRESSION* &#124; CLAUSE.IMPERATIVE*</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>other</code> | <code>CONDITIONAL.OTHER</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |

## statement.objective.imperative

<a id="grammar-statement-objective-imperative"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>STATEMENT</code> | <code>OBJECTIVE.IMPERATIVE</code> | <code>CASE</code> | partial | [Predicates and statements](../spec/grammar.md#predicates-and-statements) | [Note 1](#grammar-note-1) |

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

## statement.objective.conditional

<a id="grammar-statement-objective-conditional"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>STATEMENT</code> | <code>OBJECTIVE.CONDITIONAL</code> | <code>CASE</code> | partial | [Predicates and statements](../spec/grammar.md#predicates-and-statements) | [Note 1](#grammar-note-1) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>conditional</code> | <code>PREDICATE.CONDITIONAL.CONDITION</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>iterative</code> | <code>PREDICATE.CONDITIONAL.ITERATION</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>alternative</code> | <code>PREDICATE.CONDITIONAL.OTHER</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |

## statement.subjective.active

<a id="grammar-statement-subjective-active"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>STATEMENT</code> | <code>SUBJECTIVE.ACTIVE</code> | <code>SUBJECT &#124; CASE</code> | partial | [Predicates and statements](../spec/grammar.md#predicates-and-statements) | [Note 1](#grammar-note-1) |

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
| <code>STATEMENT</code> | <code>SUBJECTIVE.PASSIVE</code> | <code>SUBJECT &#124; CASE</code> | partial | [Predicates and statements](../spec/grammar.md#predicates-and-statements) | [Note 1](#grammar-note-1) |

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
| <code>PHRASE</code> | <code>SIGNATURE</code> | <code>PREPOSITION.CONSIGNMENT.ASSOCIATIVE &#124; CASE &#124; PHRASE.DIRECTIVE*</code> | partial | [Typed Functions](../spec/applicational-types.md#functions-properties-and-operations) | [Note 1](#grammar-note-1) |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>both</code> | <code>NOUN.FUNCTIONAL.INFLECTIVE &#124; PREPOSITION.DEFINEMENT.DEFINITIVE &#124; [SUBJECT.SIGNATURE] &#124; CONJUNCTION.MUTUAL &#124; NOUN.FUNCTIONAL.DEFLECTIVE &#124; PREPOSITION.DEFINEMENT.DEFINITIVE &#124; [OBJECT]</code> | [Typed Functions](../spec/applicational-types.md#functions-properties-and-operations) · partial |
| <code>inputs</code> | <code>NOUN.FUNCTIONAL.INFLECTIVE &#124; PREPOSITION.DEFINEMENT.DEFINITIVE &#124; [SUBJECT.SIGNATURE] &#124; CONJUNCTION.MUTUAL &#124; ARTICLE.NONE &#124; NOUN.FUNCTIONAL.DEFLECTIVE</code> | [Typed Functions](../spec/applicational-types.md#functions-properties-and-operations) · partial |
| <code>inputs_only</code> | <code>NOUN.FUNCTIONAL.INFLECTIVE &#124; PREPOSITION.DEFINEMENT.DEFINITIVE &#124; [SUBJECT.SIGNATURE]</code> | [Scoped Process inputs](../reference/word-glossary.md#process) · partial |
| <code>outputs</code> | <code>ARTICLE.NONE &#124; NOUN.FUNCTIONAL.INFLECTIVE &#124; CONJUNCTION.MUTUAL &#124; NOUN.FUNCTIONAL.DEFLECTIVE &#124; PREPOSITION.DEFINEMENT.DEFINITIVE &#124; [OBJECT]</code> | [Typed Functions](../spec/applicational-types.md#functions-properties-and-operations) · partial |
| <code>empty</code> | <code>ARTICLE.NONE &#124; NOUN.FUNCTIONAL.INFLECTIVE &#124; CONJUNCTION.MUTUAL &#124; ARTICLE.NONE &#124; NOUN.FUNCTIONAL.DEFLECTIVE</code> | [Typed Functions](../spec/applicational-types.md#functions-properties-and-operations) · partial |

## phrase.arguments

<a id="grammar-phrase-arguments"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>PHRASE</code> | <code>ARGUMENTS</code> | <code>PREPOSITION &#124; [VALUE]</code> | partial | [Typed Functions](../spec/applicational-types.md#functions-properties-and-operations) | [Note 1](#grammar-note-1) |

## phrase.results

<a id="grammar-phrase-results"></a>

| Category | Class | Structure | Support | Semantic context | Notes |
| --- | --- | --- | --- | --- | --- |
| <code>PHRASE</code> | <code>RESULTS</code> | <code>PREPOSITION.DEFINEMENT.DECLARATIVE &#124; [SUBJECT.DECLARATION]</code> | partial | [Typed Functions](../spec/applicational-types.md#functions-properties-and-operations) | [Note 1](#grammar-note-1) |

## Support notes

| Note | Reviewed explanation |
| --- | --- |
| <a id="grammar-note-1"></a>1 | Configured recognizer retained. The current CLI builds one standalone App with core Print and Text; recognition alone does not establish execution. |
