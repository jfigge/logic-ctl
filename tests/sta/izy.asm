  .org $0000
  .org $0200
  ldy #$20
  sta (#$00),y

  .org $0020
  .byte $0200

  .org $0220
  .byte $18
