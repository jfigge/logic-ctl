.phony: bpl bmi bvc bvs bcc bcs bne beq

all: bpl bmi bvc bvs bcc bcs bne beq
	rm -rf ../../.idea/runConfigurations/op*
	cp op*.xml ../../.idea/runConfigurations

clean:
	rm -rf ../../.idea/runConfigurations/op_brc*
	rm *.bin

bpl:
	../../asm/vasm -Fbin -dotdir -o bpl.bin bpl.asm; hexdump -C bpl.bin
bmi:
	../../asm/vasm -Fbin -dotdir -o bmi.bin bmi.asm; hexdump -C bmi.bin
bvc:
	../../asm/vasm -Fbin -dotdir -o bvc.bin bvc.asm; hexdump -C bvc.bin
bvs:
	../../asm/vasm -Fbin -dotdir -o bvs.bin bvs.asm; hexdump -C bvs.bin
bcc:
	../../asm/vasm -Fbin -dotdir -o bcc.bin bcc.asm; hexdump -C bcc.bin
bcs:
	../../asm/vasm -Fbin -dotdir -o bcs.bin bcs.asm; hexdump -C bcs.bin
bne:
	../../asm/vasm -Fbin -dotdir -o bne.bin bne.asm; hexdump -C bne.bin
beq:
	../../asm/vasm -Fbin -dotdir -o beq.bin beq.asm; hexdump -C beq.bin
