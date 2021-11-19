  .org $0000
  .word $0220
  .word $0221
  .word $0222
  .word $0223

  .org $0200
  lda #$00
  ldx #$00
  ora ($00,x)
  ldx #$02
  ora ($00,x)
  ldx #$04
  ora ($00,x)
  ldx #$06
  ora ($00,x)

  .org $0220
  .byte $c0
  .byte $03
  .byte $3c
  .byte $f0
