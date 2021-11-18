  .org $0000
  .word $0220
  .word $0221
  .word $0222

  .org $0200
  lda #$10
  ldx #$00
  cmp ($00,x)
  ldx #$02
  cmp ($00,x)
  ldx #$04
  cmp ($00,x)

  .org $0220
  .byte $11
  .byte $10
  .byte $08
