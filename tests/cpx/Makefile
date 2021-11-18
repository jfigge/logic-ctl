.phony: imm zpg abs

all: imm zpg abs
	rm -rf ../../.idea/runConfigurations/op*
	cp op*.xml ../../.idea/runConfigurations

clean:
	rm -rf ../../.idea/runConfigurations/op_cpx*
	rm *.bin

imm:
	../../asm/vasm -Fbin -dotdir -o imm.bin imm.asm; hexdump -C imm.bin
zpg:
	../../asm/vasm -Fbin -dotdir -o zpg.bin zpg.asm; hexdump -C zpg.bin
abs:
	../../asm/vasm -Fbin -dotdir -o abs.bin abs.asm; hexdump -C abs.bin
