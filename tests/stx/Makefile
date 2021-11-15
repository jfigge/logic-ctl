.phony: zpg zpy abs

all: zpg zpy abs
	rm -rf ../../.idea/runConfigurations/op*
	cp op*.xml ../../.idea/runConfigurations

clean:
	rm -rf ../../.idea/runConfigurations/op_stx*
	rm *.bin

zpg:
	../../asm/vasm -Fbin -dotdir -o zpg.bin zpg.asm; hexdump -C zpg.bin
zpy:
	../../asm/vasm -Fbin -dotdir -o zpy.bin zpy.asm; hexdump -C zpy.bin
abs:
	../../asm/vasm -Fbin -dotdir -o abs.bin abs.asm; hexdump -C abs.bin
