  .org $0000
  .org $0200
  ldx #$1f
  lda ($01,x)

  .org $0020
  .word $0220

  .org $0220
  .byte $40
