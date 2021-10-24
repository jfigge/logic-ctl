assemble:
	vasm -Fbin -dotdir -o rom.bin code.asm; hexdump -C rom.bin

test:
	vasm -Fbin -dotdir -o test.bin test.asm; hexdump -C test.bin

example:
	vasm -Fbin -dotdir -o example.bin example.asm; hexdump -C example.bin

tax:
	vasm -Fbin -dotdir -o tax.bin tax.asm; hexdump -C tax.bin

countdown:
	vasm -Fbin -dotdir -o countdown.bin countdown.asm; hexdump -C countdown.bin

clean:
	rm -f *.bin
