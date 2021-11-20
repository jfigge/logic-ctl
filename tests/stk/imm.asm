  .org $0000
  byte $e0

  .org $0200
  ldx #$81
  txs
  ldx #$00
  tsx

  lda #$99
  pha
  lda #$00
  pla

  lda #$00
  bit $00
  sec
  sei
  php
  ldx #$01
  cli
  clc
  clv
  plp
