.phony: imm zpg zpx abs abx aby izx izy

all: imm zpg zpx abs abx aby izx izy
	rm -rf ../../.idea/runConfigurations/op*
	cp op*.xml ../../.idea/runConfigurations

clean:
	rm -rf ../../.idea/runConfigurations/op_and*
	rm *.bin

imm:
	../../asm/vasm -Fbin -dotdir -o imm.bin imm.asm; hexdump -C imm.bin
zpg:
	../../asm/vasm -Fbin -dotdir -o zpg.bin zpg.asm; hexdump -C zpg.bin
zpx:
	../../asm/vasm -Fbin -dotdir -o zpx.bin zpx.asm; hexdump -C zpx.bin
abs:
	../../asm/vasm -Fbin -dotdir -o abs.bin abs.asm; hexdump -C abs.bin
abx:
	../../asm/vasm -Fbin -dotdir -o abx.bin abx.asm; hexdump -C abx.bin
aby:
	../../asm/vasm -Fbin -dotdir -o aby.bin aby.asm; hexdump -C aby.bin
izx:
	../../asm/vasm -Fbin -dotdir -o izx.bin izx.asm; hexdump -C zpx.bin
izy:
	../../asm/vasm -Fbin -dotdir -o izy.bin izy.asm; hexdump -C izy.bin
