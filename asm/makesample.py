rList = [1,2,3,4,5]
rom = bytearray(rList)

with open("sample.bin", "wb") as out_file:
	out_file.write(rom);
