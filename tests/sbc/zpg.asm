  .org $0200
  sec
  lda #$ff
  sbc $00
  sbc $01
  sbc $02
  sbc $03
  sbc $04

  .org $0000
  .byte $03
  .byte $fc
  .byte $00
  .byte $01
  .byte $fe
