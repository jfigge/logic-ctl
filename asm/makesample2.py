rom = bytearray([0x11] * 32)
rom.extend(bytearray([0x22] * 32))
rom.extend(bytearray([0x33] * 32))
rom.extend(bytearray([0x44] * 32))
rom.extend(bytearray([0x55] * 32))

with open("sample2.bin", "wb") as out_file:
	out_file.write(rom);
