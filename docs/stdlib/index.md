# Standard library

L's standard library supplies bundled dictionaries. Import their vocabulary explicitly. The compiler also links the selected Runtime implementation automatically for execution support. The default libraries are embedded in the compiler, so no separate installation is needed. Builtins are available without an import.

## Current catalog

<a id="foundation-and-runtime-dictionaries"></a>

All bundled dictionaries are listed below. IO provides application vocabulary. Formatting and Runtime are standard libraries with native authoring interfaces, while Storage supplies builtin foundations. Each reference explains its imports, interfaces, examples and limits.

<!-- stdlib-catalog -->

## Import model

Use selects a dictionary or its member and can bind a local alias. The common current import is:

```l
Use IO's Print from `io` as "Print".
```

The `io` key selects the indexed dictionary. A neighboring file cannot shadow it. Explicit `.l` imports instead resolve relative to their source file. See [imports](../guide/imports.md) for both forms.

<a id="public-exports-and-evidence"></a>

IO's reference lists its public application members. The implementation references above describe their native support boundaries and selected authoring facilities separately. Private helpers remain internal to their source unit. Required support loading does not expose their vocabulary globally.

<a id="catalog-maintenance"></a>
<a id="sources-and-witnesses"></a>
<a id="index-format"></a>

<details>
<summary>Dictionary registry and maintenance</summary>

The [distribution dictionary registry](../../dictionaries.json) maps import keys to entry files and namespaces across foundations, runtime support and standard libraries. Contributors should follow the [coverage and snapshot requirements](../contributing/documentation.md#update-coverage-and-snapshots).

<!-- example:stdlib-index -->

</details>
