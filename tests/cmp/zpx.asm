  .org $0000
  .byte $11
  .byte $10
  .byte $08

  .org $0200
  lda #$10
  ldx #$00
  cmp $00,x
  inx
  cmp $00,x
  inx
  cmp $00,x
