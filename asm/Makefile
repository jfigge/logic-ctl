.phony: assemble test example jmp tax branch coundown clean output-test push-pull status ror rol abx

assemble:
	./vasm -Fbin -dotdir -o rom.bin code.asm; hexdump -C rom.bin

status:
	./vasm -Fbin -dotdir -o status.bin status.asm; hexdump -C status.bin

rol:
	./vasm -Fbin -dotdir -o rol.bin rol.asm; hexdump -C rol.bin

ror:
	./vasm -Fbin -dotdir -o ror.bin ror.asm; hexdump -C ror.bin

abx:
	./vasm -Fbin -dotdir -o abx.bin abx.asm; hexdump -C abx.bin

test:
	./vasm -Fbin -dotdir -o test.bin test.asm; hexdump -C test.bin

example:
	./vasm -Fbin -dotdir -o example.bin example.asm; hexdump -C example.bin

push-pull:
	./vasm -Fbin -dotdir -o push-pull.bin push-pull.asm; hexdump -C push-pull.bin

jmp:
	./vasm -Fbin -dotdir -o jmp.bin jmp.asm; hexdump -C jmp.bin

tax:
	./vasm -Fbin -dotdir -o tax.bin tax.asm; hexdump -C tax.bin

output-test:
	./vasm -Fbin -dotdir -o output-test.bin output-test.asm; hexdump -C output-test.bin

branch:
	./vasm -Fbin -dotdir -o branch.bin branch.asm; hexdump -C branch.bin

branch-page-cross:
	./vasm -Fbin -dotdir -o branch-page-cross.bin branch-page-cross.asm; hexdump -C branch-page-cross.bin

countdown:
	./vasm -Fbin -dotdir -o countdown.bin countdown.asm; hexdump -C countdown.bin

clean:
	rm -f *.bin
