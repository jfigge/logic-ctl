  .org $0000

  .org $0200
  ldx #$10
  cpx $0220
  cpx $0221
  cpx $0222

  .org $0220
  .byte $11
  .byte $10
  .byte $08
