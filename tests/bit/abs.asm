  .org $0000

  .org $0200
  lda #$ff
  bit $0240
  bit $0241
  bit $0242
  bit $0243
  lda #$80
  bit $0240
  bit $0241
  bit $0242
  bit $0243
  lda #$03
  bit $0240
  bit $0241
  bit $0242
  bit $0243

  .org $0240
  .byte $00
  .byte $40
  .byte $80
  .byte $c0
