  .org $0200
  lda #$00
  ora $00
  ora $01
  ora $02
  ora $03

  .org $0000
  .byte $c0
  .byte $03
  .byte $3c
  .byte $f0
