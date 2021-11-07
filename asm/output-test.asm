  .org $8000
  lda #$ff
  sta $6002
reset:
  lda #$55
  sta $6000
  lda #$aa
  sta $6000
  jmp reset
