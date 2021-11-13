.phony: all imm zpg zpy abs aby

all: imm zpg zpy abs aby
	rm -rf ../../.idea/runConfigurations/op*
	cp op*.xml ../../.idea/runConfigurations

clean:
	rm -rf ../../.idea/runConfigurations/op_ld*
	rm *.bin

imm:
	../../asm/vasm -Fbin -dotdir -o imm.bin imm.asm; hexdump -C imm.bin
zpg:
	../../asm/vasm -Fbin -dotdir -o zpg.bin zpg.asm; hexdump -C zpg.bin
zpy:
	../../asm/vasm -Fbin -dotdir -o zpy.bin zpy.asm; hexdump -C zpy.bin
abs:
	../../asm/vasm -Fbin -dotdir -o abs.bin abs.asm; hexdump -C abs.bin
aby:
	../../asm/vasm -Fbin -dotdir -o aby.bin aby.asm; hexdump -C aby.bin
