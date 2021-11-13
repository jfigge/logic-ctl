.phony: acc zpg zpx abs abx

all: acc zpg zpx abs abx
	rm -rf ../../.idea/runConfigurations/op*
	cp op*.xml ../../.idea/runConfigurations

clean:
	rm -rf ../../.idea/runConfigurations/op_asl
	rm *.bin

acc:
	../../asm/vasm -Fbin -dotdir -o acc.bin acc.asm; hexdump -C acc.bin
zpg:
	../../asm/vasm -Fbin -dotdir -o zpg.bin zpg.asm; hexdump -C zpg.bin
zpx:
	../../asm/vasm -Fbin -dotdir -o zpx.bin zpx.asm; hexdump -C zpx.bin
abs:
	../../asm/vasm -Fbin -dotdir -o abs.bin abs.asm; hexdump -C abs.bin
abx:
	../../asm/vasm -Fbin -dotdir -o abx.bin abx.asm; hexdump -C abx.bin
