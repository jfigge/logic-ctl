  .org $0000

  .org $0200
  lda #$ff
  eor $0220
  eor $0221
  eor $0222
  eor $0223

  .org $0220
  .byte $55
  .byte $50
  .byte $05
  .byte $ff
