  .org $0000

  .org $0200
  lda #$01
  ldx #$02
  adc ($fa,x)
  ldx #$04
  adc ($fa,x)
  ldx #$06
  adc ($fa,x)
  ldx #$08
  adc ($fa,x)
  ldx #$0a
  adc ($fa,x)

  .org $00fc
  .word $0230
  .word $0231
  .org $0000
  .word $0232
  .word $0233
  .word $0234

  .org $0230
  .byte $02
  .byte $fc
  .byte $00
  .byte $01
  .byte $fe
