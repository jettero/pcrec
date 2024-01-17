
V := $(wildcard *.vartan)
E := $(patsubst %.vartan,%,$(V))
J := $(patsubst %,%.json,$(E))
P := $(patsubst %,%-parsed.txt,$(E))
R := $(patsubst %,%-report.json,$(E))
T := $(patsubst %-report.json,%-report.txt,$(R))

Sp := $(patsubst %,%_parser.go,$(E))
Sl := $(patsubst %,%_lexer.go,$(E))
Sa := $(patsubst %,%_semantic_action.go,$(E))
S := $(Sp) $(Sl) $(Sa)

PAGER ?= less

default: compile

generate: $(S)

compile: $(J)

parse: $(P)
	@clear
	$(PAGER) $(P)

show: $(T)
	@clear
	$(PAGER) $(T)

test: $(patsubst %,test-%,$(E))

parse-%: %-parsed.txt
	$(PAGER) $<

show-%: %-report.txt
	$(PAGER) $<

test-%: %.vartan %-test.txt
	vartan test $< $*-test.txt

setup:
	go install github.com/nihei9/vartan/cmd/vartan@latest
	go install github.com/nihei9/vartan/cmd/vartan-go@latest

%-report.json %.json: %.vartan
	vartan compile $< -o $@

%-parsed.txt: %.json $(wildcard example*.*)
	for i in example*.$*; do echo;echo "$$(< $$i)"; echo -n "$$(< $$i)" | vartan parse $< | sed -e 's/^/    /g'; done | tee $@

%-report.txt: %-report.json
	vartan show $< | tee $@

%_lexer.go %_semantic_action.go %_parser.go: %.json
	vartan-go $<

clean:
	@git clean -dfx
