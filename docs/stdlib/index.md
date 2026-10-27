# Standard library

L's standard library supplies vocabulary dictionaries that you import explicitly. They are embedded in the compiler; no separate library installation is needed. Builtins are available without an import.

## Current catalog

<!-- stdlib-catalog -->

## Import model

Use selects a dictionary or its member and can bind a local alias. The common current import is:

```l
Use IO's Print from `io` as "Print".
```

The `io` key selects the indexed dictionary; a neighboring file cannot shadow it. Explicit `.l` imports instead resolve relative to their source file. See [imports](../guide/imports.md) for both forms.

<a id="public-exports-and-evidence"></a>

Each reference lists the dictionary's public members. Private helpers remain internal to the dictionary.

<a id="catalog-maintenance"></a>
<a id="sources-and-witnesses"></a>
<a id="index-format"></a>

<details>
<summary>Dictionary registry and maintenance</summary>

The [source index](../../stdlib/index.json) maps import keys to entry files and namespaces. Contributors should follow the [coverage and snapshot requirements](../contributing/documentation.md#update-coverage-and-snapshots).

<!-- example:stdlib-index -->

</details>
