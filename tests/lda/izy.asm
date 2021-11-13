  .org $0000
  .org $0200
  ldy #$1f
  lda ($01),y
  lda ($03),y

  .org $0001
  .word $0201
  .word $02f1

  .org $0220
  .byte $80
  .org $0310
  .byte $81
