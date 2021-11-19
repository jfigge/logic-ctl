.phony: ind  abs

all: ind abs
	rm -rf ../../.idea/runConfigurations/op*
	cp op*.xml ../../.idea/runConfigurations

clean:
	rm -rf ../../.idea/runConfigurations/op_jmp*
	rm *.bin

ind:
	../../asm/vasm -Fbin -dotdir -o ind.bin ind.asm; hexdump -C ind.bin
abs:
	../../asm/vasm -Fbin -dotdir -o abs.bin abs.asm; hexdump -C abs.bin
