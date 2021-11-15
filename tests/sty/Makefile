.phony: zpg zpx abs

all: zpg zpx abs
	rm -rf ../../.idea/runConfigurations/op*
	cp op*.xml ../../.idea/runConfigurations

clean:
	rm -rf ../../.idea/runConfigurations/op_sty*
	rm *.bin

zpg:
	../../asm/vasm -Fbin -dotdir -o zpg.bin zpg.asm; hexdump -C zpg.bin
zpx:
	../../asm/vasm -Fbin -dotdir -o zpx.bin zpx.asm; hexdump -C zpx.bin
abs:
	../../asm/vasm -Fbin -dotdir -o abs.bin abs.asm; hexdump -C abs.bin
