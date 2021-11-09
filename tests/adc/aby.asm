  .org $0000

  .org $0200
  lda #$01
  ldy #$0d
  adc $02f0,y
  iny
  adc $02f0,y
  iny
  adc $02f0,y
  iny
  adc $02f0,y
  iny
  adc $02f0,y

  .org $02fd
  .byte $02
  .byte $fc
  .byte $00
  .byte $01
  .byte $fe
