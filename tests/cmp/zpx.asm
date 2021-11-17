  .org $0200
  lda #$01
  ldx #$0
  adc $00,x
  inx
  adc $00,x
  inx
  adc $00,x
  inx
  adc $00,x
  inx
  adc $00,x

  .org 00
  .byte $02
  .byte $fc
  .byte $00
  .byte $01
  .byte $fe
