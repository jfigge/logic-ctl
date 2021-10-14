rom = bytearray([0xea] * 16)

with open("one.bin", "wb") as out_file:
	out_file.write(rom);
