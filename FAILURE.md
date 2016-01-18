# coiler

Coiler is a failed experiment. It exists in this repository just to be stored offsite,
and so that it can be linked to those who might be curious in achieving similar goals.

### Why didn't it work?

Coiler's goals were threefold;

1. Be able to pull in third-party dependencies and first-party scripts into a single file, fit for shipping as one unit
(thereby obsoleting language-specific package management like `pip` or `easy_install`)
2. Be able to pull in standard library modules into that same single file, crossing python version borders and enabling...
3. Be able to bootstrap a native executable which runs an embedded python interpreter on the combined source,
all from one native file (thereby making python something that can be "statically compiled" and capable of being packaged and distributed sanely).

None of these are possible to the satisfaction of the author.

### Distinguishing third-party libraries

Python has no way to distinguish third-party libraries from standard libraries.
There is nothing about the standard library modules that identifies them as such, nor is there any path convention that
can be used to figure out where the standard library lives.

This makes it damnably hard to figure out what is "third party" and what is not. It's possible to bake and maintain lists
of modules that are considered "standard library", but coiler was intended to not need such extensive maintenance.
Such a strategy is brittle, prone to breaking at the slightest change in the python ecosystem, or at the presentation of a new distribution.

### Baking standard library

Since goal \#1 isn't attractive, it only makes sense that baking the whole standard library ought to be possible.

Unfortunately, the standard library makes use of a lot of python-specific techniques.
Most things imported in standard libraries are conditional imports; where the general form looks like this:

	# os.py
    if platform == 'nt':
		import ntpath as path
		linesep = '\r\n'
	else:
		if platform == 'posix':
			import posixpath as path
			linesep = '\n'

	...
	path.doSomething()

You cannot meaningfully combine these in a static way without significantly rewriting the logic that goes into them.
If you combine `ntpath`, but the platform is `risc`, then suddenly your standalone executable breaks on one platform but not another.
If you dedicate the time to detecting and rewriting those import statements into something more meaningful, you run afoul of the
namespace-manipulation aspects of Python, where you can _delete imported libraries_ from the namespace!

It quickly becomes an exercise in rewriting the standard library, which feels extremely outside the scope of this project and the author's ambition.

Not to mention how some files (mostly those which deal with syscalls) aren't actually available from the `sys.path`, so you can't even reliably
combine them even if you wanted to. And how exactly does this work with modules that depend on native extensions?

All of this put together means that trying to combine all necessary python source into one file is challenging at best, impossible at worst.
Which is too bad because it affects...

### Bootstrapping

The most exciting goal of `coiler` was the ability to transcend the need for users to even have a specific version of the Python runtime installed.
The majority of this is pretty standard stuff; append the compiled \*.pyc to the end of a bootstrap executable, and then you've got
a native python application. Run it, it interprets the Python section of the executable, and your application runs.

Unfortunately, if you're unable to bake in the standard library, it's impossible to not require a Python installation.
Applications that require `re`, for instance, need to find it from somewhere. If there's no Python installed, it has to be baked
into the application. But since \#2 makes that unsatisfyingly difficult, it means that there's no real way to accomplish a real bootstrap.

Even if there was, it would be even slower than Python normally is. The amount of standard library that `coiler` can currently
pull in racks up around 750kb in source, and still leaves ~30 modules un-imported since they're not available on `sys.path`.
Running an interpreter past all of that, and initializing it all in the same scope, would be sub-optimal even for Python.
 Large executable sizes that include the whole standard library are fine, but not when the language is going to evaluate the whole library on init each run.

### What's left?

Given the inherent limitations of Python, the author does not see a great way for the project to continue. The most promising step forward would
be to tackle \#1 and find a way to distinguish standard libraries from the rest. That would eliminate most of the pain of distributing
Python applications (making users have language-specific tools like `easy_install` and `pip`), and accomplish the initial goal of the project.
But those limitations are severe, and any solution will require maintenance over time. The solutions to that over-time maintenance is \#2 and \#3,
but those are even less feasible.

So this project is on hold until something changes.
