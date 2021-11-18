  .org $0000

  .org $0200
  ldy #$10
  cpy $0220
  cpy $0221
  cpy $0222

  .org $0220
  .byte $11
  .byte $10
  .byte $08
