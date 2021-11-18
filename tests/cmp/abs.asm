  .org $0000

  .org $0200
  lda #$10
  cmp $0220
  cmp $0221
  cmp $0222

  .org $0220
  .byte $11
  .byte $10
  .byte $08
