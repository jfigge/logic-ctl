  .org $0200
  lda #$ff
  ldx #$00
  eor $00,x
  inx
  eor $00,x
  inx
  eor $00,x
  inx
  eor $00,x

  .org $0000
  .byte $55
  .byte $50
  .byte $05
  .byte $ff
