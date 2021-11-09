  .org $0000

  .org $0200
  lda #$01
  adc $0220
  adc $0221
  adc $0222
  adc $0223
  adc $0224

  .org $0220
  .byte $02
  .byte $fc
  .byte $00
  .byte $01
  .byte $fe
