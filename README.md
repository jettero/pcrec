# Perl Compatible Regular Expression Compatible

The Go regular expression library sucks and I completely reject the premise --
it's true that it runs in linear time, but who cares? It can't be made to run in
non-linear time either. Most programming languages allow programmers to write
programs that don't run in linear time. I take this simply to mean the Go
authors don't like regular expressions and little else.

There's options for using pcre, but none of them are 100% satisfactory.
go.elara.ws/pcre is fanstatic. If you're reading this and nodding your head, I
suggest using that would probably be the way to go.

In grad school, one of my favorite topics in computation theory was finite
automata. This lib is my attempt to write a NFA-on-DFA regular expression engine
completely from scratch -- not because anyone needs such a thing, but because
I'll enjoy the exercise and it'll help me learn more Go.


# Side Note

I don't really like Go, but I'm trying to get better at it because it's so
popular at work. In my exasperation and rage, I'm sure I'm missing things. Feel
free to let me know where there's newb errors and better ways.

OTOH, I don't really expect anyone will ever read this.

Have a nice day.
