  .org $0000
  byte $e0

  .org $0200
  lda #$00
  bit $00
  sec
  sei
  clc
  cli
  clv
  lda #$00
  lda #$01

