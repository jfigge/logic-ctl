.phony: zpg  abs

all: zpg abs
	rm -rf ../../.idea/runConfigurations/op*
	cp op*.xml ../../.idea/runConfigurations

clean:
	rm -rf ../../.idea/runConfigurations/op_bit*
	rm *.bin

zpg:
	../../asm/vasm -Fbin -dotdir -o zpg.bin zpg.asm; hexdump -C zpg.bin
abs:
	../../asm/vasm -Fbin -dotdir -o abs.bin abs.asm; hexdump -C abs.bin
