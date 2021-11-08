  .org $8000
  lda #$ff
  sta $6002
  lda #$E0
  sta $6003
  lda #$42
  sta $6000
loop:
  rol
  sta $6000
  rol
  sta $6000
  rol
  sta $6000
  rol
  sta $6000
  rol
  sta $6000
  rol
  sta $6000
  rol
  sta $6000
  rol
  sta $6000
  rol
  sta $6000
  jmp loop
