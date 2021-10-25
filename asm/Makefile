.phony: assemble test example jmp tax branch coundown clean

assemble:
	vasm -Fbin -dotdir -o rom.bin code.asm; hexdump -C rom.bin

test:
	vasm -Fbin -dotdir -o test.bin test.asm; hexdump -C test.bin

example:
	vasm -Fbin -dotdir -o example.bin example.asm; hexdump -C example.bin

jmp:
	vasm -Fbin -dotdir -o jmp.bin jmp.asm; hexdump -C jmp.bin

tax:
	vasm -Fbin -dotdir -o tax.bin tax.asm; hexdump -C tax.bin

branch:
	vasm -Fbin -dotdir -o branch.bin branch.asm; hexdump -C branch.bin

countdown:
	vasm -Fbin -dotdir -o countdown.bin countdown.asm; hexdump -C countdown.bin

clean:
	rm -f *.bin
