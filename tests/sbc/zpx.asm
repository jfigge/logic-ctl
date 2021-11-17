  .org $0200
  sec
  lda #$ff
  ldx #$0
  sbc $00,x
  inx
  sbc $00,x
  inx
  sbc $00,x
  inx
  sbc $00,x
  inx
  sbc $00,x

  .org 00
  .byte $03
  .byte $fc
  .byte $00
  .byte $01
  .byte $fe
