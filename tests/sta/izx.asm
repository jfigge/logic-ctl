  .org $0000
  .org $0200
  lda #$20
  ldx #$02
  sta ($2e,x)

  .org $0030
  .word $0220
