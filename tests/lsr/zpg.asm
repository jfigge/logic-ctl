  .org $0000
  .org $0200
  lda #$24
  sta $00
  lda #$a2
  sta $01
  lsr $00
  lsr $00
  lsr $00
  lsr $00
  lsr $00
  lsr $00

  .org $0000
  .byte $24
  .byte $a2
