  .org $0000

  .org $0200
  lda #$00
  ldy #$0e
  ora $02f0,y
  iny
  ora $02f0,y
  iny
  ora $02f0,y
  iny
  ora $02f0,y

  .org $02fe
  .byte $c0
  .byte $03
  .byte $3c
  .byte $f0
