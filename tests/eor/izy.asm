  .org $0000
  .word $02f0

  .org $0200
  lda #$ff
  ldy #$0e
  eor ($00),y
  iny
  eor ($00),y
  iny
  eor ($00),y
  iny
  eor ($00),y

  .org $02fe
  .byte $55
  .byte $50
  .byte $05
  .byte $ff
