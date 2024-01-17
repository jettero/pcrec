
# go install github.com/nihei9/vartan/cmd/vartan@latest

default: expr.json

%.json: %.vartan
	vartan compile $< -o $@

%-report.txt: %.json
	for i in example*.$*; do echo;echo "$$(< $$i)"; echo -n "$$(< $$i)" | vartan parse $< | sed -e 's/^/    /g'; done | tee $@
