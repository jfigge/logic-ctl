  .org $0200
  lda #$00
  ldx #$00
  ora $00,x
  inx
  ora $00,x
  inx
  ora $00,x
  inx
  ora $00,x

  .org $0000
  .byte $c0
  .byte $03
  .byte $3c
  .byte $f0
