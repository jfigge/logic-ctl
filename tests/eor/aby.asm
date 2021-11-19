  .org $0000

  .org $0200
  lda #$ff
  ldy #$0e
  eor $02f0,y
  iny
  eor $02f0,y
  iny
  eor $02f0,y
  iny
  eor $02f0,y

  .org $02fe
  .byte $55
  .byte $50
  .byte $05
  .byte $ff
