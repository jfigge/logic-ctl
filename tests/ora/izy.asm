  .org $0000
  .word $02f0

  .org $0200
  lda #$00
  ldy #$0e
  ora ($00),y
  iny
  ora ($00),y
  iny
  ora ($00),y
  iny
  ora ($00),y

  .org $02fe
  .byte $c0
  .byte $03
  .byte $3c
  .byte $f0
