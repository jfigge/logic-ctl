  .org $0200
  lda #$ff
  ldx #$0
  and $00,x
  inx
  and $00,x
  inx
  and $00,x
  inx
  and $00,x
  inx
  and $00,x

  .org $0000
  .byte $ff
  .byte $0f
  .byte $ff
  .byte $f8
  .byte $00
