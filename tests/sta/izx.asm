  .org $0000
  .org $0200
  ldx #$20
  sta (#$00,x)

  .org $0020
  .byte $0220

  .org $0220
  .byte $18
