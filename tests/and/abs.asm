  .org $0000

  .org $0200
  lda #$ff
  and $0220
  and $0221
  and $0222
  and $0223
  and $0224

  .org $0220
  .byte $ff
  .byte $0f
  .byte $ff
  .byte $f8
  .byte $00
