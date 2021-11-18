  .org $0000
  .word $02f0

  .org $0200
  lda #$10
  ldy #$0e
  cmp ($00),y
  iny
  cmp ($00),y
  iny
  cmp ($00),y

  .org $02fe
  .byte $11
  .byte $10
  .byte $08
