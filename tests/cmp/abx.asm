  .org $0000

  .org $0200
  lda #$01
  ldx #$0d
  adc $02f0,x
  inx
  adc $02f0,x
  inx
  adc $02f0,x
  inx
  adc $02f0,x
  inx
  adc $02f0,x

  .org $02fd
  .byte $02
  .byte $fc
  .byte $00
  .byte $01
  .byte $fe
