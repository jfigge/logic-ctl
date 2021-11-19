  .org $0200
  lda #$ff
  eor $00
  eor $01
  eor $02
  eor $03

  .org $0000
  .byte $55
  .byte $50
  .byte $05
  .byte $ff
