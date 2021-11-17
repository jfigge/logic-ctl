  .org $0000

  .org $0200
  sec
  lda #$ff
  sbc $0220
  sbc $0221
  sbc $0222
  sbc $0223
  sbc $0224

  .org $0220
  .byte $03
  .byte $fc
  .byte $00
  .byte $01
  .byte $fe
