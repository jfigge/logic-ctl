rom = bytearray([0xea] * 32768)

rom[0] = 0xa9  # LDA load value
rom[1] = 0x42  # value 42

rom[2] = 0x8d  # store A at address
rom[3] = 0x00  # address 0x6000
rom[4] = 0x60  # address 0x6000

rom[0x7ffc] = 0x0;
rom[0x7ffd] = 0x80;

with open("rom.bin", "wb") as out_file:
	out_file.write(rom);
