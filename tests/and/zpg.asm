  .org $0200
  lda #$ff
  and $00
  and $01
  and $02
  and $03
  and $04

  .org $0000
  .byte $ff
  .byte $0f
  .byte $ff
  .byte $f8
  .byte $00
