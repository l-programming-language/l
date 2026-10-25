# Grammar rule inventory

This mechanical reference preserves the configured English recognizers. It does not claim all recognized forms can be lowered or executed. Read [grammar notation](../spec/grammar.md) before interpreting these tables: ` | ` is sequencing, not an alternative operator. Support notes are separately reviewed.

## value

<a id="grammar-value"></a>

The cases enumerate recognizers, including categories whose general executable construction is not established. Consult each type’s support boundary before using a configured value pattern.

Semantic context: [Values and operands](../spec/values.md#declarations-assignment-and-reading-values).

| Property | Configured value |
| --- | --- |
| Category | <code>NOUN</code> |
| Class | <code>VALUE</code> |
| Structure | <code>CASE</code> |
| Support | partial |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>definitive</code> | <code>ARTICLE* &#124; [ADJECTIVE]* &#124; ID</code> | [Bound references](../spec/names-and-scope.md#possessives-and-member-lookup) · supported |
| <code>text</code> | <code>`~`</code> | [Text](../spec/values.md#text) · supported |
| <code>number</code> | <code>%~%</code> | [Number](../spec/values.md#number) · supported |
| <code>truth</code> | <code>?~?</code> | [Truth](../spec/values.md#truth) · supported |
| <code>null</code> | <code>!~!</code> | [Null](../spec/values.md#null-and-internal-states) · partial |
| <code>type</code> | <code>ARTICLE* &#124; #~#</code> | [Native type categories](../spec/lexical-types.md#attributes-definitions-and-lexical-categories) · partial |
| <code>list</code> | <code>&amp;~&amp;</code> | [List](../spec/collections-and-categories.md#lists) · partial |
| <code>dict</code> | <code>{~}</code> | [Dictionary](../spec/collections-and-categories.md#dictionaries) · partial |

## subject

<a id="grammar-subject"></a>

A subject participates in declaration/binding and paragraph antecedent selection. Its cases distinguish declared names, resolved references and definitions.

Semantic context: [Subjects and declarations](../spec/names-and-scope.md#declarations-and-definitions).

| Property | Configured value |
| --- | --- |
| Category | <code>NOUN</code> |
| Class | <code>SUBJECT</code> |
| Structure | <code>ARTICLE* &#124; [ADJECTIVE]* &#124; CASE</code> |
| Support | partial |

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

## object

<a id="grammar-object"></a>

The cases enumerate recognizers, including categories whose general executable construction is not established. Consult each type’s support boundary before using a configured value pattern.

Semantic context: [Values and operands](../spec/values.md#declarations-assignment-and-reading-values).

| Property | Configured value |
| --- | --- |
| Category | <code>NOUN</code> |
| Class | <code>OBJECT</code> |
| Structure | <code>ARTICLE* &#124; [ADJECTIVE]* &#124; CASE</code> |
| Support | partial |

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
| <code>null</code> | <code>!~!</code> | [Null](../spec/values.md#null-and-internal-states) · partial |
| <code>type</code> | <code>#~# &#124; PHRASE* &#124; CLAUSE*</code> | [Native type categories](../spec/lexical-types.md#attributes-definitions-and-lexical-categories) · partial |
| <code>list</code> | <code>&amp;~&amp; &#124; PHRASE* &#124; CLAUSE*</code> | [List](../spec/collections-and-categories.md#lists) · partial |
| <code>dict</code> | <code>{~} &#124; PHRASE* &#124; CLAUSE*</code> | [Dictionary](../spec/collections-and-categories.md#dictionaries) · partial |

## phrase.generic

<a id="grammar-phrase-generic"></a>

This phrase associates grammatical objects, prepositions or clauses as given by its sequence. Its location and optional/repeated terms are defined by the configured matcher, not conventional EBNF.

Semantic context: [Phrase association](../spec/grammar.md#phrases).

| Property | Configured value |
| --- | --- |
| Category | <code>PHRASE</code> |
| Class | <code>GENERIC</code> |
| Structure | <code>[PREPOSITION] &#124; [OBJECT] &#124; CLAUSE* &#124; PHRASE*</code> |
| Support | partial |

This rule has no named case alternatives; its structure is used directly.

## phrase.directive

<a id="grammar-phrase-directive"></a>

This phrase participates in construction or argument attachment. Native placeholders have explicit type and encoding contracts; do not infer arbitrary coercions.

Semantic context: [Construction and typed arguments](../spec/definitions.md#typed-slots).

| Property | Configured value |
| --- | --- |
| Category | <code>PHRASE</code> |
| Class | <code>DIRECTIVE</code> |
| Structure | <code>PREPOSITION.CONJOINMENT.DIRECTIVE &#124; CASE</code> |
| Support | partial |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>definitive</code> | <code>[CLAUSE.APPLICATIVE]</code> | [Source action body](../spec/definitions.md#source-actions) · supported |

## phrase.declarative

<a id="grammar-phrase-declarative"></a>

This phrase participates in construction or argument attachment. Native placeholders have explicit type and encoding contracts; do not infer arbitrary coercions.

Semantic context: [Construction and typed arguments](../spec/definitions.md#typed-slots).

| Property | Configured value |
| --- | --- |
| Category | <code>PHRASE</code> |
| Class | <code>DECLARATIVE</code> |
| Structure | <code>PREPOSITION.DEFINEMENT.DECLARATIVE &#124; CASE</code> |
| Support | partial |

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

## phrase.locative

<a id="grammar-phrase-locative"></a>

This phrase associates grammatical objects, prepositions or clauses as given by its sequence. Its location and optional/repeated terms are defined by the configured matcher, not conventional EBNF.

Semantic context: [Phrase association](../spec/grammar.md#phrases).

| Property | Configured value |
| --- | --- |
| Category | <code>PHRASE</code> |
| Class | <code>LOCATIVE</code> |
| Structure | <code>PREPOSITION.DEFINEMENT.LOCATIVE &#124; OBJECT.NUMBER</code> |
| Support | partial |

This rule has no named case alternatives; its structure is used directly.

## phrase.objective

<a id="grammar-phrase-objective"></a>

This phrase associates grammatical objects, prepositions or clauses as given by its sequence. Its location and optional/repeated terms are defined by the configured matcher, not conventional EBNF.

Semantic context: [Phrase association](../spec/grammar.md#phrases).

| Property | Configured value |
| --- | --- |
| Category | <code>PHRASE</code> |
| Class | <code>OBJECTIVE</code> |
| Structure | <code>PREPOSITION.CONJOINMENT.OBJECTIVE &#124; SUBJECT.DEFINITIVE &#124; PHRASE.DECLARATIVE.DECLARATIVE*</code> |
| Support | partial |

This rule has no named case alternatives; its structure is used directly.

## phrase.inflective

<a id="grammar-phrase-inflective"></a>

These configured collection forms do not establish a supported general insertion, indexing, removal or iteration API. Storage representations and dictionary member namespaces have narrower observed contracts.

Semantic context: [Collections](../spec/collections-and-categories.md#lists).

| Property | Configured value |
| --- | --- |
| Category | <code>PHRASE</code> |
| Class | <code>INFLECTIVE</code> |
| Structure | <code>PREPOSITION.CONSIGNMENT.INFLECTIVE &#124; CASE</code> |
| Support | partial |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>list</code> | <code>SUBJECT.DEFINITIVE</code> | [Collection association](../spec/collections-and-categories.md#lists) · partial |
| <code>dictionary</code> | <code>SUBJECT.DEFINITIVE &#124; PHRASE.DECLARATIVE.DEFINITIONAL</code> | [Collection association](../spec/collections-and-categories.md#dictionaries) · partial |
| <code>locative</code> | <code>SUBJECT.DEFINITIVE &#124; PREPOSITION.DEFINEMENT.LOCATIVE</code> | [Collection association](../spec/collections-and-categories.md#lists) · partial |

## phrase.associative

<a id="grammar-phrase-associative"></a>

This phrase associates grammatical objects, prepositions or clauses as given by its sequence. Its location and optional/repeated terms are defined by the configured matcher, not conventional EBNF.

Semantic context: [Phrase association](../spec/grammar.md#phrases).

| Property | Configured value |
| --- | --- |
| Category | <code>PHRASE</code> |
| Class | <code>ASSOCIATIVE</code> |
| Structure | <code>PREPOSITION.CONSIGNMENT.ASSOCIATIVE &#124; CASE</code> |
| Support | partial |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>subjective</code> | <code>SUBJECT.DEFINITIVE</code> | [Phrase association](../spec/grammar.md#phrases) · partial |
| <code>objective</code> | <code>OBJECT.DEFINITIVE</code> | [Phrase association](../spec/grammar.md#phrases) · partial |

## phrase.indicative

<a id="grammar-phrase-indicative"></a>

This phrase associates grammatical objects, prepositions or clauses as given by its sequence. Its location and optional/repeated terms are defined by the configured matcher, not conventional EBNF.

Semantic context: [Phrase association](../spec/grammar.md#phrases).

| Property | Configured value |
| --- | --- |
| Category | <code>PHRASE</code> |
| Class | <code>INDICATIVE</code> |
| Structure | <code>PREPOSITION.CONSIGNMENT.INDICATIVE &#124; OBJECT.DEFINITIVE</code> |
| Support | partial |

This rule has no named case alternatives; its structure is used directly.

## phrase.utilization

<a id="grammar-phrase-utilization"></a>

The to phrase supplies an imperative source clause or Text body to a definition. It is distinct from Use/import dependency resolution.

Semantic context: [Definition framing](../spec/definitions.md#framing-and-escaping).

| Property | Configured value |
| --- | --- |
| Category | <code>PHRASE</code> |
| Class | <code>UTILIZATION</code> |
| Structure | <code>PREPOSITION.CONJOINMENT.SUBJECTIVE &#124; CASE</code> |
| Support | partial |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>definitive</code> | <code>CLAUSE.IMPERATIVE</code> | [Source action relation](../spec/definitions.md#source-actions) · partial |
| <code>definitional</code> | <code>OBJECT.TEXT</code> | [Native-backed Text definition](../spec/definitions.md#framing-and-escaping) · supported |

## phrase.dependency

<a id="grammar-phrase-dependency"></a>

Source-provider ownership governs relative imports and expanded declarations. Grammatical dependency attachment does not override those ownership rules.

Semantic context: [Imports and dependencies](../spec/imports.md#explicit-source-file-imports).

| Property | Configured value |
| --- | --- |
| Category | <code>PHRASE</code> |
| Class | <code>DEPENDENCY</code> |
| Structure | <code>PREPOSITION.CONJOINMENT.OBJECTIVE &#124; `~`</code> |
| Support | partial |

This rule has no named case alternatives; its structure is used directly.

## clause.assign.passive

<a id="grammar-clause-assign-passive"></a>

Assignment follows resolved scalar/storage rules and definite assignment. Runtime ordering is preserved independently of parser retry order. The passive case is the passive grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries.

Semantic context: [Assignment](../spec/control-flow.md#assignment).

| Property | Configured value |
| --- | --- |
| Category | <code>CLAUSE</code> |
| Class | <code>ASSIGN.PASSIVE</code> |
| Structure | <code>SUBJECT.DECLARATION &#124; PREDICATE.PASSIVE.ASSIGN</code> |
| Support | partial |

This rule has no named case alternatives; its structure is used directly.

## clause.generic.imperative

<a id="grammar-clause-generic-imperative"></a>

Generic clauses call source-defined vocabulary with resolved typed arguments. There is no general closure or async-method guarantee. The imperative case is the imperative grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries.

Semantic context: [Source actions](../spec/definitions.md#source-actions).

| Property | Configured value |
| --- | --- |
| Category | <code>CLAUSE</code> |
| Class | <code>GENERIC.IMPERATIVE</code> |
| Structure | <code>[ADVERB]* &#124; VERB &#124; CASE</code> |
| Support | partial |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>subjective</code> | <code>SUBJECT</code> | [Bound source-action arguments](../spec/definitions.md#source-actions) · partial |
| <code>objective</code> | <code>[OBJECT]</code> | [Bound source-action arguments](../spec/definitions.md#source-actions) · partial |

## clause.generic.active

<a id="grammar-clause-generic-active"></a>

Generic clauses call source-defined vocabulary with resolved typed arguments. There is no general closure or async-method guarantee. The active case is the active grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries.

Semantic context: [Source actions](../spec/definitions.md#source-actions).

| Property | Configured value |
| --- | --- |
| Category | <code>CLAUSE</code> |
| Class | <code>GENERIC.ACTIVE</code> |
| Structure | <code>[ADVERB]* &#124; VERB.ACTIVE &#124; OBJECT</code> |
| Support | partial |

This rule has no named case alternatives; its structure is used directly.

## clause.generic.passive

<a id="grammar-clause-generic-passive"></a>

Generic clauses call source-defined vocabulary with resolved typed arguments. There is no general closure or async-method guarantee. The passive case is the passive grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries.

Semantic context: [Source actions](../spec/definitions.md#source-actions).

| Property | Configured value |
| --- | --- |
| Category | <code>CLAUSE</code> |
| Class | <code>GENERIC.PASSIVE</code> |
| Structure | <code>[ADVERB]* &#124; VERB.PASSIVE &#124; OBJECT</code> |
| Support | partial |

This rule has no named case alternatives; its structure is used directly.

## clause.generic.applicative

<a id="grammar-clause-generic-applicative"></a>

Generic clauses call source-defined vocabulary with resolved typed arguments. There is no general closure or async-method guarantee. The applicative case is the applicative grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries.

Semantic context: [Source actions](../spec/definitions.md#source-actions).

| Property | Configured value |
| --- | --- |
| Category | <code>CLAUSE</code> |
| Class | <code>GENERIC.APPLICATIVE</code> |
| Structure | <code>[ADVERB]* &#124; VERB.APPLICATIVE &#124; CASE</code> |
| Support | partial |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>subjective</code> | <code>SUBJECT</code> | [Bound source-action arguments](../spec/definitions.md#source-actions) · partial |
| <code>objective</code> | <code>[OBJECT]</code> | [Bound source-action arguments](../spec/definitions.md#source-actions) · partial |

## clause.declarative.imperative

<a id="grammar-clause-declarative-imperative"></a>

A declaration associates a source name with the indicated category or value. Voice-specific recognizers do not imply interchangeable native behavior. The imperative case is the imperative grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries.

Semantic context: [Declarations](../spec/definitions.md#create-define-and-describe).

| Property | Configured value |
| --- | --- |
| Category | <code>CLAUSE</code> |
| Class | <code>DECLARATIVE.IMPERATIVE</code> |
| Structure | <code>[ADVERB]* &#124; VERB.DECLARATIVE &#124; CASE</code> |
| Support | partial |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>noun</code> | <code>SUBJECT.DECLARATIVE</code> | [Declarations](../spec/definitions.md#create-define-and-describe) · partial |
| <code>verb</code> | <code>SUBJECT.DIRECTIVE</code> | [Declarations](../spec/definitions.md#create-define-and-describe) · partial |
| <code>type</code> | <code>SUBJECT.TYPE</code> | [Declarations](../spec/definitions.md#create-define-and-describe) · partial |
| <code>list</code> | <code>SUBJECT.LIST</code> | [Declarations](../spec/definitions.md#create-define-and-describe) · partial |
| <code>dict</code> | <code>SUBJECT.DICT</code> | [Declarations](../spec/definitions.md#create-define-and-describe) · partial |

## clause.declarative.active

<a id="grammar-clause-declarative-active"></a>

A declaration associates a source name with the indicated category or value. Voice-specific recognizers do not imply interchangeable native behavior. The active case is the active grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries.

Semantic context: [Declarations](../spec/definitions.md#create-define-and-describe).

| Property | Configured value |
| --- | --- |
| Category | <code>CLAUSE</code> |
| Class | <code>DECLARATIVE.ACTIVE</code> |
| Structure | <code>[ADVERB]* &#124; VERB.DECLARATIVE.ACTIVE &#124; CASE</code> |
| Support | partial |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>noun</code> | <code>SUBJECT.DECLARATIVE</code> | [Declarations](../spec/definitions.md#create-define-and-describe) · partial |
| <code>verb</code> | <code>SUBJECT.DIRECTIVE</code> | [Declarations](../spec/definitions.md#create-define-and-describe) · partial |
| <code>type</code> | <code>SUBJECT.TYPE</code> | [Declarations](../spec/definitions.md#create-define-and-describe) · partial |
| <code>list</code> | <code>SUBJECT.LIST</code> | [Declarations](../spec/definitions.md#create-define-and-describe) · partial |
| <code>dict</code> | <code>SUBJECT.DICT</code> | [Declarations](../spec/definitions.md#create-define-and-describe) · partial |

## clause.declarative.passive

<a id="grammar-clause-declarative-passive"></a>

A declaration associates a source name with the indicated category or value. Voice-specific recognizers do not imply interchangeable native behavior. The passive case is the passive grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries.

Semantic context: [Declarations](../spec/definitions.md#create-define-and-describe).

| Property | Configured value |
| --- | --- |
| Category | <code>CLAUSE</code> |
| Class | <code>DECLARATIVE.PASSIVE</code> |
| Structure | <code>[ADVERB]* &#124; VERB.DECLARATIVE.PASSIVE &#124; CASE</code> |
| Support | partial |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>noun</code> | <code>PHRASE.DECLARATIVE.DEFINITIVE</code> | [Declarations](../spec/definitions.md#create-define-and-describe) · partial |
| <code>verb</code> | <code>PHRASE.DIRECTIVE</code> | [Declarations](../spec/definitions.md#create-define-and-describe) · partial |

## clause.declarative.applicative

<a id="grammar-clause-declarative-applicative"></a>

A declaration associates a source name with the indicated category or value. Voice-specific recognizers do not imply interchangeable native behavior. The applicative case is the applicative grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries.

Semantic context: [Declarations](../spec/definitions.md#create-define-and-describe).

| Property | Configured value |
| --- | --- |
| Category | <code>CLAUSE</code> |
| Class | <code>DECLARATIVE.APPLICATIVE</code> |
| Structure | <code>[ADVERB]* &#124; VERB.DECLARATIVE.APPLICATIVE &#124; CASE</code> |
| Support | partial |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>noun</code> | <code>SUBJECT.DECLARATIVE</code> | [Declarations](../spec/definitions.md#create-define-and-describe) · partial |
| <code>verb</code> | <code>SUBJECT.DIRECTIVE</code> | [Declarations](../spec/definitions.md#create-define-and-describe) · partial |
| <code>type</code> | <code>SUBJECT.TYPE</code> | [Declarations](../spec/definitions.md#create-define-and-describe) · partial |
| <code>list</code> | <code>SUBJECT.LIST</code> | [Declarations](../spec/definitions.md#create-define-and-describe) · partial |
| <code>dict</code> | <code>SUBJECT.DICT</code> | [Declarations](../spec/definitions.md#create-define-and-describe) · partial |

## clause.definitive.imperative

<a id="grammar-clause-definitive-imperative"></a>

Definition clauses establish source vocabulary or typed native-backed forms. Accepted construction and framing are bounded by definition expansion. The imperative case is the imperative grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries.

Semantic context: [Definitions](../spec/definitions.md#create-define-and-describe).

| Property | Configured value |
| --- | --- |
| Category | <code>CLAUSE</code> |
| Class | <code>DEFINITIVE.IMPERATIVE</code> |
| Structure | <code>[ADVERB]* &#124; VERB.DEFINITIVE &#124; CASE</code> |
| Support | partial |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>noun</code> | <code>SUBJECT.DEFINITIONAL</code> | [Definitions](../spec/definitions.md#create-define-and-describe) · partial |
| <code>value</code> | <code>SUBJECT.VALUE</code> | [Definitions](../spec/definitions.md#create-define-and-describe) · partial |
| <code>verb</code> | <code>SUBJECT.UTILIZATION</code> | [Definitions](../spec/definitions.md#create-define-and-describe) · partial |
| <code>interface</code> | <code>SUBJECT.DECLARATIVE</code> | [Definitions](../spec/definitions.md#create-define-and-describe) · partial |

## clause.definitive.active

<a id="grammar-clause-definitive-active"></a>

Definition clauses establish source vocabulary or typed native-backed forms. Accepted construction and framing are bounded by definition expansion. The active case is the active grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries.

Semantic context: [Definitions](../spec/definitions.md#create-define-and-describe).

| Property | Configured value |
| --- | --- |
| Category | <code>CLAUSE</code> |
| Class | <code>DEFINITIVE.ACTIVE</code> |
| Structure | <code>[ADVERB]* &#124; VERB.DEFINITIVE.ACTIVE &#124; CASE</code> |
| Support | partial |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>noun</code> | <code>SUBJECT.DEFINITIONAL</code> | [Definitions](../spec/definitions.md#create-define-and-describe) · partial |
| <code>verb</code> | <code>SUBJECT.UTILIZATION</code> | [Definitions](../spec/definitions.md#create-define-and-describe) · partial |

## clause.definitive.passive

<a id="grammar-clause-definitive-passive"></a>

Definition clauses establish source vocabulary or typed native-backed forms. Accepted construction and framing are bounded by definition expansion. The passive case is the passive grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries.

Semantic context: [Definitions](../spec/definitions.md#create-define-and-describe).

| Property | Configured value |
| --- | --- |
| Category | <code>CLAUSE</code> |
| Class | <code>DEFINITIVE.PASSIVE</code> |
| Structure | <code>[ADVERB]* &#124; VERB.DEFINITIVE.PASSIVE &#124; CASE</code> |
| Support | partial |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>noun</code> | <code>PHRASE.DECLARATIVE.DEFINITIONAL</code> | [Definitions](../spec/definitions.md#create-define-and-describe) · partial |
| <code>verb</code> | <code>PHRASE.UTILIZATION</code> | [Definitions](../spec/definitions.md#create-define-and-describe) · partial |

## clause.definitive.applicative

<a id="grammar-clause-definitive-applicative"></a>

Definition clauses establish source vocabulary or typed native-backed forms. Accepted construction and framing are bounded by definition expansion. The applicative case is the applicative grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries.

Semantic context: [Definitions](../spec/definitions.md#create-define-and-describe).

| Property | Configured value |
| --- | --- |
| Category | <code>CLAUSE</code> |
| Class | <code>DEFINITIVE.APPLICATIVE</code> |
| Structure | <code>[ADVERB]* &#124; VERB.DEFINITIVE.APPLICATIVE &#124; CASE</code> |
| Support | partial |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>noun</code> | <code>SUBJECT.DEFINITIONAL</code> | [Definitions](../spec/definitions.md#create-define-and-describe) · partial |
| <code>verb</code> | <code>SUBJECT.UTILIZATION</code> | [Definitions](../spec/definitions.md#create-define-and-describe) · partial |

## clause.descriptive.imperative

<a id="grammar-clause-descriptive-imperative"></a>

Descriptive clauses apply source-defined adjective/adverb meanings. Constructor descriptors preserve the target and source-unit ownership. The imperative case is the imperative grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries.

Semantic context: [Descriptors](../spec/definitions.md#descriptors-and-constructors).

| Property | Configured value |
| --- | --- |
| Category | <code>CLAUSE</code> |
| Class | <code>DESCRIPTIVE.IMPERATIVE</code> |
| Structure | <code>[ADVERB]* &#124; VERB.DESCRIPTIVE &#124; CASE</code> |
| Support | partial |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>adjective</code> | <code>SUBJECT.DECLARATIVE</code> | [Descriptors](../spec/definitions.md#descriptors-and-constructors) · partial |
| <code>adverb</code> | <code>SUBJECT.UTILIZATION</code> | [Descriptors](../spec/definitions.md#descriptors-and-constructors) · partial |

## clause.descriptive.active

<a id="grammar-clause-descriptive-active"></a>

Descriptive clauses apply source-defined adjective/adverb meanings. Constructor descriptors preserve the target and source-unit ownership. The active case is the active grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries.

Semantic context: [Descriptors](../spec/definitions.md#descriptors-and-constructors).

| Property | Configured value |
| --- | --- |
| Category | <code>CLAUSE</code> |
| Class | <code>DESCRIPTIVE.ACTIVE</code> |
| Structure | <code>[ADVERB]* &#124; VERB.DESCRIPTIVE.ACTIVE &#124; CASE</code> |
| Support | partial |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>adjective</code> | <code>OBJECT</code> | [Descriptors](../spec/definitions.md#descriptors-and-constructors) · partial |
| <code>adverb</code> | <code>CLAUSE.ASSOCIATIVE.APPLICATIVE</code> | [Descriptors](../spec/definitions.md#descriptors-and-constructors) · partial |

## clause.descriptive.passive

<a id="grammar-clause-descriptive-passive"></a>

Descriptive clauses apply source-defined adjective/adverb meanings. Constructor descriptors preserve the target and source-unit ownership. The passive case is the passive grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries.

Semantic context: [Descriptors](../spec/definitions.md#descriptors-and-constructors).

| Property | Configured value |
| --- | --- |
| Category | <code>CLAUSE</code> |
| Class | <code>DESCRIPTIVE.PASSIVE</code> |
| Structure | <code>[ADVERB]* &#124; VERB.DESCRIPTIVE.PASSIVE &#124; CASE</code> |
| Support | partial |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>adjective</code> | <code>PHRASE.DECLARATIVE.DECLARATIVE</code> | [Descriptors](../spec/definitions.md#descriptors-and-constructors) · partial |
| <code>adverb</code> | <code>CLAUSE.ASSOCIATIVE.APPLICATIVE</code> | [Descriptors](../spec/definitions.md#descriptors-and-constructors) · partial |

## clause.descriptive.applicative

<a id="grammar-clause-descriptive-applicative"></a>

Descriptive clauses apply source-defined adjective/adverb meanings. Constructor descriptors preserve the target and source-unit ownership. The applicative case is the applicative grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries.

Semantic context: [Descriptors](../spec/definitions.md#descriptors-and-constructors).

| Property | Configured value |
| --- | --- |
| Category | <code>CLAUSE</code> |
| Class | <code>DESCRIPTIVE.APPLICATIVE</code> |
| Structure | <code>[ADVERB]* &#124; VERB.DESCRIPTIVE.APPLICATIVE &#124; CASE</code> |
| Support | partial |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>adjective</code> | <code>[OBJECT] &#124; PHRASE.DECLARATIVE.DECLARATIVE*</code> | [Descriptors](../spec/definitions.md#descriptors-and-constructors) · partial |
| <code>adverb</code> | <code>SUBJECT.DEFINITIVE &#124; CLAUSE.ASSOCIATIVE.APPLICATIVE</code> | [Descriptors](../spec/definitions.md#descriptors-and-constructors) · partial |

## clause.expansive.imperative

<a id="grammar-clause-expansive-imperative"></a>

These configured collection forms do not establish a supported general insertion, indexing, removal or iteration API. Storage representations and dictionary member namespaces have narrower observed contracts. The imperative case is the imperative grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries.

Semantic context: [Collections](../spec/collections-and-categories.md#lists).

| Property | Configured value |
| --- | --- |
| Category | <code>CLAUSE</code> |
| Class | <code>EXPANSIVE.IMPERATIVE</code> |
| Structure | <code>[ADVERB]* &#124; VERB.EXPANSIVE &#124; OBJECT &#124; CASE</code> |
| Support | partial |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>list</code> | <code>PHRASE.INFLECTIVE.LIST</code> | [Collections](../spec/collections-and-categories.md#lists) · partial |
| <code>dictionary</code> | <code>PHRASE.INFLECTIVE.DICTIONARY</code> | [Collections](../spec/collections-and-categories.md#lists) · partial |

## clause.expansive.active

<a id="grammar-clause-expansive-active"></a>

These configured collection forms do not establish a supported general insertion, indexing, removal or iteration API. Storage representations and dictionary member namespaces have narrower observed contracts. The active case is the active grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries.

Semantic context: [Collections](../spec/collections-and-categories.md#lists).

| Property | Configured value |
| --- | --- |
| Category | <code>CLAUSE</code> |
| Class | <code>EXPANSIVE.ACTIVE</code> |
| Structure | <code>[ADVERB]* &#124; VERB.EXPANSIVE.ACTIVE &#124; OBJECT &#124; CASE</code> |
| Support | partial |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>list</code> | <code>PHRASE.INFLECTIVE.LIST</code> | [Collections](../spec/collections-and-categories.md#lists) · partial |
| <code>dictionary</code> | <code>PHRASE.INFLECTIVE.DICTIONARY</code> | [Collections](../spec/collections-and-categories.md#lists) · partial |

## clause.expansive.passive

<a id="grammar-clause-expansive-passive"></a>

These configured collection forms do not establish a supported general insertion, indexing, removal or iteration API. Storage representations and dictionary member namespaces have narrower observed contracts. The passive case is the passive grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries.

Semantic context: [Collections](../spec/collections-and-categories.md#lists).

| Property | Configured value |
| --- | --- |
| Category | <code>CLAUSE</code> |
| Class | <code>EXPANSIVE.PASSIVE</code> |
| Structure | <code>[ADVERB]* &#124; VERB.EXPANSIVE.PASSIVE &#124; CASE</code> |
| Support | partial |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>list</code> | <code>PHRASE.INFLECTIVE.LIST</code> | [Collections](../spec/collections-and-categories.md#lists) · partial |
| <code>dictionary</code> | <code>PHRASE.INFLECTIVE.DICTIONARY</code> | [Collections](../spec/collections-and-categories.md#lists) · partial |

## clause.expansive.applicative

<a id="grammar-clause-expansive-applicative"></a>

These configured collection forms do not establish a supported general insertion, indexing, removal or iteration API. Storage representations and dictionary member namespaces have narrower observed contracts. The applicative case is the applicative grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries.

Semantic context: [Collections](../spec/collections-and-categories.md#lists).

| Property | Configured value |
| --- | --- |
| Category | <code>CLAUSE</code> |
| Class | <code>EXPANSIVE.APPLICATIVE</code> |
| Structure | <code>[ADVERB]* &#124; VERB.EXPANSIVE.APPLICATIVE &#124; OBJECT &#124; CASE</code> |
| Support | partial |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>list</code> | <code>PHRASE.INFLECTIVE.LIST</code> | [Collections](../spec/collections-and-categories.md#lists) · partial |
| <code>dictionary</code> | <code>PHRASE.INFLECTIVE.DICTIONARY</code> | [Collections](../spec/collections-and-categories.md#lists) · partial |

## clause.locative.imperative

<a id="grammar-clause-locative-imperative"></a>

These configured collection forms do not establish a supported general insertion, indexing, removal or iteration API. Storage representations and dictionary member namespaces have narrower observed contracts. The imperative case is the imperative grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries.

Semantic context: [Collections](../spec/collections-and-categories.md#lists).

| Property | Configured value |
| --- | --- |
| Category | <code>CLAUSE</code> |
| Class | <code>LOCATIVE.IMPERATIVE</code> |
| Structure | <code>[ADVERB]* &#124; VERB.LOCATIVE &#124; OBJECT &#124; CASE</code> |
| Support | partial |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>list</code> | <code>PHRASE.INFLECTIVE.LOCATIVE</code> | [Collections](../spec/collections-and-categories.md#lists) · partial |
| <code>dictionary</code> | <code>PHRASE.INFLECTIVE.DICTIONARY</code> | [Collections](../spec/collections-and-categories.md#lists) · partial |

## clause.locative.active

<a id="grammar-clause-locative-active"></a>

These configured collection forms do not establish a supported general insertion, indexing, removal or iteration API. Storage representations and dictionary member namespaces have narrower observed contracts. The active case is the active grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries.

Semantic context: [Collections](../spec/collections-and-categories.md#lists).

| Property | Configured value |
| --- | --- |
| Category | <code>CLAUSE</code> |
| Class | <code>LOCATIVE.ACTIVE</code> |
| Structure | <code>[ADVERB]* &#124; VERB.LOCATIVE.ACTIVE &#124; OBJECT &#124; CASE</code> |
| Support | partial |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>list</code> | <code>PHRASE.INFLECTIVE.LOCATIVE</code> | [Collections](../spec/collections-and-categories.md#lists) · partial |
| <code>dictionary</code> | <code>PHRASE.INFLECTIVE.DICTIONARY</code> | [Collections](../spec/collections-and-categories.md#lists) · partial |

## clause.locative.passive

<a id="grammar-clause-locative-passive"></a>

These configured collection forms do not establish a supported general insertion, indexing, removal or iteration API. Storage representations and dictionary member namespaces have narrower observed contracts. The passive case is the passive grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries.

Semantic context: [Collections](../spec/collections-and-categories.md#lists).

| Property | Configured value |
| --- | --- |
| Category | <code>CLAUSE</code> |
| Class | <code>LOCATIVE.PASSIVE</code> |
| Structure | <code>[ADVERB]* &#124; VERB.LOCATIVE.PASSIVE &#124; CASE</code> |
| Support | partial |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>list</code> | <code>PHRASE.INFLECTIVE.LOCATIVE</code> | [Collections](../spec/collections-and-categories.md#lists) · partial |
| <code>dictionary</code> | <code>PHRASE.INFLECTIVE.DICTIONARY</code> | [Collections](../spec/collections-and-categories.md#lists) · partial |

## clause.locative.applicative

<a id="grammar-clause-locative-applicative"></a>

These configured collection forms do not establish a supported general insertion, indexing, removal or iteration API. Storage representations and dictionary member namespaces have narrower observed contracts. The applicative case is the applicative grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries.

Semantic context: [Collections](../spec/collections-and-categories.md#lists).

| Property | Configured value |
| --- | --- |
| Category | <code>CLAUSE</code> |
| Class | <code>LOCATIVE.APPLICATIVE</code> |
| Structure | <code>[ADVERB]* &#124; VERB.LOCATIVE.APPLICATIVE &#124; OBJECT &#124; CASE</code> |
| Support | partial |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>list</code> | <code>PHRASE.INFLECTIVE.LOCATIVE</code> | [Collections](../spec/collections-and-categories.md#lists) · partial |
| <code>dictionary</code> | <code>PHRASE.INFLECTIVE.DICTIONARY</code> | [Collections](../spec/collections-and-categories.md#lists) · partial |

## clause.receptive.imperative

<a id="grammar-clause-receptive-imperative"></a>

These configured collection forms do not establish a supported general insertion, indexing, removal or iteration API. Storage representations and dictionary member namespaces have narrower observed contracts. The imperative case is the imperative grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries.

Semantic context: [Collections](../spec/collections-and-categories.md#lists).

| Property | Configured value |
| --- | --- |
| Category | <code>CLAUSE</code> |
| Class | <code>RECEPTIVE.IMPERATIVE</code> |
| Structure | <code>[ADVERB]* &#124; VERB.RECEPTIVE &#124; CASE &#124; PHRASE.OBJECTIVE</code> |
| Support | partial |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>list</code> | <code>OBJECT.NUMBER</code> | [Collections](../spec/collections-and-categories.md#lists) · partial |
| <code>dictionary</code> | <code>OBJECT.TEXT</code> | [Collections](../spec/collections-and-categories.md#lists) · partial |

## clause.receptive.active

<a id="grammar-clause-receptive-active"></a>

These configured collection forms do not establish a supported general insertion, indexing, removal or iteration API. Storage representations and dictionary member namespaces have narrower observed contracts. The active case is the active grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries.

Semantic context: [Collections](../spec/collections-and-categories.md#lists).

| Property | Configured value |
| --- | --- |
| Category | <code>CLAUSE</code> |
| Class | <code>RECEPTIVE.ACTIVE</code> |
| Structure | <code>[ADVERB]* &#124; VERB.RECEPTIVE.ACTIVE &#124; CASE &#124; PHRASE.OBJECTIVE</code> |
| Support | partial |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>list</code> | <code>OBJECT.NUMBER</code> | [Collections](../spec/collections-and-categories.md#lists) · partial |
| <code>dictionary</code> | <code>OBJECT.TEXT</code> | [Collections](../spec/collections-and-categories.md#lists) · partial |

## clause.receptive.passive

<a id="grammar-clause-receptive-passive"></a>

These configured collection forms do not establish a supported general insertion, indexing, removal or iteration API. Storage representations and dictionary member namespaces have narrower observed contracts. The passive case is the passive grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries.

Semantic context: [Collections](../spec/collections-and-categories.md#lists).

| Property | Configured value |
| --- | --- |
| Category | <code>CLAUSE</code> |
| Class | <code>RECEPTIVE.PASSIVE</code> |
| Structure | <code>[ADVERB]* &#124; VERB.RECEPTIVE.PASSIVE &#124; PHRASE.OBJECTIVE</code> |
| Support | partial |

This rule has no named case alternatives; its structure is used directly.

## clause.receptive.applicative

<a id="grammar-clause-receptive-applicative"></a>

These configured collection forms do not establish a supported general insertion, indexing, removal or iteration API. Storage representations and dictionary member namespaces have narrower observed contracts. The applicative case is the applicative grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries.

Semantic context: [Collections](../spec/collections-and-categories.md#lists).

| Property | Configured value |
| --- | --- |
| Category | <code>CLAUSE</code> |
| Class | <code>RECEPTIVE.APPLICATIVE</code> |
| Structure | <code>[ADVERB]* &#124; VERB.RECEPTIVE.APPLICATIVE &#124; CASE &#124; PHRASE.OBJECTIVE</code> |
| Support | partial |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>list</code> | <code>OBJECT.NUMBER</code> | [Collections](../spec/collections-and-categories.md#lists) · partial |
| <code>dictionary</code> | <code>OBJECT.TEXT</code> | [Collections](../spec/collections-and-categories.md#lists) · partial |

## clause.reductive.imperative

<a id="grammar-clause-reductive-imperative"></a>

These configured collection forms do not establish a supported general insertion, indexing, removal or iteration API. Storage representations and dictionary member namespaces have narrower observed contracts. The imperative case is the imperative grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries.

Semantic context: [Collections](../spec/collections-and-categories.md#lists).

| Property | Configured value |
| --- | --- |
| Category | <code>CLAUSE</code> |
| Class | <code>REDUCTIVE.IMPERATIVE</code> |
| Structure | <code>[ADVERB]* &#124; VERB.RECEPTIVE &#124; CASE &#124; PHRASE.OBJECTIVE</code> |
| Support | partial |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>list</code> | <code>OBJECT.NUMBER</code> | [Collections](../spec/collections-and-categories.md#lists) · partial |
| <code>dictionary</code> | <code>OBJECT.TEXT</code> | [Collections](../spec/collections-and-categories.md#lists) · partial |

## clause.reductive.active

<a id="grammar-clause-reductive-active"></a>

These configured collection forms do not establish a supported general insertion, indexing, removal or iteration API. Storage representations and dictionary member namespaces have narrower observed contracts. The active case is the active grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries.

Semantic context: [Collections](../spec/collections-and-categories.md#lists).

| Property | Configured value |
| --- | --- |
| Category | <code>CLAUSE</code> |
| Class | <code>REDUCTIVE.ACTIVE</code> |
| Structure | <code>[ADVERB]* &#124; VERB.RECEPTIVE.ACTIVE &#124; CASE &#124; PHRASE.OBJECTIVE</code> |
| Support | partial |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>list</code> | <code>OBJECT.NUMBER</code> | [Collections](../spec/collections-and-categories.md#lists) · partial |
| <code>dictionary</code> | <code>OBJECT.TEXT</code> | [Collections](../spec/collections-and-categories.md#lists) · partial |

## clause.reductive.passive

<a id="grammar-clause-reductive-passive"></a>

These configured collection forms do not establish a supported general insertion, indexing, removal or iteration API. Storage representations and dictionary member namespaces have narrower observed contracts. The passive case is the passive grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries.

Semantic context: [Collections](../spec/collections-and-categories.md#lists).

| Property | Configured value |
| --- | --- |
| Category | <code>CLAUSE</code> |
| Class | <code>REDUCTIVE.PASSIVE</code> |
| Structure | <code>[ADVERB]* &#124; VERB.RECEPTIVE.PASSIVE &#124; PHRASE.OBJECTIVE</code> |
| Support | partial |

This rule has no named case alternatives; its structure is used directly.

## clause.reductive.applicative

<a id="grammar-clause-reductive-applicative"></a>

These configured collection forms do not establish a supported general insertion, indexing, removal or iteration API. Storage representations and dictionary member namespaces have narrower observed contracts. The applicative case is the applicative grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries.

Semantic context: [Collections](../spec/collections-and-categories.md#lists).

| Property | Configured value |
| --- | --- |
| Category | <code>CLAUSE</code> |
| Class | <code>REDUCTIVE.APPLICATIVE</code> |
| Structure | <code>[ADVERB]* &#124; VERB.RECEPTIVE.APPLICATIVE &#124; CASE &#124; PHRASE.OBJECTIVE</code> |
| Support | partial |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>list</code> | <code>OBJECT.NUMBER</code> | [Collections](../spec/collections-and-categories.md#lists) · partial |
| <code>dictionary</code> | <code>OBJECT.TEXT</code> | [Collections](../spec/collections-and-categories.md#lists) · partial |

## clause.associative.imperative

<a id="grammar-clause-associative-imperative"></a>

Association names an existing canonical value or definition. It does not create an independent runtime copy. The imperative case is the imperative grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries.

Semantic context: [Aliases](../spec/names-and-scope.md#aliases-and-canonical-identity).

| Property | Configured value |
| --- | --- |
| Category | <code>CLAUSE</code> |
| Class | <code>ASSOCIATIVE.IMPERATIVE</code> |
| Structure | <code>[ADVERB]* &#124; VERB.ASSOCIATIVE &#124; SUBJECT.DEFINITIVE &#124; CASE*</code> |
| Support | partial |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>alias</code> | <code>PHRASE.DECLARATIVE.ALIAS</code> | [Aliases](../spec/names-and-scope.md#aliases-and-canonical-identity) · partial |

## clause.associative.active

<a id="grammar-clause-associative-active"></a>

Association names an existing canonical value or definition. It does not create an independent runtime copy. The active case is the active grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries.

Semantic context: [Aliases](../spec/names-and-scope.md#aliases-and-canonical-identity).

| Property | Configured value |
| --- | --- |
| Category | <code>CLAUSE</code> |
| Class | <code>ASSOCIATIVE.ACTIVE</code> |
| Structure | <code>[ADVERB]* &#124; VERB.ASSOCIATIVE.ACTIVE &#124; SUBJECT.DEFINITIVE &#124; CASE*</code> |
| Support | partial |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>alias</code> | <code>PHRASE.DECLARATIVE.ALIAS</code> | [Aliases](../spec/names-and-scope.md#aliases-and-canonical-identity) · partial |

## clause.associative.passive

<a id="grammar-clause-associative-passive"></a>

Association names an existing canonical value or definition. It does not create an independent runtime copy. The passive case is the passive grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries.

Semantic context: [Aliases](../spec/names-and-scope.md#aliases-and-canonical-identity).

| Property | Configured value |
| --- | --- |
| Category | <code>CLAUSE</code> |
| Class | <code>ASSOCIATIVE.PASSIVE</code> |
| Structure | <code>[ADVERB]* &#124; VERB.ASSOCIATIVE.PASSIVE &#124; CASE</code> |
| Support | partial |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>definitive</code> | <code>PHRASE.DECLARATIVE</code> | [Aliases](../spec/names-and-scope.md#aliases-and-canonical-identity) · partial |
| <code>associative</code> | <code>PHRASE.INDICATIVE</code> | [Aliases](../spec/names-and-scope.md#aliases-and-canonical-identity) · partial |

## clause.associative.applicative

<a id="grammar-clause-associative-applicative"></a>

Association names an existing canonical value or definition. It does not create an independent runtime copy. The applicative case is the applicative grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries.

Semantic context: [Aliases](../spec/names-and-scope.md#aliases-and-canonical-identity).

| Property | Configured value |
| --- | --- |
| Category | <code>CLAUSE</code> |
| Class | <code>ASSOCIATIVE.APPLICATIVE</code> |
| Structure | <code>[ADVERB]* &#124; VERB.ASSOCIATIVE.APPLICATIVE &#124; OBJECT &#124; CASE*</code> |
| Support | partial |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>alias</code> | <code>PHRASE.DECLARATIVE.ALIAS</code> | [Aliases](../spec/names-and-scope.md#aliases-and-canonical-identity) · partial |

## clause.generative.imperative

<a id="grammar-clause-generative-imperative"></a>

Build is the runtime readiness operation. It is distinct from the CLI command that compiles native output. The imperative case is the imperative grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries.

Semantic context: [Build](../spec/execution.md#build-run-stop-and-await).

| Property | Configured value |
| --- | --- |
| Category | <code>CLAUSE</code> |
| Class | <code>GENERATIVE.IMPERATIVE</code> |
| Structure | <code>[ADVERB]* &#124; VERB.GENERATIVE &#124; OBJECT</code> |
| Support | partial |

This rule has no named case alternatives; its structure is used directly.

## clause.generative.active

<a id="grammar-clause-generative-active"></a>

Build is the runtime readiness operation. It is distinct from the CLI command that compiles native output. The active case is the active grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries.

Semantic context: [Build](../spec/execution.md#build-run-stop-and-await).

| Property | Configured value |
| --- | --- |
| Category | <code>CLAUSE</code> |
| Class | <code>GENERATIVE.ACTIVE</code> |
| Structure | <code>[ADVERB]* &#124; VERB.GENERATIVE.ACTIVE &#124; OBJECT</code> |
| Support | partial |

This rule has no named case alternatives; its structure is used directly.

## clause.generative.passive

<a id="grammar-clause-generative-passive"></a>

Build is the runtime readiness operation. It is distinct from the CLI command that compiles native output. The passive case is the passive grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries.

Semantic context: [Build](../spec/execution.md#build-run-stop-and-await).

| Property | Configured value |
| --- | --- |
| Category | <code>CLAUSE</code> |
| Class | <code>GENERATIVE.PASSIVE</code> |
| Structure | <code>[ADVERB]* &#124; VERB.GENERATIVE.PASSIVE</code> |
| Support | partial |

This rule has no named case alternatives; its structure is used directly.

## clause.generative.applicative

<a id="grammar-clause-generative-applicative"></a>

Build is the runtime readiness operation. It is distinct from the CLI command that compiles native output. The applicative case is the applicative grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries.

Semantic context: [Build](../spec/execution.md#build-run-stop-and-await).

| Property | Configured value |
| --- | --- |
| Category | <code>CLAUSE</code> |
| Class | <code>GENERATIVE.APPLICATIVE</code> |
| Structure | <code>[ADVERB]* &#124; VERB.GENERATIVE.APPLICATIVE &#124; OBJECT</code> |
| Support | partial |

This rule has no named case alternatives; its structure is used directly.

## clause.executive.imperative

<a id="grammar-clause-executive-imperative"></a>

Run launches the selected process. The grammar does not imply that every applicative category is runnable. The imperative case is the imperative grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries.

Semantic context: [Run](../spec/execution.md#build-run-stop-and-await).

| Property | Configured value |
| --- | --- |
| Category | <code>CLAUSE</code> |
| Class | <code>EXECUTIVE.IMPERATIVE</code> |
| Structure | <code>[ADVERB]* &#124; VERB.EXECUTIVE &#124; OBJECT</code> |
| Support | partial |

This rule has no named case alternatives; its structure is used directly.

## clause.executive.active

<a id="grammar-clause-executive-active"></a>

Run launches the selected process. The grammar does not imply that every applicative category is runnable. The active case is the active grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries.

Semantic context: [Run](../spec/execution.md#build-run-stop-and-await).

| Property | Configured value |
| --- | --- |
| Category | <code>CLAUSE</code> |
| Class | <code>EXECUTIVE.ACTIVE</code> |
| Structure | <code>[ADVERB]* &#124; VERB.EXECUTIVE.ACTIVE &#124; OBJECT</code> |
| Support | partial |

This rule has no named case alternatives; its structure is used directly.

## clause.executive.passive

<a id="grammar-clause-executive-passive"></a>

Run launches the selected process. The grammar does not imply that every applicative category is runnable. The passive case is the passive grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries.

Semantic context: [Run](../spec/execution.md#build-run-stop-and-await).

| Property | Configured value |
| --- | --- |
| Category | <code>CLAUSE</code> |
| Class | <code>EXECUTIVE.PASSIVE</code> |
| Structure | <code>[ADVERB]* &#124; VERB.EXECUTIVE.PASSIVE</code> |
| Support | partial |

This rule has no named case alternatives; its structure is used directly.

## clause.executive.applicative

<a id="grammar-clause-executive-applicative"></a>

Run launches the selected process. The grammar does not imply that every applicative category is runnable. The applicative case is the applicative grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries.

Semantic context: [Run](../spec/execution.md#build-run-stop-and-await).

| Property | Configured value |
| --- | --- |
| Category | <code>CLAUSE</code> |
| Class | <code>EXECUTIVE.APPLICATIVE</code> |
| Structure | <code>[ADVERB]* &#124; VERB.EXECUTIVE.APPLICATIVE &#124; OBJECT</code> |
| Support | partial |

This rule has no named case alternatives; its structure is used directly.

## clause.suspensive.imperative

<a id="grammar-clause-suspensive-imperative"></a>

Await joins a selected process invocation and activates its continuation once. A parsed clause alone does not establish valid activation placement. The imperative case is the imperative grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries.

Semantic context: [Await](../spec/execution.md#continuations-and-activation).

| Property | Configured value |
| --- | --- |
| Category | <code>CLAUSE</code> |
| Class | <code>SUSPENSIVE.IMPERATIVE</code> |
| Structure | <code>[ADVERB]* &#124; VERB.SUSPENSIVE &#124; OBJECT</code> |
| Support | partial |

This rule has no named case alternatives; its structure is used directly.

## clause.suspensive.active

<a id="grammar-clause-suspensive-active"></a>

Await joins a selected process invocation and activates its continuation once. A parsed clause alone does not establish valid activation placement. The active case is the active grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries.

Semantic context: [Await](../spec/execution.md#continuations-and-activation).

| Property | Configured value |
| --- | --- |
| Category | <code>CLAUSE</code> |
| Class | <code>SUSPENSIVE.ACTIVE</code> |
| Structure | <code>[ADVERB]* &#124; VERB.SUSPENSIVE.ACTIVE &#124; OBJECT</code> |
| Support | partial |

This rule has no named case alternatives; its structure is used directly.

## clause.suspensive.passive

<a id="grammar-clause-suspensive-passive"></a>

Await joins a selected process invocation and activates its continuation once. A parsed clause alone does not establish valid activation placement. The passive case is the passive grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries.

Semantic context: [Await](../spec/execution.md#continuations-and-activation).

| Property | Configured value |
| --- | --- |
| Category | <code>CLAUSE</code> |
| Class | <code>SUSPENSIVE.PASSIVE</code> |
| Structure | <code>[ADVERB]* &#124; VERB.SUSPENSIVE.PASSIVE</code> |
| Support | partial |

This rule has no named case alternatives; its structure is used directly.

## clause.suspensive.applicative

<a id="grammar-clause-suspensive-applicative"></a>

Await joins a selected process invocation and activates its continuation once. A parsed clause alone does not establish valid activation placement. The applicative case is the applicative grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries.

Semantic context: [Await](../spec/execution.md#continuations-and-activation).

| Property | Configured value |
| --- | --- |
| Category | <code>CLAUSE</code> |
| Class | <code>SUSPENSIVE.APPLICATIVE</code> |
| Structure | <code>[ADVERB]* &#124; VERB.SUSPENSIVE.APPLICATIVE &#124; OBJECT</code> |
| Support | partial |

This rule has no named case alternatives; its structure is used directly.

## clause.interruptive.imperative

<a id="grammar-clause-interruptive-imperative"></a>

Stop requests termination without joining. Final output still requires the Await contract. The imperative case is the imperative grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries.

Semantic context: [Stop](../spec/execution.md#build-run-stop-and-await).

| Property | Configured value |
| --- | --- |
| Category | <code>CLAUSE</code> |
| Class | <code>INTERRUPTIVE.IMPERATIVE</code> |
| Structure | <code>[ADVERB]* &#124; VERB.INTERRUPTIVE &#124; OBJECT</code> |
| Support | partial |

This rule has no named case alternatives; its structure is used directly.

## clause.interruptive.active

<a id="grammar-clause-interruptive-active"></a>

Stop requests termination without joining. Final output still requires the Await contract. The active case is the active grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries.

Semantic context: [Stop](../spec/execution.md#build-run-stop-and-await).

| Property | Configured value |
| --- | --- |
| Category | <code>CLAUSE</code> |
| Class | <code>INTERRUPTIVE.ACTIVE</code> |
| Structure | <code>[ADVERB]* &#124; VERB.INTERRUPTIVE.ACTIVE &#124; OBJECT</code> |
| Support | partial |

This rule has no named case alternatives; its structure is used directly.

## clause.interruptive.passive

<a id="grammar-clause-interruptive-passive"></a>

Stop requests termination without joining. Final output still requires the Await contract. The passive case is the passive grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries.

Semantic context: [Stop](../spec/execution.md#build-run-stop-and-await).

| Property | Configured value |
| --- | --- |
| Category | <code>CLAUSE</code> |
| Class | <code>INTERRUPTIVE.PASSIVE</code> |
| Structure | <code>[ADVERB]* &#124; VERB.INTERRUPTIVE.PASSIVE</code> |
| Support | partial |

This rule has no named case alternatives; its structure is used directly.

## clause.interruptive.applicative

<a id="grammar-clause-interruptive-applicative"></a>

Stop requests termination without joining. Final output still requires the Await contract. The applicative case is the applicative grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries.

Semantic context: [Stop](../spec/execution.md#build-run-stop-and-await).

| Property | Configured value |
| --- | --- |
| Category | <code>CLAUSE</code> |
| Class | <code>INTERRUPTIVE.APPLICATIVE</code> |
| Structure | <code>[ADVERB]* &#124; VERB.INTERRUPTIVE.APPLICATIVE &#124; OBJECT</code> |
| Support | partial |

This rule has no named case alternatives; its structure is used directly.

## clause.inflective.imperative

<a id="grammar-clause-inflective-imperative"></a>

Import/Export-style configured vocabulary is not a native outer instruction. The supported Use dictionary/source import path is documented separately. The imperative case is the imperative grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries.

Semantic context: [Import and export vocabulary](../spec/imports.md#indexed-dictionaries).

| Property | Configured value |
| --- | --- |
| Category | <code>CLAUSE</code> |
| Class | <code>INFLECTIVE.IMPERATIVE</code> |
| Structure | <code>[ADVERB]* &#124; VERB.INFLECTIVE &#124; OBJECT</code> |
| Support | partial |

This rule has no named case alternatives; its structure is used directly.

## clause.inflective.active

<a id="grammar-clause-inflective-active"></a>

Import/Export-style configured vocabulary is not a native outer instruction. The supported Use dictionary/source import path is documented separately. The active case is the active grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries.

Semantic context: [Import and export vocabulary](../spec/imports.md#indexed-dictionaries).

| Property | Configured value |
| --- | --- |
| Category | <code>CLAUSE</code> |
| Class | <code>INFLECTIVE.ACTIVE</code> |
| Structure | <code>[ADVERB]* &#124; VERB.INFLECTIVE.ACTIVE &#124; OBJECT</code> |
| Support | partial |

This rule has no named case alternatives; its structure is used directly.

## clause.inflective.passive

<a id="grammar-clause-inflective-passive"></a>

Import/Export-style configured vocabulary is not a native outer instruction. The supported Use dictionary/source import path is documented separately. The passive case is the passive grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries.

Semantic context: [Import and export vocabulary](../spec/imports.md#indexed-dictionaries).

| Property | Configured value |
| --- | --- |
| Category | <code>CLAUSE</code> |
| Class | <code>INFLECTIVE.PASSIVE</code> |
| Structure | <code>[ADVERB]* &#124; VERB.INFLECTIVE.PASSIVE</code> |
| Support | partial |

This rule has no named case alternatives; its structure is used directly.

## clause.inflective.applicative

<a id="grammar-clause-inflective-applicative"></a>

Import/Export-style configured vocabulary is not a native outer instruction. The supported Use dictionary/source import path is documented separately. The applicative case is the applicative grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries.

Semantic context: [Import and export vocabulary](../spec/imports.md#indexed-dictionaries).

| Property | Configured value |
| --- | --- |
| Category | <code>CLAUSE</code> |
| Class | <code>INFLECTIVE.APPLICATIVE</code> |
| Structure | <code>[ADVERB]* &#124; VERB.INFLECTIVE.APPLICATIVE &#124; OBJECT</code> |
| Support | partial |

This rule has no named case alternatives; its structure is used directly.

## clause.deflective.imperative

<a id="grammar-clause-deflective-imperative"></a>

Import/Export-style configured vocabulary is not a native outer instruction. The supported Use dictionary/source import path is documented separately. The imperative case is the imperative grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries.

Semantic context: [Import and export vocabulary](../spec/imports.md#indexed-dictionaries).

| Property | Configured value |
| --- | --- |
| Category | <code>CLAUSE</code> |
| Class | <code>DEFLECTIVE.IMPERATIVE</code> |
| Structure | <code>[ADVERB]* &#124; VERB.DEFLECTIVE &#124; OBJECT</code> |
| Support | partial |

This rule has no named case alternatives; its structure is used directly.

## clause.deflective.active

<a id="grammar-clause-deflective-active"></a>

Import/Export-style configured vocabulary is not a native outer instruction. The supported Use dictionary/source import path is documented separately. The active case is the active grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries.

Semantic context: [Import and export vocabulary](../spec/imports.md#indexed-dictionaries).

| Property | Configured value |
| --- | --- |
| Category | <code>CLAUSE</code> |
| Class | <code>DEFLECTIVE.ACTIVE</code> |
| Structure | <code>[ADVERB]* &#124; VERB.DEFLECTIVE.ACTIVE &#124; OBJECT</code> |
| Support | partial |

This rule has no named case alternatives; its structure is used directly.

## clause.deflective.passive

<a id="grammar-clause-deflective-passive"></a>

Import/Export-style configured vocabulary is not a native outer instruction. The supported Use dictionary/source import path is documented separately. The passive case is the passive grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries.

Semantic context: [Import and export vocabulary](../spec/imports.md#indexed-dictionaries).

| Property | Configured value |
| --- | --- |
| Category | <code>CLAUSE</code> |
| Class | <code>DEFLECTIVE.PASSIVE</code> |
| Structure | <code>[ADVERB]* &#124; VERB.DEFLECTIVE.PASSIVE</code> |
| Support | partial |

This rule has no named case alternatives; its structure is used directly.

## clause.deflective.applicative

<a id="grammar-clause-deflective-applicative"></a>

Import/Export-style configured vocabulary is not a native outer instruction. The supported Use dictionary/source import path is documented separately. The applicative case is the applicative grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries.

Semantic context: [Import and export vocabulary](../spec/imports.md#indexed-dictionaries).

| Property | Configured value |
| --- | --- |
| Category | <code>CLAUSE</code> |
| Class | <code>DEFLECTIVE.APPLICATIVE</code> |
| Structure | <code>[ADVERB]* &#124; VERB.DEFLECTIVE.APPLICATIVE &#124; OBJECT</code> |
| Support | partial |

This rule has no named case alternatives; its structure is used directly.

## clause.observative.imperative

<a id="grammar-clause-observative-imperative"></a>

Show emits the resolved operand without Print’s added newline; Output reads are subject to process finalization. The imperative case is the imperative grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries.

Semantic context: [Show](../spec/control-flow.md#show-and-output).

| Property | Configured value |
| --- | --- |
| Category | <code>CLAUSE</code> |
| Class | <code>OBSERVATIVE.IMPERATIVE</code> |
| Structure | <code>[ADVERB]* &#124; VERB.OBSERVATIVE &#124; OBJECT</code> |
| Support | partial |

This rule has no named case alternatives; its structure is used directly.

## clause.observative.active

<a id="grammar-clause-observative-active"></a>

Show emits the resolved operand without Print’s added newline; Output reads are subject to process finalization. The active case is the active grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries.

Semantic context: [Show](../spec/control-flow.md#show-and-output).

| Property | Configured value |
| --- | --- |
| Category | <code>CLAUSE</code> |
| Class | <code>OBSERVATIVE.ACTIVE</code> |
| Structure | <code>[ADVERB]* &#124; VERB.OBSERVATIVE.ACTIVE &#124; OBJECT</code> |
| Support | partial |

This rule has no named case alternatives; its structure is used directly.

## clause.observative.passive

<a id="grammar-clause-observative-passive"></a>

Show emits the resolved operand without Print’s added newline; Output reads are subject to process finalization. The passive case is the passive grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries.

Semantic context: [Show](../spec/control-flow.md#show-and-output).

| Property | Configured value |
| --- | --- |
| Category | <code>CLAUSE</code> |
| Class | <code>OBSERVATIVE.PASSIVE</code> |
| Structure | <code>[ADVERB]* &#124; VERB.OBSERVATIVE.PASSIVE</code> |
| Support | partial |

This rule has no named case alternatives; its structure is used directly.

## clause.observative.applicative

<a id="grammar-clause-observative-applicative"></a>

Show emits the resolved operand without Print’s added newline; Output reads are subject to process finalization. The applicative case is the applicative grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries.

Semantic context: [Show](../spec/control-flow.md#show-and-output).

| Property | Configured value |
| --- | --- |
| Category | <code>CLAUSE</code> |
| Class | <code>OBSERVATIVE.APPLICATIVE</code> |
| Structure | <code>[ADVERB]* &#124; VERB.OBSERVATIVE.APPLICATIVE &#124; OBJECT</code> |
| Support | partial |

This rule has no named case alternatives; its structure is used directly.

## clause.terminative.imperative

<a id="grammar-clause-terminative-imperative"></a>

Exit ends the compiled program; source Function.Return is a separate builtin expansion. The imperative case is the imperative grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries.

Semantic context: [Exit](../spec/control-flow.md#exit-and-return).

| Property | Configured value |
| --- | --- |
| Category | <code>CLAUSE</code> |
| Class | <code>TERMINATIVE.IMPERATIVE</code> |
| Structure | <code>[ADVERB]* &#124; VERB.TERMINATIVE &#124; CASE*</code> |
| Support | partial |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>objective</code> | <code>OBJECT</code> | [Exit](../spec/control-flow.md#exit-and-return) · partial |
| <code>associative</code> | <code>PHRASE.ASSOCIATIVE</code> | [Exit](../spec/control-flow.md#exit-and-return) · partial |

## clause.terminative.active

<a id="grammar-clause-terminative-active"></a>

Exit ends the compiled program; source Function.Return is a separate builtin expansion. The active case is the active grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries.

Semantic context: [Exit](../spec/control-flow.md#exit-and-return).

| Property | Configured value |
| --- | --- |
| Category | <code>CLAUSE</code> |
| Class | <code>TERMINATIVE.ACTIVE</code> |
| Structure | <code>[ADVERB]* &#124; VERB.TERMINATIVE.ACTIVE &#124; CASE</code> |
| Support | partial |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>objective</code> | <code>OBJECT</code> | [Exit](../spec/control-flow.md#exit-and-return) · partial |
| <code>associative</code> | <code>PHRASE.ASSOCIATIVE</code> | [Exit](../spec/control-flow.md#exit-and-return) · partial |

## clause.terminative.passive

<a id="grammar-clause-terminative-passive"></a>

Exit ends the compiled program; source Function.Return is a separate builtin expansion. The passive case is the passive grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries.

Semantic context: [Exit](../spec/control-flow.md#exit-and-return).

| Property | Configured value |
| --- | --- |
| Category | <code>CLAUSE</code> |
| Class | <code>TERMINATIVE.PASSIVE</code> |
| Structure | <code>[ADVERB]* &#124; VERB.TERMINATIVE.PASSIVE &#124; CASE</code> |
| Support | partial |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>objective</code> | <code>OBJECT</code> | [Exit](../spec/control-flow.md#exit-and-return) · partial |
| <code>associative</code> | <code>PHRASE.ASSOCIATIVE</code> | [Exit](../spec/control-flow.md#exit-and-return) · partial |

## clause.terminative.applicative

<a id="grammar-clause-terminative-applicative"></a>

Exit ends the compiled program; source Function.Return is a separate builtin expansion. The applicative case is the applicative grammatical structure shown below; it must still satisfy the same semantic binding and lowering boundaries.

Semantic context: [Exit](../spec/control-flow.md#exit-and-return).

| Property | Configured value |
| --- | --- |
| Category | <code>CLAUSE</code> |
| Class | <code>TERMINATIVE.APPLICATIVE</code> |
| Structure | <code>[ADVERB]* &#124; VERB.TERMINATIVE.APPLICATIVE &#124; CASE</code> |
| Support | partial |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>objective</code> | <code>OBJECT</code> | [Exit](../spec/control-flow.md#exit-and-return) · partial |
| <code>associative</code> | <code>PHRASE.ASSOCIATIVE</code> | [Exit](../spec/control-flow.md#exit-and-return) · partial |

## condition

<a id="grammar-condition"></a>

Scalar conditions have specific type constraints. Joined conditions are rejected by outer lowering; no general short-circuit Boolean expression language is established.

Semantic context: [Conditions](../spec/control-flow.md#comparisons).

| Property | Configured value |
| --- | --- |
| Category | <code>CONDITION</code> |
| Class | <code>CONDITION</code> |
| Structure | <code>CASE</code> |
| Support | partial |

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

Scalar conditions have specific type constraints. Joined conditions are rejected by outer lowering; no general short-circuit Boolean expression language is established.

Semantic context: [Conditions](../spec/control-flow.md#comparisons).

| Property | Configured value |
| --- | --- |
| Category | <code>CONDITIONAL</code> |
| Class | <code>IF</code> |
| Structure | <code>EVALUATION.CONDITIONAL &#124; [CONDITION] &#124; [CLAUSE.IMPERATIVE]</code> |
| Support | partial |

This rule has no named case alternatives; its structure is used directly.

## conditional.else

<a id="grammar-conditional-else"></a>

Scalar conditions have specific type constraints. Joined conditions are rejected by outer lowering; no general short-circuit Boolean expression language is established.

Semantic context: [Conditions](../spec/control-flow.md#comparisons).

| Property | Configured value |
| --- | --- |
| Category | <code>CONDITIONAL</code> |
| Class | <code>ELSE</code> |
| Structure | <code>EVALUATION.EXCEPTIONAL &#124; CASE</code> |
| Support | partial |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>if</code> | <code>CONDITIONAL.IF</code> | [Conditions](../spec/control-flow.md#comparisons) · partial |
| <code>else</code> | <code>[CLAUSE.IMPERATIVE]</code> | [Conditions](../spec/control-flow.md#comparisons) · partial |

## conditional.while

<a id="grammar-conditional-while"></a>

The parser recognizes this shape, but outer instruction lowering explicitly rejects while. Rust backend iteration cases do not establish source-language loop support.

Semantic context: [Iteration rejection](../spec/control-flow.md#iteration-and-collection-operations).

| Property | Configured value |
| --- | --- |
| Category | <code>CONDITIONAL</code> |
| Class | <code>WHILE</code> |
| Structure | <code>EVALUATION.ITERATIVE &#124; [CONDITION] &#124; CLAUSE.IMPERATIVE</code> |
| Support | rejected |

This rule has no named case alternatives; its structure is used directly.

## conditional.other

<a id="grammar-conditional-other"></a>

Scalar conditions have specific type constraints. Joined conditions are rejected by outer lowering; no general short-circuit Boolean expression language is established.

Semantic context: [Conditions](../spec/control-flow.md#comparisons).

| Property | Configured value |
| --- | --- |
| Category | <code>CONDITIONAL</code> |
| Class | <code>OTHER</code> |
| Structure | <code>EVALUATION.ALTERNATIVE &#124; [CLAUSE.IMPERATIVE]</code> |
| Support | partial |

This rule has no named case alternatives; its structure is used directly.

## predicate.imperative

<a id="grammar-predicate-imperative"></a>

This wrapper selects a clause or conditional family. Complete statements must consume their associated tokens; a named alternative is not proof of native operational support.

Semantic context: [Predicates and statements](../spec/grammar.md#predicates-and-statements).

| Property | Configured value |
| --- | --- |
| Category | <code>PREDICATE</code> |
| Class | <code>IMPERATIVE</code> |
| Structure | <code>CASE &#124; [CLAUSE]*</code> |
| Support | partial |

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

This wrapper selects a clause or conditional family. Complete statements must consume their associated tokens; a named alternative is not proof of native operational support.

Semantic context: [Predicates and statements](../spec/grammar.md#predicates-and-statements).

| Property | Configured value |
| --- | --- |
| Category | <code>PREDICATE</code> |
| Class | <code>ACTIVE</code> |
| Structure | <code>CASE &#124; [CLAUSE]*</code> |
| Support | partial |

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

This wrapper selects a clause or conditional family. Complete statements must consume their associated tokens; a named alternative is not proof of native operational support.

Semantic context: [Predicates and statements](../spec/grammar.md#predicates-and-statements).

| Property | Configured value |
| --- | --- |
| Category | <code>PREDICATE</code> |
| Class | <code>PASSIVE</code> |
| Structure | <code>VERB.PASSIVE &#124; CASE &#124; [CLAUSE]*</code> |
| Support | partial |

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

This wrapper selects a clause or conditional family. Complete statements must consume their associated tokens; a named alternative is not proof of native operational support.

Semantic context: [Predicates and statements](../spec/grammar.md#predicates-and-statements).

| Property | Configured value |
| --- | --- |
| Category | <code>PREDICATE</code> |
| Class | <code>APPLICATIVE</code> |
| Structure | <code>CASE &#124; [CLAUSE]*</code> |
| Support | partial |

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

This wrapper selects a clause or conditional family. Complete statements must consume their associated tokens; a named alternative is not proof of native operational support.

Semantic context: [Predicates and statements](../spec/grammar.md#predicates-and-statements).

| Property | Configured value |
| --- | --- |
| Category | <code>PREDICATE</code> |
| Class | <code>CONDITIONAL</code> |
| Structure | <code>CASE</code> |
| Support | partial |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>condition</code> | <code>CONDITIONAL.IF &#124; [CONDITIONAL.ELSE.IF]* &#124; CONDITIONAL.ELSE*</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>iteration</code> | <code>CONDITIONAL.WHILE &#124; ADVERB.PROGRESSION* &#124; CLAUSE.IMPERATIVE*</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>other</code> | <code>CONDITIONAL.OTHER</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |

## statement.objective.imperative

<a id="grammar-statement-objective-imperative"></a>

This wrapper selects a clause or conditional family. Complete statements must consume their associated tokens; a named alternative is not proof of native operational support.

Semantic context: [Predicates and statements](../spec/grammar.md#predicates-and-statements).

| Property | Configured value |
| --- | --- |
| Category | <code>STATEMENT</code> |
| Class | <code>OBJECTIVE.IMPERATIVE</code> |
| Structure | <code>CASE</code> |
| Support | partial |

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

This wrapper selects a clause or conditional family. Complete statements must consume their associated tokens; a named alternative is not proof of native operational support.

Semantic context: [Predicates and statements](../spec/grammar.md#predicates-and-statements).

| Property | Configured value |
| --- | --- |
| Category | <code>STATEMENT</code> |
| Class | <code>OBJECTIVE.CONDITIONAL</code> |
| Structure | <code>CASE</code> |
| Support | partial |

| Named case | Sequence | Semantic context |
| --- | --- | --- |
| <code>conditional</code> | <code>PREDICATE.CONDITIONAL.CONDITION</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>iterative</code> | <code>PREDICATE.CONDITIONAL.ITERATION</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |
| <code>alternative</code> | <code>PREDICATE.CONDITIONAL.OTHER</code> | [Predicates and statements](../spec/grammar.md#predicates-and-statements) · partial |

## statement.subjective.active

<a id="grammar-statement-subjective-active"></a>

This wrapper selects a clause or conditional family. Complete statements must consume their associated tokens; a named alternative is not proof of native operational support.

Semantic context: [Predicates and statements](../spec/grammar.md#predicates-and-statements).

| Property | Configured value |
| --- | --- |
| Category | <code>STATEMENT</code> |
| Class | <code>SUBJECTIVE.ACTIVE</code> |
| Structure | <code>SUBJECT &#124; CASE</code> |
| Support | partial |

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

This wrapper selects a clause or conditional family. Complete statements must consume their associated tokens; a named alternative is not proof of native operational support.

Semantic context: [Predicates and statements](../spec/grammar.md#predicates-and-statements).

| Property | Configured value |
| --- | --- |
| Category | <code>STATEMENT</code> |
| Class | <code>SUBJECTIVE.PASSIVE</code> |
| Structure | <code>SUBJECT &#124; CASE</code> |
| Support | partial |

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
