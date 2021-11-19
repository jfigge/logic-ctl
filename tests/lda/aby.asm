  .org $0000

  .org $0200
  ldy #$1f
  ldx #$00
  lda $0201,y
  ldx #$00
  lda $02f1,y

  .org $0220
  .byte $20
  .org $0310
  .byte $21
