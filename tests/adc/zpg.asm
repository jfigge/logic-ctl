  .org $0200
  lda #$01
  adc $00
  adc $01
  adc $02
  adc $03
  adc $04

  .org 00
  .byte $02
  .byte $fc
  .byte $00
  .byte $01
  .byte $fe
