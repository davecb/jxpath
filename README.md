**jxpath -- path expressions for json**

This is a path-expression program for experimenting with
people's json and xml APIs.  The real story, however, is 
in [13 Hours of Go](./Thirteen_Talk.odp)


The nominal business problem is in having a lightweight tool
for experimenting with APIs, that can generate code
to compile parsers as well as interactively interpret
the data. The latter is important when dealing
with constantly-changing and sometimes fragile APIs.

My personal interests, however, are different. They lie in
* using an elegant parser from Rob Pike
* doing xpath for the third time (the first two were too
big and too small, respectively)
* using Elliotte Rusty Harold's simplification of element
versus attribute, 
* doing an "explain" to show the production code to use, and
* working in Go rather than Java or C++

The last is the big thing: at the 13-hours-in mark, I'd decided
I very much like Go.

[Things still to do: json arrays, big refactor to make parser know what lexer does and simplify lexer, TBA]
