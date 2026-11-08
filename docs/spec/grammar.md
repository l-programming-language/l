# Grammar and vocabulary

The exact English grammar lives in `lexicon/eng/grammar.json`; the exact configured words live in `lexicon/eng/vocabulary.json`. This chapter explains their notation and semantic groupings. The generated inventory preserves every rule, named case, and vocabulary entry without treating every table combination as executable.

## Inventory and authority

The inspected English grammar has 99 rules and 306 named cases. Its categories are three NOUN rules, thirteen PHRASE rules, sixty-nine CLAUSE rules, one CONDITION rule, four CONDITIONAL rules, five PREDICATE rules, and four STATEMENT rules. The vocabulary has 72 entries across eight parts of speech. These counts describe the reviewed source snapshot; freshness checks must catch later changes.

The [complete rule and case inventory](../_generated/grammar.md) provides a stable entry for every configured recognizer, and the [vocabulary inventory](../_generated/vocabulary.md) retains each root, inflection, role, and handler description. These tables supplement the semantic explanations below.

The grammar controls recognition. Bound source declarations, descriptor construction, typed instruction lowering, and backend support independently determine whether a recognized sentence can be built. A retained grammar entry may be partial or explicitly rejected in executable contexts.

## Configuration notation

The notation is specialized and is not EBNF:

| Notation | Actual parser meaning |
| --- | --- |
| `A \| B` | Sequence: match A, then B |
| `TERM*` | Optional TERM; failure leaves the prior state intact |
| `[TERM]` | Multiple association of TERM, using the grammar's coordinated-sequence handling |
| `[TERM]*` | Optional multiple association |
| `CASE` | Select a named case for this rule |
| `ID` | A resolved/reference identifier position |
| `"ID"` | A declaration identifier position |
| `FAMILY.kind.case` | Qualified grammar or lexical role |
| backtick `~` pattern | Text token category |
| `%~%`, `?~?`, `!~!` | Number, Truth, Null token categories |
| `#~#` | Native type token category |
| `&~&`, `{~}` | List and Dictionary type categories |

Square brackets do not mean optionality here, and a vertical bar does not mean alternation. Named cases are alternatives. Symbols such as `%~%` describe already-typed tokens rather than source text the reader should type around a literal.

The decoder validates family names, qualification depth, type patterns, and declared identifier markers. PRONOUN is not admitted as a grammar rule element even though lexical pronouns participate in source binding.

## Matching and ambiguity

Nested grammar consumes a contiguous source prefix. Every candidate case begins from the same state; the parser chooses the successful candidate consuming the longest prefix. Equal lengths use sorted case names, so map iteration order does not decide the parse. Optional elements commit only on success. A full statement rejects leftover unassociated tokens.

A parse choice is different from semantic ambiguity. If a spelling resolves to more than one valid lexical declaration, binding can report AmbiguousReference even after a grammar shape has been recognized. Coordinated subjects/objects retain occurrence identity; identical token spelling is not token identity.

## Values, subjects, and objects

The value rule recognizes definitive references and Text/Number/Truth/Null/type/List/Dictionary patterns. Subject cases describe declarations, references, definitions, directives, utilization, and typed/collection declarations. Object cases additionally distinguish named objects, dependencies, definitive references, and scalar operands.

Articles and adjectives modify these positions. `a`/`an` is indefinite, `the` definite, `any` identifies variable arguments in supported action signatures, and `no` expresses absence or a supported negative condition. These roles are not interchangeable decoration.

A subject identifies what is being defined or acted upon; an object supplies the action operand. That distinction matters for `The App Prints message`: the App owns the native body effect, while message supplies Print's argument. In `Run Worker`, Worker is the outer lifecycle operand.

### Read a complete sentence in context

The [complete first program](../guide/first-program.md#read-the-program) gives those grammatical roles a concrete meaning:

| Source phrase | Role in this program |
| --- | --- |
| `"message" is` followed by Text | Declares a name and assigns its literal value |
| `Create "Hello" as a New Basic App` | Imperative declaration; `as` introduces the construction |
| `The App Prints message` | Active statement: App is the subject, Prints is the action, and message is its operand |
| `Execute Hello` | Imperative source-action call with Hello as its argument |
| `Hello's Output` | Member access, used here as the operand of Show |

`The App` associates Print with the App's generated body. The outer Execute and Show sentences control execution and display its result. Grammatical form does not erase that ownership distinction.

### Any

<!-- word:grammar-word:any -->

Declare a variable argument or nominal target in a supported source signature or descriptor.

**Usage and boundary.** Start declares `any App` in the [source-action example](definitions.md#source-actions); Main declares exact target categories. Any does not erase the supplied value's type or create unrestricted native generics.

### No

<!-- word:grammar-word:no -->

Express an absent signature component or negate a condition where that grammar permits it.

**Usage and boundary.** Main's Function refinement uses `No Inputs and No Outputs` in the [descriptor example](lexical-types.md#describe-with-an-adjective). This explicit empty signature is not an unassigned value or implicit runtime default.

### A

<!-- word:grammar-word:a -->

Introduce an indefinite declaration or construction using the declared alternative form An where appropriate.

**Usage and boundary.** `a New Basic App` in the [first program](../guide/first-program.md#read-the-program) and `an Error` in the [value program](../examples/word-values.l) use the same article family. A/An does not select a type independently of its noun.

### The

<!-- word:grammar-word:the -->

Refer definitively to the selected subject or value in its owning source context.

**Usage and boundary.** `The App Prints message` in the [first program](../guide/first-program.md#read-the-program) supplies an App body effect. The does not make a private or unresolved name visible.

## Phrases

The thirteen phrase families are generic, directive, declarative, locative, objective, inflective, associative, indicative, utilization, dependency, signature, arguments, and results.

Declarative `as` relates a subject to a type, value, definition, or alias. Directive `for` introduces action composition. Utilization `to` connects a native-backed definition or imperative use. Associative `with` composes described child structures; `using` comes from source-defined inflection and supplies parameters. Dependency `from` selects a dictionary or explicit source provider in supported import forms.

Locative `at` and inflective `in` participate in [collection updates and typed reads](collections-and-categories.md). A phrase's meaning follows its owning action, not the preposition alone.

Signature phrases declare ordered Function inputs and outputs, argument phrases bind invocation operands, and result phrases name invocation outputs. These forms retain the [typed Function checks](applicational-types.md#functions-properties-and-operations) for counts, types, and assignment after invocation.

### Compose an action with `for`

These two builtin declarations use a directive phrase:

<!-- example:lexical-action-inflections -->

`for` introduces the action body. `Building`, `Running`, `Starting`, and `Awaiting` are applicative verb forms in that body; `any App` declares the argument, and `it` refers to it. The source conjunction `and then` separates the actions. The bars inside each quoted label declare inflections; they are unrelated to the grammar JSON's sequence notation. These definitions are implicitly loaded already, so do not redeclare them in a program.

In the basic program, `from` and `as` select the IO dictionary and local alias. The [Main descriptor example](lexical-types.md#describe-with-an-adjective) uses `of`, `with`, and `using` in Function/Object/Service construction.

### As

<!-- word:grammar-word:as -->

Relate a declaration or alias to its value, category, construction or new local name.

**Usage and boundary.** `Create "Hello" as a New Basic App` and the IO import alias in the [first program](../guide/first-program.md#read-the-program) use distinct owning actions. As alone does not convert runtime types.

### Of

<!-- word:grammar-word:of -->

Supply declared contents, type arguments or operands to a supported construction or signature.

**Usage and boundary.** `New List of Number` in the [value program](../examples/word-values.l) gives the element type; `Inputs of "Value" as Text` in the [Function example](../examples/composed-functions.l) declares a parameter. The selected owner determines the contract.

### At

<!-- word:grammar-word:at -->

Select a Number index or Text key for a supported typed read or update.

**Usage and boundary.** The [value program](../examples/word-values.l) uses `items at 0` and the Text key `first` in names. The containing operation checks the key type, bounds and result type; At supplies no unchecked access.

### With

<!-- word:grammar-word:with -->

Introduce associated structures or signature components in the selected construction.

**Usage and boundary.** The [Function example](../examples/composed-functions.l) uses `with Inputs`; the [Main construction](lexical-types.md#describe-with-an-adjective) uses `with a New Main Object`. With preserves the owning construction's argument contract.

### In

<!-- word:grammar-word:in -->

Identify the destination of a supported insertion or update.

**Usage and boundary.** `Put 7 in items` in the [value program](../examples/word-values.l) appends a Number to its List. Pointer and Map updates have their own checked contracts; In does not select a lexical namespace.

### Out

<!-- word:grammar-word:out -->

Name the configured deflective preposition without an established general executable application operation.

**Usage and boundary.** The [configured vocabulary](../../lexicon/eng/vocabulary.json) recognizes Out. No runnable source form is claimed here; use the precise [Output member](../reference/builtins/lifecycle.md#output) or IO operations for their documented behavior.

### By

<!-- word:grammar-word:by -->

Introduce the actor phrase recognized by an indicative grammatical form.

**Usage and boundary.** A focused `by Actor` phrase is configuration-level syntax; its enclosing clause must supply supported semantics. The [indicative grammar](../../lexicon/eng/grammar.json) is not evidence for an arbitrary executable passive callback API.

### From

<!-- word:grammar-word:from -->

Select the source of an import or a checked typed value read.

**Usage and boundary.** The [first program](../guide/first-program.md#read-the-program) imports IO.Print from the `io` dictionary; the [value program](../examples/word-values.l) reads a Text from names at the key `first`. The owning import/read determines provider and type checks.

### To

<!-- word:grammar-word:to -->

Introduce a native-backed action definition or a target-dependent utilization/comparison phrase.

**Usage and boundary.** The [IO source](../../stdlib/eng/io/dictionary.l) defines Flush with `to` and one framed native flush call. In `less than equal to`, To belongs to the numeric comparison; it is not an independent conversion or transfer.

### Than

<!-- word:grammar-word:than -->

Connect the operands of an admitted ordered Number comparison.

**Usage and boundary.** The focused `left less than right` and `left greater than right` forms follow the [comparison contract](control-flow.md#comparisons). Both operands must be Number; Than does not enable Text ordering.

### For

<!-- word:grammar-word:for -->

Introduce the composed body of a source action or structured Process.

**Usage and boundary.** Launch in the [local definitions program](../../demo/paragraphs/local-definitions.l) uses `for Building ...`; the [selector caller](../examples/word-selection.l) uses `Process ... for Returning ...`. For is not a new loop primitive.

## Clauses and grammatical forms

Most action families have imperative, active, passive, and applicative clauses. The configured verb root acts as the imperative. Active is typically the `-s` form, passive the past form, and applicative the `-ing` form; source inflection declarations determine the actual spelling.

| Action family | Root | Current semantic boundary |
| --- | --- | --- |
| Declarative | Create | Named values, actions, and supported construction |
| Definitive | Define | Scalar/runtime definitions and native-backed definitions |
| Descriptive | Describe | Source descriptors and bounded construction refinements |
| Associative | Use | Aliases and explicit imports |
| Generative | Build | Checked App operation |
| Executive | Run | Launch without join |
| Suspensive | Await | Join/finalize and activate continuation once |
| Interruptive | Stop | Request termination without join |
| Observative | Show | Write supported value bytes to outer output |
| Terminative | Exit | Terminate the whole program |
| Expansive / locative | Put / Place | Supported collection insertion and replacement |
| Receptive / reductive | Get / Take | Recognized collection forms; general execution not established |
| Inflective / deflective | Import / Export | Configured roots; not replacements for supported Use import syntax |
| Generic | Source-defined word | Depends on bound typed definition/action |

These forms are not freely interchangeable in all owners. Imperative and outer passive Await are implemented. Await inside an active App or Service body parses but rejects unsupported ownership. A standalone applicative sentence is not an executable statement simply because applicative clauses can appear in an action body.

### Supported passive Await

This canonical example uses an imperative Run followed by passive Await:

<!-- example:grammar-passive-await -->

`Worker is awaited` is an outer join operation, with the same lifecycle role as `Await Worker`. The following Show reads finalized output. The existing LANG-02 runtime witness expects `passive` followed by one newline. The active `The App Prints message` sentence still supplies the App body; grammatical voice and effect ownership are separate concerns.

### Recognized but rejected active Await

This negative fixture deliberately puts Await in an App body:

<!-- example:grammar-active-await-rejected -->

**Rejected at executable-body lowering.** `Controller awaits Worker` can be classified grammatically, but source-level child-body process supervision is unsupported. Compilation rejects that statement; it does not silently move it to the orchestrator. Put lifecycle orchestration in the outer program, as in the preceding passive example. This negative fixture is an existing rejection witness, not a program to run or a newly observed diagnostic.

### Get

<!-- word:builtin:get -->

Recognize a configured receptive action without a general supported collection instruction.

**Supported alternative and limitation.** Use the checked typed read form in the [value program](../examples/word-values.l), such as `Define "first" as a New Number from items at 0`. Get is not an alternate supported spelling for that operation.

### Take

<!-- word:builtin:take -->

Recognize a configured reductive action without a general supported removal instruction.

**Supported alternative and limitation.** The [collection contract](collections-and-categories.md) provides insertion, replacement and typed reads, not removal. A configured Take form is not a runnable removal example.

### Import

<!-- word:builtin:import -->

Recognize a configured import action whose general native execution is not implemented.

**Supported alternative and limitation.** Use [Use](imports.md#indexed-dictionaries) with an explicit source or registered dictionary. The [file-import program](../../demo/dictionaries/file-import/main.l) demonstrates that supported spelling; Import is not an alias for it.

### Export

<!-- word:builtin:export -->

Recognize a configured export action whose general native execution is not implemented.

**Supported alternative and limitation.** Use [Public/Publicly](names-and-scope.md#public) to publish the selected declaration. The [public-values program](../../demo/paragraphs/public-values.l) demonstrates publication; Export is not an executable replacement.

## Conditions

The condition rule contains truth, has, equals, is, isnt, less, greater, less_equal, and greater_equal cases. Exact spellings arise from configured is/isnt, have/has, equal/equals, less/greater, than, not, and articles. Equality requires matching supported scalar types; ordering requires Number; has is limited to a target's Error.

The configured if/else/otherwise/while families are not equally executable. If, an attached else, and paragraph-local otherwise have bounded instruction support. While currently reaches an unsupported outer conditional boundary. Multiple joined conditions reject, so the grammar's repeated CONDITION form does not establish boolean conjunction or short-circuit evaluation.

The [Truth conditional example](values.md#truth) shows an actual `If enabled equals true, await Worker, else exit` statement with its declarations and execution context. It compares Truth operands and attaches an else branch to that If. For complete branch examples and explicitly rejected loop forms, see [operations and control flow](control-flow.md). A branch separator comma is source punctuation; it is not the grammar configuration's `|` sequence marker.

## Predicates and statements

Predicates assemble clauses in imperative, active, passive, applicative, or conditional form. Passive predicates include assignment. Statements choose objective imperative/conditional or subjective active/passive organization. Subjective App effects contribute application body code; outer instructions control orchestration. Standalone Service body effects reject. The same action spelling can therefore be valid in one ownership context and rejected in another.

Source order is retained after expansion. A sequence of native unary lifecycle operands runs each operation in order. A source-defined action with one parameter and several coordinated actuals expands its complete action body for each actual before proceeding to the next.

## Vocabulary reference

Nouns include pronouns, Input/Output/Error, List/Map/Dictionary, Null/Thing/Text/Number/Byte/Truth/Pointer, native Func/Prop/Op concepts, Definition/Attribute, and Object/Class/Service/App. The adjective entries are Public, New, and Generic. Adverbs are Publicly, Not, Less, Greater, and Then. Prepositions are as/of/at/with/in/out/by/from/to/than/for. Articles are any/no/a/the, conjunctions and/or, and evaluations if/else/otherwise/while.

The [source-defined builtins](../reference/builtins/index.md#source-defined-facilities) are implicit. Standard-library words such as Print require import. New words compose existing definitions; vocabulary recognition alone does not implement an action.

## Sources and witnesses

Exact data: [grammar](../../lexicon/eng/grammar.json), [vocabulary](../../lexicon/eng/vocabulary.json). Parser: [grammar terms](../../lyger/grammar_terms.go), [matching](../../lyger/grammar_match.go), [grammar types](../../lyger/grammar_types.go), [sequence parsing](../../lyger/grammar.go). Lowering: [instructions](../../lyger/instructions.go), [native roles](../../lyger/ir_kinds.go). Witnesses: [term decoding](../../lyger/grammar_terms_test.go), [consumption](../../lyger/grammar_consumption_test.go), [sequence ownership](../../lyger/grammar_sequence_test.go), [paragraph grammar](../../lyger/paragraph_grammar_architecture_test.go), and [phase-one language](../../lyger/phase1_language_test.go).
