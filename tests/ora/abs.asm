  .org $0000

  .org $0200
  lda #$00
  ora $0220
  ora $0221
  ora $0222
  ora $0223

  .org $0220
  .byte $c0
  .byte $03
  .byte $3c
  .byte $f0
