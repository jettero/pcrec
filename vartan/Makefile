
V := $(wildcard *.vartan)
E := $(patsubst %.vartan,%,$(V))
J := $(patsubst %,%.json,$(E))
P := $(patsubst %,%-parsed.txt,$(E))
R := $(patsubst %,%-report.json,$(E))
T := $(patsubst %-report.json,%-report.txt,$(R))

PAGER ?= less

default: compile

compile: $(J)

parse: $(P)
	@clear
	$(PAGER) $(P)

show: $(T)
	@clear
	$(PAGER) $(T)

parse-%: %-parsed.txt
	$(PAGER) $<

show-%: %-report.txt
	$(PAGER) $<

setup:
	go install github.com/nihei9/vartan/cmd/vartan@latest

%-report.json %.json: %.vartan
	vartan compile $< -o $@

%-parsed.txt: %.json $(wildcard example*.*)
	for i in example*.$*; do echo;echo "$$(< $$i)"; echo -n "$$(< $$i)" | vartan parse $< | sed -e 's/^/    /g'; done | tee $@

%-report.txt: %-report.json
	vartan show $< | tee $@

clean:
	@git clean -dfx
