coiler
====

Coiler is a tool to create a single \*.pyc file from multiple \*.py files. It can be compared to a JS minifier (though it doesn't try to minify source code, just concatenate it into a single output), or a static compiler (though it produces python bytecode, not real machinecode.)

Why would I use this?
====

Creating a single executable has a lot of operational benefits. Library version conflicts, installation issues, and module naming become problems at _build time_, and cannot be encountered at runtime. This greatly reduces the amount of fear and pain involved in deploying code to machines that were not used to develop it. It also prevents the need to rigorously maintain a production machine's libraries separately from the application(s) running on them. And of course, this eliminates the need to install a bunch of other non-core tools like `pip`, `easy_install`, `distutils`, or whatever other pseudo-packaging tool is out there - your executable becomes a single file that can be run on any machine with a Python installation (not as good as being able to run on any machine period, but still pretty good for Python).

Of course, by having a single executable file for your script, you no longer need to use python-specific packaging methods (like eggs or zips). Instead, you can take advantage of serious packaging formats like \*.rpm and \*.deb, which even further reduces your operational overhead.

As a side benefit, if you choose to use the "static" mode, the standard library functions will also be baked into the resulting \*.pyc file. Meaning that 2.7 code can be reliably run on 3x interpreters, since library incompatibilities are impossible.

*To summarize*, coiling your python code makes it easier to deploy, more reliable to run, and more repeatable across machines without the need of clients running some new packaging tool.

What is required?
====

You will need the tools `python` and the `compileall` module available on the default system search path. It doesn't matter which version of python you have, but bear in mind that `coiler` will use that default version of python to discover library search paths and compile scripts - so be sure it's the version you want to use.
