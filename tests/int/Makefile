.phony: imm

all: imm
	rm -rf ../../.idea/runConfigurations/op*
	cp op*.xml ../../.idea/runConfigurations

clean:
	rm -rf ../../.idea/runConfigurations/op_int*
	rm *.bin

imm:
	../../asm/vasm -Fbin -dotdir -o imm.bin imm.asm; hexdump -C imm.bin
