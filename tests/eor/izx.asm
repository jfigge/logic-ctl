  .org $0000
  .word $0220
  .word $0221
  .word $0222
  .word $0223

  .org $0200
  lda #$ff
  ldx #$00
  eor ($00,x)
  ldx #$02
  eor ($00,x)
  ldx #$04
  eor ($00,x)
  ldx #$06
  eor ($00,x)

  .org $0220
  .byte $55
  .byte $50
  .byte $05
  .byte $ff
