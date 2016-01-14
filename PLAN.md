How will this work?
====

All \*.py files in a run will be combined together into a single output file. Functions, constants, and their references will be namedspaced according to their original filename, instead of using the dot-separator a magic string (current thinking is *_ZC_*). Alternatives include just making a symbol table and translating names artificially.

e.g.

    import xyz.foo.bar
    print(xyz.foo.bar.MESSAGE)

Becomes

    <contents of xyz.foo.bar>
    print(xyz_ZC_foo_ZC_bar_ZC_MESSAGE)


What are the compile modes?
====

`user` mode will only combine non-system search paths, specifically the local directory and the [PYTHONPATH](https://docs.python.org/2/using/cmdline.html#envvar-PYTHONPATH). System imports are deduplicated, hoisted, and rewritten so that any `import x as y` or `from x import y` call (and the references to those functions) are translated to something sane.

`all` mode will combine _all_ search paths, including system search paths. This means doing `import re` will actually pull the regex implementation from the workstation's system python into the combined file.

How will this read search paths?
====

Ref: https://docs.python.org/2/tutorial/modules.html#the-module-search-path

Modules will be taken from the workstation's search path and included. Search path will be discovered by parsing the stdout input of `python -c "import sys; print(sys.path)"`. If a search path begins with "/" (and does not match the [PYTHONPATH](https://docs.python.org/2/using/cmdline.html#envvar-PYTHONPATH)), it will be treated as being a system-dependent search path (which is only baked into the output if the user requests it)

What are these contexts?
====

A single build has a context. This context contains a dependency graph of which files require which other files, and a "symbol table" for translating calls to other imports.

A single file has a context, too. This is primarily used to map the file's function terminology back to the build's terminology.

For instance, let's say we have `foo.py` and `bar.py`, which each reference functions from `baz.py`.

`foo.py`:
    import baz
    print(baz.findMessage())

`bar.py`:
    from baz import findMessage as message
    print(message())

Each of these files does the same thing with the same imports, but writes it differently. The _file context_ for `foo.py` will translate all occurrences of "baz.findMessage()" to whatever the _build context_ has for that symbol (probably "baz_ZC_findMessage()", see the first section). Concurrently, the _file context_ for `bar.py` will translate all occurrences of "message()" to the _build context_'s symbol table for that same entry.

Practically, this means that the _build context_ is used as a dependency graph, deduplication measure, and symbol lookup table.

All flavors of imports will cause the entire library to be imported and symbols added to the build context.
