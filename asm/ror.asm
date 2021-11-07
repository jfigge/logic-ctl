  .org $8000
  lda #$ff
  sta $6002
  lda #$E0
  sta $6003
  lda #$42
loop:
  sta $6000
  ror
  jmp loop
