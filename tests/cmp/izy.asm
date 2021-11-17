  .org $0000

  .org $0200
  lda #$01
  ldy #$01
  adc ($10),y
  iny
  adc ($10),y
  iny
  adc ($10),y
  iny
  adc ($10),y
  iny
  adc ($10),y

  .org $0010
  .word $02fc

  .org $02fd
  .byte $02
  .byte $fc
  .byte $00
  .byte $01
  .byte $fe
